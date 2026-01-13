package main

import (
	"context"
	"flag"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"gozero-rag/consumer/document_index/internal/config"
	"gozero-rag/consumer/document_index/internal/logic"
	"gozero-rag/consumer/document_index/internal/svc"
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

	logx.Infof("启动文档索引消费者")
	serviceGroup.Start()

}
