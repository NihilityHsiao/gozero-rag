// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteKnowledgeDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除文档
func NewDeleteKnowledgeDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteKnowledgeDocumentLogic {
	return &DeleteKnowledgeDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteKnowledgeDocumentLogic) DeleteKnowledgeDocument(req *types.DeleteKnowledgeDocumentReq) (resp *types.DeleteKnowledgeDocumentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
