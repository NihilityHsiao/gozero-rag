package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"gozero-rag/consumer/document_index/internal/svc"
	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/model/user_api"
	"gozero-rag/internal/model/vector"
	"gozero-rag/internal/mq"
	"gozero-rag/internal/rag_core/metric"
	"gozero-rag/internal/rag_core/types"
	"gozero-rag/internal/slicex"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/openai"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DocumentIndexLogic struct {
	svcCtx *svc.ServiceContext
	ctx    context.Context
}

func NewDocumentIndexLogic(svcCtx *svc.ServiceContext, ctx context.Context) *DocumentIndexLogic {
	return &DocumentIndexLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
	}
}

func (l *DocumentIndexLogic) Consume(_ context.Context, key, val string) (err error) {
	var msg *mq.KnowledgeDocumentIndexMsg
	err = json.Unmarshal([]byte(val), &msg)
	if err != nil {
		logx.Errorf("消费反序列化失败 val: %s, err: %v", val, err) // 增强日志
		return nil                                         // 这里的 err 不需要抛给 kq，否则会一直重试。如果是格式错误，直接丢弃即可。
	}

	if msg.UserId == "" || msg.KnowledgeBaseId == 0 || msg.DocumentId == "" {
		logx.Errorf("非法参数:%s", val)
		return nil
	}
	logx.Infof("消费 msg: %v", msg)
	return l.mainWork(l.ctx, msg)
}

// DocParserConfig 前端传入的解析配置
type DocParserConfig struct {
	Separators         []string                      `json:"separators"`
	ChunkOverlap       int                           `json:"chunk_overlap"`
	MaxChunkLength     int                           `json:"max_chunk_length"`
	EnableQAGeneration bool                          `json:"enable_qa_generation"`
	PreCleanRule       types.IndexConfigPreCleanRule `json:"pre_clean_rule"`
}

// KBModelIds 知识库绑定的模型 ID
type KBModelIds struct {
	Qa      int64 `json:"qa"`
	Chat    int64 `json:"chat"`
	Rerank  int64 `json:"rerank"`
	Rewrite int64 `json:"rewrite"`
}

