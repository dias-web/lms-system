package service

import (
	"context"
	"errors"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/repository"
	"github.com/sirupsen/logrus"
)

type LessonService interface {
	Create(ctx context.Context, req dto.CreateLessonRequest) (dto.LessonResponse, error)
	GetByID(ctx context.Context, id uint) (dto.LessonResponse, error)
	ListByChapter(ctx context.Context, chapterID uint) ([]dto.LessonResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateLessonRequest) (dto.LessonResponse, error)
	Delete(ctx context.Context, id uint) error
}

type lessonService struct {
	repo        repository.LessonRepository
	chapterRepo repository.ChapterRepository
	log         *logrus.Logger
}

func NewLessonService(repo repository.LessonRepository, chapterRepo repository.ChapterRepository, log *logrus.Logger) LessonService {
	return &lessonService{repo: repo, chapterRepo: chapterRepo, log: log}
}

func (s *lessonService) Create(ctx context.Context, req dto.CreateLessonRequest) (dto.LessonResponse, error) {
	s.log.WithField("chapter_id", req.ChapterID).Info("Creating new lesson")
	s.log.WithFields(logrus.Fields{
		"chapter_id":     req.ChapterID,
		"name":           req.Name,
		"order":          req.Order,
		"content_length": len(req.Content),
	}).Debug("lesson creation payload")

	if _, err := s.chapterRepo.GetByID(ctx, req.ChapterID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("chapter_id", req.ChapterID).Warn("parent chapter not found")
			return dto.LessonResponse{}, ErrChapterNotFound
		}
		s.log.WithError(err).WithField("chapter_id", req.ChapterID).
			Error("failed to verify parent chapter")
		return dto.LessonResponse{}, err
	}

	lesson := req.ToEntity()
	if err := s.repo.Create(ctx, &lesson); err != nil {
		s.log.WithError(err).WithField("chapter_id", req.ChapterID).
			Error("failed to create lesson")
		return dto.LessonResponse{}, err
	}
	s.log.WithFields(logrus.Fields{
		"lesson_id":  lesson.ID,
		"chapter_id": lesson.ChapterID,
		"name":       lesson.Name,
	}).Info("lesson created")
	return dto.ToLessonResponse(&lesson), nil
}

func (s *lessonService) GetByID(ctx context.Context, id uint) (dto.LessonResponse, error) {
	s.log.WithField("lesson_id", id).Debug("fetching lesson by id")
	lesson, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("lesson_id", id).Warn("lesson not found")
			return dto.LessonResponse{}, ErrLessonNotFound
		}
		s.log.WithError(err).WithField("lesson_id", id).Error("failed to fetch lesson")
		return dto.LessonResponse{}, err
	}
	return dto.ToLessonResponse(lesson), nil
}

func (s *lessonService) ListByChapter(ctx context.Context, chapterID uint) ([]dto.LessonResponse, error) {
	s.log.WithField("chapter_id", chapterID).Debug("listing lessons by chapter")
	if _, err := s.chapterRepo.GetByID(ctx, chapterID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("chapter_id", chapterID).Warn("parent chapter not found")
			return nil, ErrChapterNotFound
		}
		s.log.WithError(err).WithField("chapter_id", chapterID).
			Error("failed to verify parent chapter")
		return nil, err
	}

	lessons, err := s.repo.ListByChapter(ctx, chapterID)
	if err != nil {
		s.log.WithError(err).WithField("chapter_id", chapterID).
			Error("failed to list lessons")
		return nil, err
	}
	s.log.WithFields(logrus.Fields{
		"chapter_id": chapterID,
		"count":      len(lessons),
	}).Debug("lessons listed")
	return dto.ToLessonResponseList(lessons), nil
}

func (s *lessonService) Update(ctx context.Context, id uint, req dto.UpdateLessonRequest) (dto.LessonResponse, error) {
	s.log.WithField("lesson_id", id).Info("Updating lesson")
	s.log.WithFields(logrus.Fields{
		"lesson_id":      id,
		"name":           req.Name,
		"order":          req.Order,
		"content_length": len(req.Content),
	}).Debug("lesson update payload")

	lesson, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("lesson_id", id).Warn("lesson not found for update")
			return dto.LessonResponse{}, ErrLessonNotFound
		}
		s.log.WithError(err).WithField("lesson_id", id).
			Error("failed to load lesson for update")
		return dto.LessonResponse{}, err
	}

	req.ApplyTo(lesson)

	if err := s.repo.Update(ctx, lesson); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("lesson_id", id).Warn("lesson not found during save")
			return dto.LessonResponse{}, ErrLessonNotFound
		}
		s.log.WithError(err).WithField("lesson_id", id).Error("failed to update lesson")
		return dto.LessonResponse{}, err
	}
	s.log.WithFields(logrus.Fields{
		"lesson_id": lesson.ID,
		"name":      lesson.Name,
	}).Info("lesson updated")
	return dto.ToLessonResponse(lesson), nil
}

func (s *lessonService) Delete(ctx context.Context, id uint) error {
	s.log.WithField("lesson_id", id).Info("Deleting lesson")
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("lesson_id", id).Warn("lesson not found for delete")
			return ErrLessonNotFound
		}
		s.log.WithError(err).WithField("lesson_id", id).Error("failed to delete lesson")
		return err
	}
	s.log.WithField("lesson_id", id).Info("lesson deleted")
	return nil
}
