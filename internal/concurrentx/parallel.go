package concurrentx

import (
	"context"
	"sync"
	"time"
)

// DefaultTimeout 单个任务的默认超时时间
const DefaultTimeout = 120 * time.Second

// BatchResult 表示一个批次的处理结果
type BatchResult[Out any] struct {
	Index int   // 批次索引 (用于按顺序合并)
	Data  []Out // 处理结果
	Err   error // 错误信息
}

// ProcessFunc 处理函数类型
type ProcessFunc[In any, Out any] func(ctx context.Context, batch []In) ([]Out, error)

// ParallelProcessConfig 并发处理配置
type ParallelProcessConfig struct {
	BatchSize int           // 每批大小
	Workers   int           // 并发 worker 数量
	Timeout   time.Duration // 单个批次超时时间 (0 表示使用 DefaultTimeout)
}

// ParallelProcess 将 slice 分成多个批次，用多个 worker 并发处理
// 返回 channel 接收结果 (乱序)，调用方可根据 Index 排序
//
// 用法示例:
//
//	results := concurrent.ParallelProcess(ctx, texts, concurrent.ParallelProcessConfig{
//	    BatchSize: 10,
//	    Workers:   4,
//	    Timeout:   30 * time.Second,
//	}, func(ctx context.Context, batch []string) ([][]float64, error) {
//	    return embedder.EmbedStrings(ctx, batch)
//	})
//
//	for result := range results {
//	    if result.Err != nil {
//	        // 处理错误
//	    }
//	    // 使用 result.Data
//	}
func ParallelProcess[In any, Out any](
	ctx context.Context,
	items []In,
	config ParallelProcessConfig,
	callback ProcessFunc[In, Out],
) <-chan BatchResult[Out] {
	// 参数校验
	if config.BatchSize <= 0 {
		config.BatchSize = 1
	}
	if config.Workers <= 0 {
		config.Workers = 1
	}
	if config.Timeout <= 0 {
		config.Timeout = DefaultTimeout
	}

	// 分割批次
	batches := splitSlice(items, config.BatchSize)
	resultCh := make(chan BatchResult[Out], len(batches))

	if len(batches) == 0 {
		close(resultCh)
		return resultCh
	}

	// 任务队列
	type task struct {
		index int
		batch []In
	}
	taskCh := make(chan task, len(batches))

	// 启动 workers
	var wg sync.WaitGroup
	for i := 0; i < config.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range taskCh {
				// 为每个批次创建带超时的 context
				taskCtx, cancel := context.WithTimeout(ctx, config.Timeout)
				data, err := callback(taskCtx, t.batch)
				cancel()

				select {
				case resultCh <- BatchResult[Out]{Index: t.index, Data: data, Err: err}:
				case <-ctx.Done():
					resultCh <- BatchResult[Out]{Index: t.index, Err: ctx.Err()}
					return
				}
			}
		}()
	}

	// 分发任务
	go func() {
		defer func() {
			close(taskCh)
			wg.Wait()
			close(resultCh)
		}()
		for i, batch := range batches {
			select {
			case taskCh <- task{index: i, batch: batch}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultCh
}

// CollectOrdered 收集所有结果并按 Index 排序合并
// 返回合并后的结果和第一个错误 (如果有)
func CollectOrdered[Out any](resultCh <-chan BatchResult[Out], totalBatches int) ([]Out, error) {
	results := make([][]Out, totalBatches)
	var firstErr error

	for result := range resultCh {
		if result.Err != nil && firstErr == nil {
			firstErr = result.Err
		}
		if result.Index < totalBatches {
			results[result.Index] = result.Data
		}
	}

	// 合并结果
	var merged []Out
	for _, batch := range results {
		merged = append(merged, batch...)
	}

	return merged, firstErr
}

// ParallelProcessOrdered 并发处理并返回有序结果 (一步到位)
// 这是 ParallelProcess + CollectOrdered 的便捷封装
//
// 用法示例:
//
//	vectors, err := concurrentx.ParallelProcessOrdered(ctx, texts, concurrentx.ParallelProcessConfig{
//	    BatchSize: 10,
//	    Workers:   4,
//	}, func(ctx context.Context, batch []string) ([][]float64, error) {
//	    return embedder.EmbedStrings(ctx, batch)
//	})
func ParallelProcessOrdered[In any, Out any](
	ctx context.Context,
	items []In,
	config ParallelProcessConfig,
	callback ProcessFunc[In, Out],
) ([]Out, error) {
	if config.BatchSize <= 0 {
		config.BatchSize = 1
	}
	totalBatches := (len(items) + config.BatchSize - 1) / config.BatchSize

	resultCh := ParallelProcess(ctx, items, config, callback)
	return CollectOrdered(resultCh, totalBatches)
}

// splitSlice 将 slice 分成多个子切片，每个最多 n 个元素
func splitSlice[T any](slice []T, n int) [][]T {
	if n <= 0 {
		n = 1
	}
	if len(slice) == 0 {
		return [][]T{}
	}

	result := make([][]T, 0, (len(slice)+n-1)/n)
	for i := 0; i < len(slice); i += n {
		end := i + n
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}

	return result
}
