package doc_processor

import (
	"context"
	"gozero-rag/internal/rag_core/constant"
	"gozero-rag/internal/rag_core/loader"
	"gozero-rag/internal/rag_core/qa"
	"gozero-rag/internal/rag_core/transformer"
	"gozero-rag/internal/rag_core/types"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
)

type RunnableInput = *types.ProcessRequest
type RunnableOutput = []*schema.Document
type ProcessorService struct {
	indexer compose.Runnable[RunnableInput, RunnableOutput]
}

func NewDocProcessService(ctx context.Context) (*ProcessorService, error) {
	const (
		NodeReqToLoader = "Req2Loader"
		NodeLoader      = "Loader"
		NodeTransformer = "Transformer"
		NodeQaChecker   = "QaChecker"
		NodeIndexer     = "Indexer"
	)

	loader1, err := loader.NewLoader(ctx)
	if err != nil {
		return nil, err
	}
	transformer2, err := transformer.NewTransformer(ctx)
	if err != nil {
		return nil, err
	}

	// input: document.Source
	// output: []*schema.Document
	g := compose.NewGraph[RunnableInput, RunnableOutput]()

	_ = g.AddLambdaNode(NodeReqToLoader, compose.InvokableLambda(func(ctx context.Context, input *types.ProcessRequest) (output document.Source, err error) {
		output = document.Source{
			URI: input.URI,
		}
		return output, err
	}), compose.WithNodeName(NodeReqToLoader))
	_ = g.AddLoaderNode(NodeLoader, loader1)
	_ = g.AddDocumentTransformerNode(NodeTransformer, transformer2)
	_ = g.AddLambdaNode(NodeQaChecker, compose.InvokableLambda(func(ctx context.Context, input []*schema.Document) (output []*schema.Document, err error) {
		config, ok := ctx.Value(constant.CtxKeyIndexConfig).(types.ProcessConfig)
		if !ok || !config.EnableQACheck {
			logx.Infof("[indexer-service] QA checker disabled")
			return input, nil
		}
		checker := qa.NewQaChecker()
		return checker.Check(ctx, input)
	}), compose.WithNodeName(NodeQaChecker))

	// _ = g.AddIndexerNode(NodeIndexer)

	_ = g.AddEdge(compose.START, NodeReqToLoader)
	_ = g.AddEdge(NodeReqToLoader, NodeLoader)
	_ = g.AddEdge(NodeLoader, NodeTransformer)
	_ = g.AddEdge(NodeTransformer, NodeQaChecker)
	_ = g.AddEdge(NodeQaChecker, compose.END)

	r, err := g.Compile(ctx, compose.WithGraphName("indexer-service"))
	if err != nil {
		return nil, err
	}

	return &ProcessorService{
		indexer: r,
	}, nil
}
func (l *ProcessorService) Invoke(ctx context.Context, input RunnableInput, opts ...compose.Option) (output RunnableOutput, err error) {
	withCallbacks := compose.WithCallbacks(logCallback())

	opts = append(opts, withCallbacks)

	// 注入 config
	ctx = context.WithValue(ctx, constant.CtxKeyIndexConfig, input.IndexConfig)
	return l.indexer.Invoke(ctx, input, opts...)
}

func (l *ProcessorService) Stream(ctx context.Context, input RunnableOutput, opts ...compose.Option) (output *schema.StreamReader[RunnableOutput], err error) {
	//TODO implement me
	panic("implement me")
}

func (l *ProcessorService) Collect(ctx context.Context, input *schema.StreamReader[RunnableInput], opts ...compose.Option) (output RunnableOutput, err error) {
	//TODO implement me
	panic("implement me")
}

func (l *ProcessorService) Transform(ctx context.Context, input *schema.StreamReader[RunnableInput], opts ...compose.Option) (output *schema.StreamReader[RunnableOutput], err error) {
	//TODO implement me
	panic("implement me")
}
func logCallback() callbacks.Handler {
	builder := callbacks.NewHandlerBuilder()
	builder.OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {

		logx.Infof("[indexer-service] - %s:%s:%s, ctx:%v", info.Component, info.Type, info.Name, ctx.Value(constant.CtxKeyIndexConfig))

		return ctx
	})

	return builder.Build()
}
