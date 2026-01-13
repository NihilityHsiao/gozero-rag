package retriever

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"

	"gozero-rag/internal/rag_core/metric"
	vectorstore "gozero-rag/internal/vector_store"
)

// VectorRetriever 基于 VectorStore 的通用向量检索器
type VectorRetriever struct {
	client vectorstore.Client
}

// NewVectorRetriever 创建通用向量检索器
func NewVectorRetriever(client vectorstore.Client) (retriever.Retriever, error) {
	return &VectorRetriever{
		client: client,
	}, nil
}

// Retrieve 执行检索
// 流程:
// 1. 从 context 获取 RetrieveRequest
// 2. 根据 req.Mode 决定检索策略 (Vector / FullText / Hybrid)
// 3. 返回融合后的 Document 列表
func (r *VectorRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 1. 获取请求配置
	req := getRetrieveRequest(ctx)
	if req == nil {
		return nil, fmt.Errorf("未传递 RetrieveRequest，请使用 WithRetrieveRequest 设置")
	}

	logx.Infof("[VectorRetriever] 开始检索, query=%s, kb_id=%d, mode=%s", query, req.KnowledgeBaseId, req.Mode)

	// 2. 构建 collection 名称
	collectionName := fmt.Sprintf("kb_%d", req.KnowledgeBaseId)
	searchTopK := req.TopK * 3 // 搜索时多取一些，融合后再截取

	var searchResults []*vectorstore.SearchResult
	var err error

	// 3. 根据模式执行检索
	switch req.Mode {
	case RetrieveModeFulltext:
		// 全文检索 (BM25)
		searchResults, err = r.client.FullTextSearch(ctx, collectionName, query, searchTopK)

	case RetrieveModeHybrid:
		// 混合检索 (Dense + Sparse)
		// 需要先生成向量
		queryVector, errEmbed := r.embedQuery(ctx, req, query)
		if errEmbed != nil {
			return nil, errEmbed
		}
		searchResults, err = r.client.HybridSearch(ctx, collectionName, queryVector, query, searchTopK)

	case RetrieveModeVector:
		fallthrough
	default:
		// 向量检索 (默认)
		queryVector, errEmbed := r.embedQuery(ctx, req, query)
		if errEmbed != nil {
			return nil, errEmbed
		}
		searchResults, err = r.client.Search(ctx, collectionName, queryVector, searchTopK)
	}

	if err != nil {
		logx.Errorf("[VectorRetriever] 检索失败 (mode=%s): %v", req.Mode, err)
		return nil, fmt.Errorf("检索失败: %w", err)
	}

	logx.Infof("[VectorRetriever] 检索完成 (mode=%s), 返回 %d 条结果", req.Mode, len(searchResults))

	// 4. 组装成 document
	docs := make([]*schema.Document, 0, len(searchResults))
	for _, result := range searchResults {
		doc := &schema.Document{
			ID:      result.ChunkID,
			Content: result.Content,
			MetaData: map[string]any{
				MetaChunkID:         result.ChunkID,
				MetaDocID:           result.DocID,
				MetaKnowledgeBaseID: result.KnowledgeBaseID,
				MetaType:            result.Type,
				MetaScore:           result.Score,
				MetaSource:          req.Mode, // 简单标记来源
			},
		}
		doc.WithScore(float64(result.Score))

		docs = append(docs, doc)
	}

	return docs, nil
}

// embedQuery 生成查询向量
func (r *VectorRetriever) embedQuery(ctx context.Context, req *RetrieveRequest, query string) ([]float64, error) {
	embeddingConfig := req.EmbeddingModelConfig

	start := time.Now()

	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:  embeddingConfig.ApiKey,
		BaseURL: embeddingConfig.BaseUrl,
		Model:   embeddingConfig.ModelName,
		Timeout: 30 * time.Second,
	})
	if err != nil {
		logx.Errorf("[VectorRetriever] 创建 Embedder 失败: %v", err)
		return nil, fmt.Errorf("创建 Embedder 失败: %w", err)
	}

	// 生成 query embedding
	embeddings, err := embedder.EmbedStrings(ctx, []string{query})

	// 记录 Embedding 延迟指标
	metric.EmbeddingDuration.WithLabelValues(embeddingConfig.ModelName).Observe(time.Since(start).Seconds())

	if err != nil {
		logx.Errorf("[VectorRetriever] 生成 query embedding 失败: %v", err)
		return nil, fmt.Errorf("生成 query embedding 失败: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("embedding 结果为空")
	}

	return embeddings[0], nil
}
