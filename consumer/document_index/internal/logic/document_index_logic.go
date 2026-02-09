package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"gozero-rag/internal/concurrentx"
	"os"
	"path/filepath"
	"time"

	"gozero-rag/consumer/document_index/internal/svc"
	"gozero-rag/internal/model/chunk"
	"gozero-rag/internal/model/knowledge_base"
	"gozero-rag/internal/model/knowledge_document"
	"gozero-rag/internal/model/local_message"
	"gozero-rag/internal/mq"
	"gozero-rag/internal/rag_core/metric"
	"gozero-rag/internal/rag_core/parser"
	"gozero-rag/internal/rag_core/types"
	"gozero-rag/internal/slicex"
	"gozero-rag/internal/tools/llmx"

	"github.com/cespare/xxhash/v2"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
)

// ========================================
// Constants
// ========================================

const (
	defaultEmbeddingDim = 1024
	tokenEstimateRatio  = 4 // 每 4 个字符约等于 1 个 token
	// 并发批量生成向量
	batchSize = 10 // 每批10个文本
	workers   = 4  // 4个并发worker
)

// ========================================
// DocumentIndexLogic 核心结构体
// ========================================

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

// ========================================
// 消息消费入口
// ========================================

func (l *DocumentIndexLogic) Consume(_ context.Context, key, val string) (err error) {
	msg, err := l.parseMessage(val)
	if err != nil {
		logx.Errorf("[DocIndex] 消息反序列化失败: val=%s, err=%v", val, err)
		return nil // 格式错误直接丢弃，不重试
	}

	if !msg.isValid() {
		logx.Errorf("[DocIndex] 非法消息参数: %s", val)
		return nil
	}

	logx.Infof("[DocIndex] 开始处理文档索引: DocumentId=%s", msg.DocumentId)

	// 使用本地消息表包装业务逻辑，实现分布式一致性
	return l.svcCtx.LocalMsgExecutor.Execute(l.ctx, local_message.TaskTypeDocumentIndex, msg.KnowledgeDocumentIndexMsg, func(ctx context.Context) error {
		return l.processDocument(ctx, msg)
	})
}

// ========================================
// 内部数据结构
// ========================================

// indexMessage 索引消息的包装，提供便捷方法
type indexMessage struct {
	*mq.KnowledgeDocumentIndexMsg
}

func (m *indexMessage) isValid() bool {
	return m.UserId != "" && m.KnowledgeBaseId != "" && m.DocumentId != "" && m.TenantId != ""
}

// indexContext 索引过程中的上下文数据
type indexContext struct {
	msg       *indexMessage
	doc       *knowledge_document.KnowledgeDocument
	kb        *knowledge_base.KnowledgeBase
	config    *parser.ParserConfigGeneral
	embedder  embedding.Embedder
	qaEnabled bool
	qaConfig  *qaModelConfig
	startTime time.Time
}

// qaModelConfig QA 模型配置
type qaModelConfig struct {
	apiKey    string
	baseUrl   string
	modelName string
}

// ========================================
// 核心处理流程
// ========================================

