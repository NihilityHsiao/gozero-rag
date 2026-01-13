package agentic

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gozero-rag/internal/rag_core/constant"
	"gozero-rag/internal/rag_core/types"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 每批次处理的最大句子数，防止 LLM 上下文溢出
	maxSentencesPerBatch = 20
	// 默认最大 chunk 长度
	defaultMaxChunkLength = 500
)

// AgenticSplitter 使用 LLM 进行 Bottom-Up 智能分块
type AgenticSplitter struct {
	llm model.BaseChatModel
}

// NewAgenticSplitter 创建 AgenticSplitter 实例
func NewAgenticSplitter(llm model.BaseChatModel) *AgenticSplitter {
	return &AgenticSplitter{
		llm: llm,
	}
}

// NewAgenticSplitterFromConfig 从配置创建 AgenticSplitter
func NewAgenticSplitterFromConfig(ctx context.Context, cfg types.ProcessConfig) (*AgenticSplitter, error) {
	if cfg.LlmConfig.ChatKey == "" {
		return nil, fmt.Errorf("ApiKey is required for AgenticSplitter")
	}

	config := &openai.ChatModelConfig{
		APIKey:  cfg.LlmConfig.ChatKey,
		Model:   cfg.LlmConfig.ChatModelName,
		BaseURL: cfg.LlmConfig.ChatBaseUrl,
	}

	llm, err := openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChatModel: %w", err)
	}

	return NewAgenticSplitter(llm), nil
}

// Transform 实现 document.Transformer 接口
func (s *AgenticSplitter) Transform(ctx context.Context, src []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	if len(src) == 0 {
		return src, nil
	}

	// 从 ctx 获取配置
	cfg, ok := ctx.Value(constant.CtxKeyIndexConfig).(types.ProcessConfig)
	if !ok {
		cfg = types.ProcessConfig{
			MaxChunkLength: defaultMaxChunkLength,
			ChunkOverlap:   0,
		}
	}
	if cfg.MaxChunkLength <= 0 {
		cfg.MaxChunkLength = defaultMaxChunkLength
	}

	logx.Infof("[AgenticSplitter] Using config: MaxChunkLength=%d, ChunkOverlap=%d", cfg.MaxChunkLength, cfg.ChunkOverlap)

	var result []*schema.Document

	for _, doc := range src {
		chunks, err := s.splitDocument(ctx, doc, cfg)
		if err != nil {
			logx.Errorf("[AgenticSplitter] Failed to split document %s: %v", doc.ID, err)
			// 如果分割失败，返回原文档
			result = append(result, doc)
			continue
		}
		result = append(result, chunks...)
	}

	return result, nil
}

// atomize 将文档内容拆分为原子句子
func atomize(content string) []string {
	// 使用正则按中英文句子结束符分割
	// 支持: 。！？.!? 以及换行符
	re := regexp.MustCompile(`[。！？.!?\n]+`)

	// 分割并保留分隔符
	parts := re.Split(content, -1)

	var sentences []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if len(trimmed) > 0 {
			sentences = append(sentences, trimmed)
		}
	}

	return sentences
}

// detectBoundariesBatchWindowed 分批次判断相邻句子的话题边界
// 每次只处理 maxSentencesPerBatch 个句子，避免 LLM 上下文溢出
func (s *AgenticSplitter) detectBoundariesBatchWindowed(ctx context.Context, sentences []string) ([]BoundaryInfo, error) {
	if len(sentences) <= 1 {
		return nil, nil
	}

	allBoundaries := make([]BoundaryInfo, len(sentences)-1)

	// 分批处理
	for start := 0; start < len(sentences)-1; start += maxSentencesPerBatch - 1 {
		end := start + maxSentencesPerBatch
		if end > len(sentences) {
			end = len(sentences)
		}

		logx.Infof("提取批次 [%d:%d]", start, end)
		// 提取当前批次的句子
		batchSentences := sentences[start:end]
		if len(batchSentences) <= 1 {
			break
		}

		// 调用 LLM 判断当前批次
		batchBoundaries, err := s.detectBoundariesSingleBatch(ctx, batchSentences)
		if err != nil {
			logx.Errorf("[AgenticSplitter] Failed to detect batch boundaries at start=%d: %v", start, err)
			// 失败时默认全部作为边界（避免合并错误）
			for i := start; i < start+len(batchSentences)-1 && i < len(allBoundaries); i++ {
				allBoundaries[i] = BoundaryInfo{IsBoundary: true}
			}
			continue
		}

		// 复制结果到总数组
		for i, b := range batchBoundaries {
			if start+i < len(allBoundaries) {
				allBoundaries[start+i] = b
			}
		}
		logx.Infof("提取批次完成 [%d:%d]", start, end)

	}

	return allBoundaries, nil
}

