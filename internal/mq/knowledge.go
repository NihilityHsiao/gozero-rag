package mq

type KnowledgeDocumentIndexMsg struct {
	UserId          string `json:"user_id"`           // 用户ID (UUID)
	TenantId        string `json:"tenant_id"`         // 租户id
	KnowledgeBaseId string `json:"knowledge_base_id"` // knowledge_base.id
	DocumentId      string `json:"document_id"`       // knowledge_document.id
}
