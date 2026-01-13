package retriever

import (
	"context"
	"gozero-rag/internal/config"
	vectorstore "gozero-rag/internal/vector_store"
	"os"
	"testing"
)

func TestNewRetrieverService(t *testing.T) {
	ctx := context.Background()
	client, err := vectorstore.NewMilvusClient(config.VectorStoreConf{
		Type:     "milvus",
		Endpoint: "localhost:19530",
		Username: "",
		Password: "",
		Database: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewRetrieverService(ctx, client)
	if err != nil {
		t.Fatal(err)
	}

	apiKey := os.Getenv("EMB_API_KEY")
	modelName := os.Getenv("EMB_MODEL_NAME")
	baseUrl := os.Getenv("EMB_BASE_URL")

	if apiKey == "" || modelName == "" || baseUrl == "" {
		t.Skip("EMB_API_KEY / EMB_MODEL_NAME / EMB_BASE_URL not set")
	}

	query, err := svc.Query(ctx, &RetrieveRequest{
		Query:           "电商场景如何防止超卖？",
		KnowledgeBaseId: 31,
		TopK:            3,
		EmbeddingModelConfig: ModelConfig{
			ModelName: modelName,
			BaseUrl:   baseUrl,
			ApiKey:    apiKey,
		},
		RerankModelConfig: ModelConfig{},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, doc := range query {
		t.Logf("%#v\n\n", doc)
	}

}
