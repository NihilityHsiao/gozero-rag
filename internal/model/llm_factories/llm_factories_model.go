package llm_factories

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ LlmFactoriesModel = (*customLlmFactoriesModel)(nil)

type (
	// LlmFactoriesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customLlmFactoriesModel.
	LlmFactoriesModel interface {
		llmFactoriesModel
		// 自定义方法: 获取所有有效的厂商列表 (按rank降序)
		FindAllActive(ctx context.Context) ([]*LlmFactories, error)
	}

	customLlmFactoriesModel struct {
		*defaultLlmFactoriesModel
	}
)

// NewLlmFactoriesModel returns a model for the database table.
func NewLlmFactoriesModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) LlmFactoriesModel {
	return &customLlmFactoriesModel{
		defaultLlmFactoriesModel: newLlmFactoriesModel(conn, c, opts...),
	}
}

// FindAllActive 获取所有有效的厂商列表，按rank降序排列
func (m *customLlmFactoriesModel) FindAllActive(ctx context.Context) ([]*LlmFactories, error) {
	query := fmt.Sprintf("select %s from %s where `status` = 1 order by `rank` desc", llmFactoriesRows, m.table)

	var resp []*LlmFactories
	err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
