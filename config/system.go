package config

type System struct {
	DbType          string `mapstructure:"db-type" json:"db-type" yaml:"db-type"`                                     // Database type:mysql(default)|sqlite|sqlserver|postgresql
	OssType         string `mapstructure:"oss-type" json:"oss-type" yaml:"oss-type"`                                  // Oss type
	RouterPrefix    string `mapstructure:"router-prefix" json:"router-prefix" yaml:"router-prefix"`                   // Router prefix
	Addr            int    `mapstructure:"addr" json:"addr" yaml:"addr"`                                              // Port value
	LimitCountIP    int    `mapstructure:"iplimit-count" json:"iplimit-count" yaml:"iplimit-count"`                   // IP limit count per hour
	LimitTimeIP     int    `mapstructure:"iplimit-time" json:"iplimit-time" yaml:"iplimit-time"`                      // IP limit time per hour
	UseMultipoint   bool   `mapstructure:"use-multipoint" json:"use-multipoint" yaml:"use-multipoint"`                // Multi-point login interception
	UseRedis        bool   `mapstructure:"use-redis" json:"use-redis" yaml:"use-redis"`                               // Using redis
	UseMongo        bool   `mapstructure:"use-mongo" json:"use-mongo" yaml:"use-mongo"`                               // Using MongoDB
	UseStrictAuth   bool   `mapstructure:"use-strict-auth" json:"use-strict-auth" yaml:"use-strict-auth"`             // Using tree role allocation mode
	SuperAdminKey   string `mapstructure:"super-admin-key" json:"super-admin-key" yaml:"super-admin-key"`             // Secret key required to register a new system admin
	Env             string `mapstructure:"env" json:"env" yaml:"env"`                                                 // Deployment environment: development|production
}
