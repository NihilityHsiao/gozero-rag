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

	// 本地消息表补偿字段
	LocalMessageId uint64 `json:"local_message_id,omitempty"` // 补偿投递时设置
}

// GetLocalMessageId 实现 RetryableMsg 接口
func (m *GraphGenerateMsg) GetLocalMessageId() uint64 {
	return m.LocalMessageId
}

// SetLocalMessageId 设置本地消息 ID
func (m *GraphGenerateMsg) SetLocalMessageId(id uint64) {
	m.LocalMessageId = id
}
