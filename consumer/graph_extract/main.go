package main

import (
	"context"
	"flag"
	"gozero-rag/consumer/graph_extract/internal/config"
	"gozero-rag/consumer/graph_extract/internal/logic"
	"gozero-rag/consumer/graph_extract/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/conf.yaml", "the config file")

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

	logx.Infof("启动知识图谱提取消费者")
	serviceGroup.Start()
}
