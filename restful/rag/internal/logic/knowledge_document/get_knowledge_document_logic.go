// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"context"

	"gozero-rag/internal/model/knowledge_document"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKnowledgeDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文档详情
func NewGetKnowledgeDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKnowledgeDocumentLogic {
	return &GetKnowledgeDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetKnowledgeDocumentLogic) GetKnowledgeDocument(req *types.GetKnowledgeDocumentReq) (resp *types.GetKnowledgeDocumentResp, err error) {
	// 查询文档详情
	doc, err := l.svcCtx.KnowledgeDocumentModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == knowledge_document.ErrNotFound {
			return nil, err
		}
		l.Errorf("查询文档失败: %v", err)
		return nil, err
	}

	// 组装响应
	resp = &types.GetKnowledgeDocumentResp{
		KnowledgeDocumentInfo: types.KnowledgeDocumentInfo{
			Id:              doc.Id,
			KnowledgeBaseId: doc.KnowledgeBaseId,
			DocName:         doc.DocName.String,
			DocType:         doc.DocType,
			DocSize:         doc.DocSize,
			StoragePath:     doc.StoragePath.String,
			Description:     doc.Description.String,
			Status:          doc.Status,
			RunStatus:       doc.RunStatus,
			ChunkNum:        doc.ChunkNum,
			TokenNum:        doc.TokenNum,
			ParserConfig:    doc.ParserConfig,
			Progress:        doc.Progress,
			ProgressMsg:     doc.ProgressMsg.String,
			CreatedBy:       doc.CreatedBy,
			CreatedTime:     doc.CreatedTime,
			UpdatedTime:     doc.UpdatedTime,
		},
	}

	return resp, nil
}
