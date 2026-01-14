package user_tenant

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserTenantModel = (*customUserTenantModel)(nil)

type (
	// UserTenantModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserTenantModel.
	UserTenantModel interface {
		userTenantModel
		// 根据用户ID查询所有关联的租户信息
		FindByUserId(ctx context.Context, userId string) ([]*UserTenantWithTenant, error)
		// 根据租户ID查询所有成员信息(包含用户详情)
		FindMembersByTenantId(ctx context.Context, tenantId string) ([]*UserTenantWithUser, error)
		// 根据用户ID和租户ID查询关联记录
		FindByUserIdAndTenantId(ctx context.Context, userId, tenantId string) (*UserTenant, error)
		// 根据租户ID获取Owner信息
		FindOwnerByTenantId(ctx context.Context, tenantId string) (*UserTenantWithUser, error)
		// 查询用户关联的所有租户ID
		FindTenantsByUserId(ctx context.Context, userId string) (tenantIds []string, err error)
	}

	customUserTenantModel struct {
		*defaultUserTenantModel
	}

	// UserTenantWithTenant 用户租户关联信息（包含租户名称）
	UserTenantWithTenant struct {
		UserTenant
		TenantName string `db:"tenant_name"` // 租户名称
	}

	// UserTenantWithUser 用户租户关联信息（包含用户详情）
	UserTenantWithUser struct {
		UserTenant
		Nickname string `db:"nickname"` // 用户昵称
		Email    string `db:"email"`    // 用户邮箱
	}
)

// NewUserTenantModel returns a model for the database table.
func NewUserTenantModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserTenantModel {
	return &customUserTenantModel{
		defaultUserTenantModel: newUserTenantModel(conn, c, opts...),
	}
}

// FindByUserId 根据用户ID查询所有关联的租户信息
func (m *customUserTenantModel) FindByUserId(ctx context.Context, userId string) ([]*UserTenantWithTenant, error) {

	query := `
		SELECT 
			ut.id, ut.user_id, ut.tenant_id, ut.role, ut.invited_by, ut.status,
			ut.created_time, ut.updated_time, ut.created_date, ut.updated_date,
			COALESCE(t.name, '') as tenant_name
		FROM user_tenant ut
		LEFT JOIN tenant t ON ut.tenant_id = t.id
		WHERE ut.user_id = ? AND ut.status = 1
		ORDER BY ut.created_time ASC
	`

	var resp []*UserTenantWithTenant
	err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// FindMembersByTenantId 根据租户ID查询所有成员信息(包含用户详情)
func (m *customUserTenantModel) FindMembersByTenantId(ctx context.Context, tenantId string) ([]*UserTenantWithUser, error) {
	query := `
		SELECT 
			ut.id, ut.user_id, ut.tenant_id, ut.role, ut.invited_by, ut.status,
			ut.created_time, ut.updated_time, ut.created_date, ut.updated_date,
			COALESCE(u.nickname, '') as nickname,
			COALESCE(u.email, '') as email
		FROM user_tenant ut
		LEFT JOIN user u ON ut.user_id = u.id
		WHERE ut.tenant_id = ? AND ut.status = 1
		ORDER BY ut.created_time ASC
	`

	var resp []*UserTenantWithUser
	err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &resp, query, tenantId)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// FindByUserIdAndTenantId 根据用户ID和租户ID查询关联记录
func (m *customUserTenantModel) FindByUserIdAndTenantId(ctx context.Context, userId, tenantId string) (*UserTenant, error) {
	query := `
		SELECT id, user_id, tenant_id, role, invited_by, status,
			created_time, updated_time, created_date, updated_date
		FROM user_tenant
		WHERE user_id = ? AND tenant_id = ? AND status = 1
		LIMIT 1
	`

	var resp UserTenant
	err := m.CachedConn.QueryRowNoCacheCtx(ctx, &resp, query, userId, tenantId)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// FindOwnerByTenantId 根据租户ID获取Owner信息
func (m *customUserTenantModel) FindOwnerByTenantId(ctx context.Context, tenantId string) (*UserTenantWithUser, error) {
	query := `
		SELECT 
			ut.id, ut.user_id, ut.tenant_id, ut.role, ut.invited_by, ut.status,
			ut.created_time, ut.updated_time, ut.created_date, ut.updated_date,
			COALESCE(u.nickname, '') as nickname,
			COALESCE(u.email, '') as email
		FROM user_tenant ut
		LEFT JOIN user u ON ut.user_id = u.id
		WHERE ut.tenant_id = ? AND ut.role = 'owner' AND ut.status = 1
		LIMIT 1
	`

	var resp UserTenantWithUser
	err := m.CachedConn.QueryRowNoCacheCtx(ctx, &resp, query, tenantId)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// FindTenantsByUserId 查询用户关联的所有租户ID（用于多租户知识库查询）
func (m *customUserTenantModel) FindTenantsByUserId(ctx context.Context, userId string) ([]string, error) {
	query := `
		SELECT tenant_id 
		FROM user_tenant 
		WHERE user_id = ? AND status = 1
		ORDER BY created_time ASC
	`

	var tenantIds []string
	err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &tenantIds, query, userId)
	if err != nil {
		return nil, err
	}

	return tenantIds, nil
}
