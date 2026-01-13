package agentic

import (
	"context"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"gozero-rag/internal/rag_core/types"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgenticSplitter_Transform(t *testing.T) {
	err := godotenv.Load("../../../../.env")
	if err != nil {
		t.Fatal("Error loading .env file:", err.Error())
	}

	// 跳过测试如果没有设置 API Key
	apiKey := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseUrl := os.Getenv("OPENAI_BASE_URL")

	ctx := context.Background()
	ctx = context.WithValue(context.Background(), "split_config", types.ProcessConfig{
		Separators:     []string{"\n\n", "\n", "。"},
		MaxChunkLength: 500,
		ChunkOverlap:   100,
	})

	// 1. 初始化 ChatModel
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: baseUrl,
	})
	require.NoError(t, err)

	// 2. 创建 Splitter
	splitter := NewAgenticSplitter(llm)

	// 3. 准备输入文档
	docs := []*schema.Document{{
		ID: "test-doc",
		Content: `
凡人修仙传》/忘语

　　内容简介
　　一个普通的山村穷小子，偶然之下，进入到当地的江湖小门派，成了一名记名弟子。他以这样的身份，如何在门派中立足？又如何以平庸的资质，进入到修仙者的行列？和其他巨枭魔头，仙宗仙师并列于山海内外？从而笑傲三界之中！
　　
　　第一卷 七玄门风云 第一章 山边小村
　　	
　　二愣子睁大着双眼，直直望着茅草和烂泥糊成的黑屋顶，身上盖着的旧棉被，已呈深黄色，看不出原来的本来面目，还若有若无的散发着淡淡的霉味。
　　在他身边紧挨着的另一人，是二哥韩铸，酣睡的十分香甜，从他身上不时传来轻重不一的阵阵打呼声。
　　离床大约半丈远的地方，是一堵黄泥糊成的土墙，因为时间过久，墙壁上裂开了几丝不起眼的细长口子，从这些裂纹中，隐隐约约的传来韩母唠唠叨叨的埋怨声，偶尔还掺杂着韩父，抽旱烟杆的“啪嗒”“啪嗒”吸吮声。
　　二愣子缓缓的闭上已有些发涩的双目，迫使自己尽早进入深深的睡梦中。他心里非常清楚，再不老实入睡的话，明天就无法早起了，也就无法和其他约好的同伴一起进山拣干柴。

第一卷 七玄门风云 第二章 青牛镇
　　
　　这是一个小城，说是小城其实只是一个大点的镇子，名字也叫青牛镇，只有那些住在附近山沟里、没啥见识的土人，才“青牛城”“青牛城”的叫个不停。这是干了十几年门丁的张二心里话。
　　青牛镇的确不大，主街道只有一条东西方向的青牛街，连客栈也只有一家青牛客栈，客栈坐落在长条形状的镇子西端，所以过往的商客不想露宿野外的话，也只能住在这里。
　　现在有一辆一看就是赶了不少路的马车，从西边驶入青牛镇，飞快的驶过青牛客栈的大门前，停都不停，一直飞驰到镇子的另一端，春香酒楼的门口前，才停了下来。
　　春香酒楼不算大，甚至还有些陈旧，但却有一种古色古香的韵味。因为现在正是午饭时分，酒楼里用饭的客人还很多，几乎称得上是座无虚席。
　　从车上下来一个圆脸带着小胡子的胖男子和一个皮肤黝黑的、十来岁的小孩，男子带着孩童直接就大摇大摆地进了酒楼。有酒楼里的熟客认得胖子，知道他是这个酒楼的掌柜“韩胖子”，那个小孩是谁却无人认得。
　　“老韩，这个黑小子长的和你很像，不会是你背着家里婆娘生的儿子吧。”有个人突然打趣道。

第二卷 初踏修仙路 第一百章 嘉元城
　　
　　岚州是越国十三州中面积第八大的州府，但论富足程度却仅排在辛州之后，位列第二。它地处越国南部，土地肥沃，所辖域内又有数不清的水道、湖泊和运河，再加上一向风调雨顺，所以极为适合种植谷稻，是全国首屈一指的产粮大区。
　　而位于岚州中部的嘉元城，虽不是岚州府城，但却是货真价实的岚州第一大城。贯穿越国南北的乡鲁大运河就从此城中心穿过，再加上另外几条水陆干道也汇经此地，因此交通极为发达，可称得上是水运枢纽，商贸要道。每年从此经过的商户、旅人更是数不胜数，极大带动了此地的经贸活动，所以嘉元城成为全州第一大城，并不一件稀奇的事。
　　在嘉元城，大小车行、码头、船户极为繁多，遍布全城各处。从事这一行的车夫、苦力、船工更是多如牛毛，有数万人之多，孙二狗就是其中一位靠码头为生的人。
　　孙二狗人如其名，长的斜眉歪目，一副烂梨坏枣的痞子模样，不过因为擅长察言观色、溜须拍马，倒让他在码头上混成了一个帮派小头目，手下管着数十名苦力脚夫，靠帮过往商客搬运货物和行李为生。
　　因此当今日一早，孙二狗来到这小码头时，他的几名手下急忙凑了过来，恭敬的称呼道：“二爷早！”
　　“二爷来了！”
　　……
　　孙二狗听到这些称呼，人不禁有些飘飘然，毕竟能被人称呼一声“爷”，这也说明他在此地也算是个有身份的人物。因此他摆足了架子，从鼻子中哼了一下，就算是回应了这些手下的问候。
　　“什么二爷，不就是二狗吗？”

第四卷 风起海外 第三百六十四章 孤岛、巨舟
　　
　　“头好沉！”这是韩立清醒后的第一个感觉。
　　当他和曲魂在黄光中开始传送后，他只觉黄濛濛的四周蓦然出现了巨大的压力，但幸亏手中的大挪移令及时的发出了淡淡的青光，让其马上觉得压力全消。但他体内的那点灵力开始疯狂的流失到令牌中。
　　不过对此，韩立心里早有了准备，并没有多么惊慌，这些变化，他所看到的有关“大挪移令”的典籍中，都曾经提到过的。
　　而刹那间后，此法器就停止了吸取灵力，并且黄光消散，他和曲魂已经出现在了一个黒糊糊的地方。
　　光线太暗，韩立根本看不清四周的情形。但四周静悄悄的，应该没有其他人存在，这让韩立心里一松，一抬腿就要走出法阵。
　　但他一只脚刚刚迈出去，就觉得一阵天旋地转，双腿一软的坐在了地上，并差点难受的当场呕吐了起来。韩立知道，这是长距离传送后所造成的不适，而他会有这么大的反应，完全是因为他此时的修为太低了。
　　不过他现在顾不得此事，而是赶紧向曲魂下了破坏传送阵的命令。
　　只见曲魂面无表情的抽出他所给的银色巨剑，一剑剑的把传送阵的一角，砍得稀巴烂。
　　见此情景，韩立才正式放下心来。
`,
	}}

	// 4. 执行 Transform

	result, err := splitter.Transform(ctx, docs)
	require.NoError(t, err)

	// 5. 断言
	assert.GreaterOrEqual(t, len(result), 1, "应该至少返回一个 chunk")
	for i, doc := range result {
		t.Logf("Chunk %d (ID: %s) - len:%d - %s", i+1, doc.ID, utf8.RuneCountInString(doc.Content), doc.Content)
	}
}

