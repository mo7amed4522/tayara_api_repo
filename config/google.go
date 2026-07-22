package config

type Google struct {
	ClientID     string   `mapstructure:"client-id" json:"client-id" yaml:"client-id"`
	ClientSecret string   `mapstructure:"client-secret" json:"client-secret" yaml:"client-secret"`
	RedirectURL  string   `mapstructure:"redirect-url" json:"redirect-url" yaml:"redirect-url"`
	Scopes       []string `mapstructure:"scopes" json:"scopes" yaml:"scopes"`
	Audiences    []string `mapstructure:"audiences" json:"audiences" yaml:"audiences"`
}
