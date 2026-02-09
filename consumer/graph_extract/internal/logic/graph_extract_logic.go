package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"time"

	openaiemb "github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/zeromicro/go-zero/core/logx"

	"gozero-rag/consumer/graph_extract/internal/svc"
	"gozero-rag/internal/graphrag/types"
	"gozero-rag/internal/model/graph"
	"gozero-rag/internal/model/local_message"
	"gozero-rag/internal/mq"
	"gozero-rag/internal/tools/llmx"
)

type GraphExtractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGraphExtractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GraphExtractLogic {
	return &GraphExtractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GraphExtractLogic) Consume(ctx context.Context, key, value string) error {
	logx.Infof("consume graph extract task: %s", value)

	var msg mq.GraphGenerateMsg
	if err := json.Unmarshal([]byte(value), &msg); err != nil {
		logx.Errorf("unmarshal graph generate msg failed: %v", err)
		return nil // commit offset
	}

	// 使用本地消息表包装业务逻辑，实现分布式一致性
	return l.svcCtx.LocalMsgExecutor.Execute(ctx, local_message.TaskTypeGraphExtract, msg, func(ctx context.Context) error {
		return l.processGraphExtract(ctx, &msg)
	})
}

// processGraphExtract 图谱提取核心逻辑
func (l *GraphExtractLogic) processGraphExtract(ctx context.Context, msg *mq.GraphGenerateMsg) error {
	// get embedding config
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(ctx, msg.KnowledgeBaseId)
	if err != nil {
		logx.Errorf("find knowledge base failed: %v", err)
		return err
	}

	embModelName, embFactory := llmx.GetModelNameFactory(kb.EmbdId)
	// 1. Get LLM Config
	llmModelName, factory := llmx.GetModelNameFactory(msg.LlmId)

	tenantLlm, err := l.svcCtx.TenantLlmModel.FindByTenantFactoryName(ctx, msg.TenantId, factory, llmModelName)
	if err != nil {
		logx.Errorf("find tenant llm failed: %v", err)
		return err
	}

	tenantEmb, err := l.svcCtx.TenantLlmModel.FindByTenantFactoryName(ctx, msg.TenantId, embFactory, embModelName)
	if err != nil {
		logx.Errorf("find tenant embedding failed: %v", err)
		return err
	}

	embDim := 1024 // 后期做成可配置的
	embedder, err := openaiemb.NewEmbedder(ctx, &openaiemb.EmbeddingConfig{
		APIKey:     tenantEmb.ApiKey.String,
		BaseURL:    tenantEmb.ApiBase.String,
		Model:      tenantEmb.LlmName,
		Dimensions: &embDim,
	})
	if err != nil {
		return err
	}
	// 2. Initialize Chat Model
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  tenantLlm.ApiKey.String,
		BaseURL: tenantLlm.ApiBase.String,
		Model:   tenantLlm.LlmName,
	})
	if err != nil {
		logx.Errorf("init chat model failed: %v", err)
		return err
	}

	// 3. Get Chunks
	chunkResult, err := l.svcCtx.ChunkModel.ListByDocId(ctx, msg.DocumentId, "", 1, 1000)
	if err != nil {
		logx.Errorf("list chunks failed: %v", err)
		return err
	}
	if chunkResult.Total == 0 {
		logx.Info("no chunks found for document")
		return nil
	}

	// 4. Extract Graph
	logx.Infof("开始提取知识图谱, doc_id: %s", msg.DocumentId)

	extractResult, err := l.svcCtx.GraphExtractor.ParallelExtract(ctx, chunkResult.Chunks, llm, 10, 8)
	if err != nil {
		logx.Errorf("extract graph failed: %v", err)
		return err
	}
	logx.Infof("doc [%s] ,知识图谱提取完成, 实体数: %d, 关系数: %d", msg.DocumentId, len(extractResult.Entities), len(extractResult.Relations))

	// 4.2 对entity做embedding
	if len(extractResult.Entities) > 0 {
		logx.Infof("开始生成实体embedding, 实体数: %d", len(extractResult.Entities))

		for i, e := range extractResult.Entities {
			embVector, err := embedder.EmbedStrings(ctx, []string{e.Name})
			if err != nil {
				logx.Errorf("实体 [%s] 向量化失败, err:%v", e.Name, err)
				continue
			}
			extractResult.Entities[i].Embedding = embVector[0]
		}
	}

	// 5. Save to NebulaGraph
	if err := l.saveToNebula(ctx, extractResult, msg.KnowledgeBaseId); err != nil {
		logx.Errorf("save graph to nebula failed: %v", err)
		return err
	}

	// 6. Transform to ES Documents (with embeddings)
	esDocs := l.transformToEsDocs(extractResult, msg.KnowledgeBaseId)

	// 7. Save to ES
	if err := l.svcCtx.GraphModel.Put(ctx, esDocs); err != nil {
		logx.Errorf("save graph to es failed: %v", err)
		// 不阻塞流程
	}

	logx.Infof("graph extraction completed for doc: %s", msg.DocumentId)
	return nil
}

