package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/entity"
	"github.com/dias-web/lms-system/internal/repository"
	repomocks "github.com/dias-web/lms-system/internal/repository/mocks"
	svcmocks "github.com/dias-web/lms-system/internal/service/mocks"
	"github.com/dias-web/lms-system/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newAttachmentSvc(t *testing.T) (
	*repomocks.MockAttachmentRepository,
	*repomocks.MockLessonRepository,
	*svcmocks.MockObjectStore,
	AttachmentService,
) {
	aRepo := repomocks.NewMockAttachmentRepository(t)
	lRepo := repomocks.NewMockLessonRepository(t)
	store := svcmocks.NewMockObjectStore(t)
	return aRepo, lRepo, store, NewAttachmentService(aRepo, lRepo, store, silentLogger())
}

func sampleUpload() dto.UploadInput {
	return dto.UploadInput{
		FileName:    "notes.pdf",
		ContentType: "application/pdf",
		Size:        12,
		Body:        strings.NewReader("hello world!"),
	}
}

func TestAttachmentService_Upload_Success(t *testing.T) {
	aRepo, lRepo, store, svc := newAttachmentSvc(t)

	lRepo.EXPECT().GetByID(mock.Anything, uint(3)).Return(&entity.Lesson{ID: 3}, nil)
	store.EXPECT().
		Upload(mock.Anything, mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, "lessons/3/") && strings.HasSuffix(key, ".pdf")
		}), mock.Anything, int64(12), "application/pdf").
		Return(nil)
	aRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*entity.Attachment")).
		Run(func(_ context.Context, a *entity.Attachment) { a.ID = 9 }).
		Return(nil)

	resp, err := svc.Upload(context.Background(), 3, sampleUpload())
	require.NoError(t, err)
	assert.Equal(t, uint(9), resp.ID)
	assert.Equal(t, uint(3), resp.LessonID)
	assert.Equal(t, "notes.pdf", resp.FileName)
	assert.Equal(t, "application/pdf", resp.ContentType)
}

func TestAttachmentService_Upload_DefaultsContentType(t *testing.T) {
	aRepo, lRepo, store, svc := newAttachmentSvc(t)

	lRepo.EXPECT().GetByID(mock.Anything, uint(1)).Return(&entity.Lesson{ID: 1}, nil)
	store.EXPECT().
		Upload(mock.Anything, mock.Anything, mock.Anything, mock.Anything, "application/octet-stream").
		Return(nil)
	aRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil)

	in := sampleUpload()
	in.ContentType = ""
	resp, err := svc.Upload(context.Background(), 1, in)
	require.NoError(t, err)
	assert.Equal(t, "application/octet-stream", resp.ContentType)
}

func TestAttachmentService_Upload_LessonNotFound(t *testing.T) {
	_, lRepo, _, svc := newAttachmentSvc(t)

	lRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.Upload(context.Background(), 99, sampleUpload())
	assert.ErrorIs(t, err, ErrLessonNotFound)
}