// detectBoundariesSingleBatch 对单个批次的句子进行边界检测
func (s *AgenticSplitter) detectBoundariesSingleBatch(ctx context.Context, sentences []string) ([]BoundaryInfo, error) {
	if len(sentences) <= 1 {
		return nil, nil
	}

	// 构建句子列表
	var sb strings.Builder
	for i, sentence := range sentences {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, sentence))
	}

	systemPrompt := `你是一个专业的文档结构分析助手。给定一系列按顺序排列的句子，你的任务是分析每两个相邻句子之间是否发生了**话题转换(Topic Shift)**或**章节切换**。

如果发现了话题转换或新章节，请提取或生成该新部分的**简短标题(Header)**。

需遵守的输出规则：
- 返回一个 JSON 对象数组，长度必须严格等于 N-1（N 是句子数量）。
- 数组第 i 个对象描述第 i 个句子(Start)和第 i+1 个句子(End)之间的关系。
- JSON 对象格式：
  {
    "is_boundary": boolean, // true 表示句子 i 和 i+1 之间是话题边界，话题断开了
    "header": string        // 如果 is_boundary 为 true，请提供新话题的标题（不超过10个字）。如果是现有标题行，直接使用；否则根据下一句内容概括。如果 is_boundary 为 false，则为空字符串。
  }
- **重要**：如果句子本身看起来就像是一个标题（如 "第一章 总则" 或 "1. 背景"），那么它与前一个句子之间通常是边界，与后一个句子之间通常不是边界（除非它后面紧跟另一个标题）。

仅返回 JSON 数组，严禁包含 markdown 代码块标记或其他解释。`

	userPrompt := fmt.Sprintf("句子列表（共 %d 个句子）:\n%s\n请返回 %d 个对象的 JSON 数组。",
		len(sentences), sb.String(), len(sentences)-1)

	messages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: userPrompt},
	}

	resp, err := s.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generate failed: %w", err)
	}

	// 解析返回的 JSON 数组
	var boundaries []BoundaryInfo
	respContent := strings.TrimSpace(resp.Content)
	// 清理 markdown 标记
	respContent = strings.TrimPrefix(respContent, "```json")
	respContent = strings.TrimPrefix(respContent, "```")
	respContent = strings.TrimSpace(respContent)

	if err := json.Unmarshal([]byte(respContent), &boundaries); err != nil {
		logx.Errorf("[AgenticSplitter] Failed to parse boundaries JSON: %v, response: %s", err, respContent)
		// 解析失败时，默认全部作为边界
		boundaries = make([]BoundaryInfo, len(sentences)-1)
		for i := range boundaries {
			boundaries[i] = BoundaryInfo{IsBoundary: true}
		}
	}

	// 确保长度正确
	if len(boundaries) != len(sentences)-1 {
		logx.Errorf("[AgenticSplitter] Boundary count mismatch: expected %d, got %d", len(sentences)-1, len(boundaries))
		result := make([]BoundaryInfo, len(sentences)-1)
		for i := range result {
			if i < len(boundaries) {
				result[i] = boundaries[i]
			} else {
				result[i] = BoundaryInfo{IsBoundary: true} // 默认边界
			}
		}
		boundaries = result
	}

	return boundaries, nil
}

