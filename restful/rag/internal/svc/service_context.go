// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"gozero-rag/internal/model/chat_conversation"
	"gozero-rag/internal/model/chat_message"
	"gozero-rag/internal/model/chunk"
	"gozero-rag/internal/model/knowledge_base"
	"gozero-rag/internal/model/knowledge_document"
	"gozero-rag/internal/model/knowledge_retrieval_log"
	"gozero-rag/internal/model/llm_factories"
	"gozero-rag/internal/model/tenant"
	"gozero-rag/internal/model/tenant_llm"
	"gozero-rag/internal/model/user"
	"gozero-rag/internal/model/user_api"
	"gozero-rag/internal/model/user_tenant"

	"gozero-rag/internal/oss"
	"gozero-rag/internal/rag_core/retriever"
	"gozero-rag/restful/rag/internal/config"
	"gozero-rag/restful/rag/internal/mq"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	MqPusherClient  mq.Mq
	RedisClient     *redis.Redis
	OssClient       oss.Client
	UserModel       user.UserModel
	UserApiModel    user_api.UserApiModel
	TenantModel     tenant.TenantModel          // 租户Model
	UserTenantModel user_tenant.UserTenantModel // 用户租户关联Model

	KnowledgeBaseModel     knowledge_base.KnowledgeBaseModel
	KnowledgeDocumentModel knowledge_document.KnowledgeDocumentModel
	ChunkModel             chunk.ChunkModel

	KnowledgeRetrievalLogModel knowledge_retrieval_log.KnowledgeRetrievalLogModel

	RetrieveSvc *retriever.RetrieverService

	ChatConversationModel chat_conversation.ChatConversationModel
	ChatMessageModel      chat_message.ChatMessageModel

	// LLM 厂商和租户 LLM 配置
	LlmFactoriesModel llm_factories.LlmFactoriesModel
	TenantLlmModel    tenant_llm.TenantLlmModel
}

func NewServiceContext(c config.Config) *ServiceContext {

	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	rdb := redis.MustNewRedis(redis.RedisConf{
		Host: c.Cache[0].Host,
		Type: c.Cache[0].Type,
		User: c.Cache[0].User,
		Pass: c.Cache[0].Pass,
	})

	// Init OSS Client
	ossClient, err := oss.NewClient(c.Oss)
	if err != nil {
		panic(err)
	}

	err = ossClient.EnsureBucket(context.Background(), c.Oss.BucketName)
	if err != nil {
		logx.Error(err)
		panic(err)
	}

	// Phase 3: Switch to Elasticsearch
	esModel, err := chunk.NewEsChunkModel(c.ElasticSearch.Addresses, c.ElasticSearch.Username, c.ElasticSearch.Password)
	if err != nil {
		logx.Errorf("Failed to init Elasticsearch: %v", err)
		panic(err) // ES is mandatory
	}

	ctx := context.Background()
	retrieverSvc, err := retriever.NewRetrieverService(ctx, esModel)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config: c,

		MqPusherClient:  mq.NewKafka(kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic)),
		RedisClient:     rdb,
		OssClient:       ossClient,
		UserModel:       user.NewUserModel(sqlConn, c.Cache),
		UserApiModel:    user_api.NewUserApiModel(sqlConn, c.Cache),
		TenantModel:     tenant.NewTenantModel(sqlConn, c.Cache),
		UserTenantModel: user_tenant.NewUserTenantModel(sqlConn, c.Cache),

		KnowledgeBaseModel:     knowledge_base.NewKnowledgeBaseModel(sqlConn, c.Cache),
		KnowledgeDocumentModel: knowledge_document.NewKnowledgeDocumentModel(sqlConn, c.Cache),
		ChunkModel:             esModel,

		KnowledgeRetrievalLogModel: knowledge_retrieval_log.NewKnowledgeRetrievalLogModel(sqlConn),

		RetrieveSvc: retrieverSvc,

		ChatConversationModel: chat_conversation.NewChatConversationModel(sqlConn, c.Cache),
		ChatMessageModel:      chat_message.NewChatMessageModel(sqlConn),

		// LLM 厂商和租户 LLM 配置
		LlmFactoriesModel: llm_factories.NewLlmFactoriesModel(sqlConn, c.Cache),
		TenantLlmModel:    tenant_llm.NewTenantLlmModel(sqlConn, c.Cache),
	}
}
