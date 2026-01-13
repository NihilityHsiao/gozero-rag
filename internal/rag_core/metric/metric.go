package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "rag"

var (
	// ==================== 检索相关指标 ====================

	// RetrievalDuration 检索操作延迟
	RetrievalDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "retrieval_duration_seconds",
		Help:      "检索操作耗时（秒）",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"mode", "kb_id"})

	// RetrievalTotal 检索请求总数
	RetrievalTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "retrieval_total",
		Help:      "检索请求总数",
	}, []string{"mode", "kb_id"})

	// RetrievalErrors 检索错误总数
	RetrievalErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "retrieval_errors_total",
		Help:      "检索错误总数",
	}, []string{"mode", "kb_id", "error_type"})

	// RerankDuration Rerank 操作延迟
	RerankDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "rerank_duration_seconds",
		Help:      "Rerank 操作耗时（秒）",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"model"})

	// EmbeddingDuration Embedding API 延迟
	EmbeddingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "embedding_duration_seconds",
		Help:      "Embedding API 调用耗时（秒）",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
	}, []string{"model"})

	// ChunksReturned 返回的 Chunk 数量
	ChunksReturned = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "chunks_returned",
		Help:      "每次检索返回的 Chunk 数量",
		Buckets:   []float64{1, 5, 10, 20, 50, 100},
	}, []string{"mode"})

	// ==================== 索引相关指标 ====================

	// IndexingDuration 文档索引总耗时
	IndexingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "indexing_duration_seconds",
		Help:      "文档索引耗时（秒）",
		Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600},
	}, []string{"kb_id", "status"})

	// IndexingTotal 索引请求总数
	IndexingTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "indexing_total",
		Help:      "文档索引请求总数",
	}, []string{"kb_id"})

	// IndexingErrors 索引错误总数
	IndexingErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "indexing_errors_total",
		Help:      "文档索引错误总数",
	}, []string{"kb_id", "error_type"})

	// ChunksIndexed 索引的 Chunk 数量
	ChunksIndexed = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "chunks_indexed",
		Help:      "每次索引写入的 Chunk 数量",
		Buckets:   []float64{1, 5, 10, 20, 50, 100, 200, 500},
	}, []string{"kb_id"})
)
