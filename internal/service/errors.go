package service

import "errors"

var (
	ErrNotFound        = errors.New("entity not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrCourseNotFound  = errors.New("course not found")
	ErrChapterNotFound = errors.New("chapter not found")
	ErrLessonNotFound  = errors.New("lesson not found")
)
