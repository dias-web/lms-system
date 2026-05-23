package service

import (
	"context"
	"errors"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/repository"
	"github.com/sirupsen/logrus"
)

type CourseService interface {
	Create(ctx context.Context, req dto.CreateCourseRequest) (dto.CourseResponse, error)
	GetByID(ctx context.Context, id uint) (dto.CourseResponse, error)
	GetByIDWithChapters(ctx context.Context, id uint) (dto.CourseResponse, error)
	List(ctx context.Context) ([]dto.CourseResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateCourseRequest) (dto.CourseResponse, error)
	Delete(ctx context.Context, id uint) error
}

type courseService struct {
	repo repository.CourseRepository
	log  *logrus.Logger
}

func NewCourseService(repo repository.CourseRepository, log *logrus.Logger) CourseService {
	return &courseService{repo: repo, log: log}
}

func (s *courseService) Create(ctx context.Context, req dto.CreateCourseRequest) (dto.CourseResponse, error) {
	s.log.Info("Creating new course")
	s.log.WithFields(logrus.Fields{
		"name":               req.Name,
		"description_length": len(req.Description),
	}).Debug("course creation payload")

	course := req.ToEntity()
	if err := s.repo.Create(ctx, &course); err != nil {
		s.log.WithError(err).Error("failed to create course")
		return dto.CourseResponse{}, err
	}
	s.log.WithFields(logrus.Fields{"course_id": course.ID, "name": course.Name}).
		Info("course created")
	return dto.ToCourseResponse(&course), nil
}

func (s *courseService) GetByID(ctx context.Context, id uint) (dto.CourseResponse, error) {
	s.log.WithField("course_id", id).Debug("fetching course by id")
	course, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("course_id", id).Warn("course not found")
			return dto.CourseResponse{}, ErrCourseNotFound
		}
		s.log.WithError(err).WithField("course_id", id).Error("failed to fetch course")
		return dto.CourseResponse{}, err
	}
	return dto.ToCourseResponse(course), nil
}

func (s *courseService) GetByIDWithChapters(ctx context.Context, id uint) (dto.CourseResponse, error) {
	s.log.WithField("course_id", id).Debug("fetching course with chapters")
	course, err := s.repo.GetByIDWithChapters(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("course_id", id).Warn("course not found")
			return dto.CourseResponse{}, ErrCourseNotFound
		}
		s.log.WithError(err).WithField("course_id", id).Error("failed to fetch course with chapters")
		return dto.CourseResponse{}, err
	}
	return dto.ToCourseResponse(course), nil
}

func (s *courseService) List(ctx context.Context) ([]dto.CourseResponse, error) {
	s.log.Debug("listing all courses")
	courses, err := s.repo.List(ctx)
	if err != nil {
		s.log.WithError(err).Error("failed to list courses")
		return nil, err
	}
	s.log.WithField("count", len(courses)).Debug("courses listed")
	return dto.ToCourseResponseList(courses), nil
}

func (s *courseService) Update(ctx context.Context, id uint, req dto.UpdateCourseRequest) (dto.CourseResponse, error) {
	s.log.WithField("course_id", id).Info("Updating course")
	s.log.WithFields(logrus.Fields{
		"course_id":          id,
		"name":               req.Name,
		"description_length": len(req.Description),
	}).Debug("course update payload")

	course, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("course_id", id).Warn("course not found for update")
			return dto.CourseResponse{}, ErrCourseNotFound
		}
		s.log.WithError(err).WithField("course_id", id).Error("failed to load course for update")
		return dto.CourseResponse{}, err
	}

	req.ApplyTo(course)

	if err := s.repo.Update(ctx, course); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("course_id", id).Warn("course not found during save")
			return dto.CourseResponse{}, ErrCourseNotFound
		}
		s.log.WithError(err).WithField("course_id", id).Error("failed to update course")
		return dto.CourseResponse{}, err
	}
	s.log.WithFields(logrus.Fields{"course_id": course.ID, "name": course.Name}).
		Info("course updated")
	return dto.ToCourseResponse(course), nil
}

func (s *courseService) Delete(ctx context.Context, id uint) error {
	s.log.WithField("course_id", id).Info("Deleting course")
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("course_id", id).Warn("course not found for delete")
			return ErrCourseNotFound
		}
		s.log.WithError(err).WithField("course_id", id).Error("failed to delete course")
		return err
	}
	s.log.WithField("course_id", id).Info("course deleted")
	return nil
}
