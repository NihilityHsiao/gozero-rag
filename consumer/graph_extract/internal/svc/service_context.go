package svc

import (
	"context"
	"gozero-rag/consumer/graph_extract/internal/config"
	"gozero-rag/internal/graphrag/extractor"
	"gozero-rag/internal/model/chunk"
	"gozero-rag/internal/model/graph"
	"gozero-rag/internal/model/tenant_llm"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config           config.Config
	TenantLlmModel   tenant_llm.TenantLlmModel
	ChunkModel       chunk.ChunkModel
	GraphModel       graph.GraphModel       // ES 图数据存储
	NebulaGraphModel graph.NebulaGraphModel // Nebula 图数据存储 (新增)
	GraphExtractor   *extractor.GraphExtractor
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	// Init ES Chunk Model (Read-Only usage mostly)
	chunkModel, err := chunk.NewEsChunkModel(c.ElasticSearch.Addresses, c.ElasticSearch.Username, c.ElasticSearch.Password)
	if err != nil {
		logx.Errorf("NewEsChunkModel failed: %v", err)
		panic(err)
	}

	// Init ES Graph Model
	graphModel, err := graph.NewEsGraphModel(c.ElasticSearch.Addresses, c.ElasticSearch.Username, c.ElasticSearch.Password)
	if err != nil {
		logx.Errorf("NewEsGraphModel failed: %v", err)
		panic(err)
	}

	// Init Nebula Graph Model (新增)
	nebulaGraphModel, err := graph.NewNebulaGraphModel(c.Nebula.Addresses, c.Nebula.Username, c.Nebula.Password)
	if err != nil {
		logx.Errorf("NewNebulaGraphModel failed: %v", err)
		panic(err)
	}

	// Init Graph Extractor (Singleton)
	graphExtractor, err := extractor.NewGraphExtractor(context.Background())
	if err != nil {
		logx.Errorf("NewGraphExtractor failed: %v", err)
		panic(err)
	}

	return &ServiceContext{
		Config:           c,
		TenantLlmModel:   tenant_llm.NewTenantLlmModel(sqlConn, c.Cache),
		ChunkModel:       chunkModel,
		GraphModel:       graphModel,
		NebulaGraphModel: nebulaGraphModel,
		GraphExtractor:   graphExtractor,
	}
}
