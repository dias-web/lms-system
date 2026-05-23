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

func newLessonSvc(t *testing.T) (*repomocks.MockLessonRepository, *repomocks.MockChapterRepository, LessonService) {
	lRepo := repomocks.NewMockLessonRepository(t)
	chRepo := repomocks.NewMockChapterRepository(t)
	return lRepo, chRepo, NewLessonService(lRepo, chRepo, silentLogger())
}

func TestLessonService_Create_Success(t *testing.T) {
	lRepo, chRepo, svc := newLessonSvc(t)

	chRepo.EXPECT().GetByID(mock.Anything, uint(2)).Return(&entity.Chapter{ID: 2}, nil)
	lRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*entity.Lesson")).
		Run(func(_ context.Context, l *entity.Lesson) { l.ID = 77 }).
		Return(nil)

	resp, err := svc.Create(context.Background(), dto.CreateLessonRequest{
		Name: "Var", ChapterID: 2, Content: "...", Order: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, uint(77), resp.ID)
	assert.Equal(t, uint(2), resp.ChapterID)
}

func TestLessonService_Create_ParentChapterNotFound(t *testing.T) {
	_, chRepo, svc := newLessonSvc(t)

	chRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.Create(context.Background(), dto.CreateLessonRequest{
		Name: "X", ChapterID: 99,
	})
	assert.ErrorIs(t, err, ErrChapterNotFound)
}

func TestLessonService_Create_RepoError(t *testing.T) {
	lRepo, chRepo, svc := newLessonSvc(t)

	chRepo.EXPECT().GetByID(mock.Anything, uint(2)).Return(&entity.Chapter{ID: 2}, nil)
	boom := errors.New("insert failed")
	lRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(boom)

	_, err := svc.Create(context.Background(), dto.CreateLessonRequest{Name: "X", ChapterID: 2})
	assert.ErrorIs(t, err, boom)
}

func TestLessonService_GetByID_Success(t *testing.T) {
	lRepo, _, svc := newLessonSvc(t)

	lRepo.EXPECT().GetByID(mock.Anything, uint(11)).
		Return(&entity.Lesson{ID: 11, Name: "L"}, nil)

	resp, err := svc.GetByID(context.Background(), 11)
	require.NoError(t, err)
	assert.Equal(t, uint(11), resp.ID)
}

func TestLessonService_GetByID_NotFound(t *testing.T) {
	lRepo, _, svc := newLessonSvc(t)

	lRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.GetByID(context.Background(), 99)
	assert.ErrorIs(t, err, ErrLessonNotFound)
}

func TestLessonService_ListByChapter_Success(t *testing.T) {
	lRepo, chRepo, svc := newLessonSvc(t)

	chRepo.EXPECT().GetByID(mock.Anything, uint(2)).Return(&entity.Chapter{ID: 2}, nil)
	lRepo.EXPECT().ListByChapter(mock.Anything, uint(2)).
		Return([]entity.Lesson{{ID: 1}, {ID: 2}, {ID: 3}}, nil)

	resp, err := svc.ListByChapter(context.Background(), 2)
	require.NoError(t, err)
	assert.Len(t, resp, 3)
}

func TestLessonService_ListByChapter_ParentChapterNotFound(t *testing.T) {
	_, chRepo, svc := newLessonSvc(t)
	chRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)
	_, err := svc.ListByChapter(context.Background(), 99)
	assert.ErrorIs(t, err, ErrChapterNotFound)
}

func TestLessonService_Update_Success(t *testing.T) {
	lRepo, _, svc := newLessonSvc(t)

	existing := &entity.Lesson{ID: 5, Name: "Old", Content: "old", Order: 1, ChapterID: 1}
	lRepo.EXPECT().GetByID(mock.Anything, uint(5)).Return(existing, nil)
	lRepo.EXPECT().Update(mock.Anything, existing).Return(nil)

	resp, err := svc.Update(context.Background(), 5, dto.UpdateLessonRequest{
		Name: "New", Content: "new", Order: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, "New", resp.Name)
	assert.Equal(t, "new", resp.Content)
	assert.Equal(t, 2, resp.Order)
}

func TestLessonService_Update_NotFound(t *testing.T) {
	lRepo, _, svc := newLessonSvc(t)
	lRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)
	_, err := svc.Update(context.Background(), 99, dto.UpdateLessonRequest{Name: "x"})
	assert.ErrorIs(t, err, ErrLessonNotFound)
}

func TestLessonService_Delete_Success(t *testing.T) {
	lRepo, _, svc := newLessonSvc(t)
	lRepo.EXPECT().Delete(mock.Anything, uint(5)).Return(nil)
	assert.NoError(t, svc.Delete(context.Background(), 5))
}

func TestLessonService_Delete_NotFound(t *testing.T) {
	lRepo, _, svc := newLessonSvc(t)
	lRepo.EXPECT().Delete(mock.Anything, uint(99)).Return(repository.ErrNotFound)
	assert.ErrorIs(t, svc.Delete(context.Background(), 99), ErrLessonNotFound)
}
