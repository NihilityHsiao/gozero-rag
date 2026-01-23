package extractor

import (
	"context"
	"gozero-rag/internal/model/chunk"
	"os"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/joho/godotenv"
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

	extract, err1 := g.Extract(ctx, []*chunk.Chunk{
		{
			Id: "ch-1",
			Content: `Eino 基于明确的“组件”定义，提供强大的流程“编排”，覆盖开发全流程，旨在帮助开发者以最快的速度实现最有深度的大模型应用。

你是否曾有这种感受：想要为自己的应用添加大模型的能力，但面对这个较新的领域，不知如何入手；想持续的站在研究的最前沿，应用最新的业界成果，但使用的应用开发框架却已经数月没有更新；想看懂项目里的用 Python 写的代码，想确定一个变量或者参数的类型，需要反复查看上下文确认；不确定模型生成的效果是否足够好，想用又不太敢用；在调试、追踪、评测等开发之外的必要环节，还需要额外探索学习其他配套的工具。如果是，欢迎了解和尝试 Eino，因为 Eino 作为旨在覆盖 devops 全流程的大模型应用开发框架，具有如下特点：

内核稳定，API 简单易懂，有明确的上手路径，平滑的学习曲线。
极致的扩展性，研发工作高度活跃，长期可持续。
基于强类型语言 Golang，代码能看懂，易维护，高可靠。
背靠字节跳动核心业务线的充分实践经验。
提供开箱即用的配套工具。
Eino 已成为字节跳动内部大模型应用的首选全代码开发框架，已有包括豆包、抖音、扣子等多条业务线、数百个服务接入使用。`,
		}, {
			Id: "ch-2",
			Content: `在 Eino 编排场景中，每个组件成为了“节点”（Node），节点之间 1 对 1 的流转关系成为了“边”（Edge），N 选 1 的流转关系成为了“分支”（Branch）。基于 Eino 开发的应用，经过对各种组件的灵活编排，就像一支足球队可以采用各种阵型，能够支持无限丰富的业务场景。
足球队的战术千变万化，但却有迹可循，有的注重控球，有的简单直接。对 Eino 而言，针对不同的业务形态，也有更合适的编排方式`,
		},
	})
	if err1 != nil {
		t.Error(err1)
		return
	}
	t.Logf("extrac1 entities: %v", extract.Entities)

}
