package chat

import (
	"context"
	"database/sql"
	"fmt"
	"gozero-rag/internal/model/chat_conversation"
	"gozero-rag/internal/model/chat_message"
	"gozero-rag/internal/rag_core/retriever"
	"gozero-rag/internal/xerr"
	sse2 "gozero-rag/restful/rag/internal/sse"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// sse对话
func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatLogic) getConversation(convId string) (conv *chat_conversation.ChatConversation, err error) {
	conv, err = l.svcCtx.ChatConversationModel.FindOne(l.ctx, convId)
	return
}

func (l *ChatLogic) setConversationTitle(conv *chat_conversation.ChatConversation, title string) error {
	title = strings.TrimSpace(title)
	runes := []rune(title)
	if len(runes) > 20 {
		title = string(runes[:20])
	}

	conv.Title = title
	return l.svcCtx.ChatConversationModel.Update(l.ctx, conv)
}

func (l *ChatLogic) findEmbeddingConfigByIds(knowledgeBaseIds []string) (map[string]retriever.ModelConfig, error) {
	if len(knowledgeBaseIds) == 0 {
		return make(map[string]retriever.ModelConfig), nil
	}

	ret := make(map[string]retriever.ModelConfig)

	// Since generated model might not support bulk fetch effectively or FindByIds is missing
	// We loop and find one by one. Performance is acceptable for reasonably small # of KBs in a chat.
	for _, kbId := range knowledgeBaseIds {
		kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, kbId)
		if err != nil {
			logx.Errorf("failed to find kb %s: %v", kbId, err)
			continue // Skip missing KBs
		}

		// Parse EmbdId (string) to Int64 (UserApiModel ID)
		// Phase 3 note: If switching to TenantLLM, this changes.
		embId, err := strconv.ParseUint(kb.EmbdId, 10, 64)
		if err != nil {
			logx.Errorf("invalid embd_id %s for kb %s", kb.EmbdId, kbId)
			continue
		}

		api, err := l.svcCtx.UserApiModel.FindOne(l.ctx, embId)
		if err != nil {
			logx.Errorf("failed to find embedding model %d for kb %s", embId, kbId)
			continue
		}

		ret[kbId] = retriever.ModelConfig{
			ModelName: api.ModelName,
			BaseUrl:   api.BaseUrl,
			ApiKey:    api.ApiKey,
		}
	}

	return ret, nil
}

func (l *ChatLogic) Chat(req *types.ChatReq, client chan<- *types.ChatResp) (err error) {
	// msgId will be used for both User and Assistant message correlation in frontend if needed,
	// but strictly DB has its own IDs. Here we generate a temporary ID for the stream.
	streamMsgId := fmt.Sprintf("msg-%v", time.Now().UnixMilli())
	sse := sse2.NewSSEClient(streamMsgId, client)

	failTask := func(info string) error {
		sse.SendError(info)
		logx.Errorf("chat err:%v, info:%s", err, info)
		return err
	}

	conv, err := l.getConversation(req.ConversationId)
	if err != nil {
		return failTask("获取会话失败")
	}

	// 1. Initial Title Setting
	if conv.MessageCount == 0 {
		_ = l.setConversationTitle(conv, req.Message)
	}

	// 2. Save User Message
	userSeqId, err := l.getNextSeqId(req.ConversationId)
	if err != nil {
		logx.Errorf("failed to get seq_id: %v", err)
		return failTask("系统错误")
	}

	if err := l.saveMessage(req.ConversationId, userSeqId, "user", req.Message, "", ""); err != nil {
		logx.Errorf("failed to save user message: %v", err)
		return failTask("保存消息失败")
	}

	// Update conversation message count (async or here)
	// For simplicity, we skip updating message_count in conversation strictly here or do it later.

	// 3. Retrieval
	docs, err := l.retrieve(req)
	if err != nil {
		return failTask("检索失败")
	}

	// 4. Stream Retrieval Results (Citations)
	// In the new design, we might want to send "citation" event.
	// But legacy logic sent text. Let's send Citation event if possible, or Text as before?
	// The implementation plan says "Refactor ChatLogic for SSE... Events: message, reasoning, citation".
	// So let's try to construct citation event.
	// However, types.ChatRetrievalChunk matches API.
	retrievalChunks := make([]types.ChatRetrievalChunk, 0)
	for _, doc := range docs {
		chunk := types.ChatRetrievalChunk{
			ChunkID: doc.ID,
			DocID:   "", // chunks often don't have doc_id in schema.Document if not put there.
			Content: doc.Content,
			Score:   doc.Score(),
		}
		// Try to extract metadata if available
		if docId, ok := doc.MetaData["doc_id"].(string); ok {
			chunk.DocID = docId
		}
		if docName, ok := doc.MetaData["doc_name"].(string); ok {
			chunk.DocName = docName
		}
		retrievalChunks = append(retrievalChunks, chunk)
	}

	if len(retrievalChunks) > 0 {
		sse.SendCitation(retrievalChunks)
	}

	// 5. Mock Reasoning (DeepSeek R1 Style)
	// In a real implementation, this comes from the LLM stream.
	sse.SendReasoning("正在分析用户提问...\n")
	time.Sleep(500 * time.Millisecond)
	sse.SendReasoning("根据检索到的上下文，识别到关键信息...\n")
	time.Sleep(500 * time.Millisecond)
	sse.SendReasoning("构建回答逻辑...\n")

	// 6. Mock LLM Response
	answerConfig := "根据检索结果，"
	sse.SendText(answerConfig)
	time.Sleep(200 * time.Millisecond)
	sse.SendText("这是回答的具体内容。\n")

	finalContent := answerConfig + "这是回答的具体内容。\n"
	finalReasoning := "正在分析用户提问...\n根据检索到的上下文，识别到关键信息...\n构建回答逻辑...\n"

	// 7. Save Assistant Message
	asstSeqId := userSeqId + 1
	if err := l.saveMessage(req.ConversationId, asstSeqId, "assistant", finalContent, finalReasoning, ""); err != nil {
		logx.Errorf("failed to save assistant message: %v", err)
	}

	sse.SendFinish()

	return nil
}

