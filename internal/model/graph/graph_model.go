package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

const (
	DefaultGraphIndexName = "kg_graph"
)

type EsGraphDocument struct {
	Id        string `json:"id"`
	KbId      string `json:"kb_id"`
	GraphType string `json:"graph_type"` // entity | relation

	// 标识字段
	EntityName string `json:"entity_name,omitempty"` // for entity
	SrcName    string `json:"src_name,omitempty"`    // for relation
	DstName    string `json:"dst_name,omitempty"`    // for relation

	// 这里为了支持 Script Upsert，将 ContentWithWeight 中的关键合并字段提取出来
	Description string   `json:"description"`
	Weight      float64  `json:"weight"`
	SourceIds   []string `json:"source_ids"`

	ContentWithWeight string `json:"content_with_weight"` // 原始 JSON 备份
	UpdatedAt         string `json:"updated_at"`
}

type GraphModel interface {
	Put(ctx context.Context, docs []*EsGraphDocument) error
	// TODO: Add search methods if needed
}

type EsGraphModel struct {
	client *elasticsearch.Client
	index  string
}

func NewEsGraphModel(addresses []string, username, password string) (*EsGraphModel, error) {
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

	model := &EsGraphModel{
		client: client,
		index:  DefaultGraphIndexName,
	}

	if err := model.SetupIndex(context.Background()); err != nil {
		return nil, err
	}

	return model, nil
}

func (m *EsGraphModel) SetupIndex(ctx context.Context) error {
	// Check if index exists
	res, err := m.client.Indices.Exists([]string{m.index}, m.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}

	// Define Mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"kb_id": map[string]interface{}{
					"type": "keyword",
				},
				"graph_type": map[string]interface{}{
					"type": "keyword",
				},
				"entity_name": map[string]interface{}{
					"type": "keyword",
					"fields": map[string]interface{}{
						"text": map[string]interface{}{
							"type":     "text",
							"analyzer": "ik_max_word",
						},
					},
				},
				"src_name": map[string]interface{}{
					"type": "keyword",
				},
				"dst_name": map[string]interface{}{
					"type": "keyword",
				},
				"description": map[string]interface{}{
					"type":     "text",
					"analyzer": "ik_max_word",
				},
				"weight": map[string]interface{}{
					"type": "float",
				},
				"source_ids": map[string]interface{}{
					"type": "keyword",
				},
				"content_with_weight": map[string]interface{}{
					"type":  "text",
					"index": false,
				},
				"updated_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return err
	}

	resp, createErr := m.client.Indices.Create(
		m.index,
		m.client.Indices.Create.WithContext(ctx),
		m.client.Indices.Create.WithBody(&buf),
	)
	if createErr != nil {
		return createErr
	}
	defer resp.Body.Close()

	if resp.IsError() {
		// Ignore if it already exists (race condition)
		if resp.StatusCode == 400 {
			return nil
		}
		return fmt.Errorf("create index failed: %s", resp.String())
	}
	return nil
}

func (m *EsGraphModel) Put(ctx context.Context, docs []*EsGraphDocument) error {
	if len(docs) == 0 {
		return nil
	}

	// Use Bulk API with Script Upsert
	var buf bytes.Buffer
	for _, doc := range docs {
		// Meta line
		// Use doc.Id as _id
		meta := []byte(fmt.Sprintf(`{ "update" : { "_index" : "%s", "_id" : "%s" } }%s`, m.index, doc.Id, "\n"))

		// Script for merging
		script := map[string]interface{}{
			"source": `
// 1. Merge Source IDs
if (params.new_source_ids != null) {
	if (ctx._source.source_ids == null) {
		ctx._source.source_ids = params.new_source_ids;
	} else {
		for (item in params.new_source_ids) {
			if (!ctx._source.source_ids.contains(item)) {
				ctx._source.source_ids.add(item);
			}
		}
	}
}

// 2. Update Weight
if (ctx._source.graph_type == 'relation') {
	ctx._source.weight = (ctx._source.weight + params.new_weight) / 2.0;
} 
// For Entity, we keep existing weight or maybe max? Let's just keep existing logic (no change or average if new measurement)
// If we want Entity weight to reflect something, we can add logic here. Currently keep average for simplicity if strictly needed, 
// but usually Entity weight is frequency or relevance. Let's assume average for now to be consistent.
// Or if Entity has fixed weight 1.0, average remains 1.0.

// 3. Merge Description
if (params.new_description != null && !ctx._source.description.contains(params.new_description)) {
	ctx._source.description = ctx._source.description + "\n" + params.new_description;
}
ctx._source.updated_at = params.now;
`,
			"lang": "painless",
			"params": map[string]interface{}{
				"new_source_ids":  doc.SourceIds,
				"new_weight":      doc.Weight,
				"new_description": doc.Description,
				"now":             doc.UpdatedAt,
			},
		}

		// Update payload
		payload := map[string]interface{}{
			"script": script,
			"upsert": doc,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		payloadBytes = append(payloadBytes, "\n"...)

		buf.Write(meta)
		buf.Write(payloadBytes)
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, m.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		// Try to read body for error details
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk indexing failed: %s, body: %s", res.Status(), string(body))
	}

	return nil
}
