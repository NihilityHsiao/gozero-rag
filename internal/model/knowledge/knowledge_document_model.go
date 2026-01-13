package knowledge

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgeDocumentModel = (*customKnowledgeDocumentModel)(nil)

type (
	// KnowledgeDocumentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeDocumentModel.
	KnowledgeDocumentModel interface {
		knowledgeDocumentModel
		withSession(session sqlx.Session) KnowledgeDocumentModel
		FindList(ctx context.Context, knowledgeBaseId uint64, page, pageSize int, status string) ([]*KnowledgeDocument, error)
		FindByIds(ctx context.Context, ids []string) ([]*KnowledgeDocument, error)
		Count(ctx context.Context, status string) (int64, error)
		FindListByKbId(ctx context.Context, kbId uint64, excludeStatus string) ([]*KnowledgeDocument, error)
		UpdateStatusByKbId(ctx context.Context, kbId uint64, status string, excludeStatus string) error
		UpdateParserConfigAndStatus(ctx context.Context, ids []string, jsonRule string, fromStatus, toStatus string) error
		UpdateStatus(ctx context.Context, id string, status string, errMsg string) error
		UpdateStatusWithChunkCount(ctx context.Context, id string, status string, chunkCount int) error
		UpdateStatusAtomic(ctx context.Context, id string, expectedStatus, newStatus string, errMsg string) error
	}

	customKnowledgeDocumentModel struct {
		*defaultKnowledgeDocumentModel
	}
)

// NewKnowledgeDocumentModel returns a model for the database table.
func NewKnowledgeDocumentModel(conn sqlx.SqlConn) KnowledgeDocumentModel {
	return &customKnowledgeDocumentModel{
		defaultKnowledgeDocumentModel: newKnowledgeDocumentModel(conn),
	}
}

func (m *customKnowledgeDocumentModel) withSession(session sqlx.Session) KnowledgeDocumentModel {
	return NewKnowledgeDocumentModel(sqlx.NewSqlConnFromSession(session))
}
func (m *customKnowledgeDocumentModel) FindList(ctx context.Context, knowledgeBaseId uint64, page, pageSize int, status string) ([]*KnowledgeDocument, error) {
	query := fmt.Sprintf("select %s from %s where knowledge_base_id = ?", knowledgeDocumentRows, m.table)
	var args []interface{}
	args = append(args, knowledgeBaseId)
	if status != "" {
		query += " and status = ?"
		args = append(args, status)
	}
	query += " order by id desc limit ?, ?"
	args = append(args, (page-1)*pageSize, pageSize)

	var resp []*KnowledgeDocument
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}
func (m *customKnowledgeDocumentModel) FindByIds(ctx context.Context, ids []string) ([]*KnowledgeDocument, error) {
	query := fmt.Sprintf("select %s from %s where id in ('%s')", knowledgeDocumentRows, m.table, strings.Join(ids, "','"))
	var resp []*KnowledgeDocument
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err

}

func (m *customKnowledgeDocumentModel) Count(ctx context.Context, status string) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s", m.table)
	var args []interface{}
	if status != "" {
		query += " where status = ?"
		args = append(args, status)
	}
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customKnowledgeDocumentModel) FindListByKbId(ctx context.Context, kbId uint64, excludeStatus string) ([]*KnowledgeDocument, error) {
	query := fmt.Sprintf("select %s from %s where knowledge_base_id = ?", knowledgeDocumentRows, m.table)
	args := []interface{}{kbId}

	if excludeStatus != "" {
		query += " and status != ?"
		args = append(args, excludeStatus)
	}

	var resp []*KnowledgeDocument
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customKnowledgeDocumentModel) UpdateStatusByKbId(ctx context.Context, kbId uint64, status string, excludeStatus string) error {
	query := fmt.Sprintf("update %s set status = ? where knowledge_base_id = ?", m.table)
	args := []interface{}{status, kbId}

	if excludeStatus != "" {
		query += " and status != ?"
		args = append(args, excludeStatus)
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}
func (m *customKnowledgeDocumentModel) UpdateParserConfigAndStatus(ctx context.Context, ids []string, jsonRule string, fromStatus, toStatus string) error {
	if len(ids) == 0 {
		return nil // 如果没有 ID，直接返回，避免 SQL 语法错误
	}

	// 1. 动态构建占位符字符串，例如: "?,?,?"
	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = "?"
	}

	// 2. 拼接 SQL 语句
	// 最终变成: update table set parser_config = ? where id IN (?,?,?)
	query := fmt.Sprintf("update %s set parser_config = ? , status = ? where id IN (%s) AND status = ?",
		m.table, strings.Join(placeholders, ","))

	// 3. 构建参数列表
	// 第一个参数是 jsonRule，后面依次追加 ids
	args := make([]interface{}, 0, len(ids)+3)
	args = append(args, jsonRule, toStatus)
	for _, id := range ids {
		args = append(args, id)
	}
	args = append(args, fromStatus)

	// 4. 执行
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

func (m *customKnowledgeDocumentModel) UpdateStatus(ctx context.Context, id string, status string, errMsg string) error {
	query := fmt.Sprintf("update %s set status = ?, err_msg = ? where id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, errMsg, id)
	return err
}

// UpdateStatusWithChunkCount 更新文档状态和切片数量
func (m *customKnowledgeDocumentModel) UpdateStatusWithChunkCount(ctx context.Context, id string, status string, chunkCount int) error {
	query := fmt.Sprintf("update %s set status = ?, chunk_count = ?, err_msg = '' where id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, chunkCount, id)
	return err
}

// UpdateStatusAtomic 原子更新状态，仅当当前状态与预期匹配时才更新。
// 如果状态已被其他进程修改，返回 ErrOptimisticLock。
func (m *customKnowledgeDocumentModel) UpdateStatusAtomic(ctx context.Context, id string, expectedStatus, newStatus string, errMsg string) error {
	query := fmt.Sprintf("update %s set status = ?, err_msg = ? where id = ? and status = ?", m.table)
	result, err := m.conn.ExecCtx(ctx, query, newStatus, errMsg, id, expectedStatus)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrOptimisticLock
	}
	return nil
}
