package chunk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	DefaultIndexName = "rag_knowledge_chunks"
)

type EsChunkModel struct {
	client *elasticsearch.Client
	index  string
}

func NewEsChunkModel(addresses []string, username, password string) (*EsChunkModel, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	// Verify connection
	res, err := client.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, errors.New("elasticsearch connection failed: " + res.String())
	}

	return &EsChunkModel{
		client: client,
		index:  DefaultIndexName,
	}, nil
}

func (m *EsChunkModel) Put(ctx context.Context, chunks []*Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// Use Bulk API
	var buf bytes.Buffer
	for _, chunk := range chunks {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s", "_id" : "%s" } }%s`, m.index, chunk.Id, "\n"))
		data, err := json.Marshal(chunk)
		if err != nil {
			return err
		}
		data = append(data, "\n"...)

		buf.Write(meta)
		buf.Write(data)
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true", // Refresh immediately for consistency in this scenario, or use default
	}

	res, err := req.Do(ctx, m.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk indexing failed: %s", res.String())
	}

	return nil
}

func (m *EsChunkModel) HybridSearch(ctx context.Context, kbId string, query string, vector []float64, topK int) ([]*Chunk, error) {
	// Build ES Query
	// Must: kb_id filter
	// Should:
	//   1. Function Score with Vector cosine similarity
	//   2. Match query on content_with_weight (keyword search)
	// This is a simplified Hybrid Search combining filter + vector + keyword.
	// Note: Standard KNN search is preferred in newer ES versions for vectors, but here we use a flexible script_score or knn option.
	// Let's use standard KNN retrieval combined with lexical search if needed, or just standard bool query with vector field?
	// Given typical Hybrid RAG, we often use RRF or just weighted sum.
	// Here, let's implement a Bool query combining vector similarity (via script_score or proper knn section) and keyword match.
	// For simplicity and compatibility, we'll use Knn query with query filter.

	queryBody := map[string]interface{}{
		"knn": map[string]interface{}{
			"field":          "content_vector", // Ensure mapping matches this name
			"query_vector":   vector,
			"k":              topK,
			"num_candidates": topK * 10,
			"filter": map[string]interface{}{
				"term": map[string]interface{}{
					"kb_ids": kbId,
				},
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": map[string]interface{}{
					"term": map[string]interface{}{
						"kb_ids": kbId,
					},
				},
				"should": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"content": query, // Assumes content field is analyzed
						},
					},
				},
			},
		},
		"size": topK,
	}

	// If vector is empty, fallback to keyword search only
	if len(vector) == 0 {
		delete(queryBody, "knn")
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(queryBody); err != nil {
		return nil, err
	}

	res, err := m.client.Search(
		m.client.Search.WithContext(ctx),
		m.client.Search.WithIndex(m.index),
		m.client.Search.WithBody(&buf),
		m.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search failed: %s, body: %s", res.Status(), string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	chunks := make([]*Chunk, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		sourceBytes, _ := json.Marshal(source)
		var chunk Chunk
		if err := json.Unmarshal(sourceBytes, &chunk); err != nil {
			logx.Errorf("failed to unmarshal chunk: %v", err)
			continue
		}
		if score, ok := hit.(map[string]interface{})["_score"].(float64); ok {
			chunk.Score = score
		}
		chunks = append(chunks, &chunk)
	}

	return chunks, nil
}

func (m *EsChunkModel) DeleteByDocId(ctx context.Context, kbId string, docId string) error {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{"term": map[string]interface{}{"kb_ids": kbId}},
					{"term": map[string]interface{}{"doc_id": docId}},
				},
			},
		},
	}
	return m.deleteByQuery(ctx, query)
}

func (m *EsChunkModel) DeleteByKbId(ctx context.Context, kbId string) error {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{"term": map[string]interface{}{"kb_ids": kbId}},
				},
			},
		},
	}
	return m.deleteByQuery(ctx, query)
}

func (m *EsChunkModel) deleteByQuery(ctx context.Context, query map[string]interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return err
	}

	res, err := m.client.DeleteByQuery(
		[]string{m.index},
		&buf,
		m.client.DeleteByQuery.WithContext(ctx),
		m.client.DeleteByQuery.WithRefresh(true),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete by query failed: %s", res.String())
	}
	return nil
}
