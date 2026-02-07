// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"

	"gozero-rag/internal/response"
	"gozero-rag/restful/rag/internal/config"
	"gozero-rag/restful/rag/internal/handler"
	"gozero-rag/restful/rag/internal/svc"
)

var configFile = flag.String("f", "etc/rag.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载 .env 文件 (尝试多个路径以支持不同的运行方式)
	// GoLand 调试器可能使用项目根目录或模块目录作为工作目录
	_ = godotenv.Load(".env")       // 项目根目录运行
	_ = godotenv.Load("../../.env") // restful/rag 目录运行

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 使用拦截器, 处理返回值
	httpx.SetOkHandler(response.OkHandler)
	httpx.SetErrorHandlerCtx(response.ErrHandler(c.Name))

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
