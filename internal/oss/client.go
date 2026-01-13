package oss

import (
	"context"
	"io"
)

type Client interface {
	// PutObject uploads data. size can be -1 if unknown (for some providers)
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (string, error)
	// FGetObject downloads data to a local file path
	FGetObject(ctx context.Context, bucket, key, localPath string) error
	// RemoveObject deletes data
	RemoveObject(ctx context.Context, bucket, key string) error
	// EnsureBucket ensures the bucket exists
	EnsureBucket(ctx context.Context, bucket string) error
}
