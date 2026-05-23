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

func TestCourseService_Create_Success(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	req := dto.CreateCourseRequest{Name: "Golang Developer", Description: "intro"}

	repo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*entity.Course")).
		Run(func(_ context.Context, c *entity.Course) { c.ID = 42 }).
		Return(nil)

	resp, err := svc.Create(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, uint(42), resp.ID)
	assert.Equal(t, "Golang Developer", resp.Name)
	assert.Equal(t, "intro", resp.Description)
}

func TestCourseService_Create_RepoError(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	boom := errors.New("db is down")
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(boom)

	_, err := svc.Create(context.Background(), dto.CreateCourseRequest{Name: "X"})
	assert.ErrorIs(t, err, boom)
}

func TestCourseService_GetByID_Success(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	repo.EXPECT().GetByID(mock.Anything, uint(7)).
		Return(&entity.Course{ID: 7, Name: "Go"}, nil)

	resp, err := svc.GetByID(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, uint(7), resp.ID)
}

func TestCourseService_GetByID_NotFound(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	repo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.GetByID(context.Background(), 99)
	assert.ErrorIs(t, err, ErrCourseNotFound)
}

func TestCourseService_GetByID_RepoError(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	boom := errors.New("conn refused")
	repo.EXPECT().GetByID(mock.Anything, uint(1)).Return(nil, boom)

	_, err := svc.GetByID(context.Background(), 1)
	assert.ErrorIs(t, err, boom)
}

func TestCourseService_GetByIDWithChapters_Success(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	course := &entity.Course{
		ID:   3,
		Name: "Go",
		Chapters: []entity.Chapter{
			{ID: 10, Name: "Intro"},
		},
	}
	repo.EXPECT().GetByIDWithChapters(mock.Anything, uint(3)).Return(course, nil)

	resp, err := svc.GetByIDWithChapters(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, uint(3), resp.ID)
	require.Len(t, resp.Chapters, 1)
	assert.Equal(t, uint(10), resp.Chapters[0].ID)
}

func TestCourseService_GetByIDWithChapters_NotFound(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	repo.EXPECT().GetByIDWithChapters(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.GetByIDWithChapters(context.Background(), 99)
	assert.ErrorIs(t, err, ErrCourseNotFound)
}

func TestCourseService_List_Success(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	repo.EXPECT().List(mock.Anything).Return([]entity.Course{
		{ID: 1, Name: "Go"},
		{ID: 2, Name: "Py"},
	}, nil)

	resp, err := svc.List(context.Background())
	require.NoError(t, err)
	require.Len(t, resp, 2)
	assert.Equal(t, uint(1), resp[0].ID)
	assert.Equal(t, uint(2), resp[1].ID)
}

func TestCourseService_List_RepoError(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	boom := errors.New("boom")
	repo.EXPECT().List(mock.Anything).Return(nil, boom)

	_, err := svc.List(context.Background())
	assert.ErrorIs(t, err, boom)
}

func TestCourseService_Update_Success(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	existing := &entity.Course{ID: 5, Name: "Old", Description: "old"}
	repo.EXPECT().GetByID(mock.Anything, uint(5)).Return(existing, nil)
	repo.EXPECT().Update(mock.Anything, existing).Return(nil)

	resp, err := svc.Update(context.Background(), 5, dto.UpdateCourseRequest{
		Name: "New", Description: "new",
	})
	require.NoError(t, err)
	assert.Equal(t, "New", resp.Name)
	assert.Equal(t, "new", resp.Description)
}

func TestCourseService_Update_NotFoundOnLoad(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	repo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.Update(context.Background(), 99, dto.UpdateCourseRequest{Name: "x"})
	assert.ErrorIs(t, err, ErrCourseNotFound)
}

func TestCourseService_Update_NotFoundOnSave(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	existing := &entity.Course{ID: 5, Name: "Old"}
	repo.EXPECT().GetByID(mock.Anything, uint(5)).Return(existing, nil)
	repo.EXPECT().Update(mock.Anything, existing).Return(repository.ErrNotFound)

	_, err := svc.Update(context.Background(), 5, dto.UpdateCourseRequest{Name: "x"})
	assert.ErrorIs(t, err, ErrCourseNotFound)
}

func TestCourseService_Delete_Success(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	repo.EXPECT().Delete(mock.Anything, uint(5)).Return(nil)

	assert.NoError(t, svc.Delete(context.Background(), 5))
}

func TestCourseService_Delete_NotFound(t *testing.T) {
	repo := repomocks.NewMockCourseRepository(t)
	svc := NewCourseService(repo, silentLogger())

	repo.EXPECT().Delete(mock.Anything, uint(99)).Return(repository.ErrNotFound)

	err := svc.Delete(context.Background(), 99)
	assert.ErrorIs(t, err, ErrCourseNotFound)
}
