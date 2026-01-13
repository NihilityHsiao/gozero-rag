package knowledge

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgeDocumentChunkModel = (*customKnowledgeDocumentChunkModel)(nil)

type (
	// KnowledgeDocumentChunkModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeDocumentChunkModel.
	KnowledgeDocumentChunkModel interface {
		knowledgeDocumentChunkModel
		withSession(session sqlx.Session) KnowledgeDocumentChunkModel
		InsertBatch(ctx context.Context, chunks []*KnowledgeDocumentChunk) error
		FindListByDocId(ctx context.Context, docId string, page, pageSize int, keyword string) ([]*KnowledgeDocumentChunk, int64, error)
	}

	customKnowledgeDocumentChunkModel struct {
		*defaultKnowledgeDocumentChunkModel
	}
)

// NewKnowledgeDocumentChunkModel returns a model for the database table.
func NewKnowledgeDocumentChunkModel(conn sqlx.SqlConn) KnowledgeDocumentChunkModel {
	return &customKnowledgeDocumentChunkModel{
		defaultKnowledgeDocumentChunkModel: newKnowledgeDocumentChunkModel(conn),
	}
}

func (m *customKnowledgeDocumentChunkModel) withSession(session sqlx.Session) KnowledgeDocumentChunkModel {
	return NewKnowledgeDocumentChunkModel(sqlx.NewSqlConnFromSession(session))
}

// InsertBatch 批量插入 chunk 记录
func (m *customKnowledgeDocumentChunkModel) InsertBatch(ctx context.Context, chunks []*KnowledgeDocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// 构建批量插入 SQL
	// INSERT INTO table (col1, col2, ...) VALUES (?, ?, ...), (?, ?, ...), ...
	valueStrings := make([]string, 0, len(chunks))
	valueArgs := make([]interface{}, 0, len(chunks)*7)

	for _, chunk := range chunks {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			chunk.Id,
			chunk.KnowledgeBaseId,
			chunk.KnowledgeDocumentId,
			chunk.ChunkText,
			chunk.ChunkSize,
			chunk.Metadata,
			chunk.Status,
		)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (id, knowledge_base_id, knowledge_document_id, chunk_text, chunk_size, metadata, status) VALUES %s",
		m.table,
		strings.Join(valueStrings, ","),
	)

	_, err := m.conn.ExecCtx(ctx, query, valueArgs...)
	return err
}

func (m *customKnowledgeDocumentChunkModel) FindListByDocId(ctx context.Context, docId string, page, pageSize int, keyword string) ([]*KnowledgeDocumentChunk, int64, error) {
	var resp []*KnowledgeDocumentChunk
	var count int64

	// Count
	countQuery := fmt.Sprintf("SELECT count(*) FROM %s WHERE knowledge_document_id = ?", m.table)
	args := []interface{}{docId}
	if keyword != "" {
		countQuery += " AND chunk_text LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	err := m.conn.QueryRowCtx(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// List
	query := fmt.Sprintf("SELECT %s FROM %s WHERE knowledge_document_id = ?", knowledgeDocumentChunkRows, m.table)
	// args is already initialized with docId
	// recreate args for list query because limit/offset needed at end
	listArgs := []interface{}{docId}
	if keyword != "" {
		query += " AND chunk_text LIKE ?"
		listArgs = append(listArgs, "%"+keyword+"%")
	}

	query += " ORDER BY id ASC LIMIT ? OFFSET ?"
	listArgs = append(listArgs, pageSize, (page-1)*pageSize)

	err = m.conn.QueryRowsCtx(ctx, &resp, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}

	return resp, count, nil
}
