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

func setupLessonRouter(t *testing.T) (*svcmocks.MockLessonService, http.Handler) {
	svc := svcmocks.NewMockLessonService(t)
	r := newTestRouter()
	NewLessonHandler(svc).Register(r)
	return svc, r
}

func TestLessonHandler_ListByChapter_OK(t *testing.T) {
	svc, r := setupLessonRouter(t)
	svc.EXPECT().ListByChapter(mock.Anything, uint(2)).
		Return([]dto.LessonResponse{{ID: 1, ChapterID: 2}}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/chapters/2/lessons", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []dto.LessonResponse
	decodeJSON(t, w, &resp)
	assert.Len(t, resp, 1)
}

func TestLessonHandler_GetByID_OK(t *testing.T) {
	svc, r := setupLessonRouter(t)
	svc.EXPECT().GetByID(mock.Anything, uint(11)).
		Return(dto.LessonResponse{ID: 11, Name: "L"}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lessons/11", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLessonHandler_GetByID_NotFound(t *testing.T) {
	svc, r := setupLessonRouter(t)
	svc.EXPECT().GetByID(mock.Anything, uint(99)).
		Return(dto.LessonResponse{}, service.ErrLessonNotFound)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lessons/99", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "LESSON_NOT_FOUND", resp.Error.Code)
}

func TestLessonHandler_Create_OK(t *testing.T) {
	svc, r := setupLessonRouter(t)
	svc.EXPECT().
		Create(mock.Anything, dto.CreateLessonRequest{
			Name: "Var", ChapterID: 2, Order: 1, Content: "...",
		}).
		Return(dto.LessonResponse{ID: 77, Name: "Var", ChapterID: 2}, nil)

	body := bytes.NewBufferString(`{"name":"Var","chapter_id":2,"order":1,"content":"..."}`)
	req := httptest.NewRequest(http.MethodPost, "/lessons", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLessonHandler_Create_ParentChapterNotFound(t *testing.T) {
	svc, r := setupLessonRouter(t)
	svc.EXPECT().Create(mock.Anything, mock.Anything).
		Return(dto.LessonResponse{}, service.ErrChapterNotFound)

	body := bytes.NewBufferString(`{"name":"Valid Name","chapter_id":99}`)
	req := httptest.NewRequest(http.MethodPost, "/lessons", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "CHAPTER_NOT_FOUND", resp.Error.Code)
}

func TestLessonHandler_Create_ValidationError(t *testing.T) {
	_, r := setupLessonRouter(t)

	body := bytes.NewBufferString(`{"name":"Valid Name"}`) // missing chapter_id
	req := httptest.NewRequest(http.MethodPost, "/lessons", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLessonHandler_Update_OK(t *testing.T) {
	svc, r := setupLessonRouter(t)
	svc.EXPECT().
		Update(mock.Anything, uint(5), dto.UpdateLessonRequest{Name: "New", Content: "c", Order: 2}).
		Return(dto.LessonResponse{ID: 5, Name: "New", Content: "c", Order: 2}, nil)

	body := bytes.NewBufferString(`{"name":"New","content":"c","order":2}`)
	req := httptest.NewRequest(http.MethodPut, "/lessons/5", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLessonHandler_Delete_OK(t *testing.T) {
	svc, r := setupLessonRouter(t)
	svc.EXPECT().Delete(mock.Anything, uint(5)).Return(nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/lessons/5", nil))

	assert.Equal(t, http.StatusNoContent, w.Code)
}
