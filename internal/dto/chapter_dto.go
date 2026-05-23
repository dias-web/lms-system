package dto

import (
	"time"

	"github.com/dias-web/lms-system/internal/entity"
)

// CreateChapterRequest is the payload for POST /chapters.
// course_id is required; the parent course must exist.
// @name CreateChapterRequest
type CreateChapterRequest struct {
	Name        string `json:"name"        binding:"required,min=2,max=255" example:"Variables and Types"`
	Description string `json:"description" binding:"max=10000"              example:"Basic types, declaration, scope."`
	Order       int    `json:"order"       binding:"gte=0"                  example:"1"`
	CourseID    uint   `json:"course_id"   binding:"required"               example:"1"`
}

// UpdateChapterRequest is the payload for PUT /chapters/{id}.
// course_id is intentionally absent: chapters cannot be moved between courses via PUT.
// @name UpdateChapterRequest
type UpdateChapterRequest struct {
	Name        string `json:"name"        binding:"required,min=2,max=255" example:"Variables and Types (revised)"`
	Description string `json:"description" binding:"max=10000"              example:"Basic types, declaration, scope, examples."`
	Order       int    `json:"order"       binding:"gte=0"                  example:"1"`
}

// ChapterResponse is returned on chapter read endpoints. Lessons are
// included only when fetched via /chapters/{id}.
// @name ChapterResponse
type ChapterResponse struct {
	ID          uint             `json:"id"          example:"1"`
	Name        string           `json:"name"        example:"Variables and Types"`
	Description string           `json:"description" example:"Basic types, declaration, scope."`
	Order       int              `json:"order"       example:"1"`
	CourseID    uint             `json:"course_id"   example:"1"`
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
