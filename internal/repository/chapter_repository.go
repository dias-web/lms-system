package repository

import (
	"context"
	"errors"

	"github.com/dias-web/lms-system/internal/entity"
	"gorm.io/gorm"
)

type ChapterRepository interface {
	Create(ctx context.Context, chapter *entity.Chapter) error
	GetByID(ctx context.Context, id uint) (*entity.Chapter, error)
	GetByIDWithLessons(ctx context.Context, id uint) (*entity.Chapter, error)
	ListByCourse(ctx context.Context, courseID uint) ([]entity.Chapter, error)
	Update(ctx context.Context, chapter *entity.Chapter) error
	Delete(ctx context.Context, id uint) error
}

type chapterRepository struct {
	db *gorm.DB
}

func NewChapterRepository(db *gorm.DB) ChapterRepository {
	return &chapterRepository{db: db}
}

func (r *chapterRepository) Create(ctx context.Context, chapter *entity.Chapter) error {
	return r.db.WithContext(ctx).Create(chapter).Error
}

func (r *chapterRepository) GetByID(ctx context.Context, id uint) (*entity.Chapter, error) {
	var chapter entity.Chapter
	if err := r.db.WithContext(ctx).First(&chapter, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &chapter, nil
}

func (r *chapterRepository) GetByIDWithLessons(ctx context.Context, id uint) (*entity.Chapter, error) {
	var chapter entity.Chapter
	err := r.db.WithContext(ctx).
		Preload("Lessons", func(db *gorm.DB) *gorm.DB {
			return db.Order("lessons.\"order\" ASC")
		}).
		First(&chapter, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &chapter, nil
}

func (r *chapterRepository) ListByCourse(ctx context.Context, courseID uint) ([]entity.Chapter, error) {
	var chapters []entity.Chapter
	err := r.db.WithContext(ctx).
		Where("course_id = ?", courseID).
		Order("\"order\" ASC").
		Find(&chapters).Error
	if err != nil {
		return nil, err
	}
	return chapters, nil
}

func (r *chapterRepository) Update(ctx context.Context, chapter *entity.Chapter) error {
	res := r.db.WithContext(ctx).Save(chapter)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *chapterRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&entity.Chapter{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}