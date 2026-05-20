package dto

import (
	"time"

	"github.com/dias-web/lms-system/internal/entity"
)

type CreateCourseRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description" binding:"max=10000"`
}

type UpdateCourseRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Description string `json:"description" binding:"max=10000"`
}

type CourseResponse struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Chapters    []ChapterResponse `json:"chapters,omitempty"`
}

func (r CreateCourseRequest) ToEntity() entity.Course {
	return entity.Course{
		Name:        r.Name,
		Description: r.Description,
	}
}

func (r UpdateCourseRequest) ApplyTo(course *entity.Course) {
	course.Name = r.Name
	course.Description = r.Description
}

func ToCourseResponse(c *entity.Course) CourseResponse {
	resp := CourseResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if len(c.Chapters) > 0 {
		resp.Chapters = make([]ChapterResponse, 0, len(c.Chapters))
		for i := range c.Chapters {
			resp.Chapters = append(resp.Chapters, ToChapterResponse(&c.Chapters[i]))
		}
	}
	return resp
}

func ToCourseResponseList(courses []entity.Course) []CourseResponse {
	out := make([]CourseResponse, 0, len(courses))
	for i := range courses {
		out = append(out, ToCourseResponse(&courses[i]))
	}
	return out
}
