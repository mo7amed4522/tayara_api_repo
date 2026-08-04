package config

// Travel holds travel-marketplace business defaults. These are the deploy-time defaults that seed
// the runtime `base_system_setting` table on init; at runtime the System Settings value wins
// (see docs/ROADMAP.md — Phase 0). Nothing business-critical is hardcoded in Go.
//
// Percentages are **basis points** (1000 == 10%) to match the int64 money/percent convention.
type Travel struct {
	DefaultCurrency     string   `mapstructure:"default-currency" json:"default-currency" yaml:"default-currency"`
	SupportedCurrencies []string `mapstructure:"supported-currencies" json:"supported-currencies" yaml:"supported-currencies"`

	Commission   TravelCommission   `mapstructure:"commission" json:"commission" yaml:"commission"`
	SLA          TravelSLA          `mapstructure:"sla" json:"sla" yaml:"sla"`
	Payment      TravelPayment      `mapstructure:"payment" json:"payment" yaml:"payment"`
	Cancellation TravelCancellation `mapstructure:"cancellation" json:"cancellation" yaml:"cancellation"`
	Payout       TravelPayout       `mapstructure:"payout" json:"payout" yaml:"payout"`
}

// TravelCommission commission plan defaults (basis points; 1000 == 10%).
type TravelCommission struct {
	BasicPlanPercent    int64 `mapstructure:"basic-plan-percent" json:"basic-plan-percent" yaml:"basic-plan-percent"`
	PremiumPlanPercent  int64 `mapstructure:"premium-plan-percent" json:"premium-plan-percent" yaml:"premium-plan-percent"`
	AllowManualOverride bool  `mapstructure:"allow-manual-override" json:"allow-manual-override" yaml:"allow-manual-override"`
}

// TravelSLA response/payment windows + warning in hours.
type TravelSLA struct {
	AgentResponseHours    int  `mapstructure:"agent-response-hours" json:"agent-response-hours" yaml:"agent-response-hours"`
	UserPaymentHours      int  `mapstructure:"user-payment-hours" json:"user-payment-hours" yaml:"user-payment-hours"`
	WarningThresholdHours int  `mapstructure:"warning-threshold-hours" json:"warning-threshold-hours" yaml:"warning-threshold-hours"`
	AutoAlertOnBreach     bool `mapstructure:"auto-alert-on-breach" json:"auto-alert-on-breach" yaml:"auto-alert-on-breach"`
}

// TravelPayment Stripe hold/capture behavior.
type TravelPayment struct {
	HoldDurationHours       int  `mapstructure:"hold-duration-hours" json:"hold-duration-hours" yaml:"hold-duration-hours"`
	AutoCapture             bool `mapstructure:"auto-capture" json:"auto-capture" yaml:"auto-capture"`
	AutoRefundBeforeCapture bool `mapstructure:"auto-refund-before-capture" json:"auto-refund-before-capture" yaml:"auto-refund-before-capture"`
	PartialRefundEnabled    bool `mapstructure:"partial-refund-enabled" json:"partial-refund-enabled" yaml:"partial-refund-enabled"`
}

// TravelCancellation auto-cancel timers + force-cancel flag.
type TravelCancellation struct {
	AutoCancelPricePendingHours int  `mapstructure:"auto-cancel-price-pending-hours" json:"auto-cancel-price-pending-hours" yaml:"auto-cancel-price-pending-hours"`
	AutoCancelKYCPendingHours   int  `mapstructure:"auto-cancel-kyc-pending-hours" json:"auto-cancel-kyc-pending-hours" yaml:"auto-cancel-kyc-pending-hours"`
	AllowAdminForceCancel       bool `mapstructure:"allow-admin-force-cancel" json:"allow-admin-force-cancel" yaml:"allow-admin-force-cancel"`
}

// TravelPayout payout-gate defaults (MVP = manual, per booking).
type TravelPayout struct {
	ManualOnly               bool `mapstructure:"manual-only" json:"manual-only" yaml:"manual-only"`
	RequireAdminVerification bool `mapstructure:"require-admin-verification" json:"require-admin-verification" yaml:"require-admin-verification"`
	RequireConfirmationFile  bool `mapstructure:"require-confirmation-file" json:"require-confirmation-file" yaml:"require-confirmation-file"`
}
