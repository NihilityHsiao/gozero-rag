// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"gozero-rag/internal/model/chat_conversation"
	"gozero-rag/internal/model/chat_message"
	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/model/knowledge_retrieval_log"
	"gozero-rag/internal/model/user"
	"gozero-rag/internal/model/user_api"
	"gozero-rag/internal/model/vector" // New
	"gozero-rag/internal/oss"
	"gozero-rag/internal/rag_core/retriever"
	vectorstore "gozero-rag/internal/vector_store"
	"gozero-rag/restful/rag/internal/config"
	"gozero-rag/restful/rag/internal/mq"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	MqPusherClient mq.Mq
	RedisClient    *redis.Redis
	OssClient      oss.Client
	UserModel      user.UserModel
	UserApiModel   user_api.UserApiModel
	knowledge.KnowledgeBaseModel
	knowledge.KnowledgeDocumentModel
	knowledge.KnowledgeDocumentChunkModel
	vector.KnowledgeVectorModel // New
	KnowledgeRetrievalLogModel  knowledge_retrieval_log.KnowledgeRetrievalLogModel

	RetrieveSvc *retriever.RetrieverService

	ChatConversationModel chat_conversation.ChatConversationModel
	ChatMessageModel      chat_message.ChatMessageModel
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

	vectorClient, err := vectorstore.NewClient(c.VectorStore)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	retrieverSvc, err := retriever.NewRetrieverService(ctx, vectorClient)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config: c,

		MqPusherClient:              mq.NewKafka(kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic)),
		RedisClient:                 rdb,
		OssClient:                   ossClient,
		UserModel:                   user.NewUserModel(sqlConn, c.Cache),
		UserApiModel:                user_api.NewUserApiModel(sqlConn, c.Cache),
		KnowledgeBaseModel:          knowledge.NewKnowledgeBaseModel(sqlConn),
		KnowledgeDocumentModel:      knowledge.NewKnowledgeDocumentModel(sqlConn),
		KnowledgeDocumentChunkModel: knowledge.NewKnowledgeDocumentChunkModel(sqlConn),
		KnowledgeVectorModel:        vector.NewKnowledgeVectorModel(vectorClient), // New
		KnowledgeRetrievalLogModel:  knowledge_retrieval_log.NewKnowledgeRetrievalLogModel(sqlConn),

		RetrieveSvc: retrieverSvc,

		ChatConversationModel: chat_conversation.NewChatConversationModel(sqlConn, c.Cache),
		ChatMessageModel:      chat_message.NewChatMessageModel(sqlConn),
	}
}
