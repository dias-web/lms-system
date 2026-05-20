package dto

import (
	"time"

	"github.com/dias-web/lms-system/internal/entity"
)

type CreateLessonRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description" binding:"max=10000"`
	Content     string `json:"content"`
	Order       int    `json:"order" binding:"gte=0"`
	ChapterID   uint   `json:"chapter_id" binding:"required"`
}

type UpdateLessonRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description" binding:"max=10000"`
	Content     string `json:"content"`
	Order       int    `json:"order" binding:"gte=0"`
}

type LessonResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Order       int       `json:"order"`
	ChapterID   uint      `json:"chapter_id"`
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
