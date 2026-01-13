package vectorstore

import (
	"context"
	"fmt"

	"gozero-rag/internal/config"

	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/zeromicro/go-zero/core/logx"
)

// MilvusClient Milvus 向量数据库客户端 (新 SDK)
type MilvusClient struct {
	client *milvusclient.Client
	cfg    config.VectorStoreConf
}

// NewMilvusClient 创建 Milvus 客户端
func NewMilvusClient(cfg config.VectorStoreConf) (*MilvusClient, error) {
	ctx := context.Background()

	clientCfg := &milvusclient.ClientConfig{
		Address:  cfg.Endpoint,
		Username: cfg.Username,
		Password: cfg.Password,
	}

	c, err := milvusclient.New(ctx, clientCfg)
	if err != nil {
		return nil, fmt.Errorf("连接 Milvus 失败: %w", err)
	}

	logx.Infof("[VectorStore] Milvus 连接成功: %s", cfg.Endpoint)

	return &MilvusClient{
		client: c,
		cfg:    cfg,
	}, nil
}

// EnsureCollection 确保 Collection 存在，支持 BM25 全文检索
func (m *MilvusClient) EnsureCollection(ctx context.Context, collectionName string, dim int) error {
	has, err := m.client.HasCollection(ctx, milvusclient.NewHasCollectionOption(collectionName))
	if err != nil {
		return fmt.Errorf("检查 Collection 失败: %w", err)
	}
	if has {
		return nil
	}

	// 创建 Schema，包含 BM25 Function
	schema := entity.NewSchema().
		WithName(collectionName).
		WithField(entity.NewField().
			WithName("id").
			WithDataType(entity.FieldTypeVarChar).
			WithIsPrimaryKey(true).
			WithMaxLength(128)).
		WithField(entity.NewField().
			WithName("kb_id").
			WithDataType(entity.FieldTypeInt64)).
		WithField(entity.NewField().
			WithName("chunk_id").
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(64)).
		WithField(entity.NewField().
			WithName("doc_id").
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(64)).
		WithField(entity.NewField().
			WithName("type").
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(16)).
		WithField(entity.NewField().
			WithName("content").
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(65535).
			WithEnableAnalyzer(true)). // 启用分词器，用于 BM25
		WithField(entity.NewField().
			WithName("vector").
			WithDataType(entity.FieldTypeFloatVector).
			WithDim(int64(dim))).
		WithField(entity.NewField().
			WithName("sparse_vector").
			WithDataType(entity.FieldTypeSparseVector)).
		// BM25 函数：自动将 content 转换为 sparse_vector
		WithFunction(entity.NewFunction().
			WithName("bm25_func").
			WithType(entity.FunctionTypeBM25).
			WithInputFields("content").
			WithOutputFields("sparse_vector"))

	// 创建 Collection
	err = m.client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(collectionName, schema))
	if err != nil {
		return fmt.Errorf("创建 Collection 失败: %w", err)
	}

	// 创建向量索引 (Dense Vector)
	vectorIndex := index.NewAutoIndex(entity.COSINE)
	vectorIndexTask, err := m.client.CreateIndex(ctx, milvusclient.NewCreateIndexOption(collectionName, "vector", vectorIndex))
	if err != nil {
		return fmt.Errorf("创建向量索引失败: %w", err)
	}
	err = vectorIndexTask.Await(ctx)
	if err != nil {
		return fmt.Errorf("等待向量索引创建失败: %w", err)
	}

	// 创建稀疏向量索引 (Sparse Vector for BM25)
	sparseIndex := index.NewAutoIndex(entity.BM25)
	sparseIndexTask, err := m.client.CreateIndex(ctx, milvusclient.NewCreateIndexOption(collectionName, "sparse_vector", sparseIndex))
	if err != nil {
		return fmt.Errorf("创建稀疏向量索引失败: %w", err)
	}
	err = sparseIndexTask.Await(ctx)
	if err != nil {
		return fmt.Errorf("等待稀疏向量索引创建失败: %w", err)
	}

	// 为 doc_id 创建标量索引
	docIdIndex := index.NewAutoIndex(entity.L2)
	_, err = m.client.CreateIndex(ctx, milvusclient.NewCreateIndexOption(collectionName, "doc_id", docIdIndex))
	if err != nil {
		logx.Errorf("创建 doc_id 索引失败 (非致命): %v", err)
	}

	// 加载 Collection 到内存
	loadTask, err := m.client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collectionName))
	if err != nil {
		return fmt.Errorf("加载 Collection 失败: %w", err)
	}
	err = loadTask.Await(ctx)
	if err != nil {
		return fmt.Errorf("等待 Collection 加载失败: %w", err)
	}

	logx.Infof("[VectorStore] Collection 创建成功: %s (含 BM25 支持)", collectionName)
	return nil
}

