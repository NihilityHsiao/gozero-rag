package tenant

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantModel = (*customTenantModel)(nil)

type (
	// TenantModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantModel.
	TenantModel interface {
		tenantModel
	}

	customTenantModel struct {
		*defaultTenantModel
	}
)

// NewTenantModel returns a model for the database table.
func NewTenantModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantModel {
	return &customTenantModel{
		defaultTenantModel: newTenantModel(conn, c, opts...),
	}
}
