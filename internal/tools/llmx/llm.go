package llmx

import (
	"strings"
)

// GetModelNameFactory 解析唯一的模型标识符以提取模型名称和厂商。
// 标准格式: "ModelName@ModelFactory"
//
// 应用场景:
// 1. 云端模型: 例如 "gpt-4@openai" -> 名称: "gpt-4", 厂商: "openai"
// 2. 本地模型: 无 "@" 分隔符, 例如 "llama2-7b" -> 名称: "llama2-7b", 厂商: "Local"
func GetModelNameFactory(modelId string) (modelName string, modelFactory string) {
	if modelId == "" {
		return "", ""
	}

	// 使用 LastIndex 以支持模型名称中可能包含 '@' 的情况（虽然很少见）
	// 我们假设最后一个 '@' 用于分隔厂商。
	lastAt := strings.LastIndex(modelId, "@")

	// 情况: 本地模型或格式错误的 ID（无 '@' 分隔符）
	if lastAt == -1 {
		return modelId, "Local"
	}

	// 情况: 标准格式
	modelName = modelId[:lastAt]
	modelFactory = modelId[lastAt+1:]

	// 边界情况: "ModelName@" -> 厂商为空
	if modelFactory == "" {
		modelFactory = "Local"
	}

	// 边界情况: "@Factory" -> 名称为空？
	// 函数将返回 ("", "Factory"), 这在技术上是该字符串的正确解析。

	return modelName, modelFactory
}
