package transformer

import (
	"context"
	"gozero-rag/internal/rag_core/constant"
	"gozero-rag/internal/rag_core/tools/docx"
	"gozero-rag/internal/rag_core/transformer/structure"
	"gozero-rag/internal/rag_core/types"
	"gozero-rag/internal/xerr"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
)

// Transformer 实现分块功能
type Transformer struct {
}

func NewTransformer(ctx context.Context) (document.Transformer, error) {
	return &Transformer{}, nil
}

func (t *Transformer) getTransformer(ctx context.Context, conf types.ProcessConfig, doc *schema.Document) (document.Transformer, error) {
	if docx.IsMarkdown(doc) {
		// 使用markdown transformer
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
		logx.Infof("使用 markdown transformer, doc: %s", doc.ID)
		return mdSplitter, nil
	}

	// 判断使用哪种类型的transform, recursive 或 agentic
	// 如果要使用 agentic 要先加个配置，先不加了
	// 使用 StructureSplitter (支持启发式标题检测)
	structureSplitter, err := structure.NewStructureSplitter(ctx, &structure.Config{
		MaxChunkSize: conf.MaxChunkLength,
		OverlapSize:  conf.ChunkOverlap,
		Separators:   conf.Separators,
	})

	logx.Infof("使用 structure transformer (heuristic), doc: %s", doc.ID)
	return structureSplitter, err
}

func (t *Transformer) Transform(ctx context.Context, src []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	// 1. 获取动态配置
	value := ctx.Value(constant.CtxKeyIndexConfig)
	conf, ok := value.(types.ProcessConfig)
	if !ok {
		return nil, xerr.NewInternalErrMsg("获取分割配置失败: Context中未找到IndexConfig")
	}

	logx.Infof("Transformer收到任务, 文档数: %d, 配置: %+v", len(src), conf)

	var result []*schema.Document

	// 2. 遍历文档处理 (因为不同文档可能有不同类型，though通常一批是一样的)
	for _, doc := range src {
		// 识别文档类型，这里简单通过 ContentType 或 扩展名判断，或者默认 Recursive
		// 假设 src 的 MetaData 中有 file_type 或者直接尝试用 Markdown splitter

		// 这里简化逻辑：尝试用 Markdown 处理，如果是其他类型则用 Recursive
		// 实际项目中可能需要更明确的类型判断

		var splitter document.Transformer
		var err error

		splitter, err = t.getTransformer(ctx, conf, doc)
		if err != nil {
			return nil, err
		}

		// 执行分割
		chunks, err := splitter.Transform(ctx, []*schema.Document{doc})
		if err != nil {
			logx.Errorf("文档分割失败 doc_id=%s: %v", doc.ID, err)
			return nil, err
		}

		result = append(result, chunks...)
	}

	return result, nil
}
