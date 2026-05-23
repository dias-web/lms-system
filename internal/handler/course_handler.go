package handler

import (
	"net/http"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	svc service.CourseService
}

func NewCourseHandler(svc service.CourseService) *CourseHandler {
	return &CourseHandler{svc: svc}
}

// Register wires all course routes onto the given router group.
func (h *CourseHandler) Register(r *gin.Engine) {
	r.GET("/courses", h.List)
	r.GET("/courses/:id", h.GetByID)
	r.POST("/courses", h.Create)
	r.PUT("/courses/:id", h.Update)
	r.DELETE("/courses/:id", h.Delete)
}

// List godoc
// @Summary      List all courses
// @Description  Returns a flat list of courses without nested chapters.
// @Tags         courses
// @Produce      json
// @Success      200  {array}   dto.CourseResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /courses [get]
func (h *CourseHandler) List(c *gin.Context) {
	courses, err := h.svc.List(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, courses)
}

// GetByID godoc
// @Summary      Get course by ID (with chapters)
// @Description  Returns a course together with its chapters, ordered by chapter.order.
// @Tags         courses
// @Produce      json
// @Param        id   path      int  true  "Course ID"
// @Success      200  {object}  dto.CourseResponse
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      404  {object}  dto.ErrorResponse  "Course not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /courses/{id} [get]
func (h *CourseHandler) GetByID(c *gin.Context) {
	var id uint
	if err := parseUintParam(c.Param("id"), &id); err != nil {
		_ = c.Error(invalidIDError())
		return
	}
	course, err := h.svc.GetByIDWithChapters(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, course)
}

// Create godoc
// @Summary      Create a new course
// @Description  Creates a course. Name length 2..255. Description up to 10000 chars.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        course  body      dto.CreateCourseRequest  true  "Course payload"
// @Success      201     {object}  dto.CourseResponse
// @Failure      400     {object}  dto.ErrorResponse  "Validation failed"
// @Failure      500     {object}  dto.ErrorResponse
// @Router       /courses [post]
func (h *CourseHandler) Create(c *gin.Context) {
	var req dto.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(bindError(err))
		return
	}
	course, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, course)
}

// Update godoc
// @Summary      Update an existing course
// @Description  Replaces course name and description. Timestamps are preserved.
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        id      path      int                      true  "Course ID"
// @Param        course  body      dto.UpdateCourseRequest  true  "Course payload"
// @Success      200     {object}  dto.CourseResponse
// @Failure      400     {object}  dto.ErrorResponse  "Validation failed or invalid id"
// @Failure      404     {object}  dto.ErrorResponse  "Course not found"
// @Failure      500     {object}  dto.ErrorResponse
// @Router       /courses/{id} [put]
func (h *CourseHandler) Update(c *gin.Context) {
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
	course, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, course)
}

// Delete godoc
// @Summary      Delete a course
// @Description  Deletes a course. All chapters and lessons are cascade-deleted.
// @Tags         courses
// @Produce      json
// @Param        id   path  int  true  "Course ID"
// @Success      204  "Deleted"
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      404  {object}  dto.ErrorResponse  "Course not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /courses/{id} [delete]
func (h *CourseHandler) Delete(c *gin.Context) {
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
