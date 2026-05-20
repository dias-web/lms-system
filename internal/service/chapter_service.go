package service

import (
	"context"
	"errors"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/repository"
	"github.com/sirupsen/logrus"
)

type ChapterService interface {
	Create(ctx context.Context, req dto.CreateChapterRequest) (dto.ChapterResponse, error)
	GetByID(ctx context.Context, id uint) (dto.ChapterResponse, error)
	GetByIDWithLessons(ctx context.Context, id uint) (dto.ChapterResponse, error)
	ListByCourse(ctx context.Context, courseID uint) ([]dto.ChapterResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateChapterRequest) (dto.ChapterResponse, error)
	Delete(ctx context.Context, id uint) error
}

type chapterService struct {
	repo       repository.ChapterRepository
	courseRepo repository.CourseRepository
	log        *logrus.Logger
}

func NewChapterService(repo repository.ChapterRepository, courseRepo repository.CourseRepository, log *logrus.Logger) ChapterService {
	return &chapterService{repo: repo, courseRepo: courseRepo, log: log}
}

func (s *chapterService) Create(ctx context.Context, req dto.CreateChapterRequest) (dto.ChapterResponse, error) {
	if _, err := s.courseRepo.GetByID(ctx, req.CourseID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.ChapterResponse{}, ErrCourseNotFound
		}
		return dto.ChapterResponse{}, err
	}

	chapter := req.ToEntity()
	if err := s.repo.Create(ctx, &chapter); err != nil {
		return dto.ChapterResponse{}, err
	}
	s.log.WithFields(logrus.Fields{
		"chapter_id": chapter.ID,
		"course_id":  chapter.CourseID,
		"name":       chapter.Name,
	}).Info("chapter created")
	return dto.ToChapterResponse(&chapter), nil
}

func (s *chapterService) GetByID(ctx context.Context, id uint) (dto.ChapterResponse, error) {
	chapter, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.ChapterResponse{}, ErrChapterNotFound
		}
		return dto.ChapterResponse{}, err
	}
	return dto.ToChapterResponse(chapter), nil
}

func (s *chapterService) GetByIDWithLessons(ctx context.Context, id uint) (dto.ChapterResponse, error) {
	chapter, err := s.repo.GetByIDWithLessons(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.ChapterResponse{}, ErrChapterNotFound
		}
		return dto.ChapterResponse{}, err
	}
	return dto.ToChapterResponse(chapter), nil
}

func (s *chapterService) ListByCourse(ctx context.Context, courseID uint) ([]dto.ChapterResponse, error) {
	if _, err := s.courseRepo.GetByID(ctx, courseID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCourseNotFound
		}
		return nil, err
	}

	chapters, err := s.repo.ListByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	return dto.ToChapterResponseList(chapters), nil
}

func (s *chapterService) Update(ctx context.Context, id uint, req dto.UpdateChapterRequest) (dto.ChapterResponse, error) {
	chapter, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.ChapterResponse{}, ErrChapterNotFound
		}
		return dto.ChapterResponse{}, err
	}

	req.ApplyTo(chapter)

	if err := s.repo.Update(ctx, chapter); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.ChapterResponse{}, ErrChapterNotFound
		}
		return dto.ChapterResponse{}, err
	}
	s.log.WithFields(logrus.Fields{
		"chapter_id": chapter.ID,
		"name":       chapter.Name,
	}).Info("chapter updated")
	return dto.ToChapterResponse(chapter), nil
}

func (s *chapterService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrChapterNotFound
		}
		return err
	}
	s.log.WithField("chapter_id", id).Info("chapter deleted")
	return nil
}
