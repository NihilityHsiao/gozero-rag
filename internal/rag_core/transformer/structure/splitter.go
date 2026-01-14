package structure

import (
	"context"
	"strings"

	"gozero-rag/internal/rag_core/constant"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type Config struct {
	MaxChunkSize int
	OverlapSize  int
	Separators   []string // 用于内部递归切分
}

// StructureSplitter 结构化分片器
type StructureSplitter struct {
	config   *Config
	detector HeaderDetector
	// 内部回退使用的递归切分器
	recursiveSplitter document.Transformer
}

func NewStructureSplitter(ctx context.Context, config *Config) (*StructureSplitter, error) {
	// 初始化内部递归切分器 (用于处理超长章节)
	recSplitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   config.MaxChunkSize,
		OverlapSize: config.OverlapSize,
		Separators:  config.Separators,
		KeepType:    recursive.KeepTypeNone,
	})
	if err != nil {
		return nil, err
	}

	return &StructureSplitter{
		config:            config,
		detector:          NewRegexHeaderDetector(),
		recursiveSplitter: recSplitter,
	}, nil
}

func (s *StructureSplitter) Transform(ctx context.Context, src []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	var result []*schema.Document

	for _, doc := range src {
		// 1. 扫描标题
		headers := s.scanHeaders(doc.Content)
		logx.Infof("文档 [%s] 检测到 %d 个标题", doc.ID, len(headers))

		// 2. 按标题切分 Section
		sections := s.splitByHeaders(doc.Content, headers)

		// 3. 处理每个 Section
		for _, section := range sections {
			// 如果 Section 内容本身就很短（小于 MaxChunkSize），直接做一个 Chunk
			if len([]rune(section.Content)) <= s.config.MaxChunkSize {
				chunk, err := s.createChunk(doc, section.Content, section.ActiveHeader, section.ActiveHeaderInfo)
				if err != nil {
					return nil, err
				}
				result = append(result, chunk)
				continue
			}

			// 如果 Section 过长，进行递归切分
			subDocs, err := s.recursiveSplitter.Transform(ctx, []*schema.Document{
				{
					Content: section.Content,
				},
			})
			if err != nil {
				logx.Errorf("递归切分失败: %v", err)
				continue
			}

			// 4. 上下文注入 (Context Injection) ✨
			// 不再拼接到 Content，而是注入到 Metadata
			for _, subDoc := range subDocs {
				chunk, err := s.createChunk(doc, subDoc.Content, section.ActiveHeader, section.ActiveHeaderInfo)
				if err != nil {
					return nil, err
				}
				result = append(result, chunk)
			}
		}
	}

	return result, nil
}

type Section struct {
	Content          string
	ActiveHeader     string      // 当前生效的标题文本
	ActiveHeaderInfo *HeaderInfo // 标题详细信息 (包含 Level, Type 等)
}

// scanHeaders 扫描所有标题
func (s *StructureSplitter) scanHeaders(content string) []*HeaderInfo {
	var headers []*HeaderInfo
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if header := s.detector.Detect(line, i); header != nil {
			// 二次过滤：检查是否真的是标题 (例如检查下一行是否为空行)
			// 这里简单起见，如果 heuristic 类型，要求下一行是空行，或者已经是最后一行
			if header.Type == "heuristic" {
				if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
					continue // 可能是短句，跳过
				}
			}
			headers = append(headers, header)
		}
	}
	return headers
}

// splitByHeaders 根据标题位置切分内容
func (s *StructureSplitter) splitByHeaders(content string, headers []*HeaderInfo) []*Section {
	if len(headers) == 0 {
		return []*Section{{Content: content, ActiveHeader: "无标题", ActiveHeaderInfo: &HeaderInfo{Level: 0, Text: "无标题", Type: "implicit"}}}
	}

	var sections []*Section
	lines := strings.Split(content, "\n")

	// 处理第一个标题之前的内容
	if headers[0].LineNum > 0 {
		preLines := lines[:headers[0].LineNum]
		if len(strings.TrimSpace(strings.Join(preLines, ""))) > 0 {
			sections = append(sections, &Section{
				Content:          strings.Join(preLines, "\n"),
				ActiveHeader:     "前言/摘要",
				ActiveHeaderInfo: &HeaderInfo{Level: 0, Text: "前言/摘要", Type: "implicit"},
			})
		}
	}

	for i, header := range headers {
		startLine := header.LineNum
		var endLine int

		if i == len(headers)-1 {
			endLine = len(lines)
		} else {
			endLine = headers[i+1].LineNum
		}

		// 提取当前 Section 内容 (包含标题行本身)
		sectionLines := lines[startLine:endLine]
		sectionContent := strings.Join(sectionLines, "\n")

		sections = append(sections, &Section{
			Content:          strings.TrimSpace(sectionContent),
			ActiveHeader:     header.Text,
			ActiveHeaderInfo: header,
		})
	}

	return sections
}

func (s *StructureSplitter) createChunk(originDoc *schema.Document, content string, header string, headerInfo *HeaderInfo) (*schema.Document, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	chunk := &schema.Document{
		ID:       id.String(),
		Content:  content,
		MetaData: make(map[string]any),
	}

	// 复制原始 Metadata
	for k, v := range originDoc.MetaData {
		chunk.MetaData[k] = v
	}

	// 注入结构化元信息 (使用统一常量)
	chunk.MetaData[constant.MetaHeaderContext] = header
	if headerInfo != nil {
		chunk.MetaData[constant.MetaHeaderLevel] = headerInfo.Level
		chunk.MetaData[constant.MetaHeaderType] = headerInfo.Type
	}

	return chunk, nil
}
