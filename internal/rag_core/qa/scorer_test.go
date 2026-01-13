package qa

import (
	"testing"

	"gozero-rag/internal/rag_core/types"

	"github.com/stretchr/testify/assert"
)

func TestScorer_CalcLengthScore(t *testing.T) {
	scorer := NewScorer()

	tests := []struct {
		name          string
		contentLength int
		expectedScore float64
		hasIssue      bool
	}{
		{"最佳长度 500 字", 500, 20.0, false},
		{"最佳长度下限 300 字", 300, 20.0, false},
		{"最佳长度上限 800 字", 800, 20.0, false},
		{"稍短 200 字", 200, 12.0, false},
		{"稍长 900 字", 900, 16.0, false},
		{"过短 50 字", 50, 6.0, true},
		{"过长 1500 字", 1500, 10.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成指定长度的内容
			content := generateContent(tt.contentLength)
			score := scorer.Score(content, nil, "")

			// 由于其他分项可能影响总分，我们只检查长度分数
			assert.Equal(t, tt.expectedScore, score.LengthScore, "长度得分应该正确")

			if tt.hasIssue {
				hasChunkIssue := false
				for _, issue := range score.Issues {
					if issue == types.IssueChunkTooShort || issue == types.IssueChunkTooLong {
						hasChunkIssue = true
						break
					}
				}
				assert.True(t, hasChunkIssue, "应该有长度相关的问题")
			}
		})
	}
}

func TestScorer_CalcStructureScore(t *testing.T) {
	scorer := NewScorer()

	tests := []struct {
		name       string
		content    string
		minScore   float64
		checkIssue string
	}{
		{"完整句子", "这是一个完整的句子。", 15.0, ""},
		{"问号结尾", "这是一个问题吗？", 15.0, ""},
		{"逗号结尾", "这是一个不完整的句子，", 10.0, types.IssueTruncatedEnd},
		{"代词开头", "它是一个好产品。", 10.0, types.IssueDanglingReference},
		{"但是开头", "但是这个功能不太好。", 10.0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.Score(tt.content, nil, "")
			assert.GreaterOrEqual(t, score.StructureScore, tt.minScore, "结构得分应该符合预期")

			if tt.checkIssue != "" {
				found := false
				for _, issue := range score.Issues {
					if issue == tt.checkIssue {
						found = true
						break
					}
				}
				assert.True(t, found, "应该包含预期的问题: %s", tt.checkIssue)
			}
		})
	}
}

func TestScorer_CalcQAScore(t *testing.T) {
	scorer := NewScorer()
	content := "Go 是一种静态类型的编程语言。它由 Google 开发，具有高效的并发处理能力。"

	tests := []struct {
		name           string
		qaPairs        []types.QAItem
		minScore       float64
		hasLowCoverage bool
	}{
		{
			name:           "无 QA",
			qaPairs:        nil,
			minScore:       0,
			hasLowCoverage: true,
		},
		{
			name: "1 个 QA",
			qaPairs: []types.QAItem{
				{Question: "Go 语言是什么类型的语言？", Answer: "静态类型语言"},
			},
			minScore:       3.0,
			hasLowCoverage: true,
		},
		{
			name: "3 个高质量 QA",
			qaPairs: []types.QAItem{
				{Question: "Go 语言是什么类型的语言？", Answer: "静态类型语言"},
				{Question: "Go 语言由哪家公司开发？", Answer: "Google"},
				{Question: "Go 语言有什么特点？", Answer: "高效的并发处理能力"},
			},
			minScore:       10.0, // 覆盖率满分 + 部分多样性和相关性
			hasLowCoverage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.Score(content, tt.qaPairs, "")
			assert.GreaterOrEqual(t, score.QAScore, tt.minScore, "QA 得分应该符合预期")

			hasLowCoverage := false
			for _, issue := range score.Issues {
				if issue == types.IssueLowQACoverage {
					hasLowCoverage = true
					break
				}
			}
			assert.Equal(t, tt.hasLowCoverage, hasLowCoverage, "低覆盖率问题应该符合预期")
		})
	}
}

func TestScorer_TotalScore(t *testing.T) {
	scorer := NewScorer()

	// 一个高质量的 chunk
	content := generateContent(500) + "。这是关于 Go 语言的介绍。Go 是一种静态类型语言。它由 Google 开发。Go 具有高效的并发能力。"
	qaPairs := []types.QAItem{
		{Question: "Go 语言是什么类型的语言？", Answer: "静态类型语言"},
		{Question: "Go 语言由哪家公司开发？", Answer: "Google"},
		{Question: "Go 语言有什么特点？", Answer: "高效并发能力"},
	}

	score := scorer.Score(content, qaPairs, "")

	t.Logf("总分: %.2f", score.TotalScore)
	t.Logf("长度: %.2f, 结构: %.2f, 内容: %.2f, 语义: %.2f, QA: %.2f",
		score.LengthScore, score.StructureScore, score.ContentScore, score.SemanticScore, score.QAScore)
	t.Logf("问题: %v", score.Issues)
	t.Logf("建议: %v", score.Suggestions)

	// 总分应该是各项之和
	expectedTotal := score.LengthScore + score.StructureScore + score.ContentScore + score.SemanticScore + score.QAScore
	assert.Equal(t, expectedTotal, score.TotalScore, "总分应该等于各项之和")

	// 这个高质量 chunk 总分应该 >= 60
	assert.GreaterOrEqual(t, score.TotalScore, 60.0, "高质量 chunk 总分应该 >= 60")
}

// generateContent 生成指定字数的测试内容
func generateContent(length int) string {
	base := "这是一段测试内容"
	result := ""
	for len([]rune(result)) < length {
		result += base
	}
	runes := []rune(result)
	if len(runes) > length {
		runes = runes[:length]
	}
	return string(runes)
}
