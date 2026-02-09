package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"gozero-rag/job/compensator/internal/config"
	"gozero-rag/job/compensator/internal/logic"
	"gozero-rag/job/compensator/internal/svc"
)

var configFile = flag.String("f", "etc/config.yaml", "config file")

func main() {
	flag.Parse()

	// 加载 .env 文件
	// 加载 .env 文件 (尝试多个路径以支持不同的运行方式)
	// GoLand 调试器可能使用项目根目录或模块目录作为工作目录
	_ = godotenv.Load(".env")       // 项目根目录运行
	_ = godotenv.Load("../../.env") //

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	svcCtx := svc.NewServiceContext(c)
	compensator := logic.NewCompensator(svcCtx)

	// 计算扫描间隔
	interval := c.Compensator.Interval
	if interval <= 0 {
		interval = 30
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	logx.Infof("[Compensator] 启动成功，扫描间隔: %d秒", interval)

	// 优雅退出
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logx.Info("[Compensator] 收到退出信号，正在停止...")
		cancel()
	}()

	// 主循环
	for {
		select {
		case <-ticker.C:
			compensator.Run(ctx)
		case <-ctx.Done():
			logx.Info("[Compensator] 已停止")
			return
		}
	}
}
