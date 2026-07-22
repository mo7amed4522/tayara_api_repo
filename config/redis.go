package config

type Redis struct {
	Name         string   `mapstructure:"name" json:"name" yaml:"name"`                         // The name representing the current instance
	Addr         string   `mapstructure:"addr" json:"addr" yaml:"addr"`                         // Server address:port
	Password     string   `mapstructure:"password" json:"password" yaml:"password"`             // password
	DB           int      `mapstructure:"db" json:"db" yaml:"db"`                               // Which database in Redis is used in single-instance mode?
	UseCluster   bool     `mapstructure:"useCluster" json:"useCluster" yaml:"useCluster"`       // Should cluster mode be used?
	ClusterAddrs []string `mapstructure:"clusterAddrs" json:"clusterAddrs" yaml:"clusterAddrs"` // List of node addresses in cluster mode
}
