package chat_message

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ChatMessageModel = (*customChatMessageModel)(nil)

type (
	// ChatMessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatMessageModel.
	ChatMessageModel interface {
		chatMessageModel
		withSession(session sqlx.Session) ChatMessageModel
	}

	customChatMessageModel struct {
		*defaultChatMessageModel
	}
)

// NewChatMessageModel returns a model for the database table.
func NewChatMessageModel(conn sqlx.SqlConn) ChatMessageModel {
	return &customChatMessageModel{
		defaultChatMessageModel: newChatMessageModel(conn),
	}
}

func (m *customChatMessageModel) withSession(session sqlx.Session) ChatMessageModel {
	return NewChatMessageModel(sqlx.NewSqlConnFromSession(session))
}
