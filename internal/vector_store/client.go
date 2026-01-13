package vectorstore

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// VectorRecord 向量存储记录
type VectorRecord struct {
	ID              string         // 主键 {chunk_id}_{type}_{index}
	KnowledgeBaseID uint64         // 知识库 ID
	ChunkID         string         // 关联 MySQL chunk.id
	DocID           string         // 关联 MySQL document.id
	Type            string         // "chunk" 或 "qa"
	Content         string         // 原文内容
	Metadata        map[string]any // 元数据
	Vector          []float64      // 向量（可选，由 embedding 模型生成）
}

// SearchResult 搜索结果
type SearchResult struct {
	ID              string
	KnowledgeBaseID uint64
	ChunkID         string
	DocID           string
	Type            string
	Content         string
	Metadata        map[string]any
	Score           float32
}

// Client 向量数据库客户端接口
type Client interface {
	// EnsureCollection 确保 Collection 存在，不存在则创建（含 BM25 支持）
	EnsureCollection(ctx context.Context, collectionName string, dim int) error

	// Insert 插入文档（内部会调用 embedding）
	Insert(ctx context.Context, collectionName string, docs []*schema.Document) error

	// InsertWithVectors 直接插入已有向量的记录
	InsertWithVectors(ctx context.Context, collectionName string, records []*VectorRecord) error

	// Search 向量搜索 (Dense Vector)
	Search(ctx context.Context, collectionName string, queryVector []float64, topK int) ([]*SearchResult, error)

	// FullTextSearch 全文检索 (BM25 Sparse Vector)
	FullTextSearch(ctx context.Context, collectionName string, query string, topK int) ([]*SearchResult, error)

	// HybridSearch 混合检索 (Dense + Sparse BM25，使用 RRF 融合)
	HybridSearch(ctx context.Context, collectionName string, queryVector []float64, queryText string, topK int) ([]*SearchResult, error)

	// Delete 删除指定条件的记录
	Delete(ctx context.Context, collectionName string, expr string) error

	// DropCollection 删除整个 Collection
	DropCollection(ctx context.Context, collectionName string) error

	// Close 关闭连接
	Close() error
}
