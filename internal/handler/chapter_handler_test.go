package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	svcmocks "github.com/dias-web/lms-system/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupChapterRouter(t *testing.T) (*svcmocks.MockChapterService, http.Handler) {
	svc := svcmocks.NewMockChapterService(t)
	r := newTestRouter()
	NewChapterHandler(svc).Register(r)
	return svc, r
}

func TestChapterHandler_ListByCourse_OK(t *testing.T) {
	svc, r := setupChapterRouter(t)
	svc.EXPECT().ListByCourse(mock.Anything, uint(1)).
		Return([]dto.ChapterResponse{{ID: 1, CourseID: 1}, {ID: 2, CourseID: 1}}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/courses/1/chapters", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []dto.ChapterResponse
	decodeJSON(t, w, &resp)
	assert.Len(t, resp, 2)
}

func TestChapterHandler_ListByCourse_ParentNotFound(t *testing.T) {
	svc, r := setupChapterRouter(t)
	svc.EXPECT().ListByCourse(mock.Anything, uint(99)).
		Return(nil, service.ErrCourseNotFound)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/courses/99/chapters", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "COURSE_NOT_FOUND", resp.Error.Code)
}

func TestChapterHandler_GetByID_OK(t *testing.T) {
	svc, r := setupChapterRouter(t)
	svc.EXPECT().GetByIDWithLessons(mock.Anything, uint(5)).
		Return(dto.ChapterResponse{ID: 5}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/chapters/5", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChapterHandler_GetByID_NotFound(t *testing.T) {
	svc, r := setupChapterRouter(t)
	svc.EXPECT().GetByIDWithLessons(mock.Anything, uint(99)).
		Return(dto.ChapterResponse{}, service.ErrChapterNotFound)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/chapters/99", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "CHAPTER_NOT_FOUND", resp.Error.Code)
}

func TestChapterHandler_Create_OK(t *testing.T) {
	svc, r := setupChapterRouter(t)
	svc.EXPECT().
		Create(mock.Anything, dto.CreateChapterRequest{Name: "Intro", CourseID: 1, Order: 1}).
		Return(dto.ChapterResponse{ID: 10, Name: "Intro", CourseID: 1}, nil)

	body := bytes.NewBufferString(`{"name":"Intro","course_id":1,"order":1}`)
	req := httptest.NewRequest(http.MethodPost, "/chapters", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestChapterHandler_Create_ParentCourseNotFound(t *testing.T) {
	svc, r := setupChapterRouter(t)
	svc.EXPECT().Create(mock.Anything, mock.Anything).
		Return(dto.ChapterResponse{}, service.ErrCourseNotFound)

	body := bytes.NewBufferString(`{"name":"Intro","course_id":99,"order":1}`)
	req := httptest.NewRequest(http.MethodPost, "/chapters", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "COURSE_NOT_FOUND", resp.Error.Code)
}

func TestChapterHandler_Create_ValidationError(t *testing.T) {
	_, r := setupChapterRouter(t)

	// missing required course_id
	body := bytes.NewBufferString(`{"name":"Intro"}`)
	req := httptest.NewRequest(http.MethodPost, "/chapters", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChapterHandler_Update_OK(t *testing.T) {
	svc, r := setupChapterRouter(t)
	svc.EXPECT().
		Update(mock.Anything, uint(5), dto.UpdateChapterRequest{Name: "New", Order: 2}).
		Return(dto.ChapterResponse{ID: 5, Name: "New", Order: 2}, nil)

	body := bytes.NewBufferString(`{"name":"New","order":2}`)
	req := httptest.NewRequest(http.MethodPut, "/chapters/5", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChapterHandler_Delete_OK(t *testing.T) {
	svc, r := setupChapterRouter(t)
	svc.EXPECT().Delete(mock.Anything, uint(5)).Return(nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/chapters/5", nil))

	assert.Equal(t, http.StatusNoContent, w.Code)
}
