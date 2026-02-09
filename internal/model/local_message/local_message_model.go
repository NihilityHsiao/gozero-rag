package local_message

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ LocalMessageModel = (*customLocalMessageModel)(nil)

// 状态常量
const (
	StatusInit     = "init"
	StatusFail     = "fail"
	StatusSuccess  = "success"
	StatusRetrying = "retrying" // 补偿投递中，防止重复投递
)

// 任务类型常量
const (
	TaskTypeDocumentIndex = "document_index"
	TaskTypeGraphExtract  = "graph_extract"
)

type (
	// LocalMessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customLocalMessageModel.
	LocalMessageModel interface {
		localMessageModel
		withSession(session sqlx.Session) LocalMessageModel

		// 自定义方法
		FindRetryable(ctx context.Context, now int64, limit int) ([]*LocalMessage, error)
		UpdateSuccess(ctx context.Context, id uint64) error
		UpdateFail(ctx context.Context, id uint64, reason string, nextRetryTime int64) error
		UpdatePermanentFail(ctx context.Context, id uint64, reason string) error
		UpdateRetrying(ctx context.Context, id uint64) error
		InsertAndGetId(ctx context.Context, data *LocalMessage) (uint64, error)
	}

	customLocalMessageModel struct {
		*defaultLocalMessageModel
	}
)

// NewLocalMessageModel returns a model for the database table.
func NewLocalMessageModel(conn sqlx.SqlConn) LocalMessageModel {
	return &customLocalMessageModel{
		defaultLocalMessageModel: newLocalMessageModel(conn),
	}
}

func (m *customLocalMessageModel) withSession(session sqlx.Session) LocalMessageModel {
	return NewLocalMessageModel(sqlx.NewSqlConnFromSession(session))
}

// FindRetryable 查询需要重试的任务
// 包括：1) status=fail 且 next_retry_time <= now 2) status=init 且超过5分钟未更新
func (m *customLocalMessageModel) FindRetryable(ctx context.Context, now int64, limit int) ([]*LocalMessage, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM %s 
		WHERE deleted = 0 
		  AND retry_times < max_retry_times
		  AND (
			(status = ? AND next_retry_time <= ?) 
			OR (status = ? AND created_at < NOW() - INTERVAL 5 MINUTE)
		  )
		ORDER BY next_retry_time ASC
		LIMIT ?
	`, localMessageRows, m.table)

	var resp []*LocalMessage
	err := m.conn.QueryRowsCtx(ctx, &resp, query, StatusFail, now, StatusInit, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateSuccess 标记任务成功
func (m *customLocalMessageModel) UpdateSuccess(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("UPDATE %s SET status = ? WHERE id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, StatusSuccess, id)
	return err
}

// UpdateFail 标记任务失败并设置下次重试时间
func (m *customLocalMessageModel) UpdateFail(ctx context.Context, id uint64, reason string, nextRetryTime int64) error {
	query := fmt.Sprintf(`
		UPDATE %s SET 
			status = ?, 
			fail_reason = ?,
			next_retry_time = ?,
			retry_times = retry_times + 1,
		WHERE id = ?
	`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, StatusFail, reason, nextRetryTime, id)
	return err
}

// UpdatePermanentFail 标记任务永久失败
func (m *customLocalMessageModel) UpdatePermanentFail(ctx context.Context, id uint64, reason string) error {
	query := fmt.Sprintf(`
		UPDATE %s SET 
			status = ?,
			fail_reason = ?,
			retry_times = max_retry_times,
		WHERE id = ?
	`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, StatusFail, reason, id)
	return err
}

// UpdateRetrying 标记任务为补偿投递中，防止重复投递
func (m *customLocalMessageModel) UpdateRetrying(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("UPDATE %s SET status = ? WHERE id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, StatusRetrying, id)
	return err
}

// InsertAndGetId 插入并返回 ID
func (m *customLocalMessageModel) InsertAndGetId(ctx context.Context, data *LocalMessage) (uint64, error) {
	result, err := m.Insert(ctx, data)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

// NewLocalMessage 创建新的本地消息记录
func NewLocalMessage(taskType string, reqSnapshot string) *LocalMessage {
	return &LocalMessage{
		TaskType:      taskType,
		ReqSnapshot:   reqSnapshot,
		Status:        StatusInit,
		NextRetryTime: time.Now().Unix(),
		RetryTimes:    0,
		MaxRetryTimes: 3,
		FailReason:    sql.NullString{},
		Deleted:       0,
	}
}
