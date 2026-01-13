package rerank

import (
	"context"
	"github.com/cloudwego/eino/schema"
)

type RerankRequest struct {
	BaseUrl   string
	ApiKey    string
	ModelName string

	Query string
	Docs  []*schema.Document
	TopK  int
}
type Reranker interface {
	Rerank(ctx context.Context, req *RerankRequest) ([]*schema.Document, error)
}
