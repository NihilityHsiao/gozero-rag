package retriever

import (
	"github.com/cloudwego/eino/schema"
)

// DocMeta 提取自 Document.MetaData 的强类型结构体
type DocMeta struct {
	ChunkID         string
	DocID           string
	KnowledgeBaseID int64
	Type            string
	Score           float64
	Source          string
}

const (
	MetaChunkID         = "chunk_id"
	MetaDocID           = "doc_id"
	MetaKnowledgeBaseID = "knowledge_base_id"
	MetaType            = "type"
	MetaScore           = "score"
	MetaSource          = "source"
)

// ExtractDocMeta 从 Document 中提取元数据
// 如果 MetaData 为空或字段缺失，将返回零值
func ExtractDocMeta(doc *schema.Document) DocMeta {
	if doc == nil || doc.MetaData == nil {
		return DocMeta{}
	}

	meta := DocMeta{}

	if v, ok := doc.MetaData[MetaChunkID].(string); ok {
		meta.ChunkID = v
	}
	if v, ok := doc.MetaData[MetaDocID].(string); ok {
		meta.DocID = v
	}
	if v, ok := doc.MetaData[MetaType].(string); ok {
		meta.Type = v
	}
	if v, ok := doc.MetaData[MetaSource].(string); ok {
		meta.Source = v
	}

	// Handle numeric types which might be float64 (from JSON) or int/int64
	if v, ok := doc.MetaData[MetaKnowledgeBaseID]; ok {
		switch val := v.(type) {
		case int64:
			meta.KnowledgeBaseID = val
		case float64:
			meta.KnowledgeBaseID = int64(val)
		case int:
			meta.KnowledgeBaseID = int64(val)
		}
	}

	if v, ok := doc.MetaData[MetaScore]; ok {
		switch val := v.(type) {
		case float64:
			meta.Score = val
		case float32:
			meta.Score = float64(val)
		}
	}

	return meta
}
