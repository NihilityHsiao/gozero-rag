// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"encoding/json"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDocByDocIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取指定文档的信息
func NewGetDocByDocIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDocByDocIdLogic {
	return &GetDocByDocIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDocByDocIdLogic) GetDocByDocId(req *types.GetDocByDocIdReq) (resp *types.KnowledgeDocumentInfo, err error) {
	doc, err := l.svcCtx.KnowledgeDocumentModel.FindOne(l.ctx, req.DocId)
	if err != nil {
		return nil, err
	}

	if doc.KnowledgeBaseId != req.KnowledgeBaseId {
		return nil, nil // Or return a specific error if preferred, here treated as not found or unauthorized
	}

	var parserConfig types.SegmentationSettings
	if doc.ParserConfig != "" {
		if err := json.Unmarshal([]byte(doc.ParserConfig), &parserConfig); err != nil {
			logx.Errorf("failed to unmarshal parser config: %v", err)
			// Decide whether to return error or continue with empty config. comprehensive approach is just logging it.
		}
	}

	return &types.KnowledgeDocumentInfo{
		Id:              doc.Id,
		KnowledgeBaseId: doc.KnowledgeBaseId,
		DocName:         doc.DocName,
		DocType:         doc.DocType,
		DocSize:         doc.DocSize,
		StoragePath:     doc.StoragePath,
		Status:          doc.Status,
		ChunkCount:      doc.ChunkCount,
		ErrMsg:          doc.ErrMsg,
		ParserConfig:    parserConfig,
		CreatedAt:       doc.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       doc.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
