package config

// VectorStoreConf 向量数据库配置
type VectorStoreConf struct {
	Type     string // milvus, qdrant, es
	Endpoint string // 连接地址，如 localhost:19530
	Username string // 用户名（可选）
	Password string // 密码（可选）
	Database string // 数据库名（Milvus 2.x 支持多数据库）
}
