package concurrentx

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestParallelProcess_Basic(t *testing.T) {
	// 准备测试数据
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// 回调: 每个数乘以 2
	callback := func(ctx context.Context, batch []int) ([]int, error) {
		result := make([]int, len(batch))
		for i, v := range batch {
			result[i] = v * 2
		}
		return result, nil
	}

	resultCh := ParallelProcess(context.Background(), items, ParallelProcessConfig{
		BatchSize: 3,
		Workers:   2,
	}, callback)

	// 收集结果
	var allResults []BatchResult[int]
	for r := range resultCh {
		allResults = append(allResults, r)
	}

	// 验证批次数
	expectedBatches := 4 // ceil(10/3)
	if len(allResults) != expectedBatches {
		t.Errorf("expected %d batches, got %d", expectedBatches, len(allResults))
	}

	// 验证无错误
	for _, r := range allResults {
		if r.Err != nil {
			t.Errorf("unexpected error in batch %d: %v", r.Index, r.Err)
		}
	}
}

func TestParallelProcess_CollectOrdered(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}

	callback := func(ctx context.Context, batch []string) ([]string, error) {
		result := make([]string, len(batch))
		for i, v := range batch {
			result[i] = v + "_processed"
		}
		return result, nil
	}

	resultCh := ParallelProcess(context.Background(), items, ParallelProcessConfig{
		BatchSize: 2,
		Workers:   3,
	}, callback)

	// 使用 CollectOrdered 收集并排序
	totalBatches := (len(items) + 1) / 2 // ceil(5/2) = 3
	merged, err := CollectOrdered(resultCh, totalBatches)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := []string{"a_processed", "b_processed", "c_processed", "d_processed", "e_processed"}
	if len(merged) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(merged))
	}

	for i, v := range expected {
		if merged[i] != v {
			t.Errorf("index %d: expected %s, got %s", i, v, merged[i])
		}
	}
}

func TestParallelProcess_WithError(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	expectedErr := errors.New("batch error")

	callback := func(ctx context.Context, batch []int) ([]int, error) {
		// 第二批返回错误
		if batch[0] == 3 {
			return nil, expectedErr
		}
		return batch, nil
	}

	resultCh := ParallelProcess(context.Background(), items, ParallelProcessConfig{
		BatchSize: 2,
		Workers:   1,
	}, callback)

	var foundError bool
	for r := range resultCh {
		if r.Err != nil {
			foundError = true
		}
	}

	if !foundError {
		t.Error("expected to find an error")
	}
}

func TestParallelProcess_Timeout(t *testing.T) {
	items := []int{1, 2, 3}

	callback := func(ctx context.Context, batch []int) ([]int, error) {
		select {
		case <-time.After(500 * time.Millisecond):
			return batch, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	resultCh := ParallelProcess(context.Background(), items, ParallelProcessConfig{
		BatchSize: 1,
		Workers:   2,
		Timeout:   100 * time.Millisecond, // 超时 100ms，但任务需要 500ms
	}, callback)

	var timeoutCount int
	for r := range resultCh {
		if r.Err == context.DeadlineExceeded {
			timeoutCount++
		}
	}

	if timeoutCount != 3 {
		t.Errorf("expected 3 timeouts, got %d", timeoutCount)
	}
}

func TestParallelProcess_EmptyInput(t *testing.T) {
	var items []int

	callback := func(ctx context.Context, batch []int) ([]int, error) {
		return batch, nil
	}

	resultCh := ParallelProcess(context.Background(), items, ParallelProcessConfig{
		BatchSize: 5,
		Workers:   2,
	}, callback)

	count := 0
	for range resultCh {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 results for empty input, got %d", count)
	}
}

func TestParallelProcess_Concurrency(t *testing.T) {
	items := make([]int, 20)
	for i := range items {
		items[i] = i
	}

	var mu sync.Mutex
	var maxConcurrent, current int

	callback := func(ctx context.Context, batch []int) ([]int, error) {
		mu.Lock()
		current++
		if current > maxConcurrent {
			maxConcurrent = current
		}
		mu.Unlock()

		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		current--
		mu.Unlock()

		return batch, nil
	}

	resultCh := ParallelProcess(context.Background(), items, ParallelProcessConfig{
		BatchSize: 2,
		Workers:   4, // 最多 4 个并发
	}, callback)

	for range resultCh {
	}

	if maxConcurrent > 4 {
		t.Errorf("max concurrency exceeded: expected <= 4, got %d", maxConcurrent)
	}

	// 应该有并发 (至少 > 1)
	if maxConcurrent < 2 {
		t.Logf("warning: concurrency lower than expected: %d", maxConcurrent)
	}
}

func TestSplitSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		size     int
		expected [][]int
	}{
		{
			name:     "normal split",
			input:    []int{1, 2, 3, 4, 5},
			size:     2,
			expected: [][]int{{1, 2}, {3, 4}, {5}},
		},
		{
			name:     "exact split",
			input:    []int{1, 2, 3, 4},
			size:     2,
			expected: [][]int{{1, 2}, {3, 4}},
		},
		{
			name:     "size larger than input",
			input:    []int{1, 2},
			size:     5,
			expected: [][]int{{1, 2}},
		},
		{
			name:     "empty input",
			input:    []int{},
			size:     3,
			expected: [][]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitSlice(tt.input, tt.size)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d batches, got %d", len(tt.expected), len(result))
			}
			for i, batch := range result {
				if len(batch) != len(tt.expected[i]) {
					t.Errorf("batch %d: expected len %d, got %d", i, len(tt.expected[i]), len(batch))
				}
			}
		})
	}
}

