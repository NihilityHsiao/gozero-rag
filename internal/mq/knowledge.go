package mq

type KnowledgeDocumentIndexMsg struct {
	UserId          int64  `json:"user_id"`           // 根据user_id获取它的默认模型信息
	KnowledgeBaseId uint64 `json:"knowledge_base_id"` // knowledge_base.id
	DocumentId      string `json:"document_id"`       // knowledge_document.id
}
