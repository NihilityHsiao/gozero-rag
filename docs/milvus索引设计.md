# Milvus 索引设计文档

## 1. 核心设计理念

### 1.1 一对多映射策略

MySQL 中一条 `knowledge_document_chunk` 记录可能对应 Milvus 中 **多条** 向量记录：

```
MySQL Chunk (1条)
├── Milvus Record 1: 原始 chunk 内容向量
├── Milvus Record 2: QA pair[0].question 向量
├── Milvus Record 3: QA pair[1].question 向量
└── ...
```

**为什么这样设计？**
- 用户搜索可能是问句形式，QA question 向量与搜索 query 语义更近
- 多向量提升召回率，原始内容 + 问题都能命中
- 检索到任意向量后，通过 `chunk_id` 可找回完整内容

### 1.2 向量类型

| type    | 说明         | content 字段  |
| ------- | ------------ | ------------- |
| `chunk` | 原始切片内容 | 切片原文      |
| `qa`    | QA 问题向量  | question 文本 |

---

## 2. Milvus Collection Schema

```go
// Collection: kb_{knowledge_base_id}
// 每个知识库一个 Collection

type MilvusRecord struct {
    ID          string   // 主键，格式: {chunk_id}_{type}_{index}
    ChunkID     string   // 关联 MySQL chunk.id
    DocID       string   // 关联 MySQL document.id
    Type        string   // "chunk" 或 "qa"
    Content     string   // 原文内容（chunk原文 或 question文本）
    Vector      []float32// 向量 (维度取决于 embedding 模型)
}
```

### Schema 定义

```go
schema := &entity.Schema{
    CollectionName: fmt.Sprintf("kb_%d", knowledgeBaseId),
    Fields: []*entity.Field{
        {Name: "id", DataType: entity.FieldTypeVarChar, PrimaryKey: true, MaxLength: 128},
        {Name: "chunk_id", DataType: entity.FieldTypeVarChar, MaxLength: 64},
        {Name: "doc_id", DataType: entity.FieldTypeVarChar, MaxLength: 64},
        {Name: "type", DataType: entity.FieldTypeVarChar, MaxLength: 16},
        {Name: "content", DataType: entity.FieldTypeVarChar, MaxLength: 65535},
        {Name: "vector", DataType: entity.FieldTypeFloatVector, Dim: 1536}, // 根据模型调整
    },
}
```

---

## 3. 写入流程（状态机 + 异步重试）

### 3.1 状态定义

| 状态       | 含义                          |
| ---------- | ----------------------------- |
| `pending`  | 等待消费者处理                |
| `indexing` | 正在处理（解析+MySQL+Milvus） |
| `enable`   | 全部完成，可检索              |
| `fail`     | 失败，查看 `err_msg` 定位原因 |

### 3.2 消费者处理流程

```go
// Step 1-5: 解析文档、切片、生成 QA（与现有逻辑相同）

// Step 6: 事务写入 MySQL
err = sqlConn.TransactCtx(ctx, func(...) error {
    chunkModel.InsertBatch(ctx, saveChunks)
    docModel.UpdateStatus(ctx, documentId, "indexing")
    return nil
})
if err != nil {
    return failTask("MySQL写入失败: " + err.Error())
}

// Step 7: 写入 Milvus（非事务）
milvusRecords := expandChunksToMilvusRecords(saveChunks)
err = milvusClient.Insert(ctx, collectionName, milvusRecords)
if err != nil {
    // Milvus 失败，标记状态，等待定时任务重试
    return failTask("Milvus写入失败: " + err.Error())
}

// Step 8: 全部成功
docModel.UpdateStatusWithChunkCount(ctx, documentId, "enable", len(saveChunks))
```

### 3.3 Chunk 到 Milvus Records 展开逻辑