// processDocument 文档索引主流程
func (l *DocumentIndexLogic) processDocument(ctx context.Context, msg *indexMessage) (err error) {
	ic := &indexContext{
		msg:       msg,
		startTime: time.Now(),
	}

	// Panic Recovery
	defer l.recoverFromPanic(ic, &err)

	// 记录索引请求总数
	metric.IndexingTotal.WithLabelValues(msg.KnowledgeBaseId).Inc()

	// Step 1: 验证并获取文档
	if err := l.loadDocument(ctx, ic); err != nil {
		return err
	}

	// Step 2: 下载文件到临时目录
	tempFilePath, cleanup, err := l.downloadToTemp(ctx, ic.doc)
	if err != nil {
		return l.failTask(ctx, ic, fmt.Sprintf("文件下载失败: %v", err))
	}
	defer cleanup()

	// Step 3: 加载知识库和模型配置
	if err := l.loadKnowledgeBaseConfig(ctx, ic); err != nil {
		return l.failTask(ctx, ic, err.Error())
	}

	// Step 4: 创建 Embedder
	if err := l.createEmbedder(ctx, ic); err != nil {
		return l.failTask(ctx, ic, fmt.Sprintf("创建Embedder失败: %v", err))
	}

	// Step 5: 更新状态为索引中
	l.updateRunStatus(ctx, ic.msg.DocumentId, knowledge_document.RunStateRunning, "正在索引...")

	// Step 6: 解析文档并切片
	chunks, err := l.parseDocument(ctx, ic, tempFilePath)
	if err != nil {
		return l.failTask(ctx, ic, fmt.Sprintf("文档解析失败: %v", err))
	}

	// Step 7: 生成向量并构建 ES Chunk
	saveChunks, totalTokenNum, err := l.buildChunksWithEmbedding(ctx, ic, chunks)
	if err != nil {
		return l.failTask(ctx, ic, err.Error())
	}

	// Step 8: 写入 ES
	if err := l.svcCtx.ChunkModel.Put(ctx, saveChunks); err != nil {
		return l.failTask(ctx, ic, fmt.Sprintf("写入ES失败: %v", err))
	}

	// Step 9: 更新 MySQL 状态
	if err := l.finalizeDocument(ctx, ic, len(chunks), totalTokenNum, saveChunks); err != nil {
		return err
	}

	// Step 10: 如果开启了rag生成，发送task到消息队列
	if ic.config.GraphRag.EnableGraph {
		_ = l.svcCtx.MqPusherClient.PublishGraphGenerateMsg(l.ctx, &mq.GraphGenerateMsg{
			DocumentId:             ic.msg.DocumentId,
			KnowledgeBaseId:        ic.msg.KnowledgeBaseId,
			TenantId:               ic.msg.TenantId,
			LlmId:                  ic.config.GraphRag.GraphLlmId,
			EntityTypes:            ic.config.GraphRag.EntityTypes,
			EnableEntityResolution: ic.config.GraphRag.EnableEntityResolution,
			EnableCommunity:        ic.config.GraphRag.EnableCommunity,
		})
	}

	// 记录成功指标
	l.recordSuccessMetrics(ic, len(saveChunks), len(chunks))
	return nil
}

// ========================================
// Step 1: 文档验证
// ========================================

func (l *DocumentIndexLogic) loadDocument(ctx context.Context, ic *indexContext) error {
	doc, err := l.svcCtx.KnowledgeDocumentModel.FindOne(ctx, ic.msg.DocumentId)
	if err != nil {
		if err == knowledge_base.ErrNotFound {
			return nil // 文档不存在，静默跳过
		}
		return err // 数据库错误，触发重试
	}

	if doc.RunStatus != knowledge_document.RunStatePending {
		logx.Infof("[DocIndex] DocumentId=%s 状态为 %s，跳过处理", ic.msg.DocumentId, doc.RunStatus)
		return nil
	}

	ic.doc = doc
	return nil
}

// ========================================
// Step 2: 文件下载
// ========================================

func (l *DocumentIndexLogic) downloadToTemp(ctx context.Context, doc *knowledge_document.KnowledgeDocument) (string, func(), error) {
	ext := filepath.Ext(doc.DocName.String)
	tempFile, err := os.CreateTemp("", fmt.Sprintf("rag_doc_%s_*%s", doc.Id, ext))
	if err != nil {
		return "", nil, fmt.Errorf("创建临时文件失败: %w", err)
	}

	cleanup := func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}

	err = l.svcCtx.OssClient.FGetObject(ctx, l.svcCtx.Config.Oss.BucketName, doc.StoragePath.String, tempFile.Name())
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("从 MinIO 下载失败: %w", err)
	}

	logx.Infof("[DocIndex] 文件下载完成: %s -> %s", doc.StoragePath.String, tempFile.Name())
	return tempFile.Name(), cleanup, nil
}

// ========================================
// Step 3: 知识库配置加载
// ========================================

func (l *DocumentIndexLogic) loadKnowledgeBaseConfig(ctx context.Context, ic *indexContext) error {
	// 获取知识库
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(ctx, ic.msg.KnowledgeBaseId)
	if err != nil {
		return fmt.Errorf("知识库未找到: %w", err)
	}
	ic.kb = kb

	// 解析 Parser 配置
	if kb.ParserId != parser.ParserIdGeneral {
		return fmt.Errorf("尚不支持该解析类型: %s", kb.ParserId)
	}

	var config parser.ParserConfigGeneral
	if err := json.Unmarshal([]byte(kb.ParserConfig.String), &config); err != nil {
		return fmt.Errorf("解析配置无效: %w", err)
	}
	ic.config = &config

	// 加载 QA 模型配置（可选）
	l.loadQAConfig(ctx, ic)

	return nil
}

