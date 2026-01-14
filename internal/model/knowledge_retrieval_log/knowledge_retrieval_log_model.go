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
		FindLogList(ctx context.Context, kbId, userId string, offset, limit int) ([]*KnowledgeRetrievalLog, error)
		CountLog(ctx context.Context, kbId, userId string) (int64, error)
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

func (m *customKnowledgeRetrievalLogModel) FindLogList(ctx context.Context, kbId, userId string, offset, limit int) ([]*KnowledgeRetrievalLog, error) {
	query := fmt.Sprintf("select %s from %s where knowledge_base_id = ? and user_id = ? order by created_at desc limit ?, ?", knowledgeRetrievalLogRows, m.table)
	var resp []*KnowledgeRetrievalLog
	// No cache model
	err := m.conn.QueryRowsCtx(ctx, &resp, query, kbId, userId, offset, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customKnowledgeRetrievalLogModel) CountLog(ctx context.Context, kbId, userId string) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where knowledge_base_id = ? and user_id = ?", m.table)
	var resp int64
	err := m.conn.QueryRowCtx(ctx, &resp, query, kbId, userId)
	if err != nil {
		return 0, err
	}
	return resp, nil
}
