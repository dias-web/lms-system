package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/entity"
	"github.com/dias-web/lms-system/internal/repository"
	"github.com/dias-web/lms-system/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ObjectStore is the subset of the MinIO storage client the service needs.
// Declared here so the service can be unit-tested with a mock.
type ObjectStore interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (*storage.Object, error)
	Remove(ctx context.Context, key string) error
}

type AttachmentService interface {
	Upload(ctx context.Context, lessonID uint, in dto.UploadInput) (dto.AttachmentResponse, error)
	Download(ctx context.Context, id uint) (*dto.DownloadResult, error)
	ListByLesson(ctx context.Context, lessonID uint) ([]dto.AttachmentResponse, error)
	Delete(ctx context.Context, id uint) error
}

type attachmentService struct {
	repo       repository.AttachmentRepository
	lessonRepo repository.LessonRepository
	store      ObjectStore
	log        *logrus.Logger
}

func NewAttachmentService(
	repo repository.AttachmentRepository,
	lessonRepo repository.LessonRepository,
	store ObjectStore,
	log *logrus.Logger,
) AttachmentService {
	return &attachmentService{repo: repo, lessonRepo: lessonRepo, store: store, log: log}
}

func (s *attachmentService) Upload(ctx context.Context, lessonID uint, in dto.UploadInput) (dto.AttachmentResponse, error) {
	s.log.WithFields(logrus.Fields{
		"lesson_id": lessonID,
		"file_name": in.FileName,
		"size":      in.Size,
	}).Info("Uploading attachment")

	if strings.TrimSpace(in.FileName) == "" {
		return dto.AttachmentResponse{}, fmt.Errorf("%w: file name is required", ErrInvalidInput)
	}

	if _, err := s.lessonRepo.GetByID(ctx, lessonID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("lesson_id", lessonID).Warn("parent lesson not found")
			return dto.AttachmentResponse{}, ErrLessonNotFound
		}
		s.log.WithError(err).WithField("lesson_id", lessonID).Error("failed to verify parent lesson")
		return dto.AttachmentResponse{}, err
	}

	contentType := in.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	key := objectKey(lessonID, in.FileName)
	if err := s.store.Upload(ctx, key, in.Body, in.Size, contentType); err != nil {
		s.log.WithError(err).WithField("key", key).Error("failed to upload object to store")
		return dto.AttachmentResponse{}, err
	}

	att := entity.Attachment{
		LessonID:    lessonID,
		FileName:    in.FileName,
		ObjectKey:   key,
		ContentType: contentType,
		Size:        in.Size,
	}
	if err := s.repo.Create(ctx, &att); err != nil {
		// The bytes are already in the store; drop them so we don't orphan an
		// object with no matching row.
		if rmErr := s.store.Remove(ctx, key); rmErr != nil {
			s.log.WithError(rmErr).WithField("key", key).
				Error("failed to roll back object after metadata write failed")
		}
		s.log.WithError(err).WithField("lesson_id", lessonID).Error("failed to persist attachment metadata")
		return dto.AttachmentResponse{}, err
	}

	s.log.WithFields(logrus.Fields{
		"attachment_id": att.ID,
		"lesson_id":     att.LessonID,
		"key":           att.ObjectKey,
	}).Info("attachment uploaded")
	return dto.ToAttachmentResponse(&att), nil
}

func (s *attachmentService) Download(ctx context.Context, id uint) (*dto.DownloadResult, error) {
	s.log.WithField("attachment_id", id).Debug("fetching attachment for download")

	att, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("attachment_id", id).Warn("attachment not found")
			return nil, ErrNotFound
		}
		s.log.WithError(err).WithField("attachment_id", id).Error("failed to load attachment")
		return nil, err
	}

	obj, err := s.store.Download(ctx, att.ObjectKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			s.log.WithField("key", att.ObjectKey).Warn("object missing in store")
			return nil, ErrNotFound
		}
		s.log.WithError(err).WithField("key", att.ObjectKey).Error("failed to open object from store")
		return nil, err
	}

	return &dto.DownloadResult{
		FileName:    att.FileName,
		ContentType: att.ContentType,
		Size:        obj.Size,
		Body:        obj.Body,
	}, nil
}

func (s *attachmentService) ListByLesson(ctx context.Context, lessonID uint) ([]dto.AttachmentResponse, error) {
	s.log.WithField("lesson_id", lessonID).Debug("listing attachments by lesson")

	if _, err := s.lessonRepo.GetByID(ctx, lessonID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("lesson_id", lessonID).Warn("parent lesson not found")
			return nil, ErrLessonNotFound
		}
		s.log.WithError(err).WithField("lesson_id", lessonID).Error("failed to verify parent lesson")
		return nil, err
	}

	attachments, err := s.repo.ListByLesson(ctx, lessonID)
	if err != nil {
		s.log.WithError(err).WithField("lesson_id", lessonID).Error("failed to list attachments")
		return nil, err
	}
	return dto.ToAttachmentResponseList(attachments), nil
}

func (s *attachmentService) Delete(ctx context.Context, id uint) error {
	s.log.WithField("attachment_id", id).Info("Deleting attachment")

	att, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.WithField("attachment_id", id).Warn("attachment not found for delete")
			return ErrNotFound
		}
		s.log.WithError(err).WithField("attachment_id", id).Error("failed to load attachment for delete")
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.WithError(err).WithField("attachment_id", id).Error("failed to delete attachment row")
		return err
	}

	// Best-effort object cleanup; the row is already gone.
	if err := s.store.Remove(ctx, att.ObjectKey); err != nil {
		s.log.WithError(err).WithField("key", att.ObjectKey).
			Warn("attachment row deleted but object removal failed")
	}
	s.log.WithField("attachment_id", id).Info("attachment deleted")
	return nil
}

// objectKey builds a collision-free storage key that preserves the file
// extension while namespacing objects under their lesson.
func objectKey(lessonID uint, fileName string) string {
	ext := path.Ext(fileName)
	return fmt.Sprintf("lessons/%d/%s%s", lessonID, uuid.NewString(), ext)
}
