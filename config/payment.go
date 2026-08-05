package config

// BNPLProvider configures a Buy-Now-Pay-Later aggregator (Tabby / Tamara).
// Secrets are loaded from the environment in config.yaml — never committed.
type BNPLProvider struct {
	Enabled       bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	BaseURL       string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	SecretKey     string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key"`
	Currency      string `mapstructure:"currency" json:"currency" yaml:"currency"`
	WebhookSecret string `mapstructure:"webhook-secret" json:"webhook-secret" yaml:"webhook-secret"`
}

// Payment is the orchestrator-level payment config: the BNPL providers plus the list of
// payment methods offered to the client (card, apple_pay, google_pay, tabby, tamara).
type Payment struct {
	Tabby          BNPLProvider `mapstructure:"tabby" json:"tabby" yaml:"tabby"`
	Tamara         BNPLProvider `mapstructure:"tamara" json:"tamara" yaml:"tamara"`
	EnabledMethods []string     `mapstructure:"enabled-methods" json:"enabled-methods" yaml:"enabled-methods"`
}

// Enabled reports whether a provider is switched on.
func (b BNPLProvider) EnabledBool() bool { return b.Enabled }
