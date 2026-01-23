package extractor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
