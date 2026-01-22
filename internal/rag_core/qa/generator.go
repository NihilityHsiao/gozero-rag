package qa

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gozero-rag/internal/rag_core/types"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
)

// Generator QA 生成器
type Generator struct {
	config        types.ProcessConfig
	knowledgeName string
}

// NewGenerator 创建 QA 生成器
func NewGenerator(config types.ProcessConfig, knowledgeName string) *Generator {
	return &Generator{
		config:        config,
		knowledgeName: knowledgeName,
	}
}

// Generate 为单个 chunk 生成 QA 问答对
func (g *Generator) Generate(ctx context.Context, doc *schema.Document, qaNum int) ([]types.QAItem, error) {
	// 创建 ChatModel
	model, err := g.createChatModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	// 构建用户消息 (包含 metadata 上下文)
	userContent := g.buildUserPrompt(doc.Content, doc.MetaData, qaNum)

	// 构建消息
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: g.buildSystemPrompt(),
		},
		{
			Role:    schema.User,
			Content: userContent,
		},
	}

	// 调用 LLM
	resp, err := model.Generate(ctx, messages)
	if err != nil {
		logx.Errorf("LLM 调用失败: %v", err)
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	// 解析 JSON 响应
	qaPairs, err := g.parseResponse(resp.Content)
	if err != nil {
		logx.Errorf("解析 QA 响应失败: %v, 原始内容: %s", err, resp.Content)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	logx.Infof("成功生成 %d 个 QA 对", len(qaPairs))
	return qaPairs, nil
}

// GenerateBatch 批量为多个 chunk 生成 QA
func (g *Generator) GenerateBatch(ctx context.Context, docs []*schema.Document, qaNum int) ([][]types.QAItem, error) {
	results := make([][]types.QAItem, len(docs))

	for i, doc := range docs {
		qaPairs, err := g.Generate(ctx, doc, qaNum)
		if err != nil {
			logx.Errorf("生成第 %d 个 chunk 的 QA 失败: %v", i, err)
			// 失败不阻塞，继续处理其他 chunk
			results[i] = nil
			continue
		}
		results[i] = qaPairs
	}

	return results, nil
}

// createChatModel 创建 ChatModel 实例
func (g *Generator) createChatModel(ctx context.Context) (*openai.ChatModel, error) {
	// 使用配置中的 LLM 设置
	baseURL := g.config.LlmConfig.QaBaseUrl
	modelName := g.config.LlmConfig.QaModelName

	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  g.config.LlmConfig.QaKey,
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, err
	}

	return model, nil
}

// buildSystemPrompt 构建系统提示词
func (g *Generator) buildSystemPrompt() string {
	return fmt.Sprintf(`你是一个专业的RAG数据集生成专家。你的任务是基于提供的文本内容生成高质量的问答对。
知识库名称：《%s》

### 生成要求：
1. **格式严格**：必须输出标准的 JSON 数组格式，不要包含 Markdown 标记（如 '''json）。
2. **独立性（关键）**：生成的问题必须是"独立可理解的"。
   - ❌ 错误案例："由于这个原因，导致了什么后果？"（脱离原文不知道"这个原因"是指什么）
   - ✅ 正确案例："在《%s》中，导致服务器宕机的主要原因是什么？"（补全了主语和背景,如果知识库名称为空,就只需要生成问题，而不需要指定知识库名称）
3. **内容来源**：严格基于文本，不要使用外部知识。
4. **多样性**：
   - 包含事实型问题（What, When, Who）
   - 包含推理型问题（Why, How）
   - 包含条件型问题（If...）

### 质量控制规则：
1. **拒绝代词**：问题中绝对不能出现"他"、"它"、"这"、"该产品"等指代不明的词汇。必须替换为具体的实体名称。
   - Bad: 它有什么功能？
   - Good: iPhone 15 Pro 有什么功能？
2. **拒绝简单是非题**：尽量少生成"是不是"、"对不对"、"你认为/你觉得"等开放性问题，多生成"是什么"、"怎么做"的问题。
3. **包含元数据**：如果文本中包含特定的时间、地点、版本号，请尽量在问题中体现这些限定条件。

### 输出格式：
[
  {"question": "问题内容", "answer": "简短答案"},
  {"question": "问题内容2", "answer": "简短答案2"}
]
`, g.knowledgeName, g.knowledgeName)
}

// buildUserPrompt 构建用户消息 (包含 metadata 上下文)
func (g *Generator) buildUserPrompt(content string, metadata map[string]any, qaNum int) string {
	var sb strings.Builder

	// 如果有 metadata，先展示上下文信息
	if len(metadata) > 0 {
		sb.WriteString("### 文档上下文信息：\n")

		// 优先展示结构化标题信息
		if h1, ok := metadata["h1"].(string); ok && h1 != "" {
			sb.WriteString(fmt.Sprintf("- 一级标题 (H1): %s\n", h1))
		}
		if h2, ok := metadata["h2"].(string); ok && h2 != "" {
			sb.WriteString(fmt.Sprintf("- 二级标题 (H2): %s\n", h2))
		}
		if h3, ok := metadata["h3"].(string); ok && h3 != "" {
			sb.WriteString(fmt.Sprintf("- 三级标题 (H3): %s\n", h3))
		}

		// 展示 header_context (structure splitter 生成)
		if headerCtx, ok := metadata["header_context"].(string); ok && headerCtx != "" {
			sb.WriteString(fmt.Sprintf("- 章节标题: %s\n", headerCtx))
		}

		// 展示文件名
		if fileName, ok := metadata["_file_name"].(string); ok && fileName != "" {
			sb.WriteString(fmt.Sprintf("- 来源文件: %s\n", fileName))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("### 文本内容：\n")
	sb.WriteString(content)
	sb.WriteString(fmt.Sprintf("\n\n请基于以上内容生成 %d 个高质量的问答对。", qaNum))

	return sb.String()
}

// parseResponse 解析 LLM 返回的 JSON 响应
func (g *Generator) parseResponse(content string) ([]types.QAItem, error) {
	// 清理可能的 Markdown 代码块标记
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// 尝试找到 JSON 数组的起始位置
	startIdx := strings.Index(content, "[")
	endIdx := strings.LastIndex(content, "]")
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return nil, fmt.Errorf("无法找到有效的 JSON 数组")
	}
	content = content[startIdx : endIdx+1]

	var qaPairs []types.QAItem
	if err := json.Unmarshal([]byte(content), &qaPairs); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	return qaPairs, nil
}