func (l *DocumentIndexLogic) loadQAConfig(ctx context.Context, ic *indexContext) {
	if ic.config.QaNum <= 0 {
		return
	}

	modelName, factory := llmx.GetModelNameFactory(ic.config.QaLlmId)
	qaModel, err := l.svcCtx.TenantLlmModel.FindOneByTenantIdLlmFactoryLlmName(ctx, ic.kb.TenantId, factory, modelName)
	if err != nil {
		logx.Infof("[DocIndex] QA 模型未配置或获取失败，跳过 QA 生成")
		return
	}

	ic.qaEnabled = true
	ic.qaConfig = &qaModelConfig{
		apiKey:    qaModel.ApiKey.String,
		baseUrl:   qaModel.ApiBase.String,
		modelName: modelName,
	}
}

// ========================================
// Step 4: 创建 Embedder
// ========================================

func (l *DocumentIndexLogic) createEmbedder(ctx context.Context, ic *indexContext) error {
	modelName, factory := llmx.GetModelNameFactory(ic.kb.EmbdId)
	llmModel, err := l.svcCtx.TenantLlmModel.FindOneByTenantIdLlmFactoryLlmName(ctx, ic.kb.TenantId, factory, modelName)
	if err != nil {
		return fmt.Errorf("Embedding 模型未找到: %w", err)
	}

	embDim := defaultEmbeddingDim
	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:     llmModel.ApiKey.String,
		BaseURL:    llmModel.ApiBase.String,
		Model:      llmModel.LlmName,
		Dimensions: &embDim,
	})
	if err != nil {
		return err
	}

	ic.embedder = embedder
	return nil
}

// ========================================
// Step 5: 文档解析
// ========================================

func (l *DocumentIndexLogic) parseDocument(ctx context.Context, ic *indexContext, filePath string) ([]*schema.Document, error) {
	req := &types.ProcessRequest{
		URI: filePath,
		IndexConfig: types.ProcessConfig{
			KnowledgeName:  ic.doc.DocName.String,
			EnableQACheck:  ic.qaEnabled,
			Separators:     ic.config.Separator,
			ChunkOverlap:   ic.config.ChunkOverlapTokenNum,
			MaxChunkLength: ic.config.ChunkTokenNum,
			QaNum:          ic.config.QaNum,
			LlmConfig:      l.buildLlmConfig(ic),
		},
	}

	return l.svcCtx.DocProcessService.Invoke(ctx, req)
}

func (l *DocumentIndexLogic) buildLlmConfig(ic *indexContext) types.ProcessLlmConfig {
	modelName, factory := llmx.GetModelNameFactory(ic.kb.EmbdId)
	llmModel, _ := l.svcCtx.TenantLlmModel.FindOneByTenantIdLlmFactoryLlmName(l.ctx, ic.kb.TenantId, factory, modelName)

	config := types.ProcessLlmConfig{
		EmbeddingKey:       llmModel.ApiKey.String,
		EmbeddingBaseUrl:   llmModel.ApiBase.String,
		EmbeddingModelName: llmModel.LlmName,
	}

	if ic.qaConfig != nil {
		config.QaKey = ic.qaConfig.apiKey
		config.QaBaseUrl = ic.qaConfig.baseUrl
		config.QaModelName = ic.qaConfig.modelName
	}

	return config
}

// ========================================
// Step 6: 向量生成与 Chunk 构建
// ========================================

func (l *DocumentIndexLogic) buildChunksWithEmbedding(ctx context.Context, ic *indexContext, chunks []*schema.Document) ([]*chunk.Chunk, int64, error) {
	var (
		contentChunks []*chunk.Chunk
		qaChunks      []*chunk.Chunk
		totalTokenNum int64
		contentErr    error
		qaErr         error
	)

	// 并行构建 Content Chunks 和 QA Chunks
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		contentChunks, totalTokenNum, contentErr = l.buildContentChunks(gCtx, ic, chunks)
		return contentErr
	})

	g.Go(func() error {
		qaChunks, qaErr = l.buildQAChunks(gCtx, ic, chunks)
		// QA 是可选的降级场景，失败不阻断主流程
		if qaErr != nil {
			logx.Errorf("build qa chunks error: %v", qaErr)
		}
		return nil // 始终返回 nil，不阻断
	})

	if err := g.Wait(); err != nil {
		return nil, 0, fmt.Errorf("构建 Chunk 失败: %w", err)
	}

	saveChunks := contentChunks
	if len(qaChunks) > 0 {
		saveChunks = append(saveChunks, qaChunks...)
	}

	return saveChunks, totalTokenNum, nil
}