func (l *GraphExtractLogic) transformToEsDocs(result *types.GraphExtractionResult, kbId string) []*graph.EsGraphDocument {
	docs := make([]*graph.EsGraphDocument, 0, len(result.Entities)+len(result.Relations))

	// Entities
	for _, entity := range result.Entities {
		id := l.genEntityId(kbId, entity.Name)
		doc := &graph.EsGraphDocument{
			Id:          id, // Use deterministic ID for upsert
			KbId:        kbId,
			GraphType:   "entity",
			EntityName:  entity.Name,
			Description: entity.Description,
			Weight:      1.0, // Default weight for entity
			SourceIds:   entity.SourceId,
			Embedding:   entity.Embedding, // 向量嵌入
			UpdatedAt:   l.nowStr(),
		}
		// Content backup (不含 embedding，避免重复存储)
		entityBackup := types.Entity{
			Name:        entity.Name,
			Type:        entity.Type,
			Description: entity.Description,
			SourceId:    entity.SourceId,
		}
		contentBytes, _ := json.Marshal(entityBackup)
		doc.ContentWithWeight = string(contentBytes)
		docs = append(docs, doc)
	}

	// 只存实体
	//// Relations
	//for _, rel := range result.Relations {
	//	id := l.genRelationId(kbId, rel.SrcId, rel.DstId)
	//	doc := &graph.EsGraphDocument{
	//		Id:          id,
	//		KbId:        kbId,
	//		GraphType:   "relation",
	//		SrcName:     rel.SrcId,
	//		DstName:     rel.DstId,
	//		Description: rel.Description,
	//		Weight:      rel.Weight,
	//		SourceIds:   rel.SourceId,
	//		UpdatedAt:   l.nowStr(),
	//	}
	//	contentBytes, _ := json.Marshal(rel)
	//	doc.ContentWithWeight = string(contentBytes)
	//	docs = append(docs, doc)
	//}

	return docs
}

func (l *GraphExtractLogic) genEntityId(kbId, name string) string {
	// Prefix: entity_
	hash := md5.Sum([]byte(kbId + name))
	return "entity_" + hex.EncodeToString(hash[:])
}

func (l *GraphExtractLogic) genRelationId(kbId, src, dst string) string {
	// Prefix: relation_
	hash := md5.Sum([]byte(kbId + src + dst))
	return "relation_" + hex.EncodeToString(hash[:])
}

func (l *GraphExtractLogic) nowStr() string {
	return time.Now().Format(time.RFC3339)
}

// saveToNebula 将图谱数据存储到 NebulaGraph
// 只调用 Model 层接口，不写 nGQL 语句
func (l *GraphExtractLogic) saveToNebula(ctx context.Context, result *types.GraphExtractionResult, kbId string) error {
	// 1. 确保 Space 和 Schema 存在
	if err := l.svcCtx.NebulaGraphModel.EnsureSpaceAndSchema(ctx, kbId); err != nil {
		return err
	}

	// 2. 批量写入实体
	if err := l.svcCtx.NebulaGraphModel.BatchUpsertEntities(ctx, kbId, result.Entities); err != nil {
		return err
	}

	// 3. 批量写入关系
	if err := l.svcCtx.NebulaGraphModel.BatchInsertRelations(ctx, kbId, result.Relations); err != nil {
		return err
	}

	return nil
}
