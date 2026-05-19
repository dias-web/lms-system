package repository

import (
	"context"
	"errors"

	"github.com/dias-web/lms-system/internal/entity"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("entity not found")

type CourseRepository interface {
	Create(ctx context.Context, course *entity.Course) error
	GetByID(ctx context.Context, id uint) (*entity.Course, error)
	GetByIDWithChapters(ctx context.Context, id uint) (*entity.Course, error)
	List(ctx context.Context) ([]entity.Course, error)
	Update(ctx context.Context, course *entity.Course) error
	Delete(ctx context.Context, id uint) error
}

type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) Create(ctx context.Context, course *entity.Course) error {
	return r.db.WithContext(ctx).Create(course).Error
}

func (r *courseRepository) GetByID(ctx context.Context, id uint) (*entity.Course, error) {
	var course entity.Course
	if err := r.db.WithContext(ctx).First(&course, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &course, nil
}

func (r *courseRepository) GetByIDWithChapters(ctx context.Context, id uint) (*entity.Course, error) {
	var course entity.Course
	err := r.db.WithContext(ctx).
		Preload("Chapters", func(db *gorm.DB) *gorm.DB {
			return db.Order("chapters.\"order\" ASC")
		}).
		First(&course, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &course, nil
}

func (r *courseRepository) List(ctx context.Context) ([]entity.Course, error) {
	var courses []entity.Course
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *courseRepository) Update(ctx context.Context, course *entity.Course) error {
	res := r.db.WithContext(ctx).Save(course)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *courseRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&entity.Course{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}