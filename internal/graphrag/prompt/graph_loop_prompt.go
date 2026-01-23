package prompt

import (
	"context"
	"strings"

	einoprompt "github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func NewGraphLoopPrompt(history string) ([]*schema.Message, error) {
	variables := map[string]any{
		"entity_types":         strings.Join([]string{"organization", "person", "geo", "event", "category"}, ","),
		"tuple_delimiter":      "<|>",
		"record_delimiter":     "##",
		"completion_delimiter": "<|DONE|>",
		"history":              history,
	}

	result, err := graphLoopTemplate.Format(context.Background(), variables)
	return result, err
}

var graphLoopTemplate = einoprompt.FromMessages(schema.FString, schema.UserMessage(`
You just identified the following entities and relationships:
{history}

Are there any other entities or relationships that you missed? 
- If yes, please list them in the same format.
- If no, please output {completion_delimiter}
`))
