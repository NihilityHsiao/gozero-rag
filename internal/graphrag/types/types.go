package types

type GraphExtractionResult struct {
	Entities  []Entity
	Relations []Relation
}

type Entity struct {
	Name        string   `json:"name"`        // 实体名称
	Type        string   `json:"type"`        // 实体的分类标签, 如 organization | person | geo | event
	Description string   `json:"description"` // llm生成的关于该实体的综合描述
	SourceId    []string `json:"source_id"`   // 关键的溯源信息(chunk id), 记录这个实体在哪些chunk中出现
}

type Relation struct {
	SrcId       string   `json:"src_id"`      // 源实体id
	DstId       string   `json:"dst_id"`      // 目标实体id
	Description string   `json:"description"` // 关系的具体描述,LLM生成的,解释为什么这2个实体相关
	Weight      float64  `json:"weight"`      // 关系的权重(1-10), 表示这两个连接有多紧密
	SourceId    []string `json:"source_id"`   // 关键的溯源信息(chunk id), 表示这个关系是在哪些文本片段中被发现的
}
