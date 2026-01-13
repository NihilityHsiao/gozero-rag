package doc_processor

import (
	"context"
	"encoding/json"
	"gozero-rag/internal/rag_core/types"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestNewIndexerService(t *testing.T) {
	ctx := context.Background()

	indexSvc, err := NewDocProcessService(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_ = godotenv.Load("../../../.env")

	apiKey := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseUrl := os.Getenv("OPENAI_BASE_URL")
	output, err := indexSvc.Invoke(ctx, &types.ProcessRequest{
		URI: "/Users/lucas/Downloads/03_PM协作指南.md",
		IndexConfig: types.ProcessConfig{
			EnableQACheck:  true,
			Separators:     []string{"\n\n", "\n", "?"},
			ChunkOverlap:   500,
			MaxChunkLength: 1024,
			LlmConfig: types.ProcessLlmConfig{
				QaKey:              apiKey,
				QaBaseUrl:          baseUrl,
				QaModelName:        modelName,
				EmbeddingKey:       "",
				EmbeddingBaseUrl:   "",
				EmbeddingModelName: "",
			},
		},
	})

	for i, chunk := range output {
		prettyJSON, _ := json.MarshalIndent(chunk.MetaData, "", "  ")
		t.Logf("chunk %d - id[%s] - metadata:\n%s", i, chunk.ID, string(prettyJSON))
		//t.Logf("chunk %d - conetent:%v\n\n", i, chunk.Content)
	}

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("chunks 数量: %v", len(output))

}