func (l *DocumentIndexLogic) mainWork(ctx context.Context, msg *mq.KnowledgeDocumentIndexMsg) (err error) {
	documentId := msg.DocumentId

	// Panic Recovery: 确保 panic 不会导致文档永久停留在 indexing 状态
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("[DocIndex] DocumentId=%s panic: %v", documentId, r)
			_ = l.svcCtx.KnowledgeDocumentModel.UpdateStatus(ctx, documentId, knowledge.StatusDocumentFail, fmt.Sprintf("panic: %v", r))
			err = nil // 避免重试
		}
	}()

	start := time.Now()
	kbId := strconv.FormatUint(msg.KnowledgeBaseId, 10)

	// 记录索引请求总数
	metric.IndexingTotal.WithLabelValues(kbId).Inc()

	// 失败处理辅助函数
	failTask := func(reason string) error {
		logx.Errorf("[DocIndex] DocumentId=%s 失败: %s", documentId, reason)
		_ = l.svcCtx.KnowledgeDocumentModel.UpdateStatus(ctx, documentId, knowledge.StatusDocumentFail, reason)
		// 记录失败指标
		metric.IndexingErrors.WithLabelValues(kbId, "process_error").Inc()
		metric.IndexingDuration.WithLabelValues(kbId, "fail").Observe(time.Since(start).Seconds())
		return nil // 返回 nil 避免重试循环
	}

	// 1. 检查文档状态和是否存在
	doc, err := l.svcCtx.KnowledgeDocumentModel.FindOne(ctx, documentId)
	if err != nil {
		if err == knowledge.ErrNotFound {
			return nil // 数据不存在，跳过
		}
		return err // 数据库错误，重试
	}

	if doc.Status != knowledge.StatusDocumentPending {
		logx.Infof("[DocIndex] DocumentId=%s status is %s, skip", documentId, doc.Status)
		return nil
	}

	// 从 MinIO 下载文件到临时目录
	tempFile, err := os.CreateTemp("", fmt.Sprintf("rag_doc_%s_*%s", doc.Id, filepath.Ext(doc.DocName)))
	if err != nil {
		return failTask(fmt.Sprintf("创建临时文件失败: %v", err))
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	err = l.svcCtx.OssClient.FGetObject(ctx, l.svcCtx.Config.Oss.BucketName, doc.StoragePath, tempFile.Name())
	if err != nil {
		return failTask(fmt.Sprintf("从 MinIO 下载失败: %v", err))
	}
	logx.Infof("[DocIndex] 从 MinIO 下载 %s 到 %s", doc.StoragePath, tempFile.Name())

	// 2. 解析解析器配置
	var parserConfig DocParserConfig
	if err := json.Unmarshal([]byte(doc.ParserConfig), &parserConfig); err != nil {
		return failTask(fmt.Sprintf("解析配置无效: %v", err))
	}

	// 3. 获取模型配置
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(ctx, msg.KnowledgeBaseId)
	if err != nil {
		return failTask(fmt.Sprintf("知识库未找到: %v", err))
	}

	// 解析知识库绑定的模型 ID
	var modelIds KBModelIds
	// kb.ModelIds 在生成的模型中是字符串类型，需要根据实际类型处理
	// 假设可以兼容 json unmarshal string->struct
	if len(kb.ModelIds) > 0 && kb.ModelIds != "{}" {
		if err := json.Unmarshal([]byte(kb.ModelIds), &modelIds); err != nil {
			logx.Errorf("解析知识库 model_ids 失败: %v", err)
			// 继续执行，使用默认值或在关键时失败
		}
	}

	// 获取 Embedding 模型
	embedApiRecord, err := l.svcCtx.UserApiModel.FindOne(ctx, uint64(kb.EmbeddingModelId))
	if err != nil {
		return failTask(fmt.Sprintf("Embedding 模型 (id=%d) 未找到: %v", kb.EmbeddingModelId, err))
	}
	if embedApiRecord.UserId != msg.UserId {
		return failTask("Embedding 模型所有权不匹配")
	}

	// 获取 QA 模型（可选）
	qaKey := ""
	qaBaseUrl := ""
	qaModelName := ""

	if parserConfig.EnableQAGeneration {
		if modelIds.Qa <= 0 {
			return failTask("已启用 QA 生成但知识库未选择 QA 模型")
		}
		qaRecord, err := l.svcCtx.UserApiModel.FindOne(ctx, uint64(modelIds.Qa))
		if err != nil {
			return failTask(fmt.Sprintf("QA 模型 (id=%d) 未找到: %v", modelIds.Qa, err))
		}
		qaKey = qaRecord.ApiKey
		qaBaseUrl = qaRecord.BaseUrl
		qaModelName = qaRecord.ModelName
	}

	// 4. 更新状态为索引中
	_ = l.svcCtx.KnowledgeDocumentModel.UpdateStatus(ctx, documentId, knowledge.StatusDocumentIndexing, "")

	// 5. 调用文档处理器

	req := &types.ProcessRequest{
		URI: tempFile.Name(),
		IndexConfig: types.ProcessConfig{
			KnowledgeName:  doc.DocName,
			EnableQACheck:  parserConfig.EnableQAGeneration,
			Separators:     parserConfig.Separators,
			ChunkOverlap:   parserConfig.ChunkOverlap,
			MaxChunkLength: parserConfig.MaxChunkLength,
			PreCleanRule:   parserConfig.PreCleanRule,
			LlmConfig: types.ProcessLlmConfig{
				EmbeddingKey:       embedApiRecord.ApiKey,
				EmbeddingBaseUrl:   embedApiRecord.BaseUrl,
				EmbeddingModelName: embedApiRecord.ModelName,
				QaKey:              qaKey,
				QaBaseUrl:          qaBaseUrl,
				QaModelName:        qaModelName,
			},
		},
	}

	chunks, err := l.svcCtx.DocProcessService.Invoke(ctx, req)
	if err != nil {
		return failTask(fmt.Sprintf("处理器执行失败: %v", err))
	}

	saveChunks, err := slicex.IntoWithError(chunks, func(chunk *schema.Document) (*knowledge.KnowledgeDocumentChunk, error) {
		// 序列化 Metadata
		metaBytes, _ := json.Marshal(chunk.MetaData)

		// 生成 UUID v7
		chunkUuid, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}

		return &knowledge.KnowledgeDocumentChunk{
			Id:                  chunkUuid.String(),
			KnowledgeBaseId:     msg.KnowledgeBaseId,
			KnowledgeDocumentId: documentId,
			ChunkText:           chunk.Content,
			ChunkSize:           int64(len(chunk.Content)),
			Metadata:            string(metaBytes),
			Status:              "enable",
		}, nil
	})
	if err != nil {
		return failTask(fmt.Sprintf("UUID generation failed: %v", err))
	}

	// 6. 生成向量数据 (Embedding) & 准备 Milvus 数据
	vectorItems, err := l.generateVectorItems(ctx, saveChunks, embedApiRecord)
	if err != nil {
		return failTask(err.Error())
	}

	// 7. 写入 Milvus (Milvus First)
	if err := l.svcCtx.KnowledgeVectorModel.Insert(ctx, vectorItems); err != nil {
		return failTask(fmt.Sprintf("写入向量库失败: %v", err))
	}

	// 8. 写入 MySQL (Transaction)
	err = l.svcCtx.SqlConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		chunkModelInTx := knowledge.NewKnowledgeDocumentChunkModel(sqlx.NewSqlConnFromSession(session))
		docModelInTx := knowledge.NewKnowledgeDocumentModel(sqlx.NewSqlConnFromSession(session))

		// 批量插入切片
		if err := chunkModelInTx.InsertBatch(ctx, saveChunks); err != nil {
			return fmt.Errorf("批量插入切片失败: %w", err)
		}

		// 更新文档状态和切片数量
		if err := docModelInTx.UpdateStatusWithChunkCount(ctx, documentId, knowledge.StatusDocumentEnable, len(saveChunks)); err != nil {
			return fmt.Errorf("更新文档状态失败: %w", err)
		}

		return nil
	})

	if err != nil {
		// 补偿机制: MySQL 失败，回滚 Milvus 数据
		logx.Errorf("[DocIndex] MySQL 事务失败，回滚 Milvus 数据: %v", err)
		if delErr := l.svcCtx.KnowledgeVectorModel.DeleteByDocId(ctx, msg.KnowledgeBaseId, documentId); delErr != nil {
			logx.Errorf("[DocIndex] 回滚 Milvus 失败 (需人工介入): %v", delErr)
		}
		return failTask(err.Error())
	}

	// 记录成功指标
	metric.IndexingDuration.WithLabelValues(kbId, "success").Observe(time.Since(start).Seconds())
	metric.ChunksIndexed.WithLabelValues(kbId).Observe(float64(len(saveChunks)))

	logx.Infof("[DocIndex] DocumentId=%s 索引成功. Chunks=%d", documentId, len(saveChunks))
	return nil
}

