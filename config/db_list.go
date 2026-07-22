package config

import (
	"strings"

	"gorm.io/gorm/logger"
)

type DsnProvider interface {
	Dsn() string
}

// EmbededThe structure can be flattened to the next higher level, thus maintaining the original structure of the config file.
// See playground: https://go.dev/play/p/KIcuhqEoxmY

// GeneralDB is also used by Pgsql and Mysql as is
type GeneralDB struct {
	Prefix       string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`                         // Database prefix
	Port         string `mapstructure:"port" json:"port" yaml:"port"`                               // Database port
	Config       string `mapstructure:"config" json:"config" yaml:"config"`                         // Advanced configuration
	Dbname       string `mapstructure:"db-name" json:"db-name" yaml:"db-name"`                      // Database name
	Username     string `mapstructure:"username" json:"username" yaml:"username"`                   // Database username
	Password     string `mapstructure:"password" json:"password" yaml:"password"`                   // Database password
	Path         string `mapstructure:"path" json:"path" yaml:"path"`                               // Database address
	Engine       string `mapstructure:"engine" json:"engine" yaml:"engine" default:"InnoDB"`        // Database engine, default InnoDB
	LogMode      string `mapstructure:"log-mode" json:"log-mode" yaml:"log-mode"`                   // Whether to enable Gorm global logging
	MaxIdleConns int    `mapstructure:"max-idle-conns" json:"max-idle-conns" yaml:"max-idle-conns"` // Maximum number of idle connections
	MaxOpenConns int    `mapstructure:"max-open-conns" json:"max-open-conns" yaml:"max-open-conns"` // Maximum number of open connections to the database
	Singular     bool   `mapstructure:"singular" json:"singular" yaml:"singular"`                   // Whether to enable global disable plural, true means enable
	LogZap       bool   `mapstructure:"log-zap" json:"log-zap" yaml:"log-zap"`                      // Whether to write logs to the file using zap
}

func (c GeneralDB) LogLevel() logger.LogLevel {
	switch strings.ToLower(c.LogMode) {
	case "silent", "Silent":
		return logger.Silent
	case "error", "Error":
		return logger.Error
	case "warn", "Warn":
		return logger.Warn
	case "info", "Info":
		return logger.Info
	default:
		return logger.Info
	}
}

type SpecializedDB struct {
	Type      string `mapstructure:"type" json:"type" yaml:"type"`
	AliasName string `mapstructure:"alias-name" json:"alias-name" yaml:"alias-name"`
	GeneralDB `yaml:",inline" mapstructure:",squash"`
	Disable   bool `mapstructure:"disable" json:"disable" yaml:"disable"`
}
