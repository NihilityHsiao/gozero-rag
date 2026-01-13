package agentic

// BoundaryInfo 包含边界判断和上下文信息
type BoundaryInfo struct {
	IsBoundary bool   `json:"is_boundary"` // 是否是边界
	Header     string `json:"header"`      // 该边界开始的新话题标题 (如果存在)
}
