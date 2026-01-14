package knowledge_base

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgeBaseModel = (*customKnowledgeBaseModel)(nil)

type (
	// KnowledgeBaseModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeBaseModel.
	KnowledgeBaseModel interface {
		knowledgeBaseModel
		FindList(ctx context.Context, tenantId string, page, pageSize int) ([]*KnowledgeBase, error)
		Count(ctx context.Context, tenantId string) (int64, error)
		UpdatePermission(ctx context.Context, id, permission string) error
		// 多租户列表查询
		FindListByMultiTenants(ctx context.Context, userId string, tenantIds []string, page, pageSize int) ([]*KnowledgeBase, error)
		CountByMultiTenants(ctx context.Context, userId string, tenantIds []string) (int64, error)
	}

	customKnowledgeBaseModel struct {
		*defaultKnowledgeBaseModel
	}
)

// NewKnowledgeBaseModel returns a model for the database table.
func NewKnowledgeBaseModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) KnowledgeBaseModel {
	return &customKnowledgeBaseModel{
		defaultKnowledgeBaseModel: newKnowledgeBaseModel(conn, c, opts...),
	}
}

func (m *customKnowledgeBaseModel) FindList(ctx context.Context, tenantId string, page, pageSize int) ([]*KnowledgeBase, error) {
	query := fmt.Sprintf("select %s from %s where tenant_id = ? order by created_time desc limit ?, ?", knowledgeBaseRows, m.table)
	var resp []*KnowledgeBase
	// No cache for list query usually, or handle cache invalidation carefully.
	// Using QueryRowsNoCacheCtx for simplicity.
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, tenantId, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customKnowledgeBaseModel) Count(ctx context.Context, tenantId string) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where tenant_id = ?", m.table)
	var resp int64
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, tenantId)
	if err != nil {
		return 0, err
	}
	return resp, nil
}

// UpdatePermission 只更新知识库的 permission 字段（避免全字段更新的SQL参数问题）
func (m *customKnowledgeBaseModel) UpdatePermission(ctx context.Context, id, permission string) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	knowledgeBaseIdKey := fmt.Sprintf("%s%v", cacheKnowledgeBaseIdPrefix, id)
	knowledgeBaseTenantIdNameKey := fmt.Sprintf("%s%v:%v", cacheKnowledgeBaseTenantIdNamePrefix, data.TenantId, data.Name)

	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set `permission` = ?, `updated_time` = ?, `updated_date` = ? where `id` = ?", m.table)
		now := time.Now()
		return conn.ExecCtx(ctx, query, permission, now.UnixMilli(), now, id)
	}, knowledgeBaseIdKey, knowledgeBaseTenantIdNameKey)

	return err
}

// FindListByMultiTenants 多租户知识库列表查询
// SQL逻辑：
// 1. 用户创建的所有知识库（不限permission）
// 2. 用户加入的租户中 permission='team' 的知识库
func (m *customKnowledgeBaseModel) FindListByMultiTenants(ctx context.Context, userId string, tenantIds []string, page, pageSize int) ([]*KnowledgeBase, error) {
	if len(tenantIds) == 0 {
		return []*KnowledgeBase{}, nil
	}

	// 构建 IN 子句占位符
	placeholders := make([]string, len(tenantIds))
	for i := range tenantIds {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE status = 1
		  AND (
		    created_by = ?
		    OR
		    (tenant_id IN (%s) AND permission = 'team')
		  )
		ORDER BY created_time DESC
		LIMIT ?, ?
	`, knowledgeBaseRows, m.table, strings.Join(placeholders, ","))

	// 参数: userId, ...tenantIds, offset, pageSize
	args := []interface{}{userId}
	for _, tid := range tenantIds {
		args = append(args, tid)
	}
	offset := (page - 1) * pageSize
	args = append(args, offset, pageSize)

	var list []*KnowledgeBase
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...)
	return list, err
}

// CountByMultiTenants 多租户知识库总数统计
func (m *customKnowledgeBaseModel) CountByMultiTenants(ctx context.Context, userId string, tenantIds []string) (int64, error) {
	if len(tenantIds) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(tenantIds))
	for i := range tenantIds {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s
		WHERE status = 1
		  AND (
		    created_by = ?
		    OR
		    (tenant_id IN (%s) AND permission = 'team')
		  )
	`, m.table, strings.Join(placeholders, ","))

	args := []interface{}{userId}
	for _, tid := range tenantIds {
		args = append(args, tid)
	}

	var total int64
	err := m.QueryRowNoCacheCtx(ctx, &total, query, args...)
	return total, err
}
