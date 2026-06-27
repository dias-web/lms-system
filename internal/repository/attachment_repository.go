package repository

import (
	"context"
	"errors"

	"github.com/dias-web/lms-system/internal/entity"
	"gorm.io/gorm"
)

type AttachmentRepository interface {
	Create(ctx context.Context, a *entity.Attachment) error
	GetByID(ctx context.Context, id uint) (*entity.Attachment, error)
	ListByLesson(ctx context.Context, lessonID uint) ([]entity.Attachment, error)
	Delete(ctx context.Context, id uint) error
}

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) Create(ctx context.Context, a *entity.Attachment) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *attachmentRepository) GetByID(ctx context.Context, id uint) (*entity.Attachment, error) {
	var a entity.Attachment
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *attachmentRepository) ListByLesson(ctx context.Context, lessonID uint) ([]entity.Attachment, error) {
	var attachments []entity.Attachment
	err := r.db.WithContext(ctx).
		Where("lesson_id = ?", lessonID).
		Order("created_at ASC").
		Find(&attachments).Error
	if err != nil {
		return nil, err
	}
	return attachments, nil
}

func (r *attachmentRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&entity.Attachment{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}