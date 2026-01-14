package tenant_llm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TenantLlmModel = (*customTenantLlmModel)(nil)

type (
	// TenantLlmModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTenantLlmModel.
	TenantLlmModel interface {
		tenantLlmModel
		// 根据租户ID查询配置列表 (分页)
		FindListByTenantId(ctx context.Context, tenantId string, llmFactory, modelType string, status int64, page, pageSize int64) ([]*TenantLlm, int64, error)
		// 按厂商分组查询
		FindGroupedByFactory(ctx context.Context, tenantId string) (map[string][]*TenantLlm, error)
		// 批量插入 (忽略冲突，返回成功数量)
		BatchInsertIgnore(ctx context.Context, records []*TenantLlm) (int64, []string, error)
		// 根据ID和租户ID查询 (多租户安全查询)
		FindOneByIdAndTenantId(ctx context.Context, id uint64, tenantId string) (*TenantLlm, error)
		// 根据ID和租户ID删除 (多租户安全删除)
		DeleteByIdAndTenantId(ctx context.Context, id uint64, tenantId string) error
		// 根据ID和租户ID更新 (多租户安全更新)
		UpdateByIdAndTenantId(ctx context.Context, data *TenantLlm) error
		// 根据租户ID、厂商和模型名称查询（用于校验 embd_id）
		FindByTenantFactoryName(ctx context.Context, tenantId, llmFactory, llmName string) (*TenantLlm, error)
	}

	customTenantLlmModel struct {
		*defaultTenantLlmModel
	}
)

// NewTenantLlmModel returns a model for the database table.
func NewTenantLlmModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TenantLlmModel {
	return &customTenantLlmModel{
		defaultTenantLlmModel: newTenantLlmModel(conn, c, opts...),
	}
}

// FindListByTenantId 根据租户ID查询配置列表，支持筛选和分页
func (m *customTenantLlmModel) FindListByTenantId(ctx context.Context, tenantId string, llmFactory, modelType string, status int64, page, pageSize int64) ([]*TenantLlm, int64, error) {
	// 构建查询条件
	where := "`tenant_id` = ?"
	args := []interface{}{tenantId}

	if llmFactory != "" {
		where += " AND `llm_factory` = ?"
		args = append(args, llmFactory)
	}
	if modelType != "" {
		where += " AND `model_type` = ?"
		args = append(args, modelType)
	}
	if status > 0 {
		where += " AND `status` = ?"
		args = append(args, status)
	}

	// 查询总数
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	var total int64
	err := m.CachedConn.QueryRowNoCacheCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where %s order by `created_time` desc limit ?, ?", tenantLlmRows, m.table, where)
	args = append(args, offset, pageSize)

	var resp []*TenantLlm
	err = m.CachedConn.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}

// FindGroupedByFactory 按厂商分组查询
func (m *customTenantLlmModel) FindGroupedByFactory(ctx context.Context, tenantId string) (map[string][]*TenantLlm, error) {
	query := fmt.Sprintf("select %s from %s where `tenant_id` = ? and `status` = 1 order by `llm_factory`, `model_type`", tenantLlmRows, m.table)

	var resp []*TenantLlm
	err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &resp, query, tenantId)
	if err != nil {
		return nil, err
	}

	// 按厂商分组
	grouped := make(map[string][]*TenantLlm)
	for _, item := range resp {
		grouped[item.LlmFactory] = append(grouped[item.LlmFactory], item)
	}

	return grouped, nil
}

// BatchInsertIgnore 批量插入，忽略冲突
func (m *customTenantLlmModel) BatchInsertIgnore(ctx context.Context, records []*TenantLlm) (int64, []string, error) {
	if len(records) == 0 {
		return 0, nil, nil
	}

	var successCount int64
	var failedModels []string

	// 逐条插入，遇到冲突则跳过
	for _, record := range records {
		_, err := m.Insert(ctx, record)
		if err != nil {
			// 检查是否是唯一约束冲突
			if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "duplicate key") {
				failedModels = append(failedModels, record.LlmName)
				continue
			}
			// 其他错误，继续处理下一条
			failedModels = append(failedModels, record.LlmName)
			continue
		}
		successCount++
	}

	return successCount, failedModels, nil
}

// FindOneByIdAndTenantId 多租户安全查询：根据ID和租户ID查询
func (m *customTenantLlmModel) FindOneByIdAndTenantId(ctx context.Context, id uint64, tenantId string) (*TenantLlm, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? and `tenant_id` = ? limit 1", tenantLlmRows, m.table)

	var resp TenantLlm
	err := m.CachedConn.QueryRowNoCacheCtx(ctx, &resp, query, id, tenantId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}

// DeleteByIdAndTenantId 多租户安全删除：只能删除属于该租户的记录
func (m *customTenantLlmModel) DeleteByIdAndTenantId(ctx context.Context, id uint64, tenantId string) error {
	// 先查询确保记录属于该租户
	data, err := m.FindOneByIdAndTenantId(ctx, id, tenantId)
	if err != nil {
		return err
	}

	// 使用默认的删除方法，会自动清理缓存
	return m.Delete(ctx, data.Id)
}

// UpdateByIdAndTenantId 多租户安全更新：只能更新属于该租户的记录
func (m *customTenantLlmModel) UpdateByIdAndTenantId(ctx context.Context, data *TenantLlm) error {
	// 先查询确保记录属于该租户
	existing, err := m.FindOneByIdAndTenantId(ctx, data.Id, data.TenantId)
	if err != nil {
		return err
	}

	// 保留原有的不可变字段
	data.TenantId = existing.TenantId
	data.LlmFactory = existing.LlmFactory
	data.LlmName = existing.LlmName
	data.ModelType = existing.ModelType

	// 使用默认的更新方法
	return m.Update(ctx, data)
}

// 辅助函数：将 string 转换为 sql.NullString
func ToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// FindByTenantFactoryName 根据租户ID、厂商和模型名称查询配置
// 用于校验创建知识库时的 embd_id (格式: 模型名称@厂商)
func (m *customTenantLlmModel) FindByTenantFactoryName(ctx context.Context, tenantId, llmFactory, llmName string) (*TenantLlm, error) {
	query := fmt.Sprintf("select %s from %s where `tenant_id` = ? and `llm_factory` = ? and `llm_name` = ? and `status` = 1 limit 1", tenantLlmRows, m.table)

	var resp TenantLlm
	err := m.CachedConn.QueryRowNoCacheCtx(ctx, &resp, query, tenantId, llmFactory, llmName)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}
