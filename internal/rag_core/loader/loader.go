package loader

import (
	"context"
	"fmt"
	"gozero-rag/internal/rag_core/parser"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"

	"net/url"

	"github.com/cloudwego/eino/schema"
)

type Loader struct {
	fileLoader document.Loader
}

func NewLoader(ctx context.Context) (document.Loader, error) {
	p, err := parser.NewParser(ctx)
	if err != nil {
		return nil, err
	}

	config := &file.FileLoaderConfig{
		UseNameAsID: true,
		Parser:      p,
	}
	fldr, err := file.NewFileLoader(ctx, config)
	if err != nil {
		return nil, err
	}
	return &Loader{
		fileLoader: fldr,
	}, nil
}

func (l *Loader) isURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func (l *Loader) Load(ctx context.Context, src document.Source, opts ...document.LoaderOption) ([]*schema.Document, error) {
	if l.isURL(src.URI) {
		return nil, fmt.Errorf("暂时不支持URL解析")
	}

	// 规范化拓展名,支持大小写不敏感的解析

	return l.fileLoader.Load(ctx, src, opts...)
}
