package extractor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/cloudwego/eino/callbacks"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"

	"gozero-rag/internal/graphrag/prompt"
	"gozero-rag/internal/graphrag/types"
	"gozero-rag/internal/model/chunk"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type GraphExtractor struct {
	runnable compose.Runnable[GraphInput, GraphOutput]
}

func (l *GraphExtractor) Invoke(ctx context.Context, input GraphInput, opts ...compose.Option) (output GraphOutput, err error) {
	withCallbacks := compose.WithCallbacks(logCallback())

	opts = append(opts, withCallbacks)

	// 注入 config
	//ctx = context.WithValue(ctx, constant.CtxKeyIndexConfig, input.IndexConfig)
	return l.runnable.Invoke(ctx, input, opts...)
}

type GraphInput = *GraphExtractionState
type GraphOutput = *GraphExtractionState

// GraphExtractionState holds the state for the Eino graph loop
type GraphExtractionState struct {
	InputText     string
	History       string
	LoopCount     int
	IsDone        bool
	CurrentPrompt []*schema.Message
	CurrentOutput string
	LLM           model.ToolCallingChatModel
}

const MaxLoops = 3

func NewGraphExtractor(ctx context.Context) (*GraphExtractor, error) {
	g := compose.NewGraph[*GraphExtractionState, *GraphExtractionState]()

	_ = g.AddLambdaNode("generate_prompt", compose.InvokableLambda(func(ctx context.Context, state *GraphExtractionState) (*GraphExtractionState, error) {
		var msgs []*schema.Message
		var err error

		if state.LoopCount == 0 {
			msgs, err = prompt.NewGraphPrompt(state.InputText)
		} else {
			msgs, err = prompt.NewGraphLoopPrompt(state.History)
		}
		if err != nil {
			return nil, err
		}
		state.CurrentPrompt = msgs
		return state, nil
	}))

	_ = g.AddLambdaNode("llm", compose.InvokableLambda(func(ctx context.Context, state *GraphExtractionState) (*GraphExtractionState, error) {
		if state.LLM == nil {
			return nil, fmt.Errorf("LLM not provided in GraphExtractionState")
		}
		resp, err := state.LLM.Generate(ctx, state.CurrentPrompt)
		if err != nil {
			return nil, err
		}
		state.CurrentOutput = resp.Content
		return state, nil
	}))

	_ = g.AddLambdaNode("parse_output", compose.InvokableLambda(func(ctx context.Context, state *GraphExtractionState) (*GraphExtractionState, error) {
		output := state.CurrentOutput
		// Check for completion
		if strings.Contains(output, "<|DONE|>") {
			state.IsDone = true
			output = strings.ReplaceAll(output, "<|DONE|>", "")
		}

		// Clean up output
		output = strings.TrimSpace(output)

		if state.History != "" {
			state.History += "##" + output
		} else {
			state.History = output
		}

		state.LoopCount++
		return state, nil
	}))

	_ = g.AddEdge(compose.START, "generate_prompt")
	_ = g.AddEdge("generate_prompt", "llm")
	_ = g.AddEdge("llm", "parse_output")
	_ = g.AddEdge("parse_output", compose.END)

	r, err := g.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &GraphExtractor{runnable: r}, nil
}

func (e *GraphExtractor) Extract(ctx context.Context, chunks []*chunk.Chunk, llm model.ToolCallingChatModel) (*types.GraphExtractionResult, error) {
	runnable := e.runnable

	result := &types.GraphExtractionResult{
		Entities:  make([]types.Entity, 0),
		Relations: make([]types.Relation, 0),
	}

	// Process each chunk
	cl := len(chunks)
	for i, c := range chunks {
		state := &GraphExtractionState{
			InputText: c.Content,
			History:   "",
			LoopCount: 0,
			IsDone:    false,
			LLM:       llm,
		}

		for state.LoopCount < MaxLoops && !state.IsDone {
			// Run the graph iteration
			newState, err := runnable.Invoke(ctx, state)
			if err != nil {
				logx.Errorf("Error extracting from chunk %v: %v", c.Id, err)
				break
			}
			state = newState
		}

		// Parse the accumulated history (which contains all results)
		entities, relations := e.parseHistory(state.History, c.Id)
		result.Entities = append(result.Entities, entities...)
		result.Relations = append(result.Relations, relations...)
		logx.Infof("知识图谱提取进度: %d/%d", i+1, cl)

	}

	result = e.Merge([]*types.GraphExtractionResult{result})
	return result, nil
}

// ParallelExtract 将 chunks 分成 numParts 份，用 concurrency 个 goroutine 并行提取
// 内部合并所有结果后返回，对调用方透明
func (e *GraphExtractor) ParallelExtract(
	ctx context.Context,
	chunks []*chunk.Chunk,
	llm model.ToolCallingChatModel,
	numParts int,
	concurrency int,
) (*types.GraphExtractionResult, error) {
	if len(chunks) == 0 {
		return &types.GraphExtractionResult{
			Entities:  make([]types.Entity, 0),
			Relations: make([]types.Relation, 0),
		}, nil
	}

	// 分割 chunks
	chunkParts := splitSlice(chunks, numParts)
	logx.Infof("并行提取: 总 chunks=%d, 分成 %d 份, 并发数=%d", len(chunks), len(chunkParts), concurrency)

	// 用于收集结果
	var mu sync.Mutex
	results := make([]*types.GraphExtractionResult, 0, len(chunkParts))

	// 使用 errgroup 控制并发
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency) // 限制并发数

	for i, part := range chunkParts {
		partIndex := i
		partChunks := part
		g.Go(func() error {
			logx.Infof("开始提取第 %d 部分 (%d chunks)", partIndex+1, len(partChunks))

			// 调用原有的 Extract 处理这部分 chunks
			partResult, err := e.Extract(gctx, partChunks, llm)
			if err != nil {
				logx.Errorf("提取第 %d 部分失败: %v", partIndex+1, err)
				return err
			}

			// 收集结果
			mu.Lock()
			results = append(results, partResult)
			mu.Unlock()

			logx.Infof("第 %d 部分提取完成: %d 实体, %d 关系",
				partIndex+1, len(partResult.Entities), len(partResult.Relations))
			return nil
		})
	}

	// 等待所有 goroutine 完成
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("parallel extraction failed: %w", err)
	}

	// 合并所有结果
	finalResult := e.Merge(results)
	logx.Infof("并行提取完成: 合并后 %d 实体, %d 关系",
		len(finalResult.Entities), len(finalResult.Relations))

	return finalResult, nil
}

