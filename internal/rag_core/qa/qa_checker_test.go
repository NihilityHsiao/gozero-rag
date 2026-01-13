package qa

import (
	"context"
	"os"
	"testing"

	"gozero-rag/internal/rag_core/constant"
	"gozero-rag/internal/rag_core/types"

	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func TestQaChecker_Check(t *testing.T) {
	// 尝试加载环境变量 (可选)
	_ = godotenv.Load("../../../.env")

	apiKey := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseUrl := os.Getenv("OPENAI_BASE_URL")

	ctx := context.WithValue(context.Background(), constant.CtxKeyIndexConfig, types.ProcessConfig{
		KnowledgeName:  "测试知识库",
		EnableQACheck:  true,
		Separators:     []string{"\n\n", "\n"},
		ChunkOverlap:   100,
		MaxChunkLength: 500,
		PreCleanRule:   types.IndexConfigPreCleanRule{},
		LlmConfig: types.ProcessLlmConfig{
			QaKey:              apiKey,
			QaBaseUrl:          baseUrl,
			QaModelName:        modelName,
			EmbeddingKey:       "",
			EmbeddingBaseUrl:   "",
			EmbeddingModelName: "",
		},
	})

	// 创建多个测试 chunk，模拟真实场景
	docs := []*schema.Document{
		{
			ID: "chunk-1",
			Content: `Go 语言是一种静态类型的编译型语言。它由 Google 开发，于 2009 年首次发布。
Go 语言具有以下特点：
1. 简洁的语法
2. 高效的并发处理 (goroutine)
3. 快速的编译速度
4. 内置垃圾回收机制`,
			MetaData: map[string]any{
				constant.MetaSourceURI:     "/docs/go-intro.md",
				constant.MetaHeaderContext: "第一章 Go 语言简介",
			},
		},
		{
			ID: "chunk-2",
			Content: `goroutine 是 Go 语言中轻量级的并发执行单元。
与传统线程相比，goroutine 占用更少的内存（约 2KB），创建和销毁的开销也更小。
通过 go 关键字即可启动一个新的 goroutine。
channel 是 goroutine 之间通信的管道，遵循 CSP (Communicating Sequential Processes) 模型。`,
			MetaData: map[string]any{
				constant.MetaSourceURI:     "/docs/go-intro.md",
				constant.MetaHeaderContext: "1.1 并发模型",
			},
		},
		{
			ID:      "chunk-3",
			Content: `这是一个很短的 chunk。`, // 过短的 chunk
			MetaData: map[string]any{
				constant.MetaSourceURI: "/docs/go-intro.md",
			},
		},
		{
			ID:      "chunk-4",
			Content: `它很好用，`, // 以代词开头，逗号结尾
			MetaData: map[string]any{
				constant.MetaSourceURI: "/docs/go-intro.md",
			},
		},
	}

	qa := NewQaChecker()
	output, err := qa.Check(ctx, docs)
	if err != nil {
		t.Fatalf("check err: %v", err)
	}

	// 输出每个 chunk 的评分结果
	t.Log("========== QA Checker 结果 ==========")
	for i, doc := range output {
		t.Logf("\n--- Chunk %d [%s] ---", i+1, doc.ID)
		t.Logf("内容: %s", truncate(doc.Content, 50))

		// 输出评分
		if score, ok := doc.MetaData[constant.MetaQualityScore].(float64); ok {
			t.Logf("总分: %.2f / 100", score)
		}

		// 输出详细评分
		if details, ok := doc.MetaData[constant.MetaQualityDetails].(*types.ChunkQualityScore); ok {
			t.Logf("  长度: %.2f, 结构: %.2f, 内容: %.2f, 语义: %.2f, QA: %.2f",
				details.LengthScore, details.StructureScore, details.ContentScore,
				details.SemanticScore, details.QAScore)
			if len(details.Issues) > 0 {
				t.Logf("  问题: %v", details.Issues)
			}
			if len(details.Suggestions) > 0 {
				t.Logf("  建议: %v", details.Suggestions)
			}
		}

		// 输出 QA pairs (如果有)
		if qaPairs, ok := doc.MetaData[constant.MetaQaPairs].([]types.QAItem); ok && len(qaPairs) > 0 {
			t.Logf("  生成的 QA (%d 个):", len(qaPairs))
			for j, qa := range qaPairs {
				t.Logf("    Q%d: %s", j+1, truncate(qa.Question, 60))
				t.Logf("    A%d: %s", j+1, truncate(qa.Answer, 60))
			}
		}
	}
	t.Log("\n========== 检查完成 ==========")
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
