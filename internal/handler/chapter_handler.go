package handler

import (
	"net/http"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	"github.com/gin-gonic/gin"
)

type ChapterHandler struct {
	svc service.ChapterService
}

func NewChapterHandler(svc service.ChapterService) *ChapterHandler {
	return &ChapterHandler{svc: svc}
}

// Register wires all chapter routes onto the given router group. Any handlers
// passed as authMW guard the mutating endpoints (POST/PUT/DELETE); reads stay
// public.
func (h *ChapterHandler) Register(r *gin.Engine, authMW ...gin.HandlerFunc) {
	r.GET("/courses/:id/chapters", h.ListByCourse)
	r.GET("/chapters/:id", h.GetByID)
	r.POST("/chapters", chain(authMW, h.Create)...)
	r.PUT("/chapters/:id", chain(authMW, h.Update)...)
	r.DELETE("/chapters/:id", chain(authMW, h.Delete)...)
}

// ListByCourse godoc
// @Summary      List chapters of a course
// @Description  Returns all chapters of the given course, ordered by chapter.order.
// @Tags         chapters
// @Produce      json
// @Param        id   path      int  true  "Course ID"
// @Success      200  {array}   dto.ChapterResponse
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      404  {object}  dto.ErrorResponse  "Course not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /courses/{id}/chapters [get]
func (h *ChapterHandler) ListByCourse(c *gin.Context) {
	var courseID uint
	if err := parseUintParam(c.Param("id"), &courseID); err != nil {
		_ = c.Error(invalidIDError())
		return
	}
	chapters, err := h.svc.ListByCourse(c.Request.Context(), courseID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, chapters)
}

// GetByID godoc
// @Summary      Get chapter by ID (with lessons)
// @Description  Returns a chapter together with its lessons, ordered by lesson.order.
// @Tags         chapters
// @Produce      json
// @Param        id   path      int  true  "Chapter ID"
// @Success      200  {object}  dto.ChapterResponse
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      404  {object}  dto.ErrorResponse  "Chapter not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /chapters/{id} [get]
func (h *ChapterHandler) GetByID(c *gin.Context) {
	var id uint
	if err := parseUintParam(c.Param("id"), &id); err != nil {
		_ = c.Error(invalidIDError())
		return
	}
	chapter, err := h.svc.GetByIDWithLessons(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, chapter)
}

// Create godoc
// @Summary      Create a new chapter
// @Description  Creates a chapter under an existing course. course_id is required and validated.
// @Tags         chapters
// @Accept       json
// @Produce      json
// @Param        chapter  body      dto.CreateChapterRequest  true  "Chapter payload"
// @Success      201      {object}  dto.ChapterResponse
// @Failure      400      {object}  dto.ErrorResponse  "Validation failed"
// @Failure      404      {object}  dto.ErrorResponse  "Parent course not found"
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /chapters [post]
func (h *ChapterHandler) Create(c *gin.Context) {
	var req dto.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(bindError(err))
		return
	}
	chapter, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, chapter)
}

// Update godoc
// @Summary      Update an existing chapter
// @Description  Updates chapter name, description and order. course_id cannot be changed via PUT.
// @Tags         chapters
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true  "Chapter ID"
// @Param        chapter  body      dto.UpdateChapterRequest  true  "Chapter payload"
// @Success      200      {object}  dto.ChapterResponse
// @Failure      400      {object}  dto.ErrorResponse  "Validation failed or invalid id"
// @Failure      404      {object}  dto.ErrorResponse  "Chapter not found"
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /chapters/{id} [put]
func (h *ChapterHandler) Update(c *gin.Context) {
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
	chapter, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, chapter)
}

// Delete godoc
// @Summary      Delete a chapter
// @Description  Deletes a chapter. All lessons are cascade-deleted.
// @Tags         chapters
// @Produce      json
// @Param        id   path  int  true  "Chapter ID"
// @Success      204  "Deleted"
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      404  {object}  dto.ErrorResponse  "Chapter not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /chapters/{id} [delete]
func (h *ChapterHandler) Delete(c *gin.Context) {
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
