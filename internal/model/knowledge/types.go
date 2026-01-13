package knowledge

const (
	StatusDocumentDisable  = "disable"
	StatusDocumentPending  = "pending"
	StatusDocumentEnable   = "enable"
	StatusDocumentFail     = "fail"
	StatusDocumentIndexing = "indexing"
)

type QaPair struct {
	Answer   string `json:"answer,omitempty"`
	Question string `json:"question,omitempty"`
}

type QualityDetails struct {
	Issues         []interface{} `json:"issues,omitempty"`
	QaScore        float64       `json:"qa_score,omitempty"`
	Suggestions    []interface{} `json:"suggestions,omitempty"`
	TotalScore     float64       `json:"total_score,omitempty"`
	LengthScore    int           `json:"length_score,omitempty"`
	ContentScore   float64       `json:"content_score,omitempty"`
	SemanticScore  int           `json:"semantic_score,omitempty"`
	StructureScore int           `json:"structure_score,omitempty"`
}

type ChunkMetadata struct {
	H1             string          `json:"h1,omitempty"`
	H2             string          `json:"h2,omitempty"`
	H3             string          `json:"h3,omitempty"`
	Source         string          `json:"_source,omitempty"`
	QaPairs        []QaPair        `json:"qa_pairs,omitempty"`
	Extension      string          `json:"_extension,omitempty"`
	FileName       string          `json:"_file_name,omitempty"`
	QualityScore   float64         `json:"quality_score,omitempty"`
	QualityDetails *QualityDetails `json:"quality_details,omitempty"`
}
