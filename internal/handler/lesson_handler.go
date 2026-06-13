package handler

import (
	"net/http"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	"github.com/gin-gonic/gin"
)

type LessonHandler struct {
	svc service.LessonService
}

func NewLessonHandler(svc service.LessonService) *LessonHandler {
	return &LessonHandler{svc: svc}
}

// Register wires all lesson routes onto the given router group. Any handlers
// passed as authMW guard the mutating endpoints (POST/PUT/DELETE); reads stay
// public.
func (h *LessonHandler) Register(r *gin.Engine, authMW ...gin.HandlerFunc) {
	r.GET("/chapters/:id/lessons", h.ListByChapter)
	r.GET("/lessons/:id", h.GetByID)
	r.POST("/lessons", chain(authMW, h.Create)...)
	r.PUT("/lessons/:id", chain(authMW, h.Update)...)
	r.DELETE("/lessons/:id", chain(authMW, h.Delete)...)
}

// ListByChapter godoc
// @Summary      List lessons of a chapter
// @Description  Returns all lessons of the given chapter, ordered by lesson.order.
// @Tags         lessons
// @Produce      json
// @Param        id   path      int  true  "Chapter ID"
// @Success      200  {array}   dto.LessonResponse
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      404  {object}  dto.ErrorResponse  "Chapter not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /chapters/{id}/lessons [get]
func (h *LessonHandler) ListByChapter(c *gin.Context) {
	var chapterID uint
	if err := parseUintParam(c.Param("id"), &chapterID); err != nil {
		_ = c.Error(invalidIDError())
		return
	}
	lessons, err := h.svc.ListByChapter(c.Request.Context(), chapterID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, lessons)
}

// GetByID godoc
// @Summary      Get lesson by ID
// @Description  Returns a single lesson with its full content.
// @Tags         lessons
// @Produce      json
// @Param        id   path      int  true  "Lesson ID"
// @Success      200  {object}  dto.LessonResponse
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      404  {object}  dto.ErrorResponse  "Lesson not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /lessons/{id} [get]
func (h *LessonHandler) GetByID(c *gin.Context) {
	var id uint
	if err := parseUintParam(c.Param("id"), &id); err != nil {
		_ = c.Error(invalidIDError())
		return
	}
	lesson, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, lesson)
}

// Create godoc
// @Summary      Create a new lesson
// @Description  Creates a lesson under an existing chapter. chapter_id is required and validated.
// @Tags         lessons
// @Accept       json
// @Produce      json
// @Param        lesson  body      dto.CreateLessonRequest  true  "Lesson payload"
// @Success      201     {object}  dto.LessonResponse
// @Failure      400     {object}  dto.ErrorResponse  "Validation failed"
// @Failure      404     {object}  dto.ErrorResponse  "Parent chapter not found"
// @Failure      500     {object}  dto.ErrorResponse
// @Router       /lessons [post]
func (h *LessonHandler) Create(c *gin.Context) {
	var req dto.CreateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(bindError(err))
		return
	}
	lesson, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, lesson)
}

// Update godoc
// @Summary      Update an existing lesson
// @Description  Updates lesson name, content and order. chapter_id cannot be changed via PUT.
// @Tags         lessons
// @Accept       json
// @Produce      json
// @Param        id      path      int                      true  "Lesson ID"
// @Param        lesson  body      dto.UpdateLessonRequest  true  "Lesson payload"
// @Success      200     {object}  dto.LessonResponse
// @Failure      400     {object}  dto.ErrorResponse  "Validation failed or invalid id"
// @Failure      404     {object}  dto.ErrorResponse  "Lesson not found"
// @Failure      500     {object}  dto.ErrorResponse
// @Router       /lessons/{id} [put]
func (h *LessonHandler) Update(c *gin.Context) {
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
	lesson, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, lesson)
}

// Delete godoc
// @Summary      Delete a lesson
// @Description  Deletes a single lesson.
// @Tags         lessons
// @Produce      json
// @Param        id   path  int  true  "Lesson ID"
// @Success      204  "Deleted"
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      404  {object}  dto.ErrorResponse  "Lesson not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /lessons/{id} [delete]
func (h *LessonHandler) Delete(c *gin.Context) {
	var id uint
	if err := parseUintParam(c.Param("id"), &id); err != nil {
		_ = c.Error(invalidIDError())
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
