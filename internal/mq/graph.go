package mq

const (
	TopicGraphExtract = "prod.rag.knowledge.graph.extract"
)

type GraphGenerateMsg struct {
	DocumentId             string   `json:"document_id"`
	KnowledgeBaseId        string   `json:"knowledge_base_id"`
	TenantId               string   `json:"tenant_id"` // 租户id
	LlmId                  string   `json:"llm_id"`    // 格式: 模型名称@模型厂商
	EntityTypes            []string `json:"entity_types"`
	EnableEntityResolution bool     `json:"enable_entity_resolution"`
	EnableCommunity        bool     `json:"enable_community"`
}
