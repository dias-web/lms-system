package repository

import (
	"context"
	"errors"

	"github.com/dias-web/lms-system/internal/entity"
	"gorm.io/gorm"
)

type LessonRepository interface {
	Create(ctx context.Context, lesson *entity.Lesson) error
	GetByID(ctx context.Context, id uint) (*entity.Lesson, error)
	ListByChapter(ctx context.Context, chapterID uint) ([]entity.Lesson, error)
	Update(ctx context.Context, lesson *entity.Lesson) error
	Delete(ctx context.Context, id uint) error
}

type lessonRepository struct {
	db *gorm.DB
}

func NewLessonRepository(db *gorm.DB) LessonRepository {
	return &lessonRepository{db: db}
}

func (r *lessonRepository) Create(ctx context.Context, lesson *entity.Lesson) error {
	return r.db.WithContext(ctx).Create(lesson).Error
}

func (r *lessonRepository) GetByID(ctx context.Context, id uint) (*entity.Lesson, error) {
	var lesson entity.Lesson
	if err := r.db.WithContext(ctx).First(&lesson, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &lesson, nil
}

func (r *lessonRepository) ListByChapter(ctx context.Context, chapterID uint) ([]entity.Lesson, error) {
	var lessons []entity.Lesson
	err := r.db.WithContext(ctx).
		Where("chapter_id = ?", chapterID).
		Order("\"order\" ASC").
		Find(&lessons).Error
	if err != nil {
		return nil, err
	}
	return lessons, nil
}

func (r *lessonRepository) Update(ctx context.Context, lesson *entity.Lesson) error {
	res := r.db.WithContext(ctx).Save(lesson)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *lessonRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&entity.Lesson{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}