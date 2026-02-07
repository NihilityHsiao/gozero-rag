package slicex

// Into 将切片中的每个元素通过转换函数转换成另一个类型
// 类似 Rust 的 into() 风格
//
// 用法示例:
//
//	newChunks := slicex.Into(chunks, func(chunk *schema.Document) *knowledge.KnowledgeDocumentChunk {
//	    metaBytes, _ := json.Marshal(chunk.MetaData)
//	    return &knowledge.KnowledgeDocumentChunk{
//	        Id:        uuid.Must(uuid.NewV7()).String(),
//	        ChunkText: chunk.Content,
//	        Metadata:  string(metaBytes),
//	    }
//	})
func Into[T any, R any](slice []T, fn func(T) R) []R {
	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

// IntoWithError 将切片中的每个元素通过转换函数转换成另一个类型，支持返回错误
// 如果任意一个转换失败，立即返回错误
func IntoWithError[T any, R any](slice []T, fn func(T) (R, error)) ([]R, error) {
	result := make([]R, len(slice))
	for i, item := range slice {
		r, err := fn(item)
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}

// Filter 过滤切片中的元素
func Filter[T any](slice []T, fn func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range slice {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

// Split 将切片分隔成多个子切片，每个子切片最多 n 个元素
//
// 用法示例:
//
//	batches := slicex.Split(entities, 36)
//	for _, batch := range batches {
//	    // 每批最多36个
//	}
func Split[T any](slice []T, n int) [][]T {
	if n <= 0 {
		n = 1
	}
	if len(slice) == 0 {
		return [][]T{}
	}

	numParts := (len(slice) + n - 1) / n // 向上取整
	result := make([][]T, 0, numParts)

	for i := 0; i < len(slice); i += n {
		end := i + n
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}

	return result
}
