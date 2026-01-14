package knowledge_document

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgeDocumentModel = (*customKnowledgeDocumentModel)(nil)

type (
	// KnowledgeDocumentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeDocumentModel.
	KnowledgeDocumentModel interface {
		knowledgeDocumentModel
		DeleteByKbId(ctx context.Context, kbId string) error
		FindListByKnowledgeBaseId(ctx context.Context, kbId string, page, pageSize int) ([]*KnowledgeDocument, error)
		CountByKnowledgeBaseId(ctx context.Context, kbId string) (int64, error)
		FindManyByIdsAndKbId(ctx context.Context, ids []string, kbId string) ([]*KnowledgeDocument, error)
	}

	customKnowledgeDocumentModel struct {
		*defaultKnowledgeDocumentModel
	}
)

// NewKnowledgeDocumentModel returns a model for the database table.
func NewKnowledgeDocumentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) KnowledgeDocumentModel {
	return &customKnowledgeDocumentModel{
		defaultKnowledgeDocumentModel: newKnowledgeDocumentModel(conn, c, opts...),
	}
}

func (m *customKnowledgeDocumentModel) DeleteByKbId(ctx context.Context, kbId string) error {
	query := fmt.Sprintf("delete from %s where knowledge_base_id = ?", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, kbId)
	return err
}

// FindListByKnowledgeBaseId 根据知识库ID查询文档列表（分页）
func (m *customKnowledgeDocumentModel) FindListByKnowledgeBaseId(ctx context.Context, kbId string, page, pageSize int) ([]*KnowledgeDocument, error) {
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where knowledge_base_id = ? order by created_time desc limit ?, ?", knowledgeDocumentRows, m.table)

	var resp []*KnowledgeDocument
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, kbId, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CountByKnowledgeBaseId 统计知识库下的文档数量
func (m *customKnowledgeDocumentModel) CountByKnowledgeBaseId(ctx context.Context, kbId string) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where knowledge_base_id = ?", m.table)

	var total int64
	err := m.QueryRowNoCacheCtx(ctx, &total, query, kbId)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// FindManyByIdsAndKbId 根据文档ID列表和知识库ID查询文档
func (m *customKnowledgeDocumentModel) FindManyByIdsAndKbId(ctx context.Context, ids []string, kbId string) ([]*KnowledgeDocument, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// 构造 IN 查询 (?,?,?)
	if len(ids) == 0 {
		return nil, nil
	}
	// 简单的占位符构造
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("select %s from %s where knowledge_base_id = ? and id in (%s)",
		knowledgeDocumentRows, m.table, strings.Join(placeholders, ","))

	// args 需要包含 kbId 在第一个
	fullArgs := append([]interface{}{kbId}, args...)

	var resp []*KnowledgeDocument
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, fullArgs...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
