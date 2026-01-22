package parser

type ParserConfigGeneral struct {
	ChunkTokenNum        int      `json:"chunk_token_num"`
	ChunkOverlapTokenNum int      `json:"chunk_overlap_token_num"`
	Separator            []string `json:"separator"`
	LayoutRecognize      bool     `json:"layout_recognize"`
	QaNum                int      `json:"qa_num"`
	QaLlmId              string   `json:"qa_llm_id"`
	PdfParser            string   `json:"pdf_parser"`
}
