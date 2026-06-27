package handler

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	svcmocks "github.com/dias-web/lms-system/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAttachmentRouter(t *testing.T) (*svcmocks.MockAttachmentService, http.Handler) {
	svc := svcmocks.NewMockAttachmentService(t)
	r := newTestRouter()
	NewAttachmentHandler(svc).Register(r, nil, nil)
	return svc, r
}

// multipartUpload builds a multipart/form-data body with a lesson_id field and
// a single file part. Returns the body and its Content-Type header.
func multipartUpload(t *testing.T, lessonID, fileName, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if lessonID != "" {
		_ = mw.WriteField("lesson_id", lessonID)
	}
	if fileName != "" {
		fw, err := mw.CreateFormFile("file", fileName)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		_, _ = io.Copy(fw, strings.NewReader(content))
	}
	_ = mw.Close()
	return &buf, mw.FormDataContentType()
}

func TestAttachmentHandler_Upload_OK(t *testing.T) {
	svc, r := setupAttachmentRouter(t)
	svc.EXPECT().
		Upload(mock.Anything, uint(3), mock.MatchedBy(func(in dto.UploadInput) bool {
			return in.FileName == "syllabus.pdf" && in.Size == 5
		})).
		Return(dto.AttachmentResponse{ID: 7, LessonID: 3, FileName: "syllabus.pdf"}, nil)

	body, ct := multipartUpload(t, "3", "syllabus.pdf", "hello")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp dto.AttachmentResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, uint(7), resp.ID)
}

func TestAttachmentHandler_Upload_MissingFile(t *testing.T) {
	_, r := setupAttachmentRouter(t)

	body, ct := multipartUpload(t, "3", "", "")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "INVALID_INPUT", resp.Error.Code)
}

func TestAttachmentHandler_Upload_InvalidLessonID(t *testing.T) {
	_, r := setupAttachmentRouter(t)

	body, ct := multipartUpload(t, "abc", "f.txt", "x")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttachmentHandler_Upload_LessonNotFound(t *testing.T) {
	svc, r := setupAttachmentRouter(t)
	svc.EXPECT().Upload(mock.Anything, uint(99), mock.Anything).
		Return(dto.AttachmentResponse{}, service.ErrLessonNotFound)

	body, ct := multipartUpload(t, "99", "f.txt", "x")
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "LESSON_NOT_FOUND", resp.Error.Code)
}

func TestAttachmentHandler_Download_OK(t *testing.T) {
	svc, r := setupAttachmentRouter(t)
	svc.EXPECT().Download(mock.Anything, uint(7)).Return(&dto.DownloadResult{
		FileName:    "syllabus.pdf",
		ContentType: "application/pdf",
		Size:        5,
		Body:        io.NopCloser(strings.NewReader("hello")),
	}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/download/7", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello", w.Body.String())
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "syllabus.pdf")
}

func TestAttachmentHandler_Download_NotFound(t *testing.T) {
	svc, r := setupAttachmentRouter(t)
	svc.EXPECT().Download(mock.Anything, uint(99)).Return(nil, service.ErrNotFound)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/download/99", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAttachmentHandler_Download_InvalidID(t *testing.T) {
	_, r := setupAttachmentRouter(t)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/download/abc", nil))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAttachmentHandler_ListByLesson_OK(t *testing.T) {
	svc, r := setupAttachmentRouter(t)
	svc.EXPECT().ListByLesson(mock.Anything, uint(3)).Return([]dto.AttachmentResponse{
		{ID: 1, LessonID: 3, FileName: "a.pdf"},
		{ID: 2, LessonID: 3, FileName: "b.png"},
	}, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lessons/3/attachments", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []dto.AttachmentResponse
	decodeJSON(t, w, &resp)
	assert.Len(t, resp, 2)
}

func TestAttachmentHandler_ListByLesson_NotFound(t *testing.T) {
	svc, r := setupAttachmentRouter(t)
	svc.EXPECT().ListByLesson(mock.Anything, uint(99)).Return(nil, service.ErrLessonNotFound)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lessons/99/attachments", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
}
