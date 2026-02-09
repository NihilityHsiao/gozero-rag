package main

import (
	"context"
	"flag"

	"github.com/joho/godotenv"
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

	// 加载 .env 文件 (尝试多个路径以支持不同的运行方式)
	_ = godotenv.Load(".env")       // 项目根目录运行
	_ = godotenv.Load("../../.env") // consumer/document_index 目录运行

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := svc.NewServiceContext(c)
	ctx := context.Background()
	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	for _, mq := range logic.Consumers(ctx, svcCtx) {
		serviceGroup.Add(mq)
	}

	logx.Infof("启动文档索引消费者")
	serviceGroup.Start()

}
