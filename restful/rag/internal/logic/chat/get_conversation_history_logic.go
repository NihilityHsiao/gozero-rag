// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取会话历史消息
func NewGetConversationHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationHistoryLogic {
	return &GetConversationHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationHistoryLogic) GetConversationHistory(req *types.GetConversationHistoryReq) (resp *types.GetConversationHistoryResp, err error) {
	// Check if user has permission
	return nil, err
}
