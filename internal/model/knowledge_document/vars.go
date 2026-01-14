package knowledge_document

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var ErrNotFound = sqlx.ErrNotFound

type RunState = string

const (
	RunStatePending  = "pending"  // 待处理
	RunStateRunning  = "indexing" // 正在索引中
	RunStateSuccess  = "success"  // 处理成功
	RunStateFailed   = "failed"   // 处理失败
	RunStateCanceled = "canceled" // 主动取消
	RunStatePaused   = "paused"   // 暂停
)
