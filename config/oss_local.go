package config

type Local struct {
	Path      string `mapstructure:"path" json:"path" yaml:"path"`                   // Local file access path
	StorePath string `mapstructure:"store-path" json:"store-path" yaml:"store-path"` // Local file storage path
	BaseURL   string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`       // Public base URL for local files (e.g. http://localhost:8888)

}
