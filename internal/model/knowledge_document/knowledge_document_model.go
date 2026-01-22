package knowledge_document

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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
		FindListByKnowledgeBaseId(ctx context.Context, kbId string, page, pageSize int, keyword string) ([]*KnowledgeDocument, error)
		CountByKnowledgeBaseId(ctx context.Context, kbId string, keyword string) (int64, error)
		FindManyByIdsAndKbId(ctx context.Context, ids []string, kbId string) ([]*KnowledgeDocument, error)
		UpdateRunStatus(ctx context.Context, id, status, msg string) error
		UpdateStatusWithChunkCount(ctx context.Context, id, status string, chunkNum, tokenNum int64) error
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
func (m *customKnowledgeDocumentModel) FindListByKnowledgeBaseId(ctx context.Context, kbId string, page, pageSize int, keyword string) ([]*KnowledgeDocument, error) {
	offset := (page - 1) * pageSize

	var query string
	var args []interface{}

	if keyword != "" {
		// 模糊搜索 doc_name 或 精确匹配 id
		query = fmt.Sprintf("select %s from %s where knowledge_base_id = ? and (doc_name like ? or id = ?) order by created_time desc limit ?, ?", knowledgeDocumentRows, m.table)
		likeKeyword := "%" + keyword + "%"
		args = []interface{}{kbId, likeKeyword, keyword, offset, pageSize}
	} else {
		query = fmt.Sprintf("select %s from %s where knowledge_base_id = ? order by created_time desc limit ?, ?", knowledgeDocumentRows, m.table)
		args = []interface{}{kbId, offset, pageSize}
	}

	var resp []*KnowledgeDocument
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CountByKnowledgeBaseId 统计知识库下的文档数量
func (m *customKnowledgeDocumentModel) CountByKnowledgeBaseId(ctx context.Context, kbId string, keyword string) (int64, error) {
	var query string
	var args []interface{}

	if keyword != "" {
		query = fmt.Sprintf("select count(*) from %s where knowledge_base_id = ? and (doc_name like ? or id = ?)", m.table)
		likeKeyword := "%" + keyword + "%"
		args = []interface{}{kbId, likeKeyword, keyword}
	} else {
		query = fmt.Sprintf("select count(*) from %s where knowledge_base_id = ?", m.table)
		args = []interface{}{kbId}
	}

	var total int64
	err := m.QueryRowNoCacheCtx(ctx, &total, query, args...)
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

// UpdateRunStatus 更新文档运行状态
func (m *customKnowledgeDocumentModel) UpdateRunStatus(ctx context.Context, id, status, msg string) error {
	knowledgeDocumentIdKey := fmt.Sprintf("%s%v", cacheKnowledgeDocumentIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set run_status = ?, progress_msg = ?, updated_time = ?, updated_date = ? where `id` = ?", m.table)
		now := time.Now()
		return conn.ExecCtx(ctx, query, status, msg, now.UnixMilli(), now, id)
	}, knowledgeDocumentIdKey)
	return err
}

// UpdateStatusWithChunkCount 更新文档状态及切片统计
func (m *customKnowledgeDocumentModel) UpdateStatusWithChunkCount(ctx context.Context, id, status string, chunkNum, tokenNum int64) error {
	knowledgeDocumentIdKey := fmt.Sprintf("%s%v", cacheKnowledgeDocumentIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set run_status = ?, chunk_num = ?, token_num = ?, updated_time = ?, updated_date = ? where `id` = ?", m.table)
		now := time.Now()
		return conn.ExecCtx(ctx, query, status, chunkNum, tokenNum, now.UnixMilli(), now, id)
	}, knowledgeDocumentIdKey)
	return err
}