```go
func expandChunksToMilvusRecords(chunks []*knowledge.KnowledgeDocumentChunk) []*MilvusRecord {
    var records []*MilvusRecord
    
    for _, chunk := range chunks {
        var meta map[string]interface{}
        json.Unmarshal([]byte(chunk.Metadata), &meta)
        
        // 1. 添加原始 chunk 记录
        records = append(records, &MilvusRecord{
            ID:      fmt.Sprintf("%s_chunk_0", chunk.Id),
            ChunkID: chunk.Id,
            DocID:   chunk.KnowledgeDocumentId,
            Type:    "chunk",
            Content: chunk.ChunkText,
            Vector:  embed(chunk.ChunkText), // 需调用 embedding API
        })
        
        // 2. 添加 QA 问题记录
        if qaPairs, ok := meta["qa_pairs"].([]interface{}); ok {
            for i, qa := range qaPairs {
                pair := qa.(map[string]interface{})
                question := pair["question"].(string)
                
                records = append(records, &MilvusRecord{
                    ID:      fmt.Sprintf("%s_qa_%d", chunk.Id, i),
                    ChunkID: chunk.Id,
                    DocID:   chunk.KnowledgeDocumentId,
                    Type:    "qa",
                    Content: question,
                    Vector:  embed(question),
                })
            }
        }
    }
    
    return records
}
```

---

## 4. 定时任务重试（补偿机制）

```go
// 每 5 分钟执行一次
func RetryFailedDocuments(ctx context.Context) {
    docs := docModel.FindByStatus(ctx, "fail")
    
    for _, doc := range docs {
        // 检查是否是 Milvus 失败
        if !strings.Contains(doc.ErrMsg, "Milvus") {
            continue // 其他错误不在此处理
        }
        
        // 从 MySQL 读取已存的 chunks
        chunks := chunkModel.FindByDocId(ctx, doc.Id)
        if len(chunks) == 0 {
            continue // chunks 都没有，需要重新消费
        }
        
        // 重试写入 Milvus
        milvusRecords := expandChunksToMilvusRecords(chunks)
        err := milvusClient.Insert(ctx, collectionName, milvusRecords)
        if err != nil {
            logx.Errorf("重试 Milvus 失败: doc=%s, err=%v", doc.Id, err)
            continue
        }
        
        // 成功，更新状态
        docModel.UpdateStatusWithChunkCount(ctx, doc.Id, "enable", len(chunks))
        logx.Infof("重试成功: doc=%s", doc.Id)
    }
}
```

---

## 5. 检索流程

```go
// 用户 query -> 向量 -> Milvus 搜索 -> 返回 chunk_id -> MySQL 查原文

func Search(ctx context.Context, kbId uint64, query string, topK int) ([]*SearchResult, error) {
    // 1. Query embedding
    queryVector := embed(query)
    
    // 2. Milvus 搜索
    results, _ := milvusClient.Search(ctx, fmt.Sprintf("kb_%d", kbId), queryVector, topK)
    
    // 3. 去重（同一个 chunk 可能命中多次：chunk + qa）
    chunkIds := uniqueChunkIds(results)
    
    // 4. 从 MySQL 获取完整内容
    chunks := chunkModel.FindByIds(ctx, chunkIds)
    
    return chunks, nil
}
```

---

## 6. 删除/更新策略

当文档被删除或重新索引时：

```go
// 1. 删除 Milvus 中该文档的所有向量
milvusClient.Delete(ctx, collectionName, fmt.Sprintf("doc_id == '%s'", docId))

// 2. 删除 MySQL 中的 chunks
chunkModel.DeleteByDocId(ctx, docId)

// 3. 更新文档状态
docModel.UpdateStatus(ctx, docId, "disable", "")
```

---

## 7. 设计优势

1. **高召回率**：chunk 原文 + QA 问题双路召回
2. **语义对齐**：问句向量与用户 query 更接近
3. **最终一致性**：MySQL 事务 + Milvus 状态机 + 定时重试
4. **可追溯**：通过 `chunk_id` 随时关联回 MySQL 原始数据
5. **灵活扩展**：未来可增加更多 type（如 summary、keyword 等）