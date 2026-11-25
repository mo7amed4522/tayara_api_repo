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

// PayPalProvider configures the PayPal Orders v2 integration (AUTHORIZE intent, manual capture).
// Credentials and the webhook id come from the environment via config.yaml — never committed.
type PayPalProvider struct {
	Enabled      bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	BaseURL      string `mapstructure:"base-url" json:"base-url" yaml:"base-url"` // must be https
	ClientID     string `mapstructure:"client-id" json:"client-id" yaml:"client-id"`
	ClientSecret string `mapstructure:"client-secret" json:"client-secret" yaml:"client-secret"`
	WebhookID    string `mapstructure:"webhook-id" json:"webhook-id" yaml:"webhook-id"` // signature verification
	Currency     string `mapstructure:"currency" json:"currency" yaml:"currency"`
	// SettlementCurrency is charged when the booking currency is not PayPal-supported (e.g. AED).
	SettlementCurrency string `mapstructure:"settlement-currency" json:"settlement-currency" yaml:"settlement-currency"`
	// FXRates maps a booking currency to how many settlement-currency units one unit is worth.
	// Admin-maintained, never client-supplied; the rate used is snapshotted on the payment.
	FXRates   map[string]float64 `mapstructure:"fx-rates" json:"fx-rates" yaml:"fx-rates"`
	BrandName string             `mapstructure:"brand-name" json:"brand-name" yaml:"brand-name"`
	ReturnURL string             `mapstructure:"return-url" json:"return-url" yaml:"return-url"`
	CancelURL string             `mapstructure:"cancel-url" json:"cancel-url" yaml:"cancel-url"`
}

// Payment is the orchestrator-level payment config: the BNPL providers, PayPal, plus the list of
// payment methods offered to the client (card, apple_pay, google_pay, tabby, tamara, paypal).
type Payment struct {
	Tabby          BNPLProvider   `mapstructure:"tabby" json:"tabby" yaml:"tabby"`
	Tamara         BNPLProvider   `mapstructure:"tamara" json:"tamara" yaml:"tamara"`
	PayPal         PayPalProvider `mapstructure:"paypal" json:"paypal" yaml:"paypal"`
	EnabledMethods []string       `mapstructure:"enabled-methods" json:"enabled-methods" yaml:"enabled-methods"`
}

// Enabled reports whether a provider is switched on.
func (b BNPLProvider) EnabledBool() bool { return b.Enabled }

// EnabledBool reports whether PayPal is switched on in the deploy-time config.
func (p PayPalProvider) EnabledBool() bool { return p.Enabled }
