package retriever

type RetrieveMode = string

const (
	RetrieveModeVector   RetrieveMode = "vector"
	RetrieveModeFulltext RetrieveMode = "fulltext"
	RetrieveModeHybrid   RetrieveMode = "hybrid"

	HybridRankTypeWeighted = "weighted"
	HybridRankTypeRerank   = "rerank"
)