// ========================================
// ParallelProcessOrdered 测试
// ========================================

func TestParallelProcessOrdered_Basic(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	result, err := ParallelProcessOrdered(context.Background(), items, ParallelProcessConfig{
		BatchSize: 3,
		Workers:   2,
	}, func(ctx context.Context, batch []int) ([]int, error) {
		out := make([]int, len(batch))
		for i, v := range batch {
			out[i] = v * 2
		}
		return out, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}
	if len(result) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(result))
	}

	for i, v := range expected {
		if result[i] != v {
			t.Errorf("index %d: expected %d, got %d", i, v, result[i])
		}
	}
}

func TestParallelProcessOrdered_WithError(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	expectedErr := errors.New("batch error")

	_, err := ParallelProcessOrdered(context.Background(), items, ParallelProcessConfig{
		BatchSize: 2,
		Workers:   1,
	}, func(ctx context.Context, batch []int) ([]int, error) {
		if batch[0] == 3 {
			return nil, expectedErr
		}
		return batch, nil
	})

	if err == nil {
		t.Error("expected an error but got nil")
	}
}

func TestParallelProcessOrdered_EmptyInput(t *testing.T) {
	var items []int

	result, err := ParallelProcessOrdered(context.Background(), items, ParallelProcessConfig{
		BatchSize: 5,
		Workers:   2,
	}, func(ctx context.Context, batch []int) ([]int, error) {
		return batch, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

// ========================================
// 边界情况测试
// ========================================

func TestParallelProcess_ContextCancellation(t *testing.T) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	ctx, cancel := context.WithCancel(context.Background())

	var processedCount int
	var mu sync.Mutex

	resultCh := ParallelProcess(ctx, items, ParallelProcessConfig{
		BatchSize: 5,
		Workers:   2,
	}, func(ctx context.Context, batch []int) ([]int, error) {
		mu.Lock()
		processedCount++
		mu.Unlock()

		// 处理第 3 批后取消 context
		if processedCount >= 3 {
			cancel()
		}

		time.Sleep(50 * time.Millisecond)
		return batch, nil
	})

	// 消费所有结果
	for range resultCh {
	}

	// 不应该处理所有批次
	if processedCount >= 20 { // 100/5 = 20 batches
		t.Log("warning: context cancellation may not have stopped early enough")
	}
}

func TestParallelProcess_DefaultConfig(t *testing.T) {
	items := []int{1, 2, 3}

	// 使用零值配置，应该使用默认值
	result, err := ParallelProcessOrdered(context.Background(), items, ParallelProcessConfig{}, func(ctx context.Context, batch []int) ([]int, error) {
		return batch, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 results, got %d", len(result))
	}
}

func TestParallelProcess_SingleItem(t *testing.T) {
	items := []int{42}

	result, err := ParallelProcessOrdered(context.Background(), items, ParallelProcessConfig{
		BatchSize: 10,
		Workers:   4,
	}, func(ctx context.Context, batch []int) ([]int, error) {
		return batch, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 || result[0] != 42 {
		t.Errorf("expected [42], got %v", result)
	}
}

func TestParallelProcess_LargeDataset(t *testing.T) {
	// 测试大数据量
	const size = 1000
	items := make([]int, size)
	for i := range items {
		items[i] = i
	}

	result, err := ParallelProcessOrdered(context.Background(), items, ParallelProcessConfig{
		BatchSize: 50,
		Workers:   8,
	}, func(ctx context.Context, batch []int) ([]int, error) {
		out := make([]int, len(batch))
		for i, v := range batch {
			out[i] = v + 1
		}
		return out, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != size {
		t.Fatalf("expected %d results, got %d", size, len(result))
	}

	// 验证顺序正确
	for i, v := range result {
		if v != i+1 {
			t.Errorf("index %d: expected %d, got %d", i, i+1, v)
			break
		}
	}
}

func TestParallelProcess_PartialError(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6}

	resultCh := ParallelProcess(context.Background(), items, ParallelProcessConfig{
		BatchSize: 2,
		Workers:   1,
	}, func(ctx context.Context, batch []int) ([]int, error) {
		// 第二批 (3,4) 返回错误
		if batch[0] == 3 {
			return nil, errors.New("failed")
		}
		return batch, nil
	})

	var successCount, errorCount int
	for r := range resultCh {
		if r.Err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	if successCount != 2 {
		t.Errorf("expected 2 successful batches, got %d", successCount)
	}
	if errorCount != 1 {
		t.Errorf("expected 1 error batch, got %d", errorCount)
	}
}
