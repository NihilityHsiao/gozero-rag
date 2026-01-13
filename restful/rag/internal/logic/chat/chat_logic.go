// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"fmt"
	"gozero-rag/internal/model/chat_conversation"
	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/rag_core/retriever"
	"gozero-rag/internal/slicex"
	"gozero-rag/internal/xerr"
	sse2 "gozero-rag/restful/rag/internal/sse"
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

func (l *ChatLogic) findEmbeddingConfigByIds(knowledgeBaseIds []uint64) (map[uint64]retriever.ModelConfig, error) {
	if len(knowledgeBaseIds) == 0 {
		return make(map[uint64]retriever.ModelConfig), nil
	}

	// 1. 查询知识库列表
	kbs, err := l.svcCtx.KnowledgeBaseModel.FindByIds(l.ctx, knowledgeBaseIds)
	if err != nil {
		return nil, err
	}

	// 2. 提取所有 embedding 模型 ID
	embIds := slicex.Into(kbs, func(t *knowledge.KnowledgeBase) uint64 {
		return t.EmbeddingModelId
	})

	// 3. 批量查询 embedding 模型配置
	apis, err := l.svcCtx.UserApiModel.FindByIds(l.ctx, embIds)
	if err != nil {
		return nil, err
	}

	// 4. 构建 apiId -> ModelConfig 的映射
	apiConfigMap := make(map[uint64]retriever.ModelConfig)
	for _, api := range apis {
		apiConfigMap[api.Id] = retriever.ModelConfig{
			ModelName: api.ModelName,
			BaseUrl:   api.BaseUrl,
			ApiKey:    api.ApiKey,
		}
	}

	// 5. 构建 knowledgeBaseId -> ModelConfig 的映射
	ret := make(map[uint64]retriever.ModelConfig)
	for _, kb := range kbs {
		if config, ok := apiConfigMap[kb.EmbeddingModelId]; ok {
			ret[kb.Id] = config
		}
	}

	return ret, nil
}

func (l *ChatLogic) Chat(req *types.ChatReq, client chan<- *types.ChatResp) (err error) {
	// todo: add your logic here and delete this line
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

	for _, doc := range docs {
		sse.SendText(doc.Content)
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

	for _, kb := range req.KnowledgeBaseIds {
		emb, ok := embMap[kb]
		if !ok {
			continue
		}

		wg.Add(1)
		go func(kbId uint64, embConfig retriever.ModelConfig) {
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
				logx.Errorf("query失败, kb:%d, err:%v", kbId, retrieveErr)
				return
			}

			// 线程安全地追加结果
			mu.Lock()
			docs = append(docs, getDocs...)
			mu.Unlock()
		}(kb, emb)
	}

	wg.Wait()
	return docs, nil
}
