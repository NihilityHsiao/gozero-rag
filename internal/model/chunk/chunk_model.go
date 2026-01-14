package chunk

import "context"

type Chunk struct {
	Id            string    `json:"id"`
	DocId         string    `json:"doc_id"`
	KbIds         []string  `json:"kb_ids"` // 支持多库, 实际存储时映射为 kb_id 数组
	Content       string    `json:"content"`
	ContentVector []float64 `json:"content_vector"` // 对应ES中的 q_{dim}_vec
	DocName       string    `json:"doc_name"`
	ImportantKw   []string  `json:"important_keywords"`
	QuestionKw    []string  `json:"question_keywords"`
	ImgId         string    `json:"img_id"`
	PageNum       []int     `json:"page_num_int"`
	CreateTime    float64   `json:"create_timestamp_flt"`
	Available     int       `json:"available_int"`
	Score         float64   `json:"score,omitempty"` // Search score
}

type ChunkModel interface {
	// Put 插入或更新分片
	Put(ctx context.Context, chunks []*Chunk) error

	// HybridSearch 混合检索
	// kbId: 知识库ID, 必须指定
	// query: 文本查询
	// vector: 向量查询
	// topK: 返回条数
	HybridSearch(ctx context.Context, kbId string, query string, vector []float64, topK int) ([]*Chunk, error)

	// DeleteByDocId 按文档ID删除 (用于删除文件)
	DeleteByDocId(ctx context.Context, kbId string, docId string) error

	// DeleteByKbId 按知识库ID删除 (用于删除知识库)
	DeleteByKbId(ctx context.Context, kbId string) error
}
