// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 重试/重新解析文档
func NewRetryDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryDocumentLogic {
	return &RetryDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryDocumentLogic) RetryDocument(req *types.RetryDocumentReq) (resp *types.RetryDocumentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