// Insert 插入文档（需要外部先 embedding）
// 此方法预留给 eino indexer 使用，暂不实现
func (m *MilvusClient) Insert(ctx context.Context, collectionName string, docs []*schema.Document) error {
	return fmt.Errorf("请使用 InsertWithVectors 方法")
}

// InsertWithVectors 插入已有向量的记录
// 注意：content 字段会由 Milvus BM25 Function 自动生成 sparse_vector
func (m *MilvusClient) InsertWithVectors(ctx context.Context, collectionName string, records []*VectorRecord) error {
	if len(records) == 0 {
		return nil
	}

	// 准备列数据
	ids := make([]string, len(records))
	kbIds := make([]int64, len(records))
	chunkIds := make([]string, len(records))
	docIds := make([]string, len(records))
	types := make([]string, len(records))
	contents := make([]string, len(records))
	vectors := make([][]float32, len(records))

	for i, r := range records {
		ids[i] = r.ID
		kbIds[i] = int64(r.KnowledgeBaseID)
		chunkIds[i] = r.ChunkID
		docIds[i] = r.DocID
		types[i] = r.Type
		contents[i] = r.Content
		// float64 -> float32 转换
		vectors[i] = make([]float32, len(r.Vector))
		for j, v := range r.Vector {
			vectors[i][j] = float32(v)
		}
	}

	// 构建 columns
	// 注意：不需要传递 sparse_vector，它由 BM25 Function 自动生成
	columns := []column.Column{
		column.NewColumnVarChar("id", ids),
		column.NewColumnInt64("kb_id", kbIds),
		column.NewColumnVarChar("chunk_id", chunkIds),
		column.NewColumnVarChar("doc_id", docIds),
		column.NewColumnVarChar("type", types),
		column.NewColumnVarChar("content", contents),
		column.NewColumnFloatVector("vector", int(len(records[0].Vector)), vectors),
	}

	// 插入数据
	_, err := m.client.Insert(ctx, milvusclient.NewColumnBasedInsertOption(collectionName, columns...))
	if err != nil {
		return fmt.Errorf("插入 Milvus 失败: %w", err)
	}

	// 刷新数据
	// 不需要每次插入都 Flush，Milvus 会自动 Flush
	// 频繁 Flush 会导致 rate limit exceeded
	// _, err = m.client.Flush(ctx, milvusclient.NewFlushOption(collectionName))
	// if err != nil {
	// 	logx.Errorf("Flush 失败: %v", err)
	// }

	return nil
}

// Search 向量搜索 (Dense Vector)
func (m *MilvusClient) Search(ctx context.Context, collectionName string, queryVector []float64, topK int) ([]*SearchResult, error) {
	// float64 -> float32 转换
	queryVec32 := make([]float32, len(queryVector))
	for i, v := range queryVector {
		queryVec32[i] = float32(v)
	}

	// 构建搜索请求
	searchOpt := milvusclient.NewSearchOption(collectionName, topK, []entity.Vector{entity.FloatVector(queryVec32)}).
		WithANNSField("vector").
		WithOutputFields("id", "kb_id", "chunk_id", "doc_id", "type", "content")

	results, err := m.client.Search(ctx, searchOpt)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	return m.parseSearchResults(results)
}

