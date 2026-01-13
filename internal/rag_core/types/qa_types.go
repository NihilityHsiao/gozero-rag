package types

// ========================
// QA Checker 相关类型定义
// ========================

// QAItem QA 问答对
type QAItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// ChunkQualityScore chunk 质量评分结果
type ChunkQualityScore struct {
	TotalScore     float64  `json:"total_score"`     // 总分 (0-100)
	LengthScore    float64  `json:"length_score"`    // 长度合理性 (0-20)
	StructureScore float64  `json:"structure_score"` // 结构完整性 (0-20)
	ContentScore   float64  `json:"content_score"`   // 内容质量 (0-20)
	SemanticScore  float64  `json:"semantic_score"`  // 语义完整性 (0-20)
	QAScore        float64  `json:"qa_score"`        // QA 质量分 (0-20)
	Issues         []string `json:"issues"`          // 问题列表
	Suggestions    []string `json:"suggestions"`     // 建议列表
}

// ========================
// 问题类型常量 (Issues)
// ========================

const (
	// IssueChunkTooShort chunk 过短 (字数 < 100)
	IssueChunkTooShort = "chunk_too_short"
	// IssueChunkTooLong chunk 过长 (字数 > 1000)
	IssueChunkTooLong = "chunk_too_long"
	// IssueTruncatedEnd 句子被截断 (以逗号/连词结尾)
	IssueTruncatedEnd = "truncated_end"
	// IssueDanglingReference 指代不明 (以代词开头)
	IssueDanglingReference = "dangling_reference"
	// IssueLowQACoverage QA 覆盖不足 (QA 数量 < 2)
	IssueLowQACoverage = "low_qa_coverage"
	// IssueHighOverlap 重复度过高 (overlap > 50%)
	IssueHighOverlap = "high_overlap"
	// IssueLowInfoDensity 信息密度过低
	IssueLowInfoDensity = "low_info_density"
)

// ========================
// 建议类型常量 (Suggestions)
// ========================

const (
	// SuggestionMergeWithPrev 考虑与前一个 chunk 合并
	SuggestionMergeWithPrev = "merge_with_prev"
	// SuggestionMergeWithNext 考虑与后一个 chunk 合并
	SuggestionMergeWithNext = "merge_with_next"
	// SuggestionSplitFurther 考虑进一步拆分
	SuggestionSplitFurther = "split_further"
	// SuggestionExtendBoundary 调整边界以保留完整句子
	SuggestionExtendBoundary = "extend_boundary"
	// SuggestionGenerateQA 需要生成 QA 问答对
	SuggestionGenerateQA = "generate_qa"
)

// ========================
// 评分阈值常量
// ========================

const (
	// ScoreThresholdHigh 高质量阈值 (>= 80)
	ScoreThresholdHigh = 80.0
	// ScoreThresholdMedium 中等质量阈值 (>= 60)
	ScoreThresholdMedium = 60.0
	// ScoreThresholdLow 低质量阈值 (< 60)
	ScoreThresholdLow = 60.0

	// LengthOptimalMin 最佳长度下限 (300 字)
	LengthOptimalMin = 300
	// LengthOptimalMax 最佳长度上限 (800 字)
	LengthOptimalMax = 800
	// LengthTooShort 过短阈值 (100 字)
	LengthTooShort = 100
	// LengthTooLong 过长阈值 (1000 字)
	LengthTooLong = 1000

	// QAMinCount 触发 QA 生成的最小数量阈值
	QAMinCount = 2
	// QAOptimalCount QA 覆盖率满分数量
	QAOptimalCount = 3

	// OverlapThreshold 重复度过高阈值 (50%)
	OverlapThreshold = 0.5
)
