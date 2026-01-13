package vectorstore

import (
	"fmt"

	"gozero-rag/internal/config"
)

// NewClient 根据配置创建向量数据库客户端
func NewClient(cfg config.VectorStoreConf) (Client, error) {
	switch cfg.Type {
	case "milvus":
		return NewMilvusClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported vector store type: %s", cfg.Type)
	}
}
