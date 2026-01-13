package chat_conversation

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ChatConversationModel = (*customChatConversationModel)(nil)

type (
	// ChatConversationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatConversationModel.
	ChatConversationModel interface {
		chatConversationModel
		FindListByUserId(ctx context.Context, userId string, page, pageSize int) ([]*ChatConversation, int64, error)
	}

	customChatConversationModel struct {
		*defaultChatConversationModel
	}
)

// NewChatConversationModel returns a model for the database table.
func NewChatConversationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ChatConversationModel {
	return &customChatConversationModel{
		defaultChatConversationModel: newChatConversationModel(conn, c, opts...),
	}
}

func (m *customChatConversationModel) FindListByUserId(ctx context.Context, userId string, page, pageSize int) ([]*ChatConversation, int64, error) {
	// 1. Count Total
	queryCount := fmt.Sprintf("select count(*) from %s where user_id = ? and status != 3", m.table)
	var total int64
	err := m.QueryRowNoCacheCtx(ctx, &total, queryCount, userId)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return nil, 0, nil
	}

	// 2. Query List
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where user_id = ? and status != 3 order by updated_at desc limit ?, ?", chatConversationRows, m.table)

	var resp []*ChatConversation
	err = m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}
