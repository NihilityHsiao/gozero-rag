package llm

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ LlmModel = (*customLlmModel)(nil)

type (
	// LlmModel is an interface to be customized, add more methods here,
	// and implement the added methods in customLlmModel.
	LlmModel interface {
		llmModel
	}

	customLlmModel struct {
		*defaultLlmModel
	}
)

// NewLlmModel returns a model for the database table.
func NewLlmModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) LlmModel {
	return &customLlmModel{
		defaultLlmModel: newLlmModel(conn, c, opts...),
	}
}
