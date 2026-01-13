package parser

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino-ext/components/document/parser/xlsx"

	"github.com/cloudwego/eino/components/document/parser"
)

func NewParser(ctx context.Context) (p parser.Parser, err error) {
	textParser := parser.TextParser{}

	xlsxParser, err := xlsx.NewXlsxParser(ctx, nil)
	if err != nil {
		return nil, err
	}

	pdfParser, err := pdf.NewPDFParser(ctx, &pdf.Config{})
	if err != nil {
		return nil, err
	}

	// 创建扩展解析器，支持大小写不敏感的文件扩展名
	p, err = parser.NewExtParser(ctx, &parser.ExtParserConfig{
		// 注册特定扩展名的解析器（小写）
		Parsers: map[string]parser.Parser{
			".pdf":  pdfParser,
			".xlsx": xlsxParser,
			".xls":  xlsxParser,
		},
		// 设置默认解析器，用于处理未知格式
		FallbackParser: textParser,
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}
