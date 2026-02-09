package mq

const (
	TopicDocumentIndex = "prod.rag.knowledge.document.index"
)

type KnowledgeDocumentIndexMsg struct {
	UserId          string `json:"user_id"`           // 用户ID (UUID)
	TenantId        string `json:"tenant_id"`         // 租户id
	KnowledgeBaseId string `json:"knowledge_base_id"` // knowledge_base.id
	DocumentId      string `json:"document_id"`       // knowledge_document.id

	// 本地消息表补偿字段
	LocalMessageId uint64 `json:"local_message_id,omitempty"` // 补偿投递时设置
}

// GetLocalMessageId 实现 RetryableMsg 接口
func (m *KnowledgeDocumentIndexMsg) GetLocalMessageId() uint64 {
	return m.LocalMessageId
}

// SetLocalMessageId 设置本地消息 ID
func (m *KnowledgeDocumentIndexMsg) SetLocalMessageId(id uint64) {
	m.LocalMessageId = id
}
