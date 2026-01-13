package main

import (
	"context"
	"flag"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"gozero-rag/consumer/consumer_template/internal/config"
	"gozero-rag/consumer/consumer_template/internal/logic"
	"gozero-rag/consumer/consumer_template/internal/svc"
)

var configFile = flag.String("f", "etc/conf.yaml", "the config file")

func EvaluateInterviewRecord() {
	// todo: 读 kafka 中的任务, 生成面试评估结果
}

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := svc.NewServiceContext(c)
	ctx := context.Background()
	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()
	serviceGroup.Start()

	for _, mq := range logic.Consumers(ctx, svcCtx) {
		serviceGroup.Add(mq)
	}

	logx.Infof("启动面试评估消费者")
	serviceGroup.Start()

}
