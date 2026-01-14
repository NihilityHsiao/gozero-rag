package parser

type ParserConfigGeneral struct {
	ChunkTokenNum        int      `json:"chunk_token_num"`
	ChunkOverlapTokenNum int      `json:"chunk_overlap_token_num"`
	Separator            []string `json:"separator"`
	LayoutRecognize      bool     `json:"layout_recognize"`
	PdfParser            string   `json:"pdf_parser"`
}
