package metric

import "github.com/zeromicro/go-zero/core/metric"

const (
	UploadStatusSuccess = "success"
	UploadStatusFail    = "fail"
)

// 上传失败原因
const (
	FailReasonKbNotFound   = "kb_not_found"   // 知识库不存在
	FailReasonKbDisabled   = "kb_disabled"    // 知识库已禁用
	FailReasonFileTooLarge = "file_too_large" // 文件过大
	FailReasonInvalidType  = "invalid_type"   // 无效文件类型
	FailReasonSaveError    = "save_error"     // 保存失败
	FailReasonDbError      = "db_error"       // 数据库错误
	FailReasonParseError   = "parse_error"    // 解析失败
	FailReasonMqError      = "mq_error"       // 消息队列错误
)

var (
	// UploadTotal 上传总次数
	UploadTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "upload_total",
		Help:      "文档上传总次数",
		Labels:    []string{"status", "doc_type"},
	})

	// UploadBytesTotal 上传总字节数
	UploadBytesTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "upload_bytes_total",
		Help:      "文档上传总字节数",
		Labels:    []string{"doc_type"},
	})

	// UploadDuration 上传耗时分布
	UploadDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "upload_duration_seconds",
		Help:      "文档上传耗时分布(秒)",
		Labels:    []string{"doc_type"},
		Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
	})

	// UploadFailReason 上传失败原因统计
	UploadFailReason = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "rag",
		Subsystem: "knowledge",
		Name:      "upload_fail_reason_total",
		Help:      "上传失败原因统计",
		Labels:    []string{"reason"},
	})
)

// RecordUploadSuccess 记录上传成功
func RecordUploadSuccess(docType string, fileSize int64, durationSeconds float64) {
	UploadTotal.Inc(UploadStatusSuccess, docType)
	UploadBytesTotal.Add(float64(fileSize), docType)
	UploadDuration.Observe(int64(durationSeconds), docType)
}

// RecordUploadFail 记录上传失败
func RecordUploadFail(docType, reason string) {
	UploadTotal.Inc(UploadStatusFail, docType)
	UploadFailReason.Inc(reason)
}

// 知识库相关的metrics
