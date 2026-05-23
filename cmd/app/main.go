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
	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/middleware"
	"github.com/dias-web/lms-system/internal/repository"
	"github.com/dias-web/lms-system/internal/service"
	"github.com/dias-web/lms-system/pkg/database"
	"github.com/dias-web/lms-system/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
)

func parseUintParam(s string, dst *uint) error {
	if _, err := fmt.Sscanf(s, "%d", dst); err != nil {
		return err
	}
	return nil
}

// bindError wraps a Gin binding error so the middleware emits a 400 with the
// INVALID_INPUT code while preserving the validator's message.
func bindError(err error) error {
	return fmt.Errorf("%w: %s", service.ErrInvalidInput, err.Error())
}

// invalidIDError signals an unparsable path parameter.
func invalidIDError() error {
	return fmt.Errorf("%w: invalid id", service.ErrInvalidInput)
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

	courseSvc := service.NewCourseService(courseRepo, log)
	chapterSvc := service.NewChapterService(chapterRepo, courseRepo, log)
	lessonSvc := service.NewLessonService(lessonRepo, chapterRepo, log)

	if courses, err := courseSvc.List(context.Background()); err != nil {
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
	router.HandleMethodNotAllowed = true
	router.Use(
		gin.Logger(),
		middleware.Recovery(log),
		middleware.ErrorHandler(log),
	)
	router.NoRoute(middleware.NotFoundHandler())
	router.NoMethod(middleware.MethodNotAllowedHandler())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ---- Courses ----------------------------------------------------------
	router.GET("/courses", func(c *gin.Context) {
		courses, err := courseSvc.List(c.Request.Context())
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, courses)
	})

	router.GET("/courses/:id", func(c *gin.Context) {
		var id uint
		if err := parseUintParam(c.Param("id"), &id); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		course, err := courseSvc.GetByIDWithChapters(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, course)
	})

	router.POST("/courses", func(c *gin.Context) {
		var req dto.CreateCourseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(bindError(err))
			return
		}
		course, err := courseSvc.Create(c.Request.Context(), req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusCreated, course)
	})

	router.PUT("/courses/:id", func(c *gin.Context) {
		var id uint
		if err := parseUintParam(c.Param("id"), &id); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		var req dto.UpdateCourseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(bindError(err))
			return
		}
		course, err := courseSvc.Update(c.Request.Context(), id, req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, course)
	})

	router.DELETE("/courses/:id", func(c *gin.Context) {
		var id uint
		if err := parseUintParam(c.Param("id"), &id); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		if err := courseSvc.Delete(c.Request.Context(), id); err != nil {
			_ = c.Error(err)
			return
		}
		c.Status(http.StatusNoContent)
	})

	// ---- Chapters ---------------------------------------------------------
	router.GET("/courses/:id/chapters", func(c *gin.Context) {
		var courseID uint
		if err := parseUintParam(c.Param("id"), &courseID); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		chapters, err := chapterSvc.ListByCourse(c.Request.Context(), courseID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, chapters)
	})

	router.GET("/chapters/:id", func(c *gin.Context) {
		var id uint
		if err := parseUintParam(c.Param("id"), &id); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		chapter, err := chapterSvc.GetByIDWithLessons(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, chapter)
	})

	router.POST("/chapters", func(c *gin.Context) {
		var req dto.CreateChapterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(bindError(err))
			return
		}
		chapter, err := chapterSvc.Create(c.Request.Context(), req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusCreated, chapter)
	})

	router.PUT("/chapters/:id", func(c *gin.Context) {
		var id uint
		if err := parseUintParam(c.Param("id"), &id); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		var req dto.UpdateChapterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(bindError(err))
			return
		}
		chapter, err := chapterSvc.Update(c.Request.Context(), id, req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, chapter)
	})

	router.DELETE("/chapters/:id", func(c *gin.Context) {
		var id uint
		if err := parseUintParam(c.Param("id"), &id); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		if err := chapterSvc.Delete(c.Request.Context(), id); err != nil {
			_ = c.Error(err)
			return
		}
		c.Status(http.StatusNoContent)
	})

	// ---- Lessons ----------------------------------------------------------
	router.GET("/chapters/:id/lessons", func(c *gin.Context) {
		var chapterID uint
		if err := parseUintParam(c.Param("id"), &chapterID); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		lessons, err := lessonSvc.ListByChapter(c.Request.Context(), chapterID)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, lessons)
	})

	router.GET("/lessons/:id", func(c *gin.Context) {
		var id uint
		if err := parseUintParam(c.Param("id"), &id); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		lesson, err := lessonSvc.GetByID(c.Request.Context(), id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, lesson)
	})

	router.POST("/lessons", func(c *gin.Context) {
		var req dto.CreateLessonRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(bindError(err))
			return
		}
		lesson, err := lessonSvc.Create(c.Request.Context(), req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusCreated, lesson)
	})

	router.PUT("/lessons/:id", func(c *gin.Context) {
		var id uint
		if err := parseUintParam(c.Param("id"), &id); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		var req dto.UpdateLessonRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(bindError(err))
			return
		}
		lesson, err := lessonSvc.Update(c.Request.Context(), id, req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, lesson)
	})

	router.DELETE("/lessons/:id", func(c *gin.Context) {
		var id uint
		if err := parseUintParam(c.Param("id"), &id); err != nil {
			_ = c.Error(invalidIDError())
			return
		}
		if err := lessonSvc.Delete(c.Request.Context(), id); err != nil {
			_ = c.Error(err)
			return
		}
		c.Status(http.StatusNoContent)
	})

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
