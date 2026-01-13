package previewer

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
)

type Previewer struct {
	previewConf      PreviewConf
	markdownSplitter document.Transformer

	recursiveSplitter document.Transformer
}

func NewPreviewer(ctx context.Context, previewConf PreviewConf) (*Previewer, error) {
	mdSplitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"#":    "h1", // 一级标题
			"##":   "h2", // 二级标题
			"###":  "h3", // 三级标题
			"####": "h4", // 四级标题

		},
		TrimHeaders: false, // 是否在输出中保留标题行
	})

	if err != nil {
		return nil, err
	}

	recursiveSplitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   previewConf.MaxChunkLength,
		Separators:  previewConf.Separator,
		OverlapSize: previewConf.ChunkOverlap,

		LenFunc:  nil,                    // 可选：自定义长度计算函数
		KeepType: recursive.KeepTypeNone, // 可选：分隔符保留策略
	})

	return &Previewer{
		previewConf:       previewConf,
		markdownSplitter:  mdSplitter,
		recursiveSplitter: recursiveSplitter,
	}, nil
}

func (p *Previewer) Preview(ctx context.Context, docType DocType, docId string, content string) (result []*schema.Document, err error) {
	docs := []*schema.Document{
		{
			ID:      docId,
			Content: content,
		},
	}

	switch docType {
	case DocTypeMarkdown:
		result, err = p.markdownSplitter.Transform(ctx, docs)
		if err != nil {
			return nil, err
		}
	case DocTypeText:
		result, err = p.recursiveSplitter.Transform(ctx, docs)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的的doc type类型:%v", docType)
	}

	return result, nil
}
