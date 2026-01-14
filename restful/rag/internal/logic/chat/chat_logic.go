package chat

import (
	"context"
	"fmt"
	"gozero-rag/internal/model/chat_conversation"
	"gozero-rag/internal/rag_core/retriever"
	"gozero-rag/internal/xerr"
	sse2 "gozero-rag/restful/rag/internal/sse"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"

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
	// todo: add your logic here and delete this line
	// Note: msgId should ideally be generated or passed.
	msgId := fmt.Sprintf("id-%v", time.Now().UnixMilli())
	sse := sse2.NewSSEClient(msgId, client)

	failTask := func(info string) error {
		sse.SendError("任务执行失败")
		logx.Errorf("chat err:%v, info:%s", err, info)
		sse.SendError(info)
		return err
	}

	conv, err := l.getConversation(req.ConversationId)
	if err != nil {
		return failTask("获取会话失败")
	}

	// 首次提问, 用第一个问题作为会话标题
	if conv.MessageCount == 0 {
		_ = l.setConversationTitle(conv, req.Message)
	}

	docs, err := l.retrieve(req)

	// Just sending retrieval content for now? The original code did this.
	// Real chat logic needs LLM generation. But here we just stream docs?
	// The standard RAG flow: Retrieve -> Generate.
	// The original code only streamed docs. I will preserve that behavior for now.

	for _, doc := range docs {
		sse.SendText(doc.Content + "\n\n") // Append newline for separation
	}

	sse.SendFinish()

	return nil
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
