package previewer

// PreviewConf 定义切分规则

type DocType = string

const (
	DocTypeMarkdown DocType = "markdown"
	DocTypeText     DocType = "text"
)

type PreviewConf struct {
	Separator      []string
	MaxChunkLength int
	ChunkOverlap   int
}