func TestNewAgenticSplitterFromConfig(t *testing.T) {
	// 跳过测试如果没有设置 API Key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	ctx := context.Background()

	cfg := types.ProcessConfig{}

	splitter, err := NewAgenticSplitterFromConfig(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, splitter)
}

func TestAgenticSplitter_EmptyDocument(t *testing.T) {
	ctx := context.Background()

	// 使用 mock 或跳过
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey: apiKey,
		Model:  "gpt-4o-mini",
	})
	require.NoError(t, err)

	splitter := NewAgenticSplitter(llm)

	// 测试空文档
	docs := []*schema.Document{}
	result, err := splitter.Transform(ctx, docs)
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestAggregateWithMaxLength_ContextInjection(t *testing.T) {
	sentences := []string{
		"第一章 序言",     // 0
		"这是序言的内容。",   // 1
		"这是序言的第二句。",  // 2
		"第一章 正文",     // 3
		"这是第一章正文内容。", // 4
		"正文内容继续。",    // 5
		"第二章 结论",     // 6
		"结论部分很简短。",   // 7
	}

	// 模拟 LLM 返回的边界信息
	boundaries := []BoundaryInfo{
		{IsBoundary: false, Header: ""},      // 0-1
		{IsBoundary: false, Header: ""},      // 1-2
		{IsBoundary: true, Header: "第一章 正文"}, // 2-3: 这是一个边界，且 LLM 提取了新标题
		{IsBoundary: false, Header: ""},      // 3-4
		{IsBoundary: false, Header: ""},      // 4-5
		{IsBoundary: true, Header: "第二章 结论"}, // 5-6: 边界，新标题
		{IsBoundary: false, Header: ""},      // 6-7
	}

	// Case 1: 正常聚合，ChunkLength 足够大
	chunks := aggregateWithMaxLength(sentences, boundaries, 1000, 0)

	// 预期结果：
	// Chunk 0: 第一章 序言 这... (无Header上下文注入，或者是空Header)
	// Chunk 1: 【第一章 正文】\n第一章 正文...
	// Chunk 2: 【第二章 结论】\n第二章 结论...

	require.Len(t, chunks, 3)
	t.Logf("Chunk 0: %s", chunks[0])
	t.Logf("Chunk 1: %s", chunks[1])
	t.Logf("Chunk 2: %s", chunks[2])

	assert.Contains(t, chunks[1], "【第一章 正文】")
	assert.Contains(t, chunks[1], "这是第一章正文内容")
	assert.Contains(t, chunks[2], "【第二章 结论】")
	assert.Contains(t, chunks[2], "结论部分很简短")

	// Case 2: 强制切分（ChunkSize 很小），测试上下文继承
	// 假设 ChunkSize 只能容纳 1-2 句话
	chunksSplit := aggregateWithMaxLength(sentences, boundaries, 15, 0) // 10 chars very small

	t.Log("--- Case 2: Small Chunk Size ---")
	for i, c := range chunksSplit {
		t.Logf("Chunk %d: %s", i, c)
	}
	// 验证在长章节内部被强制切分出的 Chunk，是否也继承了 Header
	// 第4-5句属于"第一章 正文"，如果在它们之间切开，第5句所在的 chunk 也应该有 【第一章 正文】

	hasContext := false
	for _, c := range chunksSplit {
		if strings.Contains(c, "正文内容继续") {
			if strings.Contains(c, "【第一章 正文】") {
				hasContext = true
			}
		}
	}
	assert.True(t, hasContext, "强制切分的子块应该继承上一个 Header 上下文")
}

// truncate 截断字符串用于日志输出
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
