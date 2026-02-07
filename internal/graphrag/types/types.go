package types

type GraphExtractionResult struct {
	Entities  []Entity
	Relations []Relation
}

type Entity struct {
	Name        string `json:"name"`        // 实体名称
	Type        string `json:"type"`        // 实体的分类标签, 如 organization | person | geo | event
	Description string `json:"description"` // llm生成的关于该实体的综合描述
	// 关键的溯源信息, 记录这个实体在哪些chunk中出现
	// chunk id的格式:
	// 可能为 chunk-xxhash(chunk内容-chunk对应的文档的uuidv7)
	// 也可能为 qa-xxhash(chunk内容-chunk对应的文档的uuidv7)
	SourceId []string `json:"source_id"`
}

type Relation struct {
	SrcId       string  `json:"src_id"`      // 源实体id
	DstId       string  `json:"dst_id"`      // 目标实体id
	Type        string  `json:"type"`        // 关系类型, 如 owns | works_for | invested_in | collaborates_with | belongs_to
	Description string  `json:"description"` // 关系的具体描述,LLM生成的,解释为什么这2个实体相关
	Weight      float64 `json:"weight"`      // 关系的权重(1-10), 表示这两个连接有多紧密
	// 关键的溯源信息(chunk id), 表示这个关系是在哪些文本片段中被发现的
	// chunk id的格式:
	// 可能为 chunk-xxhash(chunk内容-chunk对应的文档的uuidv7)
	// 也可能为 qa-xxhash(chunk内容-chunk对应的文档的uuidv7)
	SourceId []string `json:"source_id"`
}
