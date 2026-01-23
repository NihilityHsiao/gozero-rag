package extractor

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/joho/godotenv"
	"gozero-rag/internal/model/chunk"
	"os"
	"sync"
	"testing"
)

func TestNewGraphExtractor(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	ctx := context.Background()
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		Model:   os.Getenv("OPENAI_MODEL_NAME"),
	})
	if err != nil {
		t.Fatal(err)
	}

	g, err := NewGraphExtractor(ctx, llm)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		ctx1 := context.Background()
		extract, err1 := g.Extract(ctx1, []*chunk.Chunk{
			{
				Id: "go-1",
				DocId: `Eino 基于明确的“组件”定义，提供强大的流程“编排”，覆盖开发全流程，旨在帮助开发者以最快的速度实现最有深度的大模型应用。

你是否曾有这种感受：想要为自己的应用添加大模型的能力，但面对这个较新的领域，不知如何入手；想持续的站在研究的最前沿，应用最新的业界成果，但使用的应用开发框架却已经数月没有更新；想看懂项目里的用 Python 写的代码，想确定一个变量或者参数的类型，需要反复查看上下文确认；不确定模型生成的效果是否足够好，想用又不太敢用；在调试、追踪、评测等开发之外的必要环节，还需要额外探索学习其他配套的工具。如果是，欢迎了解和尝试 Eino，因为 Eino 作为旨在覆盖 devops 全流程的大模型应用开发框架，具有如下特点：

内核稳定，API 简单易懂，有明确的上手路径，平滑的学习曲线。
极致的扩展性，研发工作高度活跃，长期可持续。
基于强类型语言 Golang，代码能看懂，易维护，高可靠。
背靠字节跳动核心业务线的充分实践经验。
提供开箱即用的配套工具。
Eino 已成为字节跳动内部大模型应用的首选全代码开发框架，已有包括豆包、抖音、扣子等多条业务线、数百个服务接入使用。`,
			},
		})
		if err1 != nil {
			t.Error(err1)
			return
		}
		t.Logf("extrac1 entities: %v", extract.Entities)
	}()

	go func() {
		defer wg.Done()
		ctx1 := context.Background()
		extract, err1 := g.Extract(ctx1, []*chunk.Chunk{
			{
				Id: "go-2",
				DocId: `LangChain 是一个用于开发由大型语言模型 (LLMs) 驱动的应用程序的框架。

LangChain 简化了 LLM 应用程序生命周期的每个阶段：

开发：使用 LangChain 的开源 构建模块、组件 和 第三方集成 构建您的应用程序。 使用 LangGraph 构建具有一流流式处理和人机协作支持的有状态代理。
生产化：使用 LangSmith 检查、监控和评估您的链，以便您可以持续优化并自信地部署。
部署：将您的 LangGraph 应用程序转变为生产就绪的 API 和助手，使用 LangGraph Cloud。`,
			},
		})
		if err1 != nil {
			t.Error(err1)
			return
		}
		t.Logf("extrac2 entities: %v", extract.Entities)
	}()

	wg.Wait()
}
