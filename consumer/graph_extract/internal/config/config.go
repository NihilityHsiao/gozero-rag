package config

import (
	commonconf "gozero-rag/internal/config"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

// NebulaConf Nebula 图数据库配置
type NebulaConf struct {
	Addresses []string // graphd 地址列表，如 ["192.168.0.6:49174", "192.168.0.6:49191"]
	Username  string
	Password  string
}

type Config struct {
	KqConsumerConf kq.KqConf
	Cache          cache.CacheConf

	Mysql struct {
		DataSource string
	}
	ElasticSearch commonconf.ElasticSearchConf
	Nebula        NebulaConf // 新增 Nebula 配置
}
