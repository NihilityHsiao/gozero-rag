package mq

import (
	"context"
	"gozero-rag/internal/mq"
)

// Mq 封装成接口，方便切换不同的实现
// 目前只实现了kafka,后期可以支持Redis等消息队列
type Mq interface {
	PublishGraphGenerateMsg(ctx context.Context, msg *mq.GraphGenerateMsg) error
}
