package chat_conversation

import (
	"encoding/json"
)

// ConversationConfig 对应 chat_conversation 表中的 config 字段 (JSON)
type ConversationConfig struct {
	LlmId                   string          `json:"llm_id"`                     // 模型名称@模型厂商 (或者 ID)
	EnableQuoteDoc          bool            `json:"enable_quote_doc"`           // 是否显示原文出处
	EnableLlmKeywordExtract bool            `json:"enable_llm_keyword_extract"` // 是否启用 LLM 关键词提取
	EnableTts               bool            `json:"enable_tts"`                 // 是否启用 TTS
	SystemPrompt            string          `json:"system_prompt"`              // 系统提示词
	KbIds                   []string        `json:"kb_ids"`                     // 知识库 ID 列表
	RetrievalConfig         RetrievalConfig `json:"retrieval_config"`           // 检索配置
	Temperature             float64         `json:"temperature"`                // 随机性
}

// RetrievalConfig 检索详细配置
type RetrievalConfig struct {
	Mode               string  `json:"mode"`                 // 检索模式: vector, fulltext, hybrid
	RerankMode         string  `json:"rerank_mode"`          // 重排序模式: weighted, rerank
	RerankVectorWeight float64 `json:"rerank_vector_weight"` // 向量权重 (0.0 - 1.0)
	TopN               int     `json:"top_n"`                // 最终传给 LLM 的切片数量
	RerankId           string  `json:"rerank_id"`            // 重排序模型 ID 或名称
	TopK               int     `json:"top_k"`                // 传给 Rerank 模型的候选数量
}

func (m *ChatConversation) GetConfig() (*ConversationConfig, error) {
	if len(m.Config) == 0 {
		return &ConversationConfig{}, nil
	}
	var cfg ConversationConfig
	err := json.Unmarshal([]byte(m.Config), &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
