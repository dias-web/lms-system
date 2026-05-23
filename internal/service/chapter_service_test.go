package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/entity"
	"github.com/dias-web/lms-system/internal/repository"
	repomocks "github.com/dias-web/lms-system/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newChapterSvc(t *testing.T) (*repomocks.MockChapterRepository, *repomocks.MockCourseRepository, ChapterService) {
	chRepo := repomocks.NewMockChapterRepository(t)
	coRepo := repomocks.NewMockCourseRepository(t)
	return chRepo, coRepo, NewChapterService(chRepo, coRepo, silentLogger())
}

func TestChapterService_Create_Success(t *testing.T) {
	chRepo, coRepo, svc := newChapterSvc(t)

	coRepo.EXPECT().GetByID(mock.Anything, uint(1)).Return(&entity.Course{ID: 1}, nil)
	chRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*entity.Chapter")).
		Run(func(_ context.Context, c *entity.Chapter) { c.ID = 99 }).
		Return(nil)

	resp, err := svc.Create(context.Background(), dto.CreateChapterRequest{
		Name: "Intro", CourseID: 1, Order: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, uint(99), resp.ID)
	assert.Equal(t, uint(1), resp.CourseID)
}

func TestChapterService_Create_ParentCourseNotFound(t *testing.T) {
	_, coRepo, svc := newChapterSvc(t)

	coRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.Create(context.Background(), dto.CreateChapterRequest{
		Name: "X", CourseID: 99,
	})
	assert.ErrorIs(t, err, ErrCourseNotFound)
}

func TestChapterService_Create_ParentCheckRepoError(t *testing.T) {
	_, coRepo, svc := newChapterSvc(t)

	boom := errors.New("db down")
	coRepo.EXPECT().GetByID(mock.Anything, uint(1)).Return(nil, boom)

	_, err := svc.Create(context.Background(), dto.CreateChapterRequest{
		Name: "X", CourseID: 1,
	})
	assert.ErrorIs(t, err, boom)
}

func TestChapterService_Create_RepoError(t *testing.T) {
	chRepo, coRepo, svc := newChapterSvc(t)

	coRepo.EXPECT().GetByID(mock.Anything, uint(1)).Return(&entity.Course{ID: 1}, nil)
	boom := errors.New("insert failed")
	chRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(boom)

	_, err := svc.Create(context.Background(), dto.CreateChapterRequest{
		Name: "X", CourseID: 1,
	})
	assert.ErrorIs(t, err, boom)
}

func TestChapterService_GetByID_Success(t *testing.T) {
	chRepo, _, svc := newChapterSvc(t)

	chRepo.EXPECT().GetByID(mock.Anything, uint(5)).
		Return(&entity.Chapter{ID: 5, Name: "Ch"}, nil)

	resp, err := svc.GetByID(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, uint(5), resp.ID)
}

func TestChapterService_GetByID_NotFound(t *testing.T) {
	chRepo, _, svc := newChapterSvc(t)

	chRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.GetByID(context.Background(), 99)
	assert.ErrorIs(t, err, ErrChapterNotFound)
}

func TestChapterService_GetByIDWithLessons_Success(t *testing.T) {
	chRepo, _, svc := newChapterSvc(t)

	chRepo.EXPECT().GetByIDWithLessons(mock.Anything, uint(5)).
		Return(&entity.Chapter{
			ID: 5,
			Lessons: []entity.Lesson{
				{ID: 11, Name: "L1"},
			},
		}, nil)

	resp, err := svc.GetByIDWithLessons(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, resp.Lessons, 1)
	assert.Equal(t, uint(11), resp.Lessons[0].ID)
}

func TestChapterService_GetByIDWithLessons_NotFound(t *testing.T) {
	chRepo, _, svc := newChapterSvc(t)

	chRepo.EXPECT().GetByIDWithLessons(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.GetByIDWithLessons(context.Background(), 99)
	assert.ErrorIs(t, err, ErrChapterNotFound)
}

func TestChapterService_ListByCourse_Success(t *testing.T) {
	chRepo, coRepo, svc := newChapterSvc(t)

	coRepo.EXPECT().GetByID(mock.Anything, uint(1)).Return(&entity.Course{ID: 1}, nil)
	chRepo.EXPECT().ListByCourse(mock.Anything, uint(1)).
		Return([]entity.Chapter{{ID: 1}, {ID: 2}}, nil)

	resp, err := svc.ListByCourse(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestChapterService_ListByCourse_ParentCourseNotFound(t *testing.T) {
	_, coRepo, svc := newChapterSvc(t)

	coRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.ListByCourse(context.Background(), 99)
	assert.ErrorIs(t, err, ErrCourseNotFound)
}

func TestChapterService_Update_Success(t *testing.T) {
	chRepo, _, svc := newChapterSvc(t)

	existing := &entity.Chapter{ID: 5, Name: "Old", Order: 1, CourseID: 1}
	chRepo.EXPECT().GetByID(mock.Anything, uint(5)).Return(existing, nil)
	chRepo.EXPECT().Update(mock.Anything, existing).Return(nil)

	resp, err := svc.Update(context.Background(), 5, dto.UpdateChapterRequest{
		Name: "New", Order: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, "New", resp.Name)
	assert.Equal(t, 2, resp.Order)
}

func TestChapterService_Update_NotFound(t *testing.T) {
	chRepo, _, svc := newChapterSvc(t)

	chRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.Update(context.Background(), 99, dto.UpdateChapterRequest{Name: "x"})
	assert.ErrorIs(t, err, ErrChapterNotFound)
}

func TestChapterService_Delete_Success(t *testing.T) {
	chRepo, _, svc := newChapterSvc(t)
	chRepo.EXPECT().Delete(mock.Anything, uint(5)).Return(nil)
	assert.NoError(t, svc.Delete(context.Background(), 5))
}

func TestChapterService_Delete_NotFound(t *testing.T) {
	chRepo, _, svc := newChapterSvc(t)
	chRepo.EXPECT().Delete(mock.Anything, uint(99)).Return(repository.ErrNotFound)
	assert.ErrorIs(t, svc.Delete(context.Background(), 99), ErrChapterNotFound)
}
