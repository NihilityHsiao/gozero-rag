package mq

import (
	"context"
)

// Mq 封装成接口，方便切换不同的实现
// 目前只实现了kafka,后期可以支持Redis等消息队列
type Mq interface {
	PublishGraphGenerateMsg(ctx context.Context, msg *GraphGenerateMsg) error
	PublishDocumentIndex(ctx context.Context, msg *KnowledgeDocumentIndexMsg) error
}
