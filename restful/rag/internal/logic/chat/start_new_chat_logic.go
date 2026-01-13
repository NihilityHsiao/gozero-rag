// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"gozero-rag/internal/model/chat_conversation"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/google/uuid"
)

type StartNewChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 开启新对话,返回一个 会话id
func NewStartNewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartNewChatLogic {
	return &StartNewChatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StartNewChatLogic) StartNewChat(req *types.StartNewChatReq) (resp *types.StartNewChatResp, err error) {
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 1. Generate new UUID (v7)
	conversationIdObj, err := uuid.NewV7()
	if err != nil {
		return nil, xerr.NewInternalErrMsg("failed to generate uuid")
	}
	conversationId := conversationIdObj.String()

	// 2. Create conversation record
	newConversation := &chat_conversation.ChatConversation{
		Id:           conversationId,
		UserId:       userId,
		Title:        "New Conversation",
		Status:       1, // Normal
		MessageCount: 0,
	}

	_, err = l.svcCtx.ChatConversationModel.Insert(l.ctx, newConversation)
	if err != nil {

		logx.Errorf("Failed to insert new conversation: %v", err)
		return nil, xerr.NewInternalErrMsg("failed to create conversation")
	}

	return &types.StartNewChatResp{
		ConversationId: conversationId,
	}, nil
}
