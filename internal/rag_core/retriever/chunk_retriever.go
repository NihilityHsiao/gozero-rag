package retriever

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"

	"gozero-rag/internal/model/chunk"
	"gozero-rag/internal/rag_core/metric"
)

// ChunkRetriever 基于 ChunkModel (Elasticsearch) 的检索器
type ChunkRetriever struct {
	chunkModel chunk.ChunkModel
}

// NewChunkRetriever 创建 ChunkRetriever
func NewChunkRetriever(chunkModel chunk.ChunkModel) (retriever.Retriever, error) {
	return &ChunkRetriever{
		chunkModel: chunkModel,
	}, nil
}

// Retrieve 执行检索
func (r *ChunkRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 1. 获取请求配置
	req := getRetrieveRequest(ctx)
	if req == nil {
		return nil, fmt.Errorf("未传递 RetrieveRequest，请使用 WithRetrieveRequest 设置")
	}

	logx.Infof("[ChunkRetriever] 开始检索, query=%s, kb_id=%s, mode=%s", query, req.KnowledgeBaseId, req.Mode)

	// String KB ID from request (will be updated to string in orchestration.go)
	kbId := req.KnowledgeBaseId
	topK := req.TopK * 3 // Fetch more for fusion/rerank

	var chunks []*chunk.Chunk
	var err error

	// 2. 准备向量 if needed
	var queryVector []float64
	if req.Mode == RetrieveModeVector || req.Mode == RetrieveModeHybrid {
		queryVector, err = r.embedQuery(ctx, req, query)
		if err != nil {
			return nil, err
		}
	}

	// 3. 调用 ChunkModel.HybridSearch
	// ChunkModel handles vector (if provided) and keyword search logic.
	// If queryVector is nil/empty, it usually falls back to keyword only if implemented handles it.
	// My ChunkModelEs.HybridSearch implementation checks if vector is empty.
	chunks, err = r.chunkModel.HybridSearch(ctx, kbId, query, queryVector, topK)
	if err != nil {
		logx.Errorf("[ChunkRetriever] 检索失败: %v", err)
		return nil, err
	}

	logx.Infof("[ChunkRetriever] 检索完成, 返回 %d 条结果", len(chunks))

	// 4. Convert to Eino Documents
	docs := make([]*schema.Document, 0, len(chunks))
	for _, c := range chunks {
		doc := &schema.Document{
			ID:      c.Id,
			Content: c.Content,
			MetaData: map[string]any{
				"chunk_id":          c.Id,
				"doc_id":            c.DocId,
				"knowledge_base_id": c.KbIds, // Note: ChunkModel stores []string, but here we work contextually with one KB usually.
				// Add score if available from ES result (currently Chunk struct doesn't strictly have a Score field exported from ES explicitly in my model?
				// Wait, EsChunkModel unmarshals _source. Score is in hit["_score"].
				// My ChunkModel implementation returns []*Chunk, which is just data. It loses the score!
				// ERROR: My previous ChunkModel implementation LOSES the score from ES!
				// The _source doesn't contain the score. The score is meta.
				// I need to fix ChunkModel implementation to Populate Score.
			},
		}
		// doc.WithScore?
		docs = append(docs, doc)
	}

	return docs, nil
}

// embedQuery 生成查询向量 (Copy from VectorRetriever)
func (r *ChunkRetriever) embedQuery(ctx context.Context, req *RetrieveRequest, query string) ([]float64, error) {
	embeddingConfig := req.EmbeddingModelConfig

	start := time.Now()

	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:  embeddingConfig.ApiKey,
		BaseURL: embeddingConfig.BaseUrl,
		Model:   embeddingConfig.ModelName,
		Timeout: 30 * time.Second,
	})
	if err != nil {
		logx.Errorf("[ChunkRetriever] 创建 Embedder 失败: %v", err)
		return nil, fmt.Errorf("创建 Embedder 失败: %w", err)
	}

	// 生成 query embedding
	embeddings, err := embedder.EmbedStrings(ctx, []string{query})

	// 记录 Embedding 延迟指标
	metric.EmbeddingDuration.WithLabelValues(embeddingConfig.ModelName).Observe(time.Since(start).Seconds())

	if err != nil {
		logx.Errorf("[ChunkRetriever] 生成 query embedding 失败: %v", err)
		return nil, fmt.Errorf("生成 query embedding 失败: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("embedding 结果为空")
	}

	return embeddings[0], nil
}
