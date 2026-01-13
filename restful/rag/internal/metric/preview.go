package metric

import "github.com/zeromicro/go-zero/core/metric"

// Preview 状态常量
const (
	PreviewStatusSuccess = "success"
	PreviewStatusFail    = "fail"
)

// Preview 失败原因常量
const (
	PreviewFailReasonDocNotFound  = "doc_not_found" // 文档不存在
	PreviewFailReasonLoadError    = "load_error"    // 加载文件失败
	PreviewFailReasonSplitError   = "split_error"   // 分片失败
	PreviewFailReasonInvalidParam = "invalid_param" // 参数无效
	PreviewFailReasonServerError  = "server_error"  // 服务器内部错误
)

var (
	// PreviewTotal 预览请求总次数
	// 用途：监控预览功能的使用频率，区分成功/失败比例
	// 标签：status（success/fail）, doc_type（md/txt/pdf等）
	PreviewTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "preview_total",
		Help:      "文档预览请求总次数",
		Labels:    []string{"status", "doc_type"},
	})

	// PreviewDuration 预览耗时分布（毫秒）
	// 用途：监控预览接口的响应时间分布，用于发现性能问题
	// 优化后预期 P99 应在 500ms 以内
	PreviewDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "preview_duration_ms",
		Help:      "文档预览耗时分布(毫秒)",
		Labels:    []string{"doc_type"},
		// 针对预览场景优化的桶：快速响应 -> 慢响应
		Buckets: []float64{50, 100, 200, 500, 1000, 2000, 5000, 10000},
	})

	// PreviewContentLength 预览内容长度分布（字符数）
	// 用途：了解用户上传文档的大小分布，验证截断优化是否生效
	PreviewContentLength = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "preview_content_length",
		Help:      "预览内容原始长度分布(字符数)",
		Labels:    []string{"doc_type"},
		Buckets:   []float64{1000, 5000, 10000, 15000, 50000, 100000, 500000},
	})

	// PreviewChunkCount 预览生成的 chunk 数量分布
	// 用途：了解分片配置对 chunk 数量的影响，优化默认参数
	PreviewChunkCount = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "preview_chunk_count",
		Help:      "预览生成的chunk数量分布",
		Labels:    []string{"doc_type"},
		Buckets:   []float64{1, 3, 5, 10, 15, 20, 30, 50},
	})

	// PreviewContentTruncated 内容被截断的次数
	// 用途：监控有多少预览请求触发了内容截断优化
	PreviewContentTruncated = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "preview_content_truncated_total",
		Help:      "预览内容被截断的次数（触发优化）",
		Labels:    []string{"doc_type"},
	})

	// PreviewFailReason 预览失败原因统计
	// 用途：快速定位失败原因，优先处理高频错误
	PreviewFailReason = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "preview_fail_reason_total",
		Help:      "预览失败原因统计",
		Labels:    []string{"reason"},
	})
)

// RecordPreviewSuccess 记录预览成功
// 参数：
//   - docType: 文档类型 (md/txt/pdf)
//   - durationMs: 处理耗时（毫秒）
//   - originalContentLength: 原始内容长度（字符数）
//   - chunkCount: 生成的 chunk 数量
//   - wasTruncated: 内容是否被截断
func RecordPreviewSuccess(docType string, durationMs float64, originalContentLength int, chunkCount int, wasTruncated bool) {
	PreviewTotal.Inc(PreviewStatusSuccess, docType)
	PreviewDuration.Observe(int64(durationMs), docType)
	PreviewContentLength.Observe(int64(originalContentLength), docType)
	PreviewChunkCount.Observe(int64(chunkCount), docType)
	if wasTruncated {
		PreviewContentTruncated.Inc(docType)
	}
}

// RecordPreviewFail 记录预览失败
// 参数：
//   - docType: 文档类型 (md/txt/pdf)，如果未知则传空字符串
//   - reason: 失败原因（使用 PreviewFailReason* 常量）
func RecordPreviewFail(docType, reason string) {
	if docType == "" {
		docType = "unknown"
	}
	PreviewTotal.Inc(PreviewStatusFail, docType)
	PreviewFailReason.Inc(reason)
}
