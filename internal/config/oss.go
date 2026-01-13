package config

type OssConf struct {
	Type       string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}
