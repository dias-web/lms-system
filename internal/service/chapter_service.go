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
	s.log.WithField("course_id", req.CourseID).Info("Creating new chapter")
	s.log.WithFields(logrus.Fields{
		"course_id":          req.CourseID,
		"name":               req.Name,
		"order":              req.Order,
		"description_length": len(req.Description),
	}).Debug("chapter creation payload")

	if _, err := s.courseRepo.GetByID(ctx, req.CourseID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("course_id", req.CourseID).Warn("parent course not found")
			return dto.ChapterResponse{}, ErrCourseNotFound
		}
		s.log.WithError(err).WithField("course_id", req.CourseID).
			Error("failed to verify parent course")
		return dto.ChapterResponse{}, err
	}

	chapter := req.ToEntity()
	if err := s.repo.Create(ctx, &chapter); err != nil {
		s.log.WithError(err).WithField("course_id", req.CourseID).
			Error("failed to create chapter")
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
	s.log.WithField("chapter_id", id).Debug("fetching chapter by id")
	chapter, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("chapter_id", id).Warn("chapter not found")
			return dto.ChapterResponse{}, ErrChapterNotFound
		}
		s.log.WithError(err).WithField("chapter_id", id).Error("failed to fetch chapter")
		return dto.ChapterResponse{}, err
	}
	return dto.ToChapterResponse(chapter), nil
}

func (s *chapterService) GetByIDWithLessons(ctx context.Context, id uint) (dto.ChapterResponse, error) {
	s.log.WithField("chapter_id", id).Debug("fetching chapter with lessons")
	chapter, err := s.repo.GetByIDWithLessons(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("chapter_id", id).Warn("chapter not found")
			return dto.ChapterResponse{}, ErrChapterNotFound
		}
		s.log.WithError(err).WithField("chapter_id", id).
			Error("failed to fetch chapter with lessons")
		return dto.ChapterResponse{}, err
	}
	return dto.ToChapterResponse(chapter), nil
}

func (s *chapterService) ListByCourse(ctx context.Context, courseID uint) ([]dto.ChapterResponse, error) {
	s.log.WithField("course_id", courseID).Debug("listing chapters by course")
	if _, err := s.courseRepo.GetByID(ctx, courseID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("course_id", courseID).Warn("parent course not found")
			return nil, ErrCourseNotFound
		}
		s.log.WithError(err).WithField("course_id", courseID).
			Error("failed to verify parent course")
		return nil, err
	}

	chapters, err := s.repo.ListByCourse(ctx, courseID)
	if err != nil {
		s.log.WithError(err).WithField("course_id", courseID).
			Error("failed to list chapters")
		return nil, err
	}
	s.log.WithFields(logrus.Fields{
		"course_id": courseID,
		"count":     len(chapters),
	}).Debug("chapters listed")
	return dto.ToChapterResponseList(chapters), nil
}

func (s *chapterService) Update(ctx context.Context, id uint, req dto.UpdateChapterRequest) (dto.ChapterResponse, error) {
	s.log.WithField("chapter_id", id).Info("Updating chapter")
	s.log.WithFields(logrus.Fields{
		"chapter_id":         id,
		"name":               req.Name,
		"order":              req.Order,
		"description_length": len(req.Description),
	}).Debug("chapter update payload")

	chapter, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("chapter_id", id).Warn("chapter not found for update")
			return dto.ChapterResponse{}, ErrChapterNotFound
		}
		s.log.WithError(err).WithField("chapter_id", id).
			Error("failed to load chapter for update")
		return dto.ChapterResponse{}, err
	}

	req.ApplyTo(chapter)

	if err := s.repo.Update(ctx, chapter); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("chapter_id", id).Warn("chapter not found during save")
			return dto.ChapterResponse{}, ErrChapterNotFound
		}
		s.log.WithError(err).WithField("chapter_id", id).Error("failed to update chapter")
		return dto.ChapterResponse{}, err
	}
	s.log.WithFields(logrus.Fields{
		"chapter_id": chapter.ID,
		"name":       chapter.Name,
	}).Info("chapter updated")
	return dto.ToChapterResponse(chapter), nil
}

func (s *chapterService) Delete(ctx context.Context, id uint) error {
	s.log.WithField("chapter_id", id).Info("Deleting chapter")
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("chapter_id", id).Warn("chapter not found for delete")
			return ErrChapterNotFound
		}
		s.log.WithError(err).WithField("chapter_id", id).Error("failed to delete chapter")
		return err
	}
	s.log.WithField("chapter_id", id).Info("chapter deleted")
	return nil
}
