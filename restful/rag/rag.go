// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpx"
	"gozero-rag/internal/response"

	"gozero-rag/restful/rag/internal/config"
	"gozero-rag/restful/rag/internal/handler"
	"gozero-rag/restful/rag/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/rag.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

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
