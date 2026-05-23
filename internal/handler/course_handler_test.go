package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	svcmocks "github.com/dias-web/lms-system/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCourseRouter(t *testing.T) (*svcmocks.MockCourseService, http.Handler) {
	svc := svcmocks.NewMockCourseService(t)
	r := newTestRouter()
	NewCourseHandler(svc).Register(r)
	return svc, r
}

func TestCourseHandler_List_OK(t *testing.T) {
	svc, r := setupCourseRouter(t)
	svc.EXPECT().List(mock.Anything).Return([]dto.CourseResponse{
		{ID: 1, Name: "Go"},
		{ID: 2, Name: "Py"},
	}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/courses", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []dto.CourseResponse
	decodeJSON(t, w, &resp)
	assert.Len(t, resp, 2)
}

func TestCourseHandler_List_InternalError(t *testing.T) {
	svc, r := setupCourseRouter(t)
	svc.EXPECT().List(mock.Anything).Return(nil, errors.New("db down"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/courses", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "INTERNAL_ERROR", resp.Error.Code)
}

func TestCourseHandler_GetByID_OK(t *testing.T) {
	svc, r := setupCourseRouter(t)
	svc.EXPECT().GetByIDWithChapters(mock.Anything, uint(7)).
		Return(dto.CourseResponse{ID: 7, Name: "Go"}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/courses/7", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.CourseResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, uint(7), resp.ID)
}

func TestCourseHandler_GetByID_NotFound(t *testing.T) {
	svc, r := setupCourseRouter(t)
	svc.EXPECT().GetByIDWithChapters(mock.Anything, uint(99)).
		Return(dto.CourseResponse{}, service.ErrCourseNotFound)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/courses/99", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "COURSE_NOT_FOUND", resp.Error.Code)
}

func TestCourseHandler_GetByID_InvalidID(t *testing.T) {
	_, r := setupCourseRouter(t)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/courses/abc", nil))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "INVALID_INPUT", resp.Error.Code)
}

func TestCourseHandler_Create_OK(t *testing.T) {
	svc, r := setupCourseRouter(t)
	svc.EXPECT().
		Create(mock.Anything, dto.CreateCourseRequest{Name: "Go", Description: "d"}).
		Return(dto.CourseResponse{ID: 5, Name: "Go", Description: "d"}, nil)

	body := bytes.NewBufferString(`{"name":"Go","description":"d"}`)
	req := httptest.NewRequest(http.MethodPost, "/courses", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp dto.CourseResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, uint(5), resp.ID)
}

func TestCourseHandler_Create_ValidationError(t *testing.T) {
	_, r := setupCourseRouter(t)

	// name = "a" violates min=2
	body := bytes.NewBufferString(`{"name":"a"}`)
	req := httptest.NewRequest(http.MethodPost, "/courses", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "INVALID_INPUT", resp.Error.Code)
}

func TestCourseHandler_Update_OK(t *testing.T) {
	svc, r := setupCourseRouter(t)
	svc.EXPECT().
		Update(mock.Anything, uint(5), dto.UpdateCourseRequest{Name: "New", Description: "x"}).
		Return(dto.CourseResponse{ID: 5, Name: "New", Description: "x"}, nil)

	body := bytes.NewBufferString(`{"name":"New","description":"x"}`)
	req := httptest.NewRequest(http.MethodPut, "/courses/5", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseHandler_Update_NotFound(t *testing.T) {
	svc, r := setupCourseRouter(t)
	svc.EXPECT().
		Update(mock.Anything, uint(99), mock.Anything).
		Return(dto.CourseResponse{}, service.ErrCourseNotFound)

	body := bytes.NewBufferString(`{"name":"Valid Name"}`)
	req := httptest.NewRequest(http.MethodPut, "/courses/99", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "COURSE_NOT_FOUND", resp.Error.Code)
}

func TestCourseHandler_Delete_OK(t *testing.T) {
	svc, r := setupCourseRouter(t)
	svc.EXPECT().Delete(mock.Anything, uint(5)).Return(nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/courses/5", nil))

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestCourseHandler_Delete_NotFound(t *testing.T) {
	svc, r := setupCourseRouter(t)
	svc.EXPECT().Delete(mock.Anything, uint(99)).Return(service.ErrCourseNotFound)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/courses/99", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
}
