package rerank

import (
	"context"
	"os"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestOpenAiReranker_Rerank(t *testing.T) {
	reranker, err := NewOpenAiReranker()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	apiKey := os.Getenv("RERANK_API_KEY")

	modelName := os.Getenv("RERANK_MODEL_NAME")
	baseUrl := os.Getenv("RERANK_BASE_URL")

	if apiKey == "" || modelName == "" || baseUrl == "" {
		t.Skip("RERANK_API_KEY / RERANK_MODEL_NAME / RERANK_BASE_URL not set")
	}

	rerank, err := reranker.Rerank(ctx, &RerankRequest{
		BaseUrl:   baseUrl,
		ApiKey:    apiKey,
		ModelName: modelName,
		Query:     "黄色的水果",
		Docs: []*schema.Document{
			{
				ID:       "4",
				Content:  "apple",
				MetaData: nil,
			}, {
				ID:       "2",
				Content:  "banana",
				MetaData: nil,
			}, {
				ID:       "1",
				Content:  "fruit",
				MetaData: nil,
			}, {
				ID:       "3",
				Content:  "苹果",
				MetaData: nil,
			},
		},
		TopK: 4,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("rerank result: %+v", rerank)

}
