package dto

import (
	"time"

	"github.com/dias-web/lms-system/internal/entity"
)

// CreateLessonRequest is the payload for POST /lessons.
// chapter_id is required; the parent chapter must exist.
// @name CreateLessonRequest
type CreateLessonRequest struct {
	Name        string `json:"name"        binding:"required,min=2,max=255" example:"Declaring variables with var"`
	Description string `json:"description" binding:"max=10000"              example:"How to declare variables and infer types."`
	Content     string `json:"content"                                       example:"In Go, var declares one or more variables..."`
	Order       int    `json:"order"       binding:"gte=0"                   example:"1"`
	ChapterID   uint   `json:"chapter_id"  binding:"required"                example:"1"`
}

// UpdateLessonRequest is the payload for PUT /lessons/{id}.
// chapter_id is intentionally absent: lessons cannot be moved between chapters via PUT.
// @name UpdateLessonRequest
type UpdateLessonRequest struct {
	Name        string `json:"name"        binding:"required,min=2,max=255" example:"Declaring variables with var (revised)"`
	Description string `json:"description" binding:"max=10000"              example:"Updated explanation."`
	Content     string `json:"content"                                       example:"Updated lesson body..."`
	Order       int    `json:"order"       binding:"gte=0"                   example:"1"`
}

// LessonResponse is returned on lesson read endpoints.
// @name LessonResponse
type LessonResponse struct {
	ID          uint      `json:"id"          example:"1"`
	Name        string    `json:"name"        example:"Declaring variables with var"`
	Description string    `json:"description" example:"How to declare variables and infer types."`
	Content     string    `json:"content"     example:"In Go, var declares one or more variables..."`
	Order       int       `json:"order"       example:"1"`
	ChapterID   uint      `json:"chapter_id"  example:"1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (r CreateLessonRequest) ToEntity() entity.Lesson {
	return entity.Lesson{
		Name:        r.Name,
		Description: r.Description,
		Content:     r.Content,
		Order:       r.Order,
		ChapterID:   r.ChapterID,
	}
}

func (r UpdateLessonRequest) ApplyTo(lesson *entity.Lesson) {
	lesson.Name = r.Name
	lesson.Description = r.Description
	lesson.Content = r.Content
	lesson.Order = r.Order
}

func ToLessonResponse(l *entity.Lesson) LessonResponse {
	return LessonResponse{
		ID:          l.ID,
		Name:        l.Name,
		Description: l.Description,
		Content:     l.Content,
		Order:       l.Order,
		ChapterID:   l.ChapterID,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

func ToLessonResponseList(lessons []entity.Lesson) []LessonResponse {
	out := make([]LessonResponse, 0, len(lessons))
	for i := range lessons {
		out = append(out, ToLessonResponse(&lessons[i]))
	}
	return out
}
