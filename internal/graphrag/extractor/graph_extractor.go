package extractor

import "gozero-rag/internal/model/chunk"

type GraphExtractor struct {
}

type GraphInput = []*chunk.Chunk

func NewGraphExtractor() (*GraphExtractor, error) {

	// todo: 流程编排

	return nil, nil
}

// Merge todo: 当一个文档提取完所有实体后,需要合并到整个知识库的图中,需要上分布式锁。
func Merge() {

}
