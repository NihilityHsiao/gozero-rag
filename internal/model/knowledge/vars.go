package knowledge

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound       = sqlx.ErrNotFound
	ErrOptimisticLock = errors.New("乐观锁冲突: 状态已被其他进程修改")
)
