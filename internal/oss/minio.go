package oss

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/core/logx"
)

type MinioClient struct {
	client *minio.Client
}

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinioClient{client: client}, nil
}

func (m *MinioClient) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (string, error) {
	opts := minio.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}
	info, err := m.client.PutObject(ctx, bucket, key, reader, size, opts)
	if err != nil {
		return "", err
	}
	return info.Key, nil
}

func (m *MinioClient) FGetObject(ctx context.Context, bucket, key, localPath string) error {
	return m.client.FGetObject(ctx, bucket, key, localPath, minio.GetObjectOptions{})
}

func (m *MinioClient) RemoveObject(ctx context.Context, bucket, key string) error {
	return m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (m *MinioClient) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		err = m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		logx.Infof("Bucket %s created.", bucket)
	}
	return nil
}
