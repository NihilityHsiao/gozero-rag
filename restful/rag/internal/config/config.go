// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	commonconf "gozero-rag/internal/config"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	Cache cache.CacheConf

	Mysql struct {
		DataSource string
	}
	Oss          commonconf.OssConf
	KqPusherConf struct {
		Brokers []string
		Topic   string
	}
	VectorStore   commonconf.VectorStoreConf
	ElasticSearch commonconf.ElasticSearchConf
}
