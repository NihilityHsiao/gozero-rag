package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"gozero-rag/internal/xerr"
)

type KafkaMq struct {
	client *kq.Pusher
	topic  string
}

func NewKafka(client *kq.Pusher, topic string) *KafkaMq {
	return &KafkaMq{
		client: client,
		topic:  topic,
	}
}
func (k *KafkaMq) PublishDocumentIndex(ctx context.Context, msg *KnowledgeDocumentIndexMsg) error {
	if k.topic != TopicDocumentIndex {
		return xerr.NewInternalErrMsg("topic: " + k.topic + " , expected topic: " + TopicDocumentIndex)
	}

	if msg == nil {
		return xerr.NewErrCodeMsg(xerr.InternalError, "msg is nil")
	}

	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		logx.Errorf("kafka 反序列化文档索引失败:%v, msg:%v", err, msg)
		return xerr.NewErrCodeMsg(xerr.InternalError, fmt.Sprintf("marshal msg error:%v", err))
	}

	err = k.client.Push(ctx, string(data))
	if err != nil {
		logx.Errorf("kafka 发送文档索引失败:%v, msg:%v", err, msg)
		return xerr.NewErrCode(xerr.InternalError)
	}

	return nil
}

func (k *KafkaMq) PublishGraphGenerateMsg(ctx context.Context, msg *GraphGenerateMsg) error {
	if k.topic != TopicGraphExtract {
		return xerr.NewInternalErrMsg("topic: " + k.topic + " , expected topic: " + TopicGraphExtract)
	}
	if msg == nil {
		return xerr.NewErrCodeMsg(xerr.InternalError, "msg is nil")
	}

	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		logx.Errorf("kafka 反序列化文档索引失败:%v, msg:%v", err, msg)
		return xerr.NewErrCodeMsg(xerr.InternalError, fmt.Sprintf("marshal msg error:%v", err))
	}

	err = k.client.Push(ctx, string(data))
	if err != nil {
		logx.Errorf("kafka 发送文档索引失败:%v, msg:%v", err, msg)
		return xerr.NewErrCode(xerr.InternalError)
	}

	return nil
}
