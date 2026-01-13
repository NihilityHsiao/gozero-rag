// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除会话
func NewDeleteConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteConversationLogic {
	return &DeleteConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteConversationLogic) DeleteConversation(req *types.DeleteConversationReq) (resp *types.DeleteConversationResp, err error) {
	// todo: add your logic here and delete this line

	return
}
