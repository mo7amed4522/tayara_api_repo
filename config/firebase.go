package config

type Firebase struct {
	Enabled         bool            `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	ProjectID       string          `mapstructure:"project-id" json:"project-id" yaml:"project-id"`
	CredentialsFile string          `mapstructure:"credentials-file" json:"credentials-file" yaml:"credentials-file"`
	APIKey          string          `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	AuthDomain      string          `mapstructure:"auth-domain" json:"auth-domain" yaml:"auth-domain"`
	StorageBucket   string          `mapstructure:"storage-bucket" json:"storage-bucket" yaml:"storage-bucket"`
	MessagingSender string          `mapstructure:"messaging-sender-id" json:"messaging-sender-id" yaml:"messaging-sender-id"`
	AppID           string          `mapstructure:"app-id" json:"app-id" yaml:"app-id"`
	Providers       FirebaseProviders `mapstructure:"providers" json:"providers" yaml:"providers"`
}

type FirebaseProviders struct {
	Google   bool `mapstructure:"google" json:"google" yaml:"google"`
	Apple    bool `mapstructure:"apple" json:"apple" yaml:"apple"`
	Facebook bool `mapstructure:"facebook" json:"facebook" yaml:"facebook"`
	Phone    bool `mapstructure:"phone" json:"phone" yaml:"phone"`
}
