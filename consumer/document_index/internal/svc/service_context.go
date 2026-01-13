package svc

import (
	"context"
	"gozero-rag/consumer/document_index/internal/config"
	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/model/user_api"
	"gozero-rag/internal/model/vector"
	"gozero-rag/internal/oss"
	"gozero-rag/internal/rag_core/doc_processor"
	vectorstore "gozero-rag/internal/vector_store"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                      config.Config
	SqlConn                     sqlx.SqlConn
	OssClient                   oss.Client
	VectorClient                vectorstore.Client
	KnowledgeBaseModel          knowledge.KnowledgeBaseModel
	KnowledgeDocumentModel      knowledge.KnowledgeDocumentModel
	KnowledgeDocumentChunkModel knowledge.KnowledgeDocumentChunkModel
	KnowledgeVectorModel        vector.KnowledgeVectorModel // New
	UserApiModel                user_api.UserApiModel

	DocProcessService *doc_processor.ProcessorService
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	// Init OSS Client
	ossClient, err := oss.NewClient(c.Oss)
	if err != nil {
		panic(err)
	}

	err = ossClient.EnsureBucket(context.Background(), c.Oss.BucketName)
	if err != nil {
		logx.Error(err)
	}

	vectorClient, err := vectorstore.NewClient(c.VectorStore)
	if err != nil {
		panic(err)
	}

	docProcessService, err := doc_processor.NewDocProcessService(context.Background())
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:                      c,
		SqlConn:                     sqlConn,
		OssClient:                   ossClient,
		VectorClient:                vectorClient,
		KnowledgeBaseModel:          knowledge.NewKnowledgeBaseModel(sqlConn),
		KnowledgeDocumentModel:      knowledge.NewKnowledgeDocumentModel(sqlConn),
		KnowledgeDocumentChunkModel: knowledge.NewKnowledgeDocumentChunkModel(sqlConn),
		KnowledgeVectorModel:        vector.NewKnowledgeVectorModel(vectorClient), // New
		UserApiModel:                user_api.NewUserApiModel(sqlConn, c.Cache),
		DocProcessService:           docProcessService,
	}
}
