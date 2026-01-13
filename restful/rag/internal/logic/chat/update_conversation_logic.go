// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新会话标题
func NewUpdateConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConversationLogic {
	return &UpdateConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateConversationLogic) UpdateConversation(req *types.UpdateConversationReq) (resp *types.UpdateConversationResp, err error) {
	// todo: add your logic here and delete this line

	return
}
