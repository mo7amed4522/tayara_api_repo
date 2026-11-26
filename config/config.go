package config

type Server struct {
	JWT       JWT     `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Zap       Zap     `mapstructure:"zap" json:"zap" yaml:"zap"`
	Redis     Redis   `mapstructure:"redis" json:"redis" yaml:"redis"`
	RedisList []Redis `mapstructure:"redis-list" json:"redis-list" yaml:"redis-list"`
	Mongo     Mongo   `mapstructure:"mongo" json:"mongo" yaml:"mongo"`
	Email     Email   `mapstructure:"email" json:"email" yaml:"email"`
	System    System  `mapstructure:"system" json:"system" yaml:"system"`
	Captcha   Captcha `mapstructure:"captcha" json:"captcha" yaml:"captcha"`
	// auto
	AutoCode Autocode `mapstructure:"autocode" json:"autocode" yaml:"autocode"`
	// gorm
	Mysql  Mysql           `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Mssql  Mssql           `mapstructure:"mssql" json:"mssql" yaml:"mssql"`
	Pgsql  Pgsql           `mapstructure:"pgsql" json:"pgsql" yaml:"pgsql"`
	Oracle Oracle          `mapstructure:"oracle" json:"oracle" yaml:"oracle"`
	Sqlite Sqlite          `mapstructure:"sqlite" json:"sqlite" yaml:"sqlite"`
	DBList []SpecializedDB `mapstructure:"db-list" json:"db-list" yaml:"db-list"`
	// oss
	Local        Local        `mapstructure:"local" json:"local" yaml:"local"`
	Qiniu        Qiniu        `mapstructure:"qiniu" json:"qiniu" yaml:"qiniu"`
	AliyunOSS    AliyunOSS    `mapstructure:"aliyun-oss" json:"aliyun-oss" yaml:"aliyun-oss"`
	HuaWeiObs    HuaWeiObs    `mapstructure:"hua-wei-obs" json:"hua-wei-obs" yaml:"hua-wei-obs"`
	TencentCOS   TencentCOS   `mapstructure:"tencent-cos" json:"tencent-cos" yaml:"tencent-cos"`
	AwsS3        AwsS3        `mapstructure:"aws-s3" json:"aws-s3" yaml:"aws-s3"`
	CloudflareR2 CloudflareR2 `mapstructure:"cloudflare-r2" json:"cloudflare-r2" yaml:"cloudflare-r2"`
	Minio        Minio        `mapstructure:"minio" json:"minio" yaml:"minio"`

	Excel Excel `mapstructure:"excel" json:"excel" yaml:"excel"`

	DiskList []DiskList `mapstructure:"disk-list" json:"disk-list" yaml:"disk-list"`

	// Cross-domain configuration
	Cors CORS `mapstructure:"cors" json:"cors" yaml:"cors"`

	// Task configuration
	Task   Task  `mapstructure:"task" json:"task" yaml:"task"`
	Merch1 Merch `mapstructure:"merch1" json:"merch1" yaml:"merch1"`
	Merch2 Merch `mapstructure:"merch2" json:"merch2" yaml:"merch2"`
	Merch3 Merch `mapstructure:"merch3" json:"merch3" yaml:"merch3"`

	OrderPay OrderPay `mapstructure:"orderPay" json:"orderPay" yaml:"orderPay"`
	Stripe   Stripe   `mapstructure:"stripe" json:"stripe" yaml:"stripe"`
	// Google OAuth
	Google Google `mapstructure:"google" json:"google" yaml:"google"`
	// Firebase Authentication (bxapi customer social/phone login)
	Firebase Firebase `mapstructure:"firebase" json:"firebase" yaml:"firebase"`
	// Travel marketplace business defaults (seed the runtime system_settings table)
	Travel Travel `mapstructure:"travel" json:"travel" yaml:"travel"`
	// Payment orchestrator config (BNPL providers + enabled payment methods)
	Payment Payment `mapstructure:"payment" json:"payment" yaml:"payment"`
	// Real-time support chat (Kafka transport, presence, last seen, attachments)
	Support SupportConfig `mapstructure:"support" json:"support" yaml:"support"`
}
type OrderPay struct {
	PayHost string `mapstructure:"payHost" json:"payHost" yaml:"payHost"`
}

/*
task:

	  payout:
		jobtime: 5  # Unit: minutes

merch1:

	userName: paymer01
	loginName: zhangsan
	loginpwd: 123456
	queryOrderUrl: http://baidu.com  # //Domain and port
	loginUrl: : http://baidu.com  # //Domain and port

merch2:

	userName: paymer02
	loginName: zhangsan
	loginpwd: 123456
	queryOrderUrl: http://baidu.com
	loginUrl: : http://baidu.com

merch3:

	userName: paymer03
	loginName: zhangsan
	loginpwd: 123456
	queryOrderUrl: http://baidu.com
	loginUrl: : http://baidu.com
*/

type Task struct {
	Payout Payout `mapstructure:"payout" json:"payout" yaml:"payout"`
}

type Payout struct {
	JobTime int `mapstructure:"jobtime" json:"jobtime" yaml:"jobtime"`
}

type Merch struct {
	UserName      string `mapstructure:"userName" json:"userName" yaml:"userName"`
	LoginName     string `mapstructure:"loginName" json:"loginName" yaml:"loginName"`
	LoginPwd      string `mapstructure:"loginpwd" json:"loginpwd" yaml:"loginpwd"`
	QueryOrderUrl string `mapstructure:"queryOrderUrl" json:"queryOrderUrl" yaml:"queryOrderUrl"`
	LoginUrl      string `mapstructure:"loginUrl" json:"loginUrl" yaml:"loginUrl"`
}
