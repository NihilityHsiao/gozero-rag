package metric

import "github.com/zeromicro/go-zero/core/metric"

var (
	// UserLoginCount 只增不减的计数器
	UserLoginCount = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "rag",
		Subsystem: "user",
		Name:      "login_total",
		Help:      "用户登录次数",
		Labels:    []string{"status"}, // status: success/fail
	})

	ApiLatencyHistogram = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: "rag",
		Subsystem: "api",
		Name:      "request_duration_ms",
		Help:      "API请求延迟分布(毫秒)",
		Labels:    []string{"api_name"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000}, // 延迟桶
	})
)

func LoginSuccess() {
	UserLoginCount.Inc("success")
}
func LoginFail() {
	UserLoginCount.Inc("fail")
}