func (l *DocumentIndexLogic) buildContentChunks(ctx context.Context, ic *indexContext, chunks []*schema.Document) ([]*chunk.Chunk, int64, error) {

	// 批量生成向量
	vectors, err := concurrentx.ParallelProcessOrdered(ctx, chunks, concurrentx.ParallelProcessConfig{
		BatchSize: batchSize,
		Workers:   workers,
		Timeout:   60 * time.Second,
	}, func(ctx context.Context, batch []*schema.Document) ([][]float64, error) {
		contents := slicex.Into(batch, func(d *schema.Document) string {
			return d.Content
		})
		return ic.embedder.EmbedStrings(ctx, contents)
	})
	if err != nil {
		return nil, 0, fmt.Errorf("生成 Embedding 失败: %w", err)
	}

	if len(vectors) != len(chunks) {
		return nil, 0, fmt.Errorf("embedding 数量(%d)与 Chunk 数量(%d)不一致", len(vectors), len(chunks))
	}

	saveChunks := make([]*chunk.Chunk, 0, len(chunks))
	var totalTokenNum int64

	now := float64(time.Now().Unix())

	for i, doc := range chunks {
		chunkId := l.generateChunkId("chunk", doc.Content, ic.msg.DocumentId)
		tokenNum := int64(len(doc.Content) / tokenEstimateRatio)
		totalTokenNum += tokenNum

		saveChunks = append(saveChunks, &chunk.Chunk{
			Id:            chunkId,
			DocId:         ic.msg.DocumentId,
			KbIds:         []string{ic.msg.KnowledgeBaseId},
			Content:       doc.Content,
			ContentVector: vectors[i],
			DocName:       ic.doc.DocName.String,
			CreateTime:    now,
			Available:     1,
		})
	}

	return saveChunks, totalTokenNum, nil
}

func (l *DocumentIndexLogic) buildQAChunks(ctx context.Context, ic *indexContext, docs []*schema.Document) ([]*chunk.Chunk, error) {
	// Step 1: 收集所有 QA 对
	type qaWithMeta struct {
		qa       types.QAItem
		question string
	}
	var allQAs []qaWithMeta

	for _, doc := range docs {
		qaPairs, ok := doc.MetaData["qa_pairs"].([]types.QAItem)
		if !ok || len(qaPairs) == 0 {
			continue
		}
		for _, qa := range qaPairs {
			allQAs = append(allQAs, qaWithMeta{qa: qa, question: qa.Question})
		}
	}

	if len(allQAs) == 0 {
		return nil, nil
	}

	// Step 2: 提取所有问题
	questions := slicex.Into(allQAs, func(q qaWithMeta) string {
		return q.question
	})

	// Step 3: 一次性并发批量生成所有 QA 向量
	qaVectors, err := concurrentx.ParallelProcessOrdered(ctx, questions, concurrentx.ParallelProcessConfig{
		BatchSize: batchSize,
		Workers:   workers,
		Timeout:   60 * time.Second,
	}, func(ctx context.Context, batch []string) ([][]float64, error) {
		return ic.embedder.EmbedStrings(ctx, batch)
	})
	if err != nil {
		return nil, fmt.Errorf("QA 向量生成失败: %w", err)
	}

	if len(qaVectors) != len(allQAs) {
		return nil, fmt.Errorf("QA 向量数量(%d)与 QA 对数量(%d)不一致", len(qaVectors), len(allQAs))
	}

	// Step 4: 构建 QA Chunks
	now := float64(time.Now().Unix())
	qaChunks := make([]*chunk.Chunk, 0, len(allQAs))

	for i, qaMeta := range allQAs {
		qa := qaMeta.qa
		qaId := l.generateChunkId("qa", qa.Question+qa.Answer, ic.msg.DocumentId)
		qaContent := fmt.Sprintf("Question: %s\nAnswer: %s", qa.Question, qa.Answer)

		qaChunks = append(qaChunks, &chunk.Chunk{
			Id:            qaId,
			DocId:         ic.msg.DocumentId,
			KbIds:         []string{ic.msg.KnowledgeBaseId},
			Content:       qaContent,
			ContentVector: qaVectors[i],
			DocName:       ic.doc.DocName.String,
			CreateTime:    now,
			Available:     1,
		})
	}

	return qaChunks, nil
}

