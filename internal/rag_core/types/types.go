package types

type ProcessRequest struct {
	URI         string        // 可以是文件路径，也可以是web url
	IndexConfig ProcessConfig // 索引相关的配置
}
type IndexConfigPreCleanRule struct {
	CleanWhitespace  bool
	RemoveUrlsEmails bool
}

type ProcessLlmConfig struct {
	QaKey       string
	QaBaseUrl   string
	QaModelName string

	EmbeddingKey       string
	EmbeddingBaseUrl   string
	EmbeddingModelName string

	ChatKey       string
	ChatBaseUrl   string
	ChatModelName string
}

type ProcessConfig struct {
	KnowledgeName string // 知识库名称
	EnableQACheck bool   // 是否启用 QA 检查
	QaNum         int

	Separators     []string
	ChunkOverlap   int
	MaxChunkLength int
	PreCleanRule   IndexConfigPreCleanRule

	LlmConfig ProcessLlmConfig
}
