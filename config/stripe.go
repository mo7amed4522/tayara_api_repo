package config

type Stripe struct {
	SecretKey        string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key"`
	PublishableKey   string `mapstructure:"publishable-key" json:"publishable-key" yaml:"publishable-key"`
	WebhookSecret    string `mapstructure:"webhook-secret" json:"webhook-secret" yaml:"webhook-secret"`
	Currency         string `mapstructure:"currency" json:"currency" yaml:"currency"`
	CaptureMethod    string `mapstructure:"capture-method" json:"capture-method" yaml:"capture-method"` // manual | automatic
	AutoCapture      bool   `mapstructure:"auto-capture" json:"auto-capture" yaml:"auto-capture"`
	PaymentHoldHours int    `mapstructure:"payment-hold-hours" json:"payment-hold-hours" yaml:"payment-hold-hours"`
	// PaymentMethods are the card/wallet methods enabled on the PaymentIntent:
	// visa, mastercard, apple_pay, google_pay (mapped to Stripe payment_method_types card/apple_pay/google_pay).
	PaymentMethods []string `mapstructure:"payment-methods" json:"payment-methods" yaml:"payment-methods"`
	// FeeBPS is the platform's Stripe fee as basis points (e.g. 250 == 2.5%), applied on capture.
	FeeBPS int64 `mapstructure:"fee-bps" json:"fee-bps" yaml:"fee-bps"`
}