func TestAttachmentService_Upload_EmptyFileName(t *testing.T) {
	_, _, _, svc := newAttachmentSvc(t)

	in := sampleUpload()
	in.FileName = "   "
	_, err := svc.Upload(context.Background(), 3, in)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestAttachmentService_Upload_StoreError(t *testing.T) {
	_, lRepo, store, svc := newAttachmentSvc(t)

	lRepo.EXPECT().GetByID(mock.Anything, uint(3)).Return(&entity.Lesson{ID: 3}, nil)
	boom := errors.New("minio down")
	store.EXPECT().Upload(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(boom)

	_, err := svc.Upload(context.Background(), 3, sampleUpload())
	assert.ErrorIs(t, err, boom)
}

func TestAttachmentService_Upload_MetadataFailRollsBackObject(t *testing.T) {
	aRepo, lRepo, store, svc := newAttachmentSvc(t)

	lRepo.EXPECT().GetByID(mock.Anything, uint(3)).Return(&entity.Lesson{ID: 3}, nil)
	store.EXPECT().Upload(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	boom := errors.New("insert failed")
	aRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(boom)
	// The object must be removed so we don't orphan it.
	store.EXPECT().Remove(mock.Anything, mock.AnythingOfType("string")).Return(nil)

	_, err := svc.Upload(context.Background(), 3, sampleUpload())
	assert.ErrorIs(t, err, boom)
}

func TestAttachmentService_Download_Success(t *testing.T) {
	aRepo, _, store, svc := newAttachmentSvc(t)

	aRepo.EXPECT().GetByID(mock.Anything, uint(5)).Return(&entity.Attachment{
		ID: 5, FileName: "notes.pdf", ObjectKey: "lessons/3/x.pdf", ContentType: "application/pdf",
	}, nil)
	body := io.NopCloser(strings.NewReader("data"))
	store.EXPECT().Download(mock.Anything, "lessons/3/x.pdf").Return(&storage.Object{
		Body: body, Size: 4, ContentType: "application/pdf",
	}, nil)

	res, err := svc.Download(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, "notes.pdf", res.FileName)
	assert.Equal(t, int64(4), res.Size)
	_ = res.Body.Close()
}

func TestAttachmentService_Download_AttachmentNotFound(t *testing.T) {
	aRepo, _, _, svc := newAttachmentSvc(t)

	aRepo.EXPECT().GetByID(mock.Anything, uint(5)).Return(nil, repository.ErrNotFound)

	_, err := svc.Download(context.Background(), 5)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestAttachmentService_Download_ObjectMissing(t *testing.T) {
	aRepo, _, store, svc := newAttachmentSvc(t)

	aRepo.EXPECT().GetByID(mock.Anything, uint(5)).Return(&entity.Attachment{
		ID: 5, ObjectKey: "lessons/3/x.pdf",
	}, nil)
	store.EXPECT().Download(mock.Anything, "lessons/3/x.pdf").Return(nil, storage.ErrObjectNotFound)

	_, err := svc.Download(context.Background(), 5)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestAttachmentService_ListByLesson_Success(t *testing.T) {
	aRepo, lRepo, _, svc := newAttachmentSvc(t)

	lRepo.EXPECT().GetByID(mock.Anything, uint(3)).Return(&entity.Lesson{ID: 3}, nil)
	aRepo.EXPECT().ListByLesson(mock.Anything, uint(3)).Return([]entity.Attachment{
		{ID: 1, LessonID: 3, FileName: "a.pdf"},
		{ID: 2, LessonID: 3, FileName: "b.png"},
	}, nil)

	list, err := svc.ListByLesson(context.Background(), 3)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestAttachmentService_ListByLesson_LessonNotFound(t *testing.T) {
	_, lRepo, _, svc := newAttachmentSvc(t)

	lRepo.EXPECT().GetByID(mock.Anything, uint(99)).Return(nil, repository.ErrNotFound)

	_, err := svc.ListByLesson(context.Background(), 99)
	assert.ErrorIs(t, err, ErrLessonNotFound)
}

func TestAttachmentService_Delete_Success(t *testing.T) {
	aRepo, _, store, svc := newAttachmentSvc(t)

	aRepo.EXPECT().GetByID(mock.Anything, uint(5)).Return(&entity.Attachment{
		ID: 5, ObjectKey: "lessons/3/x.pdf",
	}, nil)
	aRepo.EXPECT().Delete(mock.Anything, uint(5)).Return(nil)
	store.EXPECT().Remove(mock.Anything, "lessons/3/x.pdf").Return(nil)

	err := svc.Delete(context.Background(), 5)
	require.NoError(t, err)
}

func TestAttachmentService_Delete_NotFound(t *testing.T) {
	aRepo, _, _, svc := newAttachmentSvc(t)

	aRepo.EXPECT().GetByID(mock.Anything, uint(5)).Return(nil, repository.ErrNotFound)

	err := svc.Delete(context.Background(), 5)
	assert.ErrorIs(t, err, ErrNotFound)
}