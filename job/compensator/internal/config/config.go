package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

type Config struct {
	Mysql struct {
		DataSource string
	}
	Cache cache.CacheConf

	// Kafka 配置
	KqPusherConf kq.KqConf

	// 补偿任务配置
	Compensator struct {
		Interval  int // 扫描间隔 (秒)
		BatchSize int // 每次扫描数量
	}
}
