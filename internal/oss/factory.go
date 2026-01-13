package oss

import (
	"fmt"
	"gozero-rag/internal/config"
)

func NewClient(c config.OssConf) (Client, error) {
	switch c.Type {
	case "minio":
		return NewMinioClient(c.Endpoint, c.AccessKey, c.SecretKey, c.UseSSL)
	default:
		return nil, fmt.Errorf("unknown oss type: %s", c.Type)
	}
}
