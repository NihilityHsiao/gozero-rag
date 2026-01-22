package config

import (
	commonconf "gozero-rag/internal/config"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
	KqConsumerConf kq.KqConf
	Cache          cache.CacheConf

	Mysql struct {
		DataSource string
	}
	Oss commonconf.OssConf

	ElasticSearch commonconf.ElasticSearchConf

	VectorStore commonconf.VectorStoreConf
}
