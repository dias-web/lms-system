package dto

import (
	"time"

	"github.com/dias-web/lms-system/internal/entity"
)

// CreateCourseRequest is the payload for POST /courses.
// @name CreateCourseRequest
type CreateCourseRequest struct {
	Name        string `json:"name"        binding:"required,min=2,max=255" example:"Golang Developer"`
	Description string `json:"description" binding:"max=10000"              example:"Introductory course on Go programming."`
}

// UpdateCourseRequest is the payload for PUT /courses/{id}.
// @name UpdateCourseRequest
type UpdateCourseRequest struct {
	Name        string `json:"name"        binding:"required,min=2,max=255" example:"Golang Developer (v2)"`
	Description string `json:"description" binding:"max=10000"              example:"Refreshed course on Go programming."`
}

// CourseResponse is returned on course read endpoints. Chapters are
// included only when fetched via /courses/{id}.
// @name CourseResponse
type CourseResponse struct {
	ID          uint              `json:"id"          example:"1"`
	Name        string            `json:"name"        example:"Golang Developer"`
	Description string            `json:"description" example:"Introductory course on Go programming."`
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
