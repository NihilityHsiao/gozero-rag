package extractor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/callbacks"
	"github.com/zeromicro/go-zero/core/logx"

	"gozero-rag/internal/graphrag/prompt"
	"gozero-rag/internal/graphrag/types"
	"gozero-rag/internal/model/chunk"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type GraphExtractor struct {
	llm      model.ToolCallingChatModel
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
}

const MaxLoops = 3

func NewGraphExtractor(ctx context.Context, llm model.ToolCallingChatModel) (*GraphExtractor, error) {
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
		resp, err := llm.Generate(ctx, state.CurrentPrompt)
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

	return &GraphExtractor{llm: llm, runnable: r}, nil
}

func (e *GraphExtractor) Extract(ctx context.Context, chunks []*chunk.Chunk) (*types.GraphExtractionResult, error) {
	runnable := e.runnable

	result := &types.GraphExtractionResult{
		Entities:  make([]types.Entity, 0),
		Relations: make([]types.Relation, 0),
	}

	// Process each chunk
	for _, c := range chunks {
		state := &GraphExtractionState{
			InputText: c.Content,
			History:   "",
			LoopCount: 0,
			IsDone:    false,
		}

		for state.LoopCount < MaxLoops && !state.IsDone {
			// Run the graph iteration
			newState, err := runnable.Invoke(ctx, state)
			if err != nil {
				fmt.Printf("Error extracting from chunk %v: %v\n", c.Id, err)
				break
			}
			state = newState
		}

		// Parse the accumulated history (which contains all results)
		entities, relations := e.parseHistory(state.History, c.Id)
		result.Entities = append(result.Entities, entities...)
		result.Relations = append(result.Relations, relations...)
	}

	result = e.Merge([]*types.GraphExtractionResult{result})

	return result, nil
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
		} else if typeTag == "relationship" && len(parts) >= 5 {
			// ("relationship"<|>"Src"<|>"Tgt"<|>"Desc"<|>Weight)
			// parts: [relationship, Src, Tgt, Desc, Weight]
			weight, _ := strconv.ParseFloat(parts[4], 64)
			relations = append(relations, types.Relation{
				SrcId:       parts[1],
				DstId:       parts[2],
				Description: parts[3],
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