// ========================================
// Step 7: 完成索引
// ========================================

func (l *DocumentIndexLogic) finalizeDocument(ctx context.Context, ic *indexContext, chunkCount int, totalTokenNum int64, saveChunks []*chunk.Chunk) error {
	err := l.svcCtx.KnowledgeDocumentModel.UpdateStatusWithChunkCount(ctx, ic.msg.DocumentId, knowledge_document.RunStateSuccess, int64(chunkCount), totalTokenNum)
	if err != nil {
		logx.Errorf("[DocIndex] MySQL 更新失败，回滚 ES 数据: %v", err)
		_ = l.svcCtx.ChunkModel.DeleteByDocId(ctx, ic.msg.KnowledgeBaseId, ic.msg.DocumentId)
		return l.failTask(ctx, ic, fmt.Sprintf("更新数据库失败: %v", err))
	}

	return nil
}

// ========================================
// 辅助函数
// ========================================

func (l *DocumentIndexLogic) parseMessage(val string) (*indexMessage, error) {
	var msg mq.KnowledgeDocumentIndexMsg
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		return nil, err
	}
	return &indexMessage{&msg}, nil
}

func (l *DocumentIndexLogic) generateChunkId(prefix, content, docId string) string {
	hashStr := fmt.Sprintf("%s-%s", content, docId)
	hash := xxhash.Sum64String(hashStr)
	return fmt.Sprintf("%s-%x", prefix, hash)
}

func (l *DocumentIndexLogic) updateRunStatus(ctx context.Context, docId, status, msg string) {
	_ = l.svcCtx.KnowledgeDocumentModel.UpdateRunStatus(ctx, docId, status, msg)
}

func (l *DocumentIndexLogic) failTask(ctx context.Context, ic *indexContext, reason string) error {
	logx.Errorf("[DocIndex] DocumentId=%s 失败: %s", ic.msg.DocumentId, reason)
	l.updateRunStatus(ctx, ic.msg.DocumentId, knowledge_document.RunStateFailed, reason)

	// 记录失败指标
	metric.IndexingErrors.WithLabelValues(ic.msg.KnowledgeBaseId, "process_error").Inc()
	metric.IndexingDuration.WithLabelValues(ic.msg.KnowledgeBaseId, "fail").Observe(time.Since(ic.startTime).Seconds())

	return nil // 返回 nil 避免 Kafka 重试
}

func (l *DocumentIndexLogic) recoverFromPanic(ic *indexContext, err *error) {
	if r := recover(); r != nil {
		logx.Errorf("[DocIndex] DocumentId=%s panic recovered: %v", ic.msg.DocumentId, r)
		l.updateRunStatus(l.ctx, ic.msg.DocumentId, knowledge_document.RunStateFailed, fmt.Sprintf("panic: %v", r))
		*err = nil
	}
}

func (l *DocumentIndexLogic) recordSuccessMetrics(ic *indexContext, totalChunks, contentChunks int) {
	qaChunks := totalChunks - contentChunks
	metric.IndexingDuration.WithLabelValues(ic.msg.KnowledgeBaseId, "success").Observe(time.Since(ic.startTime).Seconds())
	metric.ChunksIndexed.WithLabelValues(ic.msg.KnowledgeBaseId).Observe(float64(totalChunks))

	logx.Infof("[DocIndex] DocumentId=%s 索引成功. ContentChunks=%d, QAChunks=%d, Total=%d",
		ic.msg.DocumentId, contentChunks, qaChunks, totalChunks)
}

// embedStringsBatched 分批生成向量，避免 413 Payload Too Large
func (l *DocumentIndexLogic) embedStringsBatched(ctx context.Context, ic *indexContext, texts []string) ([][]float64, error) {
	var allVectors [][]float64

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batchTexts := texts[i:end]
		vectors, err := ic.embedder.EmbedStrings(ctx, batchTexts)
		if err != nil {
			return nil, fmt.Errorf("batch embedding failed at index %d: %w", i, err)
		}
		allVectors = append(allVectors, vectors...)
	}

	return allVectors, nil
}
