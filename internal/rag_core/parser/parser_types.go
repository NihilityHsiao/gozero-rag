package parser

type ParserConfigGeneral struct {
	ChunkTokenNum        int             `json:"chunk_token_num"`
	ChunkOverlapTokenNum int             `json:"chunk_overlap_token_num"`
	Separator            []string        `json:"separator"`
	LayoutRecognize      bool            `json:"layout_recognize"`
	QaNum                int             `json:"qa_num"`
	QaLlmId              string          `json:"qa_llm_id"`
	PdfParser            string          `json:"pdf_parser"`
	GraphRag             *GraphRagConfig `json:"graph_rag,omitempty"` // 可选，兼容旧数据
}

// GraphRagConfig 知识图谱配置
type GraphRagConfig struct {
	EnableGraph            bool     `json:"enable_graph"`
	GraphLlmId             string   `json:"graph_llm_id,omitempty"`             // 格式: model@factory
	EntityTypes            []string `json:"entity_types,omitempty"`             // 实体类型列表
	EnableEntityResolution bool     `json:"enable_entity_resolution,omitempty"` // 实体归一化
	EnableCommunity        bool     `json:"enable_community,omitempty"`         // 社区报告
}
