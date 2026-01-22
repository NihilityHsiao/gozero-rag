package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gozero-rag/consumer/document_index/internal/svc"
	"gozero-rag/internal/model/chunk"
	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/model/knowledge_document"
	"gozero-rag/internal/mq"
	"gozero-rag/internal/rag_core/metric"
	"gozero-rag/internal/rag_core/parser"
	"gozero-rag/internal/rag_core/types"
	"gozero-rag/internal/slicex"
	"gozero-rag/internal/tools/llmx"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/openai"

	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
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

	if msg.UserId == "" || msg.KnowledgeBaseId == "" || msg.DocumentId == "" || msg.TenantId == "" {
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

func (l *DocumentIndexLogic) updateFail(ctx context.Context, docId string, reason string) {
	// todo:需要完善 UpdateRunStatus, 修改 updated_time,updated_date, run_state, progress_msg为reason
	_ = l.svcCtx.KnowledgeDocumentModel.UpdateRunStatus(ctx, docId, knowledge_document.RunStateFailed, reason)

}

func (l *DocumentIndexLogic) updateIndexing(ctx context.Context, docId string) {
	// todo:
}

func (l *DocumentIndexLogic) mainWork(ctx context.Context, msg *mq.KnowledgeDocumentIndexMsg) (err error) {
	documentId := msg.DocumentId

	// Panic Recovery: 确保 panic 不会导致文档永久停留在 indexing 状态
	defer func() {
		if r := recover(); r != nil {
			//logx.Errorf("[DocIndex] DocumentId=%s panic: %v", documentId, r)
			//_ = l.svcCtx.KnowledgeDocumentModel.UpdateStatus(ctx, documentId, knowledge_document.RunStateFailed, fmt.Sprintf("panic: %v", r))
			err = nil // 避免重试
		}
	}()

	start := time.Now()
	kbId := msg.KnowledgeBaseId

	// 记录索引请求总数
	metric.IndexingTotal.WithLabelValues(kbId).Inc()

	// 失败处理辅助函数
	failTask := func(reason string) error {
		logx.Errorf("[DocIndex] DocumentId=%s 失败: %s", documentId, reason)
		l.updateFail(ctx, documentId, reason)
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

	if doc.RunStatus != knowledge_document.RunStatePending {
		logx.Infof("[DocIndex] DocumentId=%s status is %s, skip", documentId, doc.RunStatus)
		return nil
	}

	// 从 MinIO 下载文件到临时目录
	tempFile, err := os.CreateTemp("", fmt.Sprintf("rag_doc_%s_*%s", doc.Id, filepath.Ext(doc.DocName.String)))
	if err != nil {
		return failTask(fmt.Sprintf("创建临时文件失败: %v", err))
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	err = l.svcCtx.OssClient.FGetObject(ctx, l.svcCtx.Config.Oss.BucketName, doc.StoragePath.String, tempFile.Name())
	if err != nil {
		return failTask(fmt.Sprintf("从 MinIO 下载失败: %v", err))
	}
	logx.Infof("[DocIndex] 从 MinIO 下载 %s 到 %s", doc.StoragePath.String, tempFile.Name())

	// 3. 获取模型配置

	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(ctx, msg.KnowledgeBaseId)
	if err != nil {
		return failTask(fmt.Sprintf("知识库未找到: %v", err))
	}

	// 解析知识库绑定的模型 ID
	// embid: 模型名称@厂商
	embModelName, embModelFactory := llmx.GetModelNameFactory(kb.EmbdId)

	llmModel, err := l.svcCtx.TenantLlmModel.FindOneByTenantIdLlmFactoryLlmName(ctx, kb.TenantId, embModelFactory, embModelName)
	if err != nil {
		return failTask(fmt.Sprintf("模型未找到: %v", err))
	}

	var conf parser.ParserConfigGeneral
	if kb.ParserId == parser.ParserIdGeneral {
		err = json.Unmarshal([]byte(kb.ParserConfig.String), &conf)
		if err != nil {
			return failTask(fmt.Sprintf("解析配置无效: %v", err))
		}
	} else {
		return failTask(fmt.Sprintf("尚不支持该解析类型: %s", kb.ParserId))
	}

	// 获取 QA 模型（可选）
	enableQa := false
	qaKey := ""
	qaBaseUrl := ""
	qaModelName := ""
	if conf.QaNum > 0 {
		name, qaModelFactory := llmx.GetModelNameFactory(conf.QaLlmId)
		qaModel, qaErr := l.svcCtx.TenantLlmModel.FindOneByTenantIdLlmFactoryLlmName(ctx, kb.TenantId, qaModelFactory, name)
		if qaErr == nil {
			enableQa = true
			qaKey = qaModel.ApiKey.String
			qaBaseUrl = qaModel.ApiBase.String
			qaModelName = name
		}
	}

	// 4. 更新状态为索引中
	l.updateIndexing(ctx, documentId)

	// 5. 调用文档处理器

	req := &types.ProcessRequest{
		URI: tempFile.Name(),
		IndexConfig: types.ProcessConfig{
			KnowledgeName:  doc.DocName.String,
			EnableQACheck:  enableQa,
			Separators:     conf.Separator,
			ChunkOverlap:   conf.ChunkOverlapTokenNum,
			MaxChunkLength: conf.ChunkTokenNum,
			QaNum:          conf.QaNum,
			LlmConfig: types.ProcessLlmConfig{
				EmbeddingKey:       llmModel.ApiKey.String,
				EmbeddingBaseUrl:   llmModel.ApiBase.String,
				EmbeddingModelName: llmModel.LlmName,
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

	// 6. 生成向量 (Embedding)
	embDim := 1024 // 默认向量维度
	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:     llmModel.ApiKey.String,
		BaseURL:    llmModel.ApiBase.String,
		Model:      llmModel.LlmName,
		Dimensions: &embDim,
	})
	if err != nil {
		return failTask(fmt.Sprintf("创建embedder失败: %v", err))
	}

	// 提取所有 Chunk 的文本内容
	contents := slicex.Into(chunks, func(c *schema.Document) string {
		return c.Content
	})
	contentVectors, err := embedder.EmbedStrings(ctx, contents)
	if err != nil {
		return failTask(fmt.Sprintf("生成embedding失败: %v", err))
	}

	if len(contentVectors) != len(chunks) {
		return failTask("embedding数量与chunk数量不一致")
	}

	// 7. 构建 ES Chunks (包含 QA 向量化)
	saveChunks := make([]*chunk.Chunk, 0, len(chunks))
	totalTokenNum := int64(0)

	for i, c := range chunks {
		// 7.1 生成唯一 ID (Hash based)
		// id = hash("chunk-" + chunk原文 + doc_id)
		hashStr := fmt.Sprintf("chunk-%s-%s", c.Content, msg.DocumentId)
		hash := md5.Sum([]byte(hashStr))
		chunkId := hex.EncodeToString(hash[:])

		// 7.2 解析元数据
		meta := c.MetaData
		var questionKw []string

		// 7.3 提取 QA 问题关键词
		if qaPairs, ok := meta["qa_pairs"].([]types.QAItem); ok {
			for _, qa := range qaPairs {
				questionKw = append(questionKw, qa.Question)
			}
		}

		// 7.4 计算 token 数 (简单估算: content 长度 / 4)
		tokenNum := int64(len(c.Content) / 4)
		totalTokenNum += tokenNum

		// 7.5 构造 ES Chunk 对象
		saveChunks = append(saveChunks, &chunk.Chunk{
			Id:            chunkId,
			DocId:         msg.DocumentId,
			KbIds:         []string{msg.KnowledgeBaseId},
			Content:       c.Content,
			ContentVector: contentVectors[i],
			DocName:       doc.DocName.String,
			ImportantKw:   nil, // 可后续提取
			QuestionKw:    questionKw,
			ImgId:         "",
			PageNum:       nil,
			CreateTime:    float64(time.Now().Unix()),
			Available:     1,
			Score:         0,
		})
	}

	// 8. 为 QA 问题生成向量并作为独立 Chunk 写入 (可选优化)
	// 遍历所有 chunks，找到有 qa_pairs 的，为每个 question 生成向量
	for _, c := range chunks {
		meta := c.MetaData
		if qaPairs, ok := meta["qa_pairs"].([]types.QAItem); ok && len(qaPairs) > 0 {
			// 提取所有 questions
			questions := make([]string, 0, len(qaPairs))
			questions = slicex.Into(qaPairs, func(t types.QAItem) string {
				return t.Question
			})

			if len(questions) > 0 {
				// 批量生成 QA 向量
				qaVectors, qaErr := embedder.EmbedStrings(ctx, questions)
				if qaErr != nil {
					logx.Errorf("生成QA向量失败: %v", qaErr)
					continue // 跳过，不影响主流程
				}

				// 为每个 question 创建独立 chunk
				for j, question := range questions {
					// id = hash("qa-" + question + answer + doc_id)
					answer := qaPairs[j].Answer
					qaHashStr := fmt.Sprintf("qa-%s-%s-%s", question, answer, msg.DocumentId)
					qaHash := md5.Sum([]byte(qaHashStr))
					qaId := hex.EncodeToString(qaHash[:])

					// content: "Question:xxx? Answer:xxxx"
					qaContent := fmt.Sprintf("Question:%s? Answer:%s", question, answer)

					saveChunks = append(saveChunks, &chunk.Chunk{
						Id:            qaId,
						DocId:         msg.DocumentId,
						KbIds:         []string{msg.KnowledgeBaseId},
						Content:       qaContent, // QA 问题作为内容
						ContentVector: qaVectors[j],
						DocName:       doc.DocName.String,
						ImportantKw:   nil,
						QuestionKw:    nil,
						ImgId:         "",
						PageNum:       nil,
						CreateTime:    float64(time.Now().Unix()),
						Available:     1,
						Score:         0,
					})
				}
			}
		}
	}

	// 9. 写入 ES
	if err := l.svcCtx.ChunkModel.Put(ctx, saveChunks); err != nil {
		return failTask(fmt.Sprintf("写入ES失败: %v", err))
	}

	// 10. 更新 MySQL 状态
	err = l.svcCtx.KnowledgeDocumentModel.UpdateStatusWithChunkCount(ctx, documentId, knowledge_document.RunStateSuccess, int64(len(chunks)), totalTokenNum)
	if err != nil {
		// 回滚 ES 数据
		logx.Errorf("[DocIndex] MySQL 更新失败，回滚 ES 数据: %v", err)
		_ = l.svcCtx.ChunkModel.DeleteByDocId(ctx, msg.KnowledgeBaseId, documentId)
		return failTask(fmt.Sprintf("更新数据库失败: %v", err))
	}

	// 记录成功指标
	metric.IndexingDuration.WithLabelValues(kbId, "success").Observe(time.Since(start).Seconds())
	metric.ChunksIndexed.WithLabelValues(kbId).Observe(float64(len(saveChunks)))

	logx.Infof("[DocIndex] DocumentId=%s 索引成功. Chunks=%d, QAChunks=%d", documentId, len(chunks), len(saveChunks)-len(chunks))
	return nil
}

//func (l *DocumentIndexLogic) generateVectorItems(ctx context.Context, chunks []*chunk.Chunk, embKey, embBaseUrl, embModelName string, dim int) ([]*vector.KnowledgeVectorItem, error) {
//	if len(chunks) == 0 {
//		return nil, nil
//	}
//
//	// openai
//	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
//		APIKey:     embKey,
//		BaseURL:    embBaseUrl,
//		Model:      embModelName,
//		Dimensions: &dim,
//	})
//
//	if err != nil {
//		return nil, fmt.Errorf("创建embedding失败: %w", err)
//	}
//
//	embedStrings := slicex.Into(chunks, func(c *chunk.Chunk) string {
//		return c.Content
//	})
//
//	embeddings, err := embedder.EmbedStrings(ctx, embedStrings)
//	if err != nil {
//		chunkIds := slicex.Into(chunks, func(chunk *chunk.Chunk) string {
//			return chunk.Id
//		})
//		logx.Errorf("embedder.EmbedStrings err:%v, chunk_ids:%v", err, chunkIds)
//		return nil, fmt.Errorf("生成embedding向量失败: %w", err)
//	}
//
//	if len(embeddings) != len(chunks) {
//		logx.Errorf("embeddings长度(%d) != chunks长度(%d)", len(embeddings), len(chunks))
//		return nil, fmt.Errorf("embedding向量数量与chunk数量不一致")
//	}
//
//	// 构建向量记录
//	vectorItems := make([]*vector.KnowledgeVectorItem, 0, len(chunks))
//
//	for i, c := range chunks {
//		// 解析 metadata 为类型化结构
//		var metadata knowledge.ChunkMetadata
//		if c.Metadata != "" {
//			_ = json.Unmarshal([]byte(c.Metadata), &metadata)
//		}
//
//		// 添加 chunk 原文记录
//		vectorItems = append(vectorItems, &vector.KnowledgeVectorItem{
//			ID:      fmt.Sprintf("%s_chunk_0", chunk.Id),
//			ChunkID: c.Id,
//			Content: c.Content,
//			DocID:   c.DocId,
//			Vector:  embeddings[i],
//		})
//
//		questions := slicex.Into(metadata.QaPairs, func(qa knowledge.QaPair) string {
//			return qa.Question
//		})
//		if len(questions) == 0 {
//			continue
//		}
//
//		questionsEmbeddings, err := embedder.EmbedStrings(ctx, questions)
//		if err != nil {
//			logx.Errorf("embedder.EmbedStrings err:%v, chunk_ids:%v", err, chunk.Id)
//			return nil, fmt.Errorf("生成QA embedding向量失败: %w", err)
//		}
//
//		if len(questionsEmbeddings) > 0 {
//			for j, q := range metadata.QaPairs {
//				vectorItems = append(vectorItems, &vector.KnowledgeVectorItem{
//					ID:              fmt.Sprintf("%s_qa_%d", chunk.Id, j),
//					KnowledgeBaseID: chunk.KnowledgeBaseId,
//					ChunkID:         chunk.Id,
//					DocID:           chunk.KnowledgeDocumentId,
//					Content:         q.Question,
//					Vector:          questionsEmbeddings[j],
//				})
//			}
//		}
//	}
//
//	return vectorItems, nil
//}
