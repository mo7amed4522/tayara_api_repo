package config

type Stripe struct {
	SecretKey      string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key"`
	PublishableKey string `mapstructure:"publishable-key" json:"publishable-key" yaml:"publishable-key"`
	WebhookSecret  string `mapstructure:"webhook-secret" json:"webhook-secret" yaml:"webhook-secret"`
	Currency       string `mapstructure:"currency" json:"currency" yaml:"currency"`
	CaptureMethod  string `mapstructure:"capture-method" json:"capture-method" yaml:"capture-method"` // manual | automatic
	AutoCapture    bool   `mapstructure:"auto-capture" json:"auto-capture" yaml:"auto-capture"`
	PaymentHoldHours int  `mapstructure:"payment-hold-hours" json:"payment-hold-hours" yaml:"payment-hold-hours"`
}
