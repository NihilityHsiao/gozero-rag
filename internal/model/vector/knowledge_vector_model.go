package vector

import (
	"context"
	"fmt"
	vectorstore "gozero-rag/internal/vector_store"
	"gozero-rag/internal/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

// KnowledgeVectorItem 知识库向量数据项
type KnowledgeVectorItem struct {
	ID              string    `json:"id"`
	KnowledgeBaseID uint64    `json:"knowledge_base_id"`
	ChunkID         string    `json:"chunk_id"`
	DocID           string    `json:"doc_id"`
	Content         string    `json:"content"`
	Type            string    `json:"type"`
	Vector          []float64 `json:"vector"`
}

// KnowledgeVectorModel 知识库向量模型接口
type KnowledgeVectorModel interface {
	Insert(ctx context.Context, items []*KnowledgeVectorItem) error
	Search(ctx context.Context, kbId uint64, vector []float64, topK int) ([]*vectorstore.SearchResult, error)
	FullTextSearch(ctx context.Context, kbId uint64, query string, topK int) ([]*vectorstore.SearchResult, error)
	HybridSearch(ctx context.Context, kbId uint64, vector []float64, query string, topK int) ([]*vectorstore.SearchResult, error)
	DeleteByDocId(ctx context.Context, kbId uint64, docId string) error
	DropCollection(ctx context.Context, kbId uint64) error
}

type customKnowledgeVectorModel struct {
	client vectorstore.Client
}

func NewKnowledgeVectorModel(client vectorstore.Client) KnowledgeVectorModel {
	return &customKnowledgeVectorModel{
		client: client,
	}
}

// collectionName 生成集合名称 kb_{id}
func (m *customKnowledgeVectorModel) collectionName(kbId uint64) string {
	return fmt.Sprintf("kb_%d", kbId)
}

func (m *customKnowledgeVectorModel) Insert(ctx context.Context, items []*KnowledgeVectorItem) error {
	if len(items) == 0 {
		return nil
	}

	kbId := items[0].KnowledgeBaseID
	collection := m.collectionName(kbId)
	logx.Infof("[VectorModel] Insert into %s, count: %d", collection, len(items))

	// 1. Ensure Collection Exists
	// 假设所有 item 维度一致，取第一个
	dim := len(items[0].Vector)
	if dim == 0 {
		return xerr.NewErrCodeMsg(xerr.VectorStoreInsertError, "Vector dimension is 0")
	}

	if err := m.client.EnsureCollection(ctx, collection, dim); err != nil {
		logx.Errorf("[VectorModel] EnsureCollection failed: %v", err)
		return xerr.NewErrCodeMsg(xerr.VectorStoreCollectionError, err.Error())
	}

	// 2. Convert to VectorRecord
	records := make([]*vectorstore.VectorRecord, len(items))
	for i, item := range items {
		records[i] = &vectorstore.VectorRecord{
			ID:              item.ID,
			KnowledgeBaseID: item.KnowledgeBaseID,
			ChunkID:         item.ChunkID,
			DocID:           item.DocID,
			Type:            item.Type,
			Content:         item.Content,
			Vector:          item.Vector,
		}
	}

	// 3. Insert
	if err := m.client.InsertWithVectors(ctx, collection, records); err != nil {
		logx.Errorf("[VectorModel] InsertWithVectors failed: %v", err)
		return xerr.NewErrCodeMsg(xerr.VectorStoreInsertError, err.Error())
	}

	return nil
}

func (m *customKnowledgeVectorModel) Search(ctx context.Context, kbId uint64, vector []float64, topK int) ([]*vectorstore.SearchResult, error) {
	collection := m.collectionName(kbId)
	return m.client.Search(ctx, collection, vector, topK)
}

func (m *customKnowledgeVectorModel) FullTextSearch(ctx context.Context, kbId uint64, query string, topK int) ([]*vectorstore.SearchResult, error) {
	collection := m.collectionName(kbId)
	return m.client.FullTextSearch(ctx, collection, query, topK)
}

func (m *customKnowledgeVectorModel) HybridSearch(ctx context.Context, kbId uint64, vector []float64, query string, topK int) ([]*vectorstore.SearchResult, error) {
	collection := m.collectionName(kbId)
	return m.client.HybridSearch(ctx, collection, vector, query, topK)
}

func (m *customKnowledgeVectorModel) DeleteByDocId(ctx context.Context, kbId uint64, docId string) error {
	collection := m.collectionName(kbId)
	expr := fmt.Sprintf("doc_id == \"%s\"", docId)
	return m.client.Delete(ctx, collection, expr)
}

func (m *customKnowledgeVectorModel) DropCollection(ctx context.Context, kbId uint64) error {
	collection := m.collectionName(kbId)
	// 忽略删除不存在的集合错误
	_ = m.client.DropCollection(ctx, collection)
	return nil
}
