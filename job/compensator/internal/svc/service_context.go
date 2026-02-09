package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"gozero-rag/internal/model/local_message"
	"gozero-rag/job/compensator/internal/config"
)

type ServiceContext struct {
	Config            config.Config
	LocalMessageModel local_message.LocalMessageModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:            c,
		LocalMessageModel: local_message.NewLocalMessageModel(sqlConn),
	}
}
