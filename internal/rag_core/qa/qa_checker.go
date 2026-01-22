package qa

import (
	"context"
	"encoding/json"
	"errors"

	"gozero-rag/internal/rag_core/constant"
	"gozero-rag/internal/rag_core/types"

	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
)

// QaChecker QA 质量检查器
type QaChecker struct {
	scorer *Scorer
}

// NewQaChecker 创建 QA 检查器
func NewQaChecker() *QaChecker {
	return &QaChecker{
		scorer: NewScorer(),
	}
}

// Check 对所有 chunk 进行质量检查和 QA 生成
func (c *QaChecker) Check(ctx context.Context, input []*schema.Document) (output []*schema.Document, err error) {
	conf, ok := ctx.Value(constant.CtxKeyIndexConfig).(types.ProcessConfig)
	if !ok {
		return nil, errors.New("index config not found in context")
	}

	// 获取知识库名称 (从第一个 chunk 的 metadata 获取，或使用默认值)
	knowledgeName := conf.KnowledgeName

	logx.Infof("[QaChecker] 开始检查 %d 个 chunk, 知识库: %s", len(input), knowledgeName)

	// 第一轮: 评分并识别需要生成 QA 的 chunk
	var needQAChunks []*schema.Document
	var needQAIndices []int

	for i, doc := range input {
		// 获取已有的 QA pairs
		qaPairs := c.getQAPairs(doc)

		// 获取前一个 chunk 的内容 (用于计算重复度)
		prevContent := ""
		if i > 0 {
			prevContent = input[i-1].Content
		}

		// 计算评分
		score := c.scorer.Score(doc.Content, qaPairs, prevContent)

		// 存储评分结果到 Metadata
		c.storeScore(doc, score)

		// 检查是否需要生成 QA
		// 条件: QA 数量不足 且 内容长度足够 (过短的 chunk 不值得生成 QA)
		contentLen := len([]rune(doc.Content))
		if len(qaPairs) < types.QAMinCount && contentLen >= types.LengthTooShort {
			needQAChunks = append(needQAChunks, doc)
			needQAIndices = append(needQAIndices, i)
		}
	}

	logx.Infof("[QaChecker] 需要生成 QA 的 chunk 数量: %d", len(needQAChunks))

	// 第二轮: 批量生成 QA (如果有 API Key)
	if conf.LlmConfig.QaKey != "" && len(needQAChunks) > 0 {
		generator := NewGenerator(conf, knowledgeName)

		qaPairsResults, err := generator.GenerateBatch(ctx, needQAChunks, conf.QaNum)
		if err != nil {
			logx.Errorf("[QaChecker] 批量生成 QA 失败: %v", err)
			// 不阻塞流程，继续处理
		} else {
			// 更新生成的 QA 到对应 chunk
			for i, qaPairs := range qaPairsResults {
				if qaPairs != nil {
					c.storeQAPairs(needQAChunks[i], qaPairs)

					// 重新计算 QA 分数
					idx := needQAIndices[i]
					prevContent := ""
					if idx > 0 {
						prevContent = input[idx-1].Content
					}
					score := c.scorer.Score(needQAChunks[i].Content, qaPairs, prevContent)
					c.storeScore(needQAChunks[i], score)
				}
			}
		}
	} else if conf.LlmConfig.QaKey == "" && len(needQAChunks) > 0 {
		logx.Infof("[QaChecker] 未配置 ApiKey，跳过 QA 生成")
	}

	// 统计结果
	c.logStats(input)

	return input, nil
}

// getKnowledgeName 获取知识库名称
func (c *QaChecker) getKnowledgeName(docs []*schema.Document) string {
	if len(docs) == 0 {
		return "未命名知识库"
	}

	// 尝试从 metadata 获取
	if name, ok := docs[0].MetaData[constant.MetaSourceURI].(string); ok && name != "" {
		return name
	}

	return "未命名知识库"
}

// getQAPairs 从 chunk 的 Metadata 中获取已有的 QA pairs
func (c *QaChecker) getQAPairs(doc *schema.Document) []types.QAItem {
	if doc.MetaData == nil {
		return nil
	}

	raw, ok := doc.MetaData[constant.MetaQaPairs]
	if !ok {
		return nil
	}

	// 尝试类型断言
	if qaPairs, ok := raw.([]types.QAItem); ok {
		return qaPairs
	}

	// 尝试 JSON 反序列化 (如果存储为 JSON 字符串)
	if jsonStr, ok := raw.(string); ok {
		var qaPairs []types.QAItem
		if err := json.Unmarshal([]byte(jsonStr), &qaPairs); err == nil {
			return qaPairs
		}
	}

	return nil
}

// storeQAPairs 存储 QA pairs 到 Metadata
func (c *QaChecker) storeQAPairs(doc *schema.Document, qaPairs []types.QAItem) {
	if doc.MetaData == nil {
		doc.MetaData = make(map[string]any)
	}
	doc.MetaData[constant.MetaQaPairs] = qaPairs
}

// storeScore 存储评分结果到 Metadata
func (c *QaChecker) storeScore(doc *schema.Document, score *types.ChunkQualityScore) {
	if doc.MetaData == nil {
		doc.MetaData = make(map[string]any)
	}

	// 存储总分
	doc.MetaData[constant.MetaQualityScore] = score.TotalScore

	// 存储详细评分
	doc.MetaData[constant.MetaQualityDetails] = score
}

// logStats 输出统计信息
func (c *QaChecker) logStats(docs []*schema.Document) {
	var highQuality, mediumQuality, lowQuality int
	var totalScore float64

	for _, doc := range docs {
		score, ok := doc.MetaData[constant.MetaQualityScore].(float64)
		if !ok {
			continue
		}
		totalScore += score

		switch {
		case score >= types.ScoreThresholdHigh:
			highQuality++
		case score >= types.ScoreThresholdMedium:
			mediumQuality++
		default:
			lowQuality++
		}
	}

	avgScore := 0.0
	if len(docs) > 0 {
		avgScore = totalScore / float64(len(docs))
	}

	logx.Infof("[QaChecker] 质量统计 - 总计: %d, 高质量: %d, 中等: %d, 低质量: %d, 平均分: %.2f",
		len(docs), highQuality, mediumQuality, lowQuality, avgScore)
}
