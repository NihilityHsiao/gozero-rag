package structure

import (
	"regexp"
	"strings"
	"unicode"
)

// HeaderInfo 标题信息
type HeaderInfo struct {
	Level   int    // 标题级别 (1: 章节, 2: 小节)
	Text    string // 标题文本
	LineNum int    // 行号 (从0开始)
	Type    string // 标题类型 (numbered/chinese/heuristic)
}

// HeaderDetector 标题检测器接口
type HeaderDetector interface {
	Detect(line string, lineNum int) *HeaderInfo
}

// RegexHeaderDetector 基于正则和启发式规则的检测器
type RegexHeaderDetector struct {
	numberHeaderRe  *regexp.Regexp
	chineseHeaderRe *regexp.Regexp
}

func NewRegexHeaderDetector() *RegexHeaderDetector {
	return &RegexHeaderDetector{
		// 匹配 "1. 标题", "1.1 标题", "1.1.1 标题"
		numberHeaderRe: regexp.MustCompile(`^(\d+\.)+\s+(.+)$`),
		// 匹配 "第一章 标题", "第1节 标题"
		chineseHeaderRe: regexp.MustCompile(`^第[一二三四五六七八九十百千0-9]+[章节部分条款]\s*(.*)$`),
	}
}

func (d *RegexHeaderDetector) Detect(line string, lineNum int) *HeaderInfo {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil
	}

	// 1. 数字编号检测 (如: 1.1 背景)
	if matches := d.numberHeaderRe.FindStringSubmatch(line); len(matches) > 0 {
		// 计算层级: "1." -> Level 1, "1.1." -> Level 2 - 这里简化逻辑，点号数量
		// matches[1] 是最后一个匹配组，如 "1."
		// 更好的方式是统计原始字符串中 "." 的数量
		dotCount := strings.Count(strings.Split(line, " ")[0], ".")
		level := dotCount
		if level < 1 {
			level = 1
		}

		return &HeaderInfo{
			Level:   level,
			Text:    matches[2], // 标题内容
			LineNum: lineNum,
			Type:    "numbered",
		}
	}

	// 2. 中文序号检测 (如: 第一章 总则)
	if matches := d.chineseHeaderRe.FindStringSubmatch(line); len(matches) > 0 {
		return &HeaderInfo{
			Level:   1, // 默认为最高层级
			Text:    line,
			LineNum: lineNum,
			Type:    "chinese",
		}
	}

	// 3. 启发式检测 (短行 + 可能的全大写/特殊格式)
	// 规则：长度小于50，且不是句子（不以标点结尾）
	if len([]rune(line)) < 50 && !d.isSentence(line) {
		// 这里无法单独判断是否为标题，通常需要结合上下文（如后跟空行）
		// 为了简单起见，这里作为 Level 2 候选
		// 实际使用中，Splitter 会结合上下文进一步过滤

		return &HeaderInfo{
			Level:   2,
			Text:    line,
			LineNum: lineNum,
			Type:    "heuristic",
		}
	}

	return nil
}

// isSentence 判断是否是句子（以标点结尾）
func (d *RegexHeaderDetector) isSentence(line string) bool {
	runes := []rune(line)
	lastChar := runes[len(runes)-1]

	// 常见的句末标点
	punctuations := []rune{'。', '？', '！', '.', '?', '!', ';', '；', ':', '：'}
	for _, p := range punctuations {
		if lastChar == p {
			return true
		}
	}

	// 如果包含主要动词或过长，也可能是句子，这里简化处理
	return false
}

// 辅助函数：判断是否全大写 (针对英文标题)
func isAllUpper(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