// FullTextSearch 全文检索 (BM25 Sparse Vector)
func (m *MilvusClient) FullTextSearch(ctx context.Context, collectionName string, query string, topK int) ([]*SearchResult, error) {
	// 使用文本查询，Milvus 会自动转换为 sparse vector
	searchOpt := milvusclient.NewSearchOption(collectionName, topK, []entity.Vector{entity.Text(query)}).
		WithANNSField("sparse_vector").
		WithOutputFields("id", "kb_id", "chunk_id", "doc_id", "type", "content")

	results, err := m.client.Search(ctx, searchOpt)
	if err != nil {
		return nil, fmt.Errorf("全文搜索失败: %w", err)
	}

	return m.parseSearchResults(results)
}

// HybridSearch 混合检索 (Dense + Sparse BM25)
func (m *MilvusClient) HybridSearch(ctx context.Context, collectionName string, queryVector []float64, queryText string, topK int) ([]*SearchResult, error) {
	// float64 -> float32 转换
	queryVec32 := make([]float32, len(queryVector))
	for i, v := range queryVector {
		queryVec32[i] = float32(v)
	}

	// Dense Vector 搜索请求
	denseRequest := milvusclient.NewAnnRequest("vector", topK, entity.FloatVector(queryVec32))

	// Sparse Vector (BM25) 搜索请求
	sparseRequest := milvusclient.NewAnnRequest("sparse_vector", topK, entity.Text(queryText))

	// 混合搜索，使用 RRF 融合
	results, err := m.client.HybridSearch(ctx,
		milvusclient.NewHybridSearchOption(collectionName, topK, denseRequest, sparseRequest).
			WithReranker(milvusclient.NewRRFReranker()).
			WithOutputFields("id", "kb_id", "chunk_id", "doc_id", "type", "content"))
	if err != nil {
		return nil, fmt.Errorf("混合搜索失败: %w", err)
	}

	return m.parseSearchResults(results)
}

// parseSearchResults 解析搜索结果
func (m *MilvusClient) parseSearchResults(results []milvusclient.ResultSet) ([]*SearchResult, error) {
	var searchResults []*SearchResult

	for _, resultSet := range results {
		for i := 0; i < resultSet.ResultCount; i++ {
			var id, chunkId, docId, typ, content string
			var kbId int64

			if col := resultSet.GetColumn("id"); col != nil {
				id, _ = col.GetAsString(i)
			}
			if col := resultSet.GetColumn("kb_id"); col != nil {
				kbId, _ = col.GetAsInt64(i)
			}
			if col := resultSet.GetColumn("chunk_id"); col != nil {
				chunkId, _ = col.GetAsString(i)
			}
			if col := resultSet.GetColumn("doc_id"); col != nil {
				docId, _ = col.GetAsString(i)
			}
			if col := resultSet.GetColumn("type"); col != nil {
				typ, _ = col.GetAsString(i)
			}
			if col := resultSet.GetColumn("content"); col != nil {
				content, _ = col.GetAsString(i)
			}

			searchResults = append(searchResults, &SearchResult{
				ID:              id,
				KnowledgeBaseID: uint64(kbId),
				ChunkID:         chunkId,
				DocID:           docId,
				Type:            typ,
				Content:         content,
				Score:           resultSet.Scores[i],
			})
		}
	}

	return searchResults, nil
}

// Delete 删除记录
func (m *MilvusClient) Delete(ctx context.Context, collectionName string, expr string) error {
	_, err := m.client.Delete(ctx, milvusclient.NewDeleteOption(collectionName).WithExpr(expr))
	if err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}
	return nil
}

// DropCollection 删除 Collection
func (m *MilvusClient) DropCollection(ctx context.Context, collectionName string) error {
	return m.client.DropCollection(ctx, milvusclient.NewDropCollectionOption(collectionName))
}

// Close 关闭连接
func (m *MilvusClient) Close() error {
	return m.client.Close(context.Background())
}