// aggregateWithMaxLength 根据边界判断结果聚合句子，同时尊重 MaxChunkLength 和 ChunkOverlap
func aggregateWithMaxLength(sentences []string, boundaries []BoundaryInfo, maxChunkLength, chunkOverlap int) []string {
	if len(sentences) == 0 {
		return nil
	}
	if len(sentences) == 1 {
		return sentences
	}

	var chunks []string

	// 当前 Chunk 状态
	var currentChunkBuilder strings.Builder
	// 当前 Chunk 的上下文标题 (来自上一个边界的 header)
	currentHeader := ""

	// 初始化第一个 Chunk
	currentChunkBuilder.WriteString(sentences[0])

	for i := 1; i < len(sentences); i++ {
		boundaryInfo := boundaries[i-1]
		nextSentence := sentences[i]

		// 判断是否需要开启新 Chunk
		// 1. LLM 认为是话题边界
		// 2. 当前 Chunk 长度超过限制

		isTopicBoundary := boundaryInfo.IsBoundary
		currentLen := len([]rune(currentChunkBuilder.String()))
		willExceedMax := currentLen+len([]rune(nextSentence))+1 > maxChunkLength

		if isTopicBoundary || willExceedMax {
			// === 结束当前 Chunk ===
			finalContent := currentChunkBuilder.String()

			// 如果有检测到的标题上下文，注入到 Chunk 头部
			if currentHeader != "" {
				finalContent = fmt.Sprintf("【%s】\n%s", currentHeader, finalContent)
			}
			chunks = append(chunks, finalContent)

			// === 开启新 Chunk ===
			currentChunkBuilder.Reset()

			// 更新上下文标题：如果是话题边界且有新标题，则更新；否则沿用（针对超长切分的情况，沿用是一个策略，或者清空）
			// 策略：如果是 TopicBoundary，使用新的 Header。如果 Header 为空，则清空上下文。
			// 如果是 LengthBoundary (非 Topic)，则沿用上一个 Header（即把长章节切开时，Context 保持）
			if isTopicBoundary {
				currentHeader = boundaryInfo.Header
			}

			// 处理 Overlap (仅在非 Topic 边界切分时更有意义，但在 Topic 边界也可以保留一点上下文)
			// 为了简化，这里统一处理Overalp
			if chunkOverlap > 0 && currentLen > chunkOverlap && !isTopicBoundary {
				// 仅当非话题强制切分时，才做重叠，保持连贯性
				// 如果是话题切换，通常不需要 overlap
				prevText := sentences[i-1] // 简单取上一句，或者取最后 N 个字符
				// 如果上一句太长，截取
				if len([]rune(prevText)) > chunkOverlap {
					prevText = string([]rune(prevText)[len([]rune(prevText))-chunkOverlap:])
				}
				currentChunkBuilder.WriteString(prevText)
				currentChunkBuilder.WriteString(" ")
			}

			currentChunkBuilder.WriteString(nextSentence)

		} else {
			// === 合并到当前 Chunk ===
			currentChunkBuilder.WriteString(" ")
			currentChunkBuilder.WriteString(nextSentence)
		}
	}

	// 保存最后一个 Chunk
	if currentChunkBuilder.Len() > 0 {
		finalContent := currentChunkBuilder.String()
		if currentHeader != "" {
			finalContent = fmt.Sprintf("【%s】\n%s", currentHeader, finalContent)
		}
		chunks = append(chunks, finalContent)
	}

	return chunks
}

// splitDocument 使用 Bottom-Up 方法分割单个文档
func (s *AgenticSplitter) splitDocument(ctx context.Context, doc *schema.Document, cfg types.ProcessConfig) ([]*schema.Document, error) {
	content := doc.Content
	if len(content) == 0 {
		return []*schema.Document{doc}, nil
	}

	// Step 1: 原子化拆分
	sentences := atomize(content)
	logx.Infof("[AgenticSplitter] Atomized document %s into %d sentences", doc.ID, len(sentences))

	if len(sentences) <= 1 {
		return []*schema.Document{doc}, nil
	}

	// Step 2: 分批次检测话题边界 (避免上下文溢出)
	boundaries, err := s.detectBoundariesBatchWindowed(ctx, sentences)
	if err != nil {
		logx.Errorf("[AgenticSplitter] Failed to detect boundaries: %v", err)
		// 失败时返回原文档
		return []*schema.Document{doc}, nil
	}

	// Step 3: 聚合同话题句子，同时尊重 MaxChunkLength 和 ChunkOverlap
	chunks := aggregateWithMaxLength(sentences, boundaries, cfg.MaxChunkLength, cfg.ChunkOverlap)
	logx.Infof("[AgenticSplitter] Aggregated into %d semantic chunks", len(chunks))

	if len(chunks) == 0 {
		return []*schema.Document{doc}, nil
	}

	// Step 4: 构造结果文档
	var result []*schema.Document
	for i, chunk := range chunks {
		newDoc := &schema.Document{
			ID:       fmt.Sprintf("%s_chunk_%d", doc.ID, i+1),
			Content:  chunk,
			MetaData: doc.MetaData,
		}
		result = append(result, newDoc)
	}

	logx.Infof("[AgenticSplitter] Split document %s into %d chunks", doc.ID, len(result))
	return result, nil
}
