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
	if _, err := s.chapterRepo.GetByID(ctx, req.ChapterID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.LessonResponse{}, ErrChapterNotFound
		}
		return dto.LessonResponse{}, err
	}

	lesson := req.ToEntity()
	if err := s.repo.Create(ctx, &lesson); err != nil {
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
	lesson, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.LessonResponse{}, ErrLessonNotFound
		}
		return dto.LessonResponse{}, err
	}
	return dto.ToLessonResponse(lesson), nil
}

func (s *lessonService) ListByChapter(ctx context.Context, chapterID uint) ([]dto.LessonResponse, error) {
	if _, err := s.chapterRepo.GetByID(ctx, chapterID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrChapterNotFound
		}
		return nil, err
	}

	lessons, err := s.repo.ListByChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	return dto.ToLessonResponseList(lessons), nil
}

func (s *lessonService) Update(ctx context.Context, id uint, req dto.UpdateLessonRequest) (dto.LessonResponse, error) {
	lesson, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.LessonResponse{}, ErrLessonNotFound
		}
		return dto.LessonResponse{}, err
	}

	req.ApplyTo(lesson)

	if err := s.repo.Update(ctx, lesson); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.LessonResponse{}, ErrLessonNotFound
		}
		return dto.LessonResponse{}, err
	}
	s.log.WithFields(logrus.Fields{
		"lesson_id": lesson.ID,
		"name":      lesson.Name,
	}).Info("lesson updated")
	return dto.ToLessonResponse(lesson), nil
}

func (s *lessonService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLessonNotFound
		}
		return err
	}
	s.log.WithField("lesson_id", id).Info("lesson deleted")
	return nil
}
