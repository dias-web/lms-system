package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dias-web/lms-system/internal/config"
	"github.com/dias-web/lms-system/internal/repository"
	"github.com/dias-web/lms-system/pkg/database"
	"github.com/dias-web/lms-system/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
)

func fmtSscanID(s string, id *uint) (int, error) {
	return fmt.Sscanf(s, "%d", id)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.App.Env)
	log.Infof("Starting LMS Main Service on port %s (env=%s)", cfg.App.Port, cfg.App.Env)

	sqlDB, err := database.NewSQL(cfg.Postgres)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Fatalf("goose up: %v", err)
	}
	log.Info("Migrations applied successfully")

	gormDB, err := database.NewGorm(cfg.Postgres)
	if err != nil {
		log.Fatalf("gorm init: %v", err)
	}

	courseRepo := repository.NewCourseRepository(gormDB)
	chapterRepo := repository.NewChapterRepository(gormDB)
	lessonRepo := repository.NewLessonRepository(gormDB)

	if courses, err := courseRepo.List(context.Background()); err != nil {
		log.Warnf("smoke test: list courses failed: %v", err)
	} else {
		log.Infof("smoke test: %d course(s) found in database", len(courses))
		for _, c := range courses {
			log.Infof("  - [%d] %s", c.ID, c.Name)
		}
	}

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/courses", func(c *gin.Context) {
		courses, err := courseRepo.List(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, courses)
	})

	router.GET("/courses/:id", func(c *gin.Context) {
		var id uint
		if _, err := fmtSscanID(c.Param("id"), &id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		course, err := courseRepo.GetByIDWithChapters(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, course)
	})

	router.GET("/chapters/:id/lessons", func(c *gin.Context) {
		var id uint
		if _, err := fmtSscanID(c.Param("id"), &id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		lessons, err := lessonRepo.ListByChapter(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, lessons)
	})

	_ = chapterRepo

	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("server shutdown: %v", err)
	}
	log.Info("Server stopped")
}
