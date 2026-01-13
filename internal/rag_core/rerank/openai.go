package rerank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
	"gozero-rag/internal/slicex"
	"io"
	"net/http"
)

type OpenAiReranker struct {
}

func NewOpenAiReranker() (*OpenAiReranker, error) {
	return &OpenAiReranker{}, nil
}

type reqData struct {
	Model           string   `json:"model"`
	Query           string   `json:"query"`
	Documents       []string `json:"documents"`
	Instruction     string   `json:"instruction"`
	TopN            int      `json:"top_n"`
	ReturnDocuments bool     `json:"return_documents"`
	MaxChunksPerDoc int      `json:"max_chunks_per_doc"`
	OverlapTokens   int      `json:"overlap_tokens"`
}

type Result struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}
type RespData struct {
	Id      string    `json:"id"`
	Results []*Result `json:"results"`
}

func (r *OpenAiReranker) intoRequestData(req *RerankRequest) *reqData {

	docs := slicex.Into(req.Docs, func(t *schema.Document) string {
		return t.Content
	})

	data := &reqData{
		Model:           req.ModelName,
		Query:           req.Query,
		Documents:       docs,
		Instruction:     "根据query对documents进行重排",
		TopN:            req.TopK,
		ReturnDocuments: false,
	}

	return data
}

func (r *OpenAiReranker) Rerank(ctx context.Context, input *RerankRequest) ([]*schema.Document, error) {
	req := r.intoRequestData(input)
	body, err := json.Marshal(req)
	if err != nil {
		logx.Errorf("rerank json err:%v, input:%+v", err, input)
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, input.BaseUrl, bytes.NewReader(body))
	if err != nil {
		logx.Errorf("rerank http request err:%v, input:%+v", err, input)
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", input.ApiKey))
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		logx.Errorf("rerank http request err:%v, input:%+v", err, input)
		return nil, err
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		logx.Errorf("rerank http request err:%v, input:%+v", err, input)
		return nil, err
	}

	var resp RespData
	err = json.Unmarshal(content, &resp)
	if err != nil {
		logx.Errorf("rerank http unmarshal err:%v, content:%+v", err, content)
		return nil, err
	}

	// 组装数据

	output := make([]*schema.Document, 0, len(resp.Results))

	for _, res := range resp.Results {
		doc := input.Docs[res.Index]
		doc.WithScore(res.RelevanceScore)
		output = append(output, doc)

	}

	return output, nil
}
