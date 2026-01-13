// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"errors"
	"fmt"
	"gozero-rag/internal/rag_core/loader"
	"gozero-rag/internal/rag_core/previewer"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/metric"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zeromicro/go-zero/core/logx"
)

// 预览优化常量
const (
	// 预览时截取的最大内容长度（字符数）
	// 按 MaxChunkLength=1000, 10个chunk计算，取 15000 留有余量
	previewMaxContentLength = 15000
	// 预览返回的最大 chunk 数量
	previewMaxChunks = 10
)

type PreviewChunkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type docIdName struct {
	DocId   string
	DocName string
	Docs    []*schema.Document
	ErrMsg  string
}

// 查看分片效果
func NewPreviewChunkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewChunkLogic {
	return &PreviewChunkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreviewChunkLogic) getDoctype(doctype string) previewer.DocType {
	doctype = strings.ToLower(doctype)
	switch doctype {
	case "md", "markdown":
		return previewer.DocTypeMarkdown
	}
	return previewer.DocTypeText
}
func (l *PreviewChunkLogic) PreviewChunk(req *types.PreviewChunkReq) (resp *types.PreviewChunkResp, err error) {
	// 开始计时
	startTime := time.Now()

	// 用于记录监控的变量
	var docType string = "unknown"
	var originalContentLength int
	var wasTruncated bool

	if len(req.Settings.Separators) == 0 {
		metric.RecordPreviewFail(docType, metric.PreviewFailReasonInvalidParam)
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "Separator cannot be empty")
	}

	// 1. Get document info
	// FindOne expects string ID according to defaultKnowledgeDocumentModel in generated code
	doc, err := l.svcCtx.KnowledgeDocumentModel.FindOne(l.ctx, req.DocId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			metric.RecordPreviewFail(docType, metric.PreviewFailReasonDocNotFound)
			return nil, xerr.NewErrCodeMsg(xerr.KnowledgeDocNotFoundError, "Document not found")
		}
		metric.RecordPreviewFail(docType, metric.PreviewFailReasonServerError)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, fmt.Sprintf("Query document failed: %v", err))
	}

	// 更新文档类型用于监控
	docType = strings.ToLower(doc.DocType)

	// 2. Load document
	fileLoader, err := loader.NewLoader(l.ctx)
	if err != nil {
		metric.RecordPreviewFail(docType, metric.PreviewFailReasonLoadError)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, fmt.Sprintf("Init loader failed: %v", err))
	}

	docs, err := fileLoader.Load(l.ctx, document.Source{
		URI: doc.StoragePath, // Correct field name is StoragePath
	})
	if err != nil {
		metric.RecordPreviewFail(docType, metric.PreviewFailReasonLoadError)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, fmt.Sprintf("Load file failed: %v", err))
	}
	if len(docs) == 0 {
		metric.RecordPreviewFail(docType, metric.PreviewFailReasonLoadError)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "No content loaded from file")
	}
	content := docs[0].Content

	// 记录原始内容长度
	contentRunes := []rune(content)
	originalContentLength = len(contentRunes)

	// 优化：截取前 N 个字符用于预览，避免大文件处理耗时
	if originalContentLength > previewMaxContentLength {
		content = string(contentRunes[:previewMaxContentLength])
		wasTruncated = true
	}

	// 3. Preview/Split
	preConf := previewer.PreviewConf{
		Separator:      req.Settings.Separators,
		MaxChunkLength: req.Settings.MaxChunkLength,
		ChunkOverlap:   req.Settings.ChunkOverlap,
	}

	p, err := previewer.NewPreviewer(l.ctx, preConf)
	if err != nil {
		metric.RecordPreviewFail(docType, metric.PreviewFailReasonSplitError)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, fmt.Sprintf("Init previewer failed: %v", err))
	}

	docTypeEnum := l.getDoctype(doc.DocType) // Assuming doc has DocType field like "md", "txt", "pdf" etc.
	chunks, err := p.Preview(l.ctx, docTypeEnum, req.DocId, content)
	if err != nil {
		metric.RecordPreviewFail(docType, metric.PreviewFailReasonSplitError)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, fmt.Sprintf("Preview chunks failed: %v", err))
	}

	// 4. Assemble response（优化：限制返回的 chunk 数量）
	totalChunks := len(chunks)
	maxChunks := previewMaxChunks
	if totalChunks < maxChunks {
		maxChunks = totalChunks
	}

	respChunks := make([]types.PreviewChunk, 0, maxChunks)
	for i := 0; i < maxChunks; i++ {
		respChunks = append(respChunks, types.PreviewChunk{
			Index:   strconv.Itoa(i + 1),
			Content: chunks[i].Content,
			Length:  len([]rune(chunks[i].Content)), // Character count
		})
	}

	resp = &types.PreviewChunkResp{
		DocId:       req.DocId,
		DocName:     doc.DocName,
		TotalChunks: totalChunks, // 返回预览内容的实际 chunk 数（非全文）
		Chunks:      respChunks,
	}

	// 记录成功监控指标
	durationMs := float64(time.Since(startTime).Milliseconds())
	metric.RecordPreviewSuccess(docType, durationMs, originalContentLength, totalChunks, wasTruncated)

	return resp, nil
}
