package logic

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"gozero-rag/consumer/consumer_template/internal/svc"
	"gozero-rag/internal/mq"
)

type DemoLogic struct {
	svcCtx *svc.ServiceContext
	ctx    context.Context
}

func NewDemoConsumerLogic(svcCtx *svc.ServiceContext, ctx context.Context) *DemoLogic {
	return &DemoLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
	}
}

func (l *DemoLogic) Consume(_ context.Context, key, val string) (err error) {
	var msg *mq.KnowledgeDocumentIndexMsg
	// 模拟发送topic:
	// 进入Kafka的容器,执行: kafka-console-producer --bootstrap-server localhost:9092 --topic prod.interview.evaluation.generate
	// 终端输入: {"user_id": 1, "resume_id": 1, "record_id": 49}

	err = json.Unmarshal([]byte(val), &msg)
	if err != nil {
		logx.Errorf("消费 val: %s, 失败: %v", val, err)
		return err
	}

	logx.Infof("消费 msg: %v", msg)
	return l.mainWork(l.ctx, msg)
}

func (l *DemoLogic) mainWork(ctx context.Context, msg *mq.KnowledgeDocumentIndexMsg) error {
	panic("implement me")
}
