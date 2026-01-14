package user_api

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserApiModel = (*customUserApiModel)(nil)

type (
	// UserApiModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserApiModel.
	UserApiModel interface {
		userApiModel
		// 自定义方法：根据用户ID查询模型列表（支持分页和筛选）
		FindListByUserId(ctx context.Context, userId string, modelType string, status int, page, pageSize int) ([]*UserApi, int64, error)
		// 自定义方法：根据用户ID和模型类型查询模型列表
		FindByUserIdAndModelType(ctx context.Context, userId string, modelType string) ([]*UserApi, error)
		// 自定义方法：更新默认状态 (事务)
		UpdateDefaultStatus(ctx context.Context, userId string, modelType string, targetId int64) error
		FindByIds(ctx context.Context, ids []uint64) ([]*UserApi, error)
	}

	customUserApiModel struct {
		*defaultUserApiModel
	}
)

// NewUserApiModel returns a model for the database table.
func NewUserApiModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserApiModel {
	return &customUserApiModel{
		defaultUserApiModel: newUserApiModel(conn, c, opts...),
	}
}

// FindListByUserId 根据用户ID查询模型列表（支持分页和筛选）
func (m *customUserApiModel) FindListByUserId(ctx context.Context, userId string, modelType string, status int, page, pageSize int) ([]*UserApi, int64, error) {
	// 构建查询条件
	where := "`user_id` = ?"
	args := []interface{}{userId}

	if modelType != "" {
		where += " AND `model_type` = ?"
		args = append(args, modelType)
	}
	if status > 0 {
		where += " AND `status` = ?"
		args = append(args, status)
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", m.table, where)
	var total int64
	err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY `id` DESC LIMIT ?, ?", userApiRows, m.table, where)
	args = append(args, offset, pageSize)

	var list []*UserApi
	err = m.QueryRowsNoCacheCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindByUserIdAndModelType 根据用户ID和模型类型查询所有模型
func (m *customUserApiModel) FindByUserIdAndModelType(ctx context.Context, userId string, modelType string) ([]*UserApi, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `user_id` = ? AND `model_type` = ? AND `status` = 1 ORDER BY `is_default` DESC, `id` DESC", userApiRows, m.table)

	var list []*UserApi
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, userId, modelType)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// UpdateDefaultStatus 更新默认状态 (事务)
func (m *customUserApiModel) UpdateDefaultStatus(ctx context.Context, userId string, modelType string, targetId int64) error {
	// 1. 获取目标模型信息 (用于校验和清除缓存)
	target, err := m.FindOne(ctx, uint64(targetId))
	if err != nil {
		return err
	}
	if target.UserId != userId {
		return fmt.Errorf("model does not belong to user")
	}

	// 2. 查找当前该类型下的默认模型 (用于清除缓存)
	var oldDefault UserApi
	// 假设存在多个默认(异常情况)，我们也只需要让它们失效，这里Limit 1取一个做缓存清除即可，或者更严谨一点不用查，直接DelCache
	// 为了简单起见，我们尽量清除已知的。
	queryOld := fmt.Sprintf("SELECT %s FROM %s WHERE `user_id` = ? AND `model_type` = ? AND `is_default` = 1 LIMIT 1", userApiRows, m.table)
	err = m.QueryRowNoCacheCtx(ctx, &oldDefault, queryOld, userId, modelType)
	var oldDefaultId int64 = 0
	if err == nil {
		oldDefaultId = int64(oldDefault.Id)
	}

	// 3. 准备需要清除的缓存键
	keys := make([]string, 0)
	// 目标模型缓存
	keys = append(keys, cacheUserApiIdPrefix+fmt.Sprint(targetId))
	keys = append(keys, fmt.Sprintf("%s%v:%v:%v", cacheUserApiUserIdModelTypeModelNamePrefix, target.UserId, target.ModelType, target.ModelName))

	// 旧默认模型缓存 (如果存在且不是同一个)
	if oldDefaultId != 0 && oldDefaultId != targetId {
		keys = append(keys, cacheUserApiIdPrefix+fmt.Sprint(oldDefaultId))
		keys = append(keys, fmt.Sprintf("%s%v:%v:%v", cacheUserApiUserIdModelTypeModelNamePrefix, oldDefault.UserId, oldDefault.ModelType, oldDefault.ModelName))
	}

	// 4. 执行事务
	err = m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 4.1 重置该用户该类型下的所有默认模型 -> 0
		resetQuery := fmt.Sprintf("UPDATE %s SET `is_default` = 0 WHERE `user_id` = ? AND `model_type` = ?", m.table)
		if _, err := session.ExecCtx(ctx, resetQuery, userId, modelType); err != nil {
			return err
		}

		// 4.2 设置目标模型 -> 1
		setQuery := fmt.Sprintf("UPDATE %s SET `is_default` = 1 WHERE `id` = ?", m.table)
		if _, err := session.ExecCtx(ctx, setQuery, targetId); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 5. 删除缓存
	// m.DelCacheCtx 是 defaultUserApiModel (sqlc.CachedConn) 的方法
	return m.DelCacheCtx(ctx, keys...)
}

// FindByIds 根据ID列表批量查询用户API配置
func (m *customUserApiModel) FindByIds(ctx context.Context, ids []uint64) ([]*UserApi, error) {
	if len(ids) == 0 {
		return []*UserApi{}, nil
	}

	// 构建 IN 查询的占位符
	placeholders := ""
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` IN (%s)", userApiRows, m.table, placeholders)

	var list []*UserApi
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...)
	if err != nil {
		return nil, err
	}

	return list, nil
}
