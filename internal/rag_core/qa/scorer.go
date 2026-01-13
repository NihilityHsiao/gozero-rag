package qa

import (
	"strings"
	"unicode"

	"gozero-rag/internal/rag_core/types"
)

// ========================
// 评分器 - 计算 chunk 质量分数
// 总分 100.0，5 个分项各 20.0
// ========================

// 截断词列表 (连词)
var truncationWords = []string{"但是", "并且", "因此", "所以", "然而", "不过", "而且", "以及", "或者"}

// 指代词列表
var pronounWords = []string{"它", "他", "她", "这", "那", "该", "其", "此"}

// 句末完整标点
var sentenceEndPuncs = []rune{'。', '？', '！', '.', '?', '!'}

// 截断式结尾标点
var truncationEndPuncs = []rune{',', '，', '、', ';', '；', ':', '：'}

// Scorer chunk 质量评分器
type Scorer struct{}

// NewScorer 创建评分器实例
func NewScorer() *Scorer {
	return &Scorer{}
}

// Score 对单个 chunk 进行完整评分
func (s *Scorer) Score(content string, qaPairs []types.QAItem, prevContent string) *types.ChunkQualityScore {
	result := &types.ChunkQualityScore{
		Issues:      []string{},
		Suggestions: []string{},
	}

	// 1. 长度合理性
	result.LengthScore, result.Issues, result.Suggestions = s.calcLengthScore(content, result.Issues, result.Suggestions)

	// 2. 结构完整性
	result.StructureScore, result.Issues, result.Suggestions = s.calcStructureScore(content, result.Issues, result.Suggestions)

	// 3. 内容质量 (信息密度 + 重复度)
	result.ContentScore, result.Issues, result.Suggestions = s.calcContentScore(content, prevContent, result.Issues, result.Suggestions)

	// 4. 语义完整性
	result.SemanticScore, result.Issues, result.Suggestions = s.calcSemanticScore(content, result.Issues, result.Suggestions)

	// 5. QA 质量分
	result.QAScore, result.Issues, result.Suggestions = s.calcQAScore(qaPairs, content, result.Issues, result.Suggestions)

	// 计算总分
	result.TotalScore = result.LengthScore + result.StructureScore + result.ContentScore + result.SemanticScore + result.QAScore

	return result
}

// calcLengthScore 计算长度合理性得分 (满分 20.0)
func (s *Scorer) calcLengthScore(content string, issues, suggestions []string) (float64, []string, []string) {
	length := len([]rune(content))
	var score float64

	switch {
	case length >= types.LengthOptimalMin && length <= types.LengthOptimalMax:
		// 最佳长度 300-800
		score = 20.0
	case length >= types.LengthTooShort && length < types.LengthOptimalMin:
		// 稍短 100-299
		score = 12.0
	case length > types.LengthOptimalMax && length <= types.LengthTooLong:
		// 稍长 801-1000
		score = 16.0
	case length < types.LengthTooShort:
		// 过短 < 100
		score = 6.0
		issues = append(issues, types.IssueChunkTooShort)
		suggestions = append(suggestions, types.SuggestionMergeWithPrev)
	default:
		// 过长 > 1000
		score = 10.0
		issues = append(issues, types.IssueChunkTooLong)
		suggestions = append(suggestions, types.SuggestionSplitFurther)
	}

	return score, issues, suggestions
}

// calcStructureScore 计算结构完整性得分 (满分 20.0)
func (s *Scorer) calcStructureScore(content string, issues, suggestions []string) (float64, []string, []string) {
	var score float64
	runes := []rune(strings.TrimSpace(content))
	if len(runes) == 0 {
		return 0, issues, suggestions
	}

	lastChar := runes[len(runes)-1]
	firstChars := string(runes[:min(10, len(runes))])

	// 检查句末是否为完整标点 (+10.0)
	endsWithSentencePunc := false
	for _, p := range sentenceEndPuncs {
		if lastChar == p {
			endsWithSentencePunc = true
			score += 10.0
			break
		}
	}

	// 检查是否以截断式标点结尾 (-5.0 如果以逗号等结尾)
	if !endsWithSentencePunc {
		for _, p := range truncationEndPuncs {
			if lastChar == p {
				issues = append(issues, types.IssueTruncatedEnd)
				suggestions = append(suggestions, types.SuggestionExtendBoundary)
				break
			}
		}
		// 不是句末标点，给 5 分基础分
		score += 5.0
	}

	// 检查是否以截断词/连词开头 (+5.0 / 0)
	startsWithTruncation := false
	for _, word := range truncationWords {
		if strings.HasPrefix(firstChars, word) {
			startsWithTruncation = true
			break
		}
	}
	if !startsWithTruncation {
		score += 5.0
	}

	// 检查是否以代词开头 (+5.0 / 0)
	startsWithPronoun := false
	for _, pronoun := range pronounWords {
		if strings.HasPrefix(firstChars, pronoun) {
			startsWithPronoun = true
			issues = append(issues, types.IssueDanglingReference)
			break
		}
	}
	if !startsWithPronoun {
		score += 5.0
	}

	return score, issues, suggestions
}

