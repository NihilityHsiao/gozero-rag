package retriever

import (
	"context"
	"fmt"
	"gozero-rag/internal/rag_core/metric"
	"gozero-rag/internal/rag_core/rerank"
	vectorstore "gozero-rag/internal/vector_store"
	"sort"
	"strconv"
	"time"

	"github.com/cloudwego/eino/compose"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

const CtxRetrieveRequestKey = "retrieve_request"

type ModelConfig struct {
	ModelName string
	BaseUrl   string
	ApiKey    string
}

type RetrieveRequest struct {
	Query           string
	KnowledgeBaseId uint64 // 知识库id
	TopK            int

	EmbeddingModelConfig ModelConfig
	RerankModelConfig    ModelConfig

	Mode           RetrieveMode
	ScoreThreshold float64

	HybridRankType string // hybrid 重排序类型: rerank / weighted

	VectorWeight  float64
	KeywordWeight float64
}

func NewRetrieverService(ctx context.Context, client vectorstore.Client) (*RetrieverService, error) {
	const (
		NodeRetriever = "Retriever"
		NodeRerank    = "Rerank"
		NodeFilter    = "Filter"
	)

	g := compose.NewGraph[string, []*schema.Document]()

	rtr, err := NewVectorRetriever(client)
	if err != nil {
		return nil, err
	}

	reranker, err := rerank.NewOpenAiReranker()
	if err != nil {
		return nil, err
	}

	_ = g.AddRetrieverNode(NodeRetriever, rtr)
	_ = g.AddLambdaNode(NodeRerank, compose.InvokableLambda(func(ctx context.Context, docs []*schema.Document) ([]*schema.Document, error) {
		conf := getRetrieveRequest(ctx)
		if conf == nil {
			return nil, fmt.Errorf("未传递request")
		}

		start := time.Now()
		result, err := reranker.Rerank(ctx, &rerank.RerankRequest{
			BaseUrl:   conf.RerankModelConfig.BaseUrl,
			ApiKey:    conf.RerankModelConfig.ApiKey,
			ModelName: conf.RerankModelConfig.ModelName,
			Query:     conf.Query,
			Docs:      docs,
			TopK:      conf.TopK,
		})
		// 记录 Rerank 延迟指标
		metric.RerankDuration.WithLabelValues(conf.RerankModelConfig.ModelName).Observe(time.Since(start).Seconds())
		return result, err

	}))
	_ = g.AddLambdaNode(NodeFilter, compose.InvokableLambda(filterDocs))

	_ = g.AddEdge(compose.START, NodeRetriever)
	_ = g.AddEdge(NodeRetriever, NodeRerank)
	_ = g.AddEdge(NodeRerank, NodeFilter)
	_ = g.AddEdge(NodeFilter, compose.END)

	r, err := g.Compile(ctx, compose.WithGraphName(NodeRetriever))

	if err != nil {
		return nil, err
	}

	return &RetrieverService{vectorStore: client, runner: r}, nil
}

type RetrieverService struct {
	vectorStore vectorstore.Client
	runner      compose.Runnable[string, []*schema.Document]
}

// 从 context 中提取 RetrieveRequest
func getRetrieveRequest(ctx context.Context) *RetrieveRequest {
	if req, ok := ctx.Value(CtxRetrieveRequestKey).(*RetrieveRequest); ok {
		return req
	}
	return nil // 或返回默认配置
}

func (r *RetrieverService) checkReq(req *RetrieveRequest) error {
	if req.TopK <= 0 {
		return fmt.Errorf("invalid req.TopK: %d", req.TopK)
	}

	if req.EmbeddingModelConfig.ModelName == "" {
		return fmt.Errorf("empty embedding model name")
	}

	if req.EmbeddingModelConfig.BaseUrl == "" {
		return fmt.Errorf("empty embdding base url")
	}

	if req.EmbeddingModelConfig.ApiKey == "" {

	}

	if req.RerankModelConfig.ModelName == "" {

	}

	if req.RerankModelConfig.BaseUrl == "" {

	}

	if req.RerankModelConfig.ApiKey == "" {

	}

	return nil
}
func (r *RetrieverService) Query(ctx context.Context, req *RetrieveRequest, opts ...retriever.Option) ([]*schema.Document, error) {
	if err := r.checkReq(req); err != nil {
		return nil, err
	}
	ctx = r.WithRetrieveRequest(ctx, req)

	start := time.Now()
	kbId := strconv.FormatUint(req.KnowledgeBaseId, 10)
	mode := string(req.Mode)

	// 记录请求总数
	metric.RetrievalTotal.WithLabelValues(mode, kbId).Inc()

	docs, err := r.Retrieve(ctx, req.Query, opts...)

	// 记录延迟
	metric.RetrievalDuration.WithLabelValues(mode, kbId).Observe(time.Since(start).Seconds())

	if err != nil {
		metric.RetrievalErrors.WithLabelValues(mode, kbId, "retrieve_error").Inc()
		return nil, err
	}

	// 记录返回的 Chunk 数量
	metric.ChunksReturned.WithLabelValues(mode).Observe(float64(len(docs)))

	return docs, nil
}

func (r *RetrieverService) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	retrieveRequest := getRetrieveRequest(ctx)
	if retrieveRequest == nil {
		return nil, fmt.Errorf("未传递request")
	}

	data, err := r.runner.Invoke(ctx, query, compose.WithRetrieverOption(opts...))

	return data, err

}

func (r *RetrieverService) WithRetrieveRequest(ctx context.Context, req *RetrieveRequest) context.Context {
	return context.WithValue(ctx, CtxRetrieveRequestKey, req)
}

func filterDocs(ctx context.Context, docs []*schema.Document) ([]*schema.Document, error) {
	conf := getRetrieveRequest(ctx)
	if conf == nil {
		return nil, fmt.Errorf("retrieve request not found in context")
	}

	// 1. Filter by ScoreThreshold
	var filteredDocs []*schema.Document
	for _, doc := range docs {
		if doc.Score() >= conf.ScoreThreshold {
			filteredDocs = append(filteredDocs, doc)
		}
	}

	// 2. Sort by Score desc (Reranker usually sorts, but ensure it)
	sort.Slice(filteredDocs, func(i, j int) bool {
		return filteredDocs[i].Score() > filteredDocs[j].Score()
	})

	// 3. Truncate by TopK
	if conf.TopK > 0 && len(filteredDocs) > conf.TopK {
		filteredDocs = filteredDocs[:conf.TopK]
	}

	return filteredDocs, nil
}
