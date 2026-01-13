package knowledge

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgeBaseModel = (*customKnowledgeBaseModel)(nil)

type (
	// KnowledgeBaseModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeBaseModel.
	KnowledgeBaseModel interface {
		knowledgeBaseModel
		withSession(session sqlx.Session) KnowledgeBaseModel
		FindList(ctx context.Context, page, pageSize int, status int64) ([]*KnowledgeBase, error)
		Count(ctx context.Context, status int64) (int64, error)
		FindByIds(ctx context.Context, ids []uint64) ([]*KnowledgeBase, error)
	}

	customKnowledgeBaseModel struct {
		*defaultKnowledgeBaseModel
	}
)

// NewKnowledgeBaseModel returns a model for the database table.
func NewKnowledgeBaseModel(conn sqlx.SqlConn) KnowledgeBaseModel {
	return &customKnowledgeBaseModel{
		defaultKnowledgeBaseModel: newKnowledgeBaseModel(conn),
	}
}

func (m *customKnowledgeBaseModel) withSession(session sqlx.Session) KnowledgeBaseModel {
	return NewKnowledgeBaseModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customKnowledgeBaseModel) FindList(ctx context.Context, page, pageSize int, status int64) ([]*KnowledgeBase, error) {
	query := fmt.Sprintf("select %s from %s", knowledgeBaseRows, m.table)
	var args []interface{}
	if status != 0 {
		query += " where status = ?"
		args = append(args, status)
	}
	query += " order by id desc limit ?, ?"
	args = append(args, (page-1)*pageSize, pageSize)

	var resp []*KnowledgeBase
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customKnowledgeBaseModel) Count(ctx context.Context, status int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s", m.table)
	var args []interface{}
	if status != 0 {
		query += " where status = ?"
		args = append(args, status)
	}
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

// FindByIds 根据ID列表批量查询知识库
func (m *customKnowledgeBaseModel) FindByIds(ctx context.Context, ids []uint64) ([]*KnowledgeBase, error) {
	if len(ids) == 0 {
		return []*KnowledgeBase{}, nil
	}

	// 构建 IN 查询的占位符
	placeholders := ""
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` IN (%s)", knowledgeBaseRows, m.table, placeholders)

	var list []*KnowledgeBase
	err := m.conn.QueryRowsCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, err
	}

	return list, nil
}