// calcContentScore 计算内容质量得分 (满分 20.0)
// 包含: 信息密度 (10.0) + 重复度 (10.0)
func (s *Scorer) calcContentScore(content string, prevContent string, issues, suggestions []string) (float64, []string, []string) {
	var score float64

	// 1. 信息密度: 非空白字符比例 (满分 10.0)
	totalChars := len([]rune(content))
	nonWhitespace := 0
	for _, r := range content {
		if !unicode.IsSpace(r) {
			nonWhitespace++
		}
	}
	if totalChars > 0 {
		density := float64(nonWhitespace) / float64(totalChars)
		densityScore := density * 10.0
		if densityScore < 5.0 {
			issues = append(issues, types.IssueLowInfoDensity)
		}
		score += densityScore
	}

	// 2. 重复度: 与前一个 chunk 的重叠程度 (满分 10.0)
	if prevContent == "" {
		// 没有前一个 chunk，满分
		score += 10.0
	} else {
		overlapRatio := s.calcOverlapRatio(content, prevContent)
		if overlapRatio > types.OverlapThreshold {
			issues = append(issues, types.IssueHighOverlap)
		}
		repeatScore := (1 - overlapRatio) * 10.0
		score += repeatScore
	}

	return score, issues, suggestions
}

// calcOverlapRatio 计算两个字符串的重叠比例 (简化算法: 最长公共子串)
func (s *Scorer) calcOverlapRatio(content, prevContent string) float64 {
	// 简化实现: 检查 content 开头是否与 prevContent 结尾重叠
	runesCurr := []rune(content)
	runesPrev := []rune(prevContent)

	minLen := min(len(runesCurr), len(runesPrev))
	if minLen == 0 {
		return 0
	}

	// 检查前后重叠区域
	maxOverlap := 0
	for i := 1; i <= minLen; i++ {
		// prevContent 的后 i 个字符 vs content 的前 i 个字符
		if string(runesPrev[len(runesPrev)-i:]) == string(runesCurr[:i]) {
			maxOverlap = i
		}
	}

	return float64(maxOverlap) / float64(minLen)
}

// calcSemanticScore 计算语义完整性得分 (满分 20.0)
func (s *Scorer) calcSemanticScore(content string, issues, suggestions []string) (float64, []string, []string) {
	score := 20.0

	// 统计代词数量，每个代词扣 2.0 (最多扣 10.0)
	pronounCount := 0
	for _, pronoun := range pronounWords {
		pronounCount += strings.Count(content, pronoun)
	}
	pronounPenalty := float64(min(pronounCount, 5)) * 2.0
	score -= pronounPenalty

	// 统计截断词数量，每个扣 3.0 (最多扣 9.0)
	truncationCount := 0
	for _, word := range truncationWords {
		truncationCount += strings.Count(content, word)
	}
	truncationPenalty := float64(min(truncationCount, 3)) * 3.0
	score -= truncationPenalty

	// 确保不低于 0
	if score < 0 {
		score = 0
	}

	return score, issues, suggestions
}

// calcQAScore 计算 QA 质量分 (满分 20.0)
// 覆盖率 (10.0) + 多样性 (5.0) + 相关性 (5.0)
func (s *Scorer) calcQAScore(qaPairs []types.QAItem, content string, issues, suggestions []string) (float64, []string, []string) {
	var score float64
	qaCount := len(qaPairs)

	// 1. 覆盖率: min(QA数量 / 3, 1.0) * 10.0
	coverageScore := float64(min(qaCount, types.QAOptimalCount)) / float64(types.QAOptimalCount) * 10.0
	score += coverageScore

	if qaCount < types.QAMinCount {
		issues = append(issues, types.IssueLowQACoverage)
		suggestions = append(suggestions, types.SuggestionGenerateQA)
	}

	if qaCount == 0 {
		// 没有 QA，直接返回 0
		return score, issues, suggestions
	}

	// 2. 多样性: 1 - (重复开头数 / 总问题数) * 5.0
	prefixCount := make(map[string]int)
	for _, qa := range qaPairs {
		// 取问题的前 4 个字作为开头
		runes := []rune(qa.Question)
		prefix := string(runes[:min(4, len(runes))])
		prefixCount[prefix]++
	}
	duplicatePrefixes := 0
	for _, count := range prefixCount {
		if count > 1 {
			duplicatePrefixes += count - 1
		}
	}
	diversityScore := (1 - float64(duplicatePrefixes)/float64(qaCount)) * 5.0
	if diversityScore < 0 {
		diversityScore = 0
	}
	score += diversityScore

	// 3. 相关性: (问题关键词匹配数 / 总问题数) * 5.0
	matchedCount := 0
	for _, qa := range qaPairs {
		// 简化: 检查问题中的关键词是否出现在 content 中
		// 提取问题中的中文词 (长度 >= 2)
		words := extractKeywords(qa.Question)
		for _, word := range words {
			if strings.Contains(content, word) {
				matchedCount++
				break
			}
		}
	}
	relevanceScore := float64(matchedCount) / float64(qaCount) * 5.0
	score += relevanceScore

	return score, issues, suggestions
}

// extractKeywords 从文本中提取关键词 (简化版: 提取连续中文字符)
func extractKeywords(text string) []string {
	var keywords []string
	var current []rune

	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			current = append(current, r)
		} else {
			if len(current) >= 2 {
				keywords = append(keywords, string(current))
			}
			current = nil
		}
	}
	if len(current) >= 2 {
		keywords = append(keywords, string(current))
	}

	return keywords
}
