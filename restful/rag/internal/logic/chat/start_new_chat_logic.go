// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"gozero-rag/internal/model/chat_conversation"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

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

	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 1. Generate new UUID (v7)
	conversationIdObj, err := uuid.NewV7()
	if err != nil {
		return nil, xerr.NewInternalErrMsg("failed to generate uuid")
	}
	conversationId := conversationIdObj.String()

	// 2. Prepare Config JSON
	retrievalConfig := chat_conversation.RetrievalConfig{
		Mode:               req.RetrievalConfig.Mode,
		RerankMode:         req.RetrievalConfig.RerankMode,
		RerankVectorWeight: req.RetrievalConfig.RerankVectorWeight,
		TopN:               req.RetrievalConfig.TopN,
		RerankId:           req.RetrievalConfig.RerankId,
		TopK:               req.RetrievalConfig.TopK,
		Score:              req.RetrievalConfig.Score,
	}

	config := chat_conversation.ConversationConfig{
		LlmId:                   req.LlmId,
		EnableQuoteDoc:          req.EnableQuoteDoc,
		EnableLlmKeywordExtract: req.EnableLlmKeywordExtract,
		EnableTts:               req.EnableTts,
		SystemPrompt:            req.SystemPrompt,
		KbIds:                   req.KbIds,
		Temperature:             req.Temperature,
		RetrievalConfig:         retrievalConfig,
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		logx.Errorf("failed to marshal conversation config: %v", err)
		return nil, xerr.NewInternalErrMsg("failed to process configuration")
	}

	// 3. Create conversation record
	newConversation := &chat_conversation.ChatConversation{
		Id:           conversationId,
		UserId:       userId,
		TenantId:     tenantId,
		Title:        "New Conversation",
		Status:       1, // Normal
		MessageCount: 0,
		Config:       sql.NullString{String: string(configBytes), Valid: true},
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
