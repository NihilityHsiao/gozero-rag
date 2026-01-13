package structure

import (
	"context"
	"testing"

	"gozero-rag/internal/rag_core/constant"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
)

func TestStructureSplitter_Transform(t *testing.T) {
	// 模拟一段包含标题的文本
	content := `前言
这是一个RAG系统的介绍。

第一章 系统架构
系统由前端和后端组成。
后端使用Go-Zero框架。

1.1 核心模块
核心模块包括Transformer和Retriever。
Transformer负责分片。

第二章 部署说明
使用Docker Compose部署。
`

	ctx := context.Background()
	config := &Config{
		MaxChunkSize: 50, // 设置较小的 chunk size 以触发切分
		OverlapSize:  10,
		Separators:   []string{"\n"},
	}

	splitter, err := NewStructureSplitter(ctx, config)
	assert.NoError(t, err)

	src := []*schema.Document{
		{
			ID:      "doc-1",
			Content: content,
			MetaData: map[string]interface{}{
				"source": "test.txt",
			},
		},
	}

	chunks, err := splitter.Transform(ctx, src)
	assert.NoError(t, err)

	// 验证分片结果
	// 预期分割出的 sections:
	// 1. 前言/摘要
	// 2. 第一章 系统架构
	// 3. 1.1 核心模块
	// 4. 第二章 部署说明

	for i, chunk := range chunks {
		t.Logf("Chunk %d:\n%s\nMetaData: %+v\n-------------------", i, chunk.Content, chunk.MetaData)
	}

	// 验证 Metadata 注入 (不再注入到 Content 中)
	// 检查第二个 chunk (第一章 系统架构)
	assert.Equal(t, "第一章 系统架构", chunks[1].MetaData[constant.MetaHeaderContext])
	assert.Equal(t, 1, chunks[1].MetaData[constant.MetaHeaderLevel])
	assert.Equal(t, "chinese", chunks[1].MetaData[constant.MetaHeaderType])

	// 验证前言/摘要 section 的 metadata
	assert.Equal(t, "前言/摘要", chunks[0].MetaData[constant.MetaHeaderContext])
	assert.Equal(t, 0, chunks[0].MetaData[constant.MetaHeaderLevel])
	assert.Equal(t, "implicit", chunks[0].MetaData[constant.MetaHeaderType])

	// 验证最后一个 chunk 属于 "第二章 部署说明"
	lastChunk := chunks[len(chunks)-1]
	assert.Equal(t, "第二章 部署说明", lastChunk.MetaData[constant.MetaHeaderContext])
	assert.Equal(t, "chinese", lastChunk.MetaData[constant.MetaHeaderType])

	// 验证原始 Metadata 被保留
	for _, chunk := range chunks {
		assert.Equal(t, "test.txt", chunk.MetaData["source"], "原始 Metadata 应该被保留")
	}
}
