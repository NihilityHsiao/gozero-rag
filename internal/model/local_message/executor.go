package local_message

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// Executor 本地消息表执行器
// 用于包装业务逻辑，实现零侵入的分布式一致性
type Executor struct {
	model LocalMessageModel
}

// BizFunc 业务函数类型
type BizFunc func(ctx context.Context) error

// RetryableMsg 可重试消息接口
type RetryableMsg interface {
	GetLocalMessageId() uint64
}

// NewExecutor 创建执行器
func NewExecutor(model LocalMessageModel) *Executor {
	return &Executor{model: model}
}

// Execute 执行业务逻辑并记录到本地消息表
// 流程：
// 1. 检查是否是补偿重试 (LocalMessageId > 0)
// 2. 新任务: 写入 local_message (status=init)
// 3. 执行 bizFunc
// 4. 成功 -> status=success
// 5. 失败 -> status=fail, 设置 next_retry_time (等待 Compensator 重试)
func (e *Executor) Execute(ctx context.Context, taskType string, reqSnapshot any, bizFunc BizFunc) error {
	var msgId uint64
	var isRetry bool

	// 检查是否是补偿重试任务
	if retryable, ok := reqSnapshot.(RetryableMsg); ok && retryable.GetLocalMessageId() > 0 {
		msgId = retryable.GetLocalMessageId()
		isRetry = true
		logx.Infof("[LocalMessage] 补偿重试任务: task_id=%d, type=%s", msgId, taskType)
	} else {
		// 新任务: 序列化请求快照并插入
		reqJson, err := json.Marshal(reqSnapshot)
		if err != nil {
			return fmt.Errorf("serialize req_snapshot failed: %w", err)
		}

		msg := NewLocalMessage(taskType, string(reqJson))
		msgId, err = e.model.InsertAndGetId(ctx, msg)
		if err != nil {
			// 插入失败，返回 error 让 Kafka 重试
			return fmt.Errorf("insert local_message failed: %w", err)
		}
	}

	// 执行业务逻辑
	if err := bizFunc(ctx); err != nil {
		// 业务失败 -> 更新状态为 fail，设置下次重试时间 (指数退避)
		var nextRetry int64
		if isRetry {
			// 重试任务: 获取当前 retry_times 计算退避时间
			task, _ := e.model.FindOne(ctx, msgId)
			if task != nil {
				nextRetry = time.Now().Add(time.Duration(30<<task.RetryTimes) * time.Second).Unix()
			} else {
				nextRetry = time.Now().Add(30 * time.Second).Unix()
			}
		} else {
			nextRetry = time.Now().Add(30 * time.Second).Unix()
		}

		if updateErr := e.model.UpdateFail(ctx, msgId, err.Error(), nextRetry); updateErr != nil {
			logx.Errorf("[LocalMessage] update fail status error: %v", updateErr)
		}
		// 返回 nil，不触发 Kafka 重试 (由 Compensator 处理)
		return nil
	}

	// 成功 -> 更新状态为 success
	if err := e.model.UpdateSuccess(ctx, msgId); err != nil {
		logx.Errorf("[LocalMessage] update success status error: %v", err)
	}

	return nil
}
