package config

type Email struct {
	To       string `mapstructure:"to" json:"to" yaml:"to"`                   // Recipients: multiple separated by English comma, e.g.: a@qq.com b@qq.com Please use this project as a parameter when developing
	From     string `mapstructure:"from" json:"from" yaml:"from"`             // Sender, the email you want to send an email to
	Host     string `mapstructure:"host" json:"host" yaml:"host"`             // Server address, e.g. smtp.qq.com Please go to QQ or the email you want to send an email to view its smtp protocol
	Secret   string `mapstructure:"secret" json:"secret" yaml:"secret"`       // A secret key used for login. It's best not to use your email password; instead, request a login key from your email address via SMTP.
	Nickname string `mapstructure:"nickname" json:"nickname" yaml:"nickname"` // Nickname, the nickname of the sender, usually their own email
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`             // Port, please go to QQ or the email you want to send an email to view its smtp protocol, mostly 465
	IsSSL    bool   `mapstructure:"is-ssl" json:"is-ssl" yaml:"is-ssl"`       // Whether SSL is enabled
}
