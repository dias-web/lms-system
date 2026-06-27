// Package storage wraps the MinIO (S3-compatible) object store used to keep
// uploaded files. Metadata about each object lives in Postgres; the bytes live
// here.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/dias-web/lms-system/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ErrObjectNotFound is returned when a requested object key does not exist.
var ErrObjectNotFound = errors.New("object not found")

// Client is a thin wrapper over the MinIO SDK scoped to a single bucket.
type Client struct {
	mc     *minio.Client
	bucket string
}

// Object is a readable object stream together with its size and content type.
// Callers must Close the Body once done.
type Object struct {
	Body        io.ReadCloser
	Size        int64
	ContentType string
}

// New builds a MinIO client from config. It does not perform any network call;
// EnsureBucket should be invoked at startup to create the bucket if missing.
func New(cfg config.MinIOConfig) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	return &Client{mc: mc, bucket: cfg.Bucket}, nil
}

// EnsureBucket creates the configured bucket if it does not already exist. It
// is safe to call repeatedly and on every startup.
func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", c.bucket, err)
	}
	if exists {
		return nil
	}
	if err := c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket %q: %w", c.bucket, err)
	}
	return nil
}

// Upload streams r (of the given size and content type) to the object key.
func (c *Client) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := c.mc.PutObject(ctx, c.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("upload object %q: %w", key, err)
	}
	return nil
}

// Download opens the object for reading. The caller owns the returned Body and
// must Close it.
func (c *Client) Download(ctx context.Context, key string) (*Object, error) {
	obj, err := c.mc.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("open object %q: %w", key, err)
	}
	// GetObject is lazy; Stat forces the request so a missing key surfaces here
	// rather than on the first Read.
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		if isNotFound(err) {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("stat object %q: %w", key, err)
	}
	return &Object{Body: obj, Size: info.Size, ContentType: info.ContentType}, nil
}

// Remove deletes the object key. A missing key is not treated as an error.
func (c *Client) Remove(ctx context.Context, key string) error {
	if err := c.mc.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object %q: %w", key, err)
	}
	return nil
}

func isNotFound(err error) bool {
	return minio.ToErrorResponse(err).Code == "NoSuchKey"
}
