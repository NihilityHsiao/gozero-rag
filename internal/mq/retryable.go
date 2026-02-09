package mq

// RetryableMsg 可重试消息接口
type RetryableMsg interface {
	GetLocalMessageId() uint64
	SetLocalMessageId(id uint64)
}
