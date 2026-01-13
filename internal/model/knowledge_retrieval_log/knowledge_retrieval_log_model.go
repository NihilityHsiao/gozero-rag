package knowledge_retrieval_log

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgeRetrievalLogModel = (*customKnowledgeRetrievalLogModel)(nil)

type (
	// KnowledgeRetrievalLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeRetrievalLogModel.
	KnowledgeRetrievalLogModel interface {
		knowledgeRetrievalLogModel
		withSession(session sqlx.Session) KnowledgeRetrievalLogModel
		FindLogList(ctx context.Context, kbId uint64, userId uint64, offset int, limit int) ([]*KnowledgeRetrievalLog, error)
		CountLog(ctx context.Context, kbId uint64, userId uint64) (int64, error)
	}

	customKnowledgeRetrievalLogModel struct {
		*defaultKnowledgeRetrievalLogModel
	}
)

// NewKnowledgeRetrievalLogModel returns a model for the database table.
func NewKnowledgeRetrievalLogModel(conn sqlx.SqlConn) KnowledgeRetrievalLogModel {
	return &customKnowledgeRetrievalLogModel{
		defaultKnowledgeRetrievalLogModel: newKnowledgeRetrievalLogModel(conn),
	}
}

func (m *customKnowledgeRetrievalLogModel) withSession(session sqlx.Session) KnowledgeRetrievalLogModel {
	return NewKnowledgeRetrievalLogModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customKnowledgeRetrievalLogModel) FindLogList(ctx context.Context, kbId uint64, userId uint64, offset int, limit int) ([]*KnowledgeRetrievalLog, error) {
	query := fmt.Sprintf("select %s from %s where `knowledge_base_id` = ? and `user_id` = ? order by `id` desc limit ?, ?", knowledgeRetrievalLogRows, m.table)
	var resp []*KnowledgeRetrievalLog
	err := m.conn.QueryRowsCtx(ctx, &resp, query, kbId, userId, offset, limit)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customKnowledgeRetrievalLogModel) CountLog(ctx context.Context, kbId uint64, userId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `knowledge_base_id` = ? and `user_id` = ?", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, kbId, userId)
	switch err {
	case nil:
		return count, nil
	case sqlx.ErrNotFound:
		return 0, nil
	default:
		return 0, err
	}
}
