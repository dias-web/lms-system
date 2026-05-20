package dto

import (
	"time"

	"github.com/dias-web/lms-system/internal/entity"
)

type CreateChapterRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description" binding:"max=10000"`
	Order       int    `json:"order" binding:"gte=0"`
	CourseID    uint   `json:"course_id" binding:"required"`
}

type UpdateChapterRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description" binding:"max=10000"`
	Order       int    `json:"order" binding:"gte=0"`
}

type ChapterResponse struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Order       int              `json:"order"`
	CourseID    uint             `json:"course_id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Lessons     []LessonResponse `json:"lessons,omitempty"`
}

func (r CreateChapterRequest) ToEntity() entity.Chapter {
	return entity.Chapter{
		Name:        r.Name,
		Description: r.Description,
		Order:       r.Order,
		CourseID:    r.CourseID,
	}
}

func (r UpdateChapterRequest) ApplyTo(chapter *entity.Chapter) {
	chapter.Name = r.Name
	chapter.Description = r.Description
	chapter.Order = r.Order
}

func ToChapterResponse(c *entity.Chapter) ChapterResponse {
	resp := ChapterResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Order:       c.Order,
		CourseID:    c.CourseID,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if len(c.Lessons) > 0 {
		resp.Lessons = make([]LessonResponse, 0, len(c.Lessons))
		for i := range c.Lessons {
			resp.Lessons = append(resp.Lessons, ToLessonResponse(&c.Lessons[i]))
		}
	}
	return resp
}

func ToChapterResponseList(chapters []entity.Chapter) []ChapterResponse {
	out := make([]ChapterResponse, 0, len(chapters))
	for i := range chapters {
		out = append(out, ToChapterResponse(&chapters[i]))
	}
	return out
}
