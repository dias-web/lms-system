package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	"github.com/gin-gonic/gin"
)

type AttachmentHandler struct {
	svc service.AttachmentService
}

func NewAttachmentHandler(svc service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{svc: svc}
}

// Register wires the attachment routes. uploadMW guards POST /upload (admin
// only); readMW guards GET /download/{id} (any authenticated user). Listing a
// lesson's attachments is a public read, consistent with the catalog.
func (h *AttachmentHandler) Register(r *gin.Engine, uploadMW, readMW []gin.HandlerFunc) {
	r.POST("/upload", chain(uploadMW, h.Upload)...)
	r.GET("/download/:id", chain(readMW, h.Download)...)
	r.GET("/lessons/:id/attachments", h.ListByLesson)
}

// Upload godoc
// @Summary      Upload a file attachment (admin only)
// @Description  Stores a file in MinIO and links it to a lesson. Multipart form: field "file" is the file, "lesson_id" the target lesson. Requires ROLE_ADMIN.
// @Tags         attachments
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        lesson_id  formData  int   true  "Target lesson ID"
// @Param        file       formData  file  true  "File to upload"
// @Success      201  {object}  dto.AttachmentResponse
// @Failure      400  {object}  dto.ErrorResponse  "Missing file or invalid lesson_id"
// @Failure      401  {object}  dto.ErrorResponse  "Missing or invalid token"
// @Failure      403  {object}  dto.ErrorResponse  "Requires ROLE_ADMIN"
// @Failure      404  {object}  dto.ErrorResponse  "Lesson not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /upload [post]
func (h *AttachmentHandler) Upload(c *gin.Context) {
	lessonID, err := strconv.ParseUint(c.PostForm("lesson_id"), 10, 64)
	if err != nil || lessonID == 0 {
		_ = c.Error(fmt.Errorf("%w: lesson_id is required and must be a positive integer", service.ErrInvalidInput))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(fmt.Errorf("%w: form field \"file\" is required", service.ErrInvalidInput))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		_ = c.Error(err)
		return
	}
	defer file.Close()

	resp, err := h.svc.Upload(c.Request.Context(), uint(lessonID), dto.UploadInput{
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
		Body:        file,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// Download godoc
// @Summary      Download a file attachment
// @Description  Streams the stored file by attachment ID. Requires a valid token (any authenticated user).
// @Tags         attachments
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Param        id   path  int  true  "Attachment ID"
// @Success      200  {file}    binary
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      401  {object}  dto.ErrorResponse  "Missing or invalid token"
// @Failure      404  {object}  dto.ErrorResponse  "Attachment not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /download/{id} [get]
func (h *AttachmentHandler) Download(c *gin.Context) {
	var id uint
	if err := parseUintParam(c.Param("id"), &id); err != nil {
		_ = c.Error(invalidIDError())
		return
	}

	res, err := h.svc.Download(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	defer res.Body.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", res.FileName))
	c.DataFromReader(http.StatusOK, res.Size, res.ContentType, res.Body, nil)
}

// ListByLesson godoc
// @Summary      List attachments of a lesson
// @Description  Returns metadata for all files attached to the given lesson.
// @Tags         attachments
// @Produce      json
// @Param        id   path      int  true  "Lesson ID"
// @Success      200  {array}   dto.AttachmentResponse
// @Failure      400  {object}  dto.ErrorResponse  "Invalid id"
// @Failure      404  {object}  dto.ErrorResponse  "Lesson not found"
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /lessons/{id}/attachments [get]
func (h *AttachmentHandler) ListByLesson(c *gin.Context) {
	var lessonID uint
	if err := parseUintParam(c.Param("id"), &lessonID); err != nil {
		_ = c.Error(invalidIDError())
		return
	}
	attachments, err := h.svc.ListByLesson(c.Request.Context(), lessonID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, attachments)
}