// splitSlice 将 slice 分成 n 份
func splitSlice[T any](slice []T, n int) [][]T {
	if n <= 0 {
		n = 1
	}
	if n > len(slice) {
		n = len(slice)
	}

	result := make([][]T, n)
	partSize := len(slice) / n
	remainder := len(slice) % n

	start := 0
	for i := 0; i < n; i++ {
		end := start + partSize
		if i < remainder {
			end++ // 前 remainder 份多分一个
		}
		result[i] = slice[start:end]
		start = end
	}

	return result
}

func (e *GraphExtractor) parseHistory(history string, sourceId string) ([]types.Entity, []types.Relation) {
	entities := make([]types.Entity, 0)
	relations := make([]types.Relation, 0)

	// Split by record delimiter ##
	records := strings.Split(history, "##")

	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		// Remove parens
		record = strings.TrimPrefix(record, "(")
		record = strings.TrimSuffix(record, ")")

		// Split by tuple delimiter <|>
		parts := strings.Split(record, "<|>")
		if len(parts) < 2 {
			continue
		}

		// Remove quotes from parts
		for i := range parts {
			// Handle "entity", etc.
			parts[i] = strings.Trim(parts[i], "\"")
		}

		typeTag := parts[0]
		// ("entity"<|>"Name"<|>"Type"<|>"Desc")
		// parts: [entity, Name, Type, Desc]
		if typeTag == "entity" && len(parts) >= 4 {
			entities = append(entities, types.Entity{
				Name:        parts[1],
				Type:        parts[2],
				Description: parts[3],
				SourceId:    []string{sourceId},
			})
		} else if typeTag == "relationship" && len(parts) >= 6 {
			// ("relationship"<|>"Src"<|>"Tgt"<|>"RelationType"<|>"Desc"<|>Weight)
			// parts: [relationship, Src, Tgt, RelationType, Desc, Weight]
			weight, _ := strconv.ParseFloat(parts[5], 64)
			relations = append(relations, types.Relation{
				SrcId:       parts[1],
				DstId:       parts[2],
				Type:        parts[3],
				Description: parts[4],
				Weight:      weight,
				SourceId:    []string{sourceId},
			})
		}
	}
	return entities, relations
}

func logCallback() callbacks.Handler {
	builder := callbacks.NewHandlerBuilder()
	builder.OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {

		logx.Infof("[graph extractor] - %s:%s:%s", info.Component, info.Type, info.Name)

		return ctx
	})

	return builder.Build()
}

// Merge merges multiple GraphExtractionResults into one
func (e *GraphExtractor) Merge(results []*types.GraphExtractionResult) *types.GraphExtractionResult {
	mergedEntities := make(map[string]types.Entity)
	mergedRelations := make(map[string]types.Relation)

	for _, res := range results {
		for _, entity := range res.Entities {
			if existing, ok := mergedEntities[entity.Name]; ok {
				// Merge logic
				existing.SourceId = append(existing.SourceId, entity.SourceId...)
				// Deduplicate SourceId
				existing.SourceId = uniqueStrings(existing.SourceId)
				// Merge Description (simple concatenation for now, or keep longest?)
				// Let's keep the longest description as it might be more detailed,
				// or concatenate if they are different.
				if len(entity.Description) > len(existing.Description) {
					existing.Description = entity.Description
				}
				mergedEntities[entity.Name] = existing
			} else {
				mergedEntities[entity.Name] = entity
			}
		}

		for _, relation := range res.Relations {
			// Key: SrcId + DstId
			key := fmt.Sprintf("%s->%s", relation.SrcId, relation.DstId)
			if existing, ok := mergedRelations[key]; ok {
				existing.SourceId = append(existing.SourceId, relation.SourceId...)
				existing.SourceId = uniqueStrings(existing.SourceId)
				// Average weight
				existing.Weight = (existing.Weight + relation.Weight) / 2
				// Merge Description
				if relation.Description != "" && !strings.Contains(existing.Description, relation.Description) {
					if existing.Description != "" {
						existing.Description += "\n" + relation.Description
					} else {
						existing.Description = relation.Description
					}
				}
				mergedRelations[key] = existing
			} else {
				mergedRelations[key] = relation
			}
		}
	}

	finalResult := &types.GraphExtractionResult{
		Entities:  make([]types.Entity, 0, len(mergedEntities)),
		Relations: make([]types.Relation, 0, len(mergedRelations)),
	}

	for _, v := range mergedEntities {
		finalResult.Entities = append(finalResult.Entities, v)
	}
	for _, v := range mergedRelations {
		finalResult.Relations = append(finalResult.Relations, v)
	}

	return finalResult
}

func uniqueStrings(input []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
