// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"context"

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
	// todo: add your logic here and delete this line

	return
}
