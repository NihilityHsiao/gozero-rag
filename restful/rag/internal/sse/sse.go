package sse

import "gozero-rag/restful/rag/internal/types"

type SseType = string

const (
	SSETypeText      = "text"
	SSETypeReasoning = "reasoning"
	SSETypeToolUse   = "tool_use"
	SSETypeCitation  = "citation"
	SSETypeFinish    = "finish"
	SSETypeError     = "error"
)

type SseClient struct {
	msgId  string
	client chan<- *types.ChatResp
}

func NewSSEClient(msgId string, client chan<- *types.ChatResp) *SseClient {
	return &SseClient{
		msgId:  msgId,
		client: client,
	}
}

func (s *SseClient) SendText(text string) {
	s.client <- &types.ChatResp{
		Type:    SSETypeText,
		MsgId:   s.msgId,
		Content: text,
	}
}

func (s *SseClient) SendReasoning(content string) {
	s.client <- &types.ChatResp{
		Type:             SSETypeReasoning,
		MsgId:            s.msgId,
		ReasoningContent: content,
	}
}

func (s *SseClient) SendFinish() {
	s.client <- &types.ChatResp{
		Type:  SSETypeFinish,
		MsgId: s.msgId,
	}
}

func (s *SseClient) SendError(errMsg string) {
	s.client <- &types.ChatResp{
		MsgId:    s.msgId,
		Type:     SSETypeError,
		ErrorMsg: errMsg,
	}
}

func (s *SseClient) SendCitation(chunks []types.ChatRetrievalChunk) {
	s.client <- &types.ChatResp{
		Type:          SSETypeCitation,
		MsgId:         s.msgId,
		RetrievalDocs: chunks,
	}
}