func (l *ChatLogic) getNextSeqId(conversationId string) (int, error) {
	// Basic implementation: Count + 1 or Max + 1.
	// Query: select coalesce(max(seq_id), 0) + 1 from chat_message where conversation_id = ?
	// Since we don't have custom query method in generated model, use FindOne/QueryRow.
	// But `chat_message_model.go` is generated. We can use QueryRowNoCacheCtx manually using SqlConn.
	// However, accessing SqlConn directly from logic is via l.svcCtx.ChatMessageModel... which encapsulates it.
	// Inspect ChatMessageModel definition. It usually embeds `defaultChatMessageModel`.
	// We might need to add `FindMaxSeqId` to the model interface or run raw query via l.svcCtx.ChatMessageModel.
	// Since I cannot modify model interface easily without re-generating (and overwriting custom code),
	// I might use a hacky way or just assume 0 for now if too hard? No, user asked for seq_id.
	// I can assume MessageCount in conversation is roughly seq_id * 2?
	// Let's use `MessageCount` from conversation as base?
	// Conversation `MessageCount` is updated.
	// Better: use l.svcCtx.ChatConversationModel.FindOne to get MessageCount, then +1 for user, +2 for assistant.
	conv, err := l.svcCtx.ChatConversationModel.FindOne(l.ctx, conversationId)
	if err != nil {
		return 0, err
	}
	// Update MessageCount logic should be consistent.
	// Let's assume current count is the last one.
	// Optimistic locking or simple increment.
	return int(conv.MessageCount) + 1, nil
}

func (l *ChatLogic) saveMessage(convId string, seqId int, role, content, reasoning, toolCallId string) error {
	uuidStr, _ := uuid.NewV7()

	msg := &chat_message.ChatMessage{
		Id:             uuidStr.String(),
		ConversationId: convId,
		SeqId:          int64(seqId),
		Role:           role,
		Content:        content,
		Type:           "text",
		TokenCount:     0, // TODO: calculate
	}

	// Manual field assignment for new fields if not in struct yet?
	// Wait, generated model should have them.
	// `ReasoningContent`, `ToolCallId`.
	msg.ReasoningContent = sql.NullString{String: reasoning, Valid: reasoning != ""}
	msg.ToolCallId = sql.NullString{String: toolCallId, Valid: toolCallId != ""}

	_, err := l.svcCtx.ChatMessageModel.Insert(l.ctx, msg)

	// Also update conversation message count
	if err == nil {
		_ = l.incrementMessageCount(convId)
	}

	return err
}

func (l *ChatLogic) incrementMessageCount(convId string) error {
	conv, err := l.svcCtx.ChatConversationModel.FindOne(l.ctx, convId)
	if err != nil {
		return err
	}
	conv.MessageCount += 1
	return l.svcCtx.ChatConversationModel.Update(l.ctx, conv)
}

func (l *ChatLogic) retrieve(req *types.ChatReq) (docs []*schema.Document, err error) {
	if len(req.KnowledgeBaseIds) == 0 {
		return []*schema.Document{}, nil
	}

	embMap, err := l.findEmbeddingConfigByIds(req.KnowledgeBaseIds)
	if err != nil {
		return nil, xerr.NewInternalErrMsg("获取emb模型失败")
	}

	rerankConfig, err := l.svcCtx.UserApiModel.FindOne(l.ctx, req.ChatRetrieveConfig.RerankModelId)
	if err != nil {
		return nil, xerr.NewInternalErrMsg("获取rerank模型失败")
	}

	rerankModelConfig := retriever.ModelConfig{
		ModelName: rerankConfig.ModelName,
		BaseUrl:   rerankConfig.BaseUrl,
		ApiKey:    rerankConfig.ApiKey,
	}

	// 使用 WaitGroup 和 Mutex 进行并发查询
	var wg sync.WaitGroup
	var mu sync.Mutex
	docs = make([]*schema.Document, 0)

	for _, kbId := range req.KnowledgeBaseIds {
		embConfig, ok := embMap[kbId]
		if !ok {
			continue
		}

		wg.Add(1)
		go func(kbId string, embConfig retriever.ModelConfig) {
			defer wg.Done()

			getDocs, retrieveErr := l.svcCtx.RetrieveSvc.Query(l.ctx, &retriever.RetrieveRequest{
				Query:                req.Message,
				KnowledgeBaseId:      kbId,
				TopK:                 req.ChatRetrieveConfig.TopK,
				EmbeddingModelConfig: embConfig,
				RerankModelConfig:    rerankModelConfig,
				Mode:                 req.ChatRetrieveConfig.Mode,
				ScoreThreshold:       req.ChatRetrieveConfig.Score,
				HybridRankType:       req.ChatRetrieveConfig.RerankMode,
				VectorWeight:         req.ChatRetrieveConfig.RerankVectorWeight,
				KeywordWeight:        req.ChatRetrieveConfig.RerankKeywordWeight,
			})

			if retrieveErr != nil {
				logx.Errorf("query失败, kb:%s, err:%v", kbId, retrieveErr)
				return
			}

			// 线程安全地追加结果
			mu.Lock()
			docs = append(docs, getDocs...)
			mu.Unlock()
		}(kbId, embConfig)
	}

	wg.Wait()
	return docs, nil
}
