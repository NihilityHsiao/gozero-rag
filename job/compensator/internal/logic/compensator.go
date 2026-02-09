package logic

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"

	"gozero-rag/internal/model/local_message"
	"gozero-rag/internal/mq"
	"gozero-rag/job/compensator/internal/svc"
)

// Compensator 补偿任务逻辑
// 只负责发现失败任务并重新投递到 Kafka，不执行重试逻辑
type Compensator struct {
	svcCtx              *svc.ServiceContext
	documentIndexPusher mq.Mq
	graphExtractPusher  mq.Mq
}

// NewCompensator 创建补偿任务
func NewCompensator(svcCtx *svc.ServiceContext) *Compensator {
	return &Compensator{
		svcCtx:              svcCtx,
		documentIndexPusher: mq.NewKafka(kq.NewPusher(svcCtx.Config.KqPusherConf.Brokers, mq.TopicDocumentIndex), mq.TopicDocumentIndex),
		graphExtractPusher:  mq.NewKafka(kq.NewPusher(svcCtx.Config.KqPusherConf.Brokers, mq.TopicGraphExtract), mq.TopicGraphExtract),
	}
}

// Run 执行补偿任务
func (c *Compensator) Run(ctx context.Context) {
	batchSize := c.svcCtx.Config.Compensator.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	// 查询需要重试的任务
	tasks, err := c.svcCtx.LocalMessageModel.FindRetryable(ctx, time.Now().Unix(), batchSize)
	if err != nil {
		logx.Errorf("[Compensator] 查询待重试任务失败: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	logx.Infof("[Compensator] 发现 %d 个待重试任务", len(tasks))

	for _, task := range tasks {
		c.pushToKafka(ctx, task)
	}
}

// pushToKafka 将任务重新投递到 Kafka
func (c *Compensator) pushToKafka(ctx context.Context, task *local_message.LocalMessage) {
	var err error

	switch task.TaskType {
	case local_message.TaskTypeDocumentIndex:
		err = c.pushDocumentIndex(ctx, task)
	case local_message.TaskTypeGraphExtract:
		err = c.pushGraphExtract(ctx, task)
	default:
		logx.Errorf("[Compensator] 未知任务类型: %s", task.TaskType)
		return
	}

	if err != nil {
		logx.Errorf("[Compensator] 投递 Kafka 失败: task_id=%d, type=%s, err=%v", task.Id, task.TaskType, err)
		return
	}

	// 投递成功，标记为 retrying 状态防止重复投递
	if updateErr := c.svcCtx.LocalMessageModel.UpdateRetrying(ctx, task.Id); updateErr != nil {
		logx.Errorf("[Compensator] 更新 retrying 状态失败: task_id=%d, err=%v", task.Id, updateErr)
	}

	logx.Infof("[Compensator] 任务已投递: task_id=%d, type=%s", task.Id, task.TaskType)
}

// pushDocumentIndex 投递到 document_index topic
func (c *Compensator) pushDocumentIndex(ctx context.Context, task *local_message.LocalMessage) error {
	var msg mq.KnowledgeDocumentIndexMsg
	if err := json.Unmarshal([]byte(task.ReqSnapshot), &msg); err != nil {
		return err
	}

	// 设置 LocalMessageId 标识这是补偿任务
	msg.SetLocalMessageId(task.Id)

	return c.documentIndexPusher.PublishDocumentIndex(ctx, &msg)
}

// pushGraphExtract 投递到 graph_extract topic
func (c *Compensator) pushGraphExtract(ctx context.Context, task *local_message.LocalMessage) error {
	var msg mq.GraphGenerateMsg
	if err := json.Unmarshal([]byte(task.ReqSnapshot), &msg); err != nil {
		return err
	}

	// 设置 LocalMessageId 标识这是补偿任务
	msg.SetLocalMessageId(task.Id)

	return c.graphExtractPusher.PublishGraphGenerateMsg(ctx, &msg)
}
