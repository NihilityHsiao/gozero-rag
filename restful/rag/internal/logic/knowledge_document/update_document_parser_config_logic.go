// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDocumentParserConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新文档解析配置
func NewUpdateDocumentParserConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDocumentParserConfigLogic {
	return &UpdateDocumentParserConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDocumentParserConfigLogic) UpdateDocumentParserConfig(req *types.UpdateDocumentParserConfigReq) (resp *types.UpdateDocumentParserConfigResp, err error) {
	// todo: add your logic here and delete this line

	return
}