func (l *DocumentIndexLogic) generateVectorItems(ctx context.Context, chunks []*knowledge.KnowledgeDocumentChunk, embeddingConfig *user_api.UserApi) ([]*vector.KnowledgeVectorItem, error) {
	if len(chunks) == 0 {
		return nil, nil
	}

	dim := int(embeddingConfig.ModelDim)

	// openai
	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:     embeddingConfig.ApiKey,
		BaseURL:    embeddingConfig.BaseUrl,
		Model:      embeddingConfig.ModelName,
		Dimensions: &dim,
	})

	if err != nil {
		logx.Errorf("openai.NewEmbedder err:%v, embedingConf:%v", err, embeddingConfig)
		return nil, fmt.Errorf("创建embedding失败: %w", err)
	}

	embedStrings := slicex.Into(chunks, func(chunk *knowledge.KnowledgeDocumentChunk) string {
		return chunk.ChunkText
	})

	embeddings, err := embedder.EmbedStrings(ctx, embedStrings)
	if err != nil {
		chunkIds := slicex.Into(chunks, func(chunk *knowledge.KnowledgeDocumentChunk) string {
			return chunk.Id
		})
		logx.Errorf("embedder.EmbedStrings err:%v, chunk_ids:%v", err, chunkIds)
		return nil, fmt.Errorf("生成embedding向量失败: %w", err)
	}

	if len(embeddings) != len(chunks) {
		logx.Errorf("embeddings长度(%d) != chunks长度(%d)", len(embeddings), len(chunks))
		return nil, fmt.Errorf("embedding向量数量与chunk数量不一致")
	}

	// 构建向量记录
	vectorItems := make([]*vector.KnowledgeVectorItem, 0, len(chunks))

	for i, chunk := range chunks {
		// 解析 metadata 为类型化结构
		var metadata knowledge.ChunkMetadata
		if chunk.Metadata != "" {
			_ = json.Unmarshal([]byte(chunk.Metadata), &metadata)
		}

		// 添加 chunk 原文记录
		vectorItems = append(vectorItems, &vector.KnowledgeVectorItem{
			ID:              fmt.Sprintf("%s_chunk_0", chunk.Id),
			KnowledgeBaseID: chunk.KnowledgeBaseId,
			ChunkID:         chunk.Id,
			DocID:           chunk.KnowledgeDocumentId,
			Content:         chunk.ChunkText,
			Vector:          embeddings[i],
		})

		questions := slicex.Into(metadata.QaPairs, func(qa knowledge.QaPair) string {
			return qa.Question
		})
		if len(questions) == 0 {
			continue
		}

		questionsEmbeddings, err := embedder.EmbedStrings(ctx, questions)
		if err != nil {
			logx.Errorf("embedder.EmbedStrings err:%v, chunk_ids:%v", err, chunk.Id)
			return nil, fmt.Errorf("生成QA embedding向量失败: %w", err)
		}

		if len(questionsEmbeddings) > 0 {
			for j, q := range metadata.QaPairs {
				vectorItems = append(vectorItems, &vector.KnowledgeVectorItem{
					ID:              fmt.Sprintf("%s_qa_%d", chunk.Id, j),
					KnowledgeBaseID: chunk.KnowledgeBaseId,
					ChunkID:         chunk.Id,
					DocID:           chunk.KnowledgeDocumentId,
					Content:         q.Question,
					Vector:          questionsEmbeddings[j],
				})
			}
		}
	}

	return vectorItems, nil
}
