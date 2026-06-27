package dto

import (
	"io"
	"time"

	"github.com/dias-web/lms-system/internal/entity"
)

// UploadInput carries a single file to be stored as a lesson attachment. It
// crosses the handler→service boundary; Body is the open request file.
type UploadInput struct {
	FileName    string
	ContentType string
	Size        int64
	Body        io.Reader
}

// DownloadResult is a readable attachment stream plus the metadata needed to
// build the HTTP response. The caller must Close Body.
type DownloadResult struct {
	FileName    string
	ContentType string
	Size        int64
	Body        io.ReadCloser
}

// AttachmentResponse is returned on upload and attachment read endpoints.
// @name AttachmentResponse
type AttachmentResponse struct {
	ID          uint      `json:"id"           example:"1"`
	LessonID    uint      `json:"lesson_id"    example:"1"`
	FileName    string    `json:"file_name"    example:"syllabus.pdf"`
	ContentType string    `json:"content_type" example:"application/pdf"`
	Size        int64     `json:"size"         example:"20480"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToAttachmentResponse(a *entity.Attachment) AttachmentResponse {
	return AttachmentResponse{
		ID:          a.ID,
		LessonID:    a.LessonID,
		FileName:    a.FileName,
		ContentType: a.ContentType,
		Size:        a.Size,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func ToAttachmentResponseList(attachments []entity.Attachment) []AttachmentResponse {
	out := make([]AttachmentResponse, 0, len(attachments))
	for i := range attachments {
		out = append(out, ToAttachmentResponse(&attachments[i]))
	}
	return out
}
