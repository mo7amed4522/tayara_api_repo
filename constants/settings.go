package constants

// System-settings section keys (PRD §6.10 System Settings). Used to group the settings API
// response and mark which config belongs to which dashboard section.
const (
	SettingSectionCurrency     = "currency"
	SettingSectionCommission   = "commission"
	SettingSectionSLA          = "sla"
	SettingSectionPayment      = "payment"
	SettingSectionCancellation = "cancellation"
	SettingSectionPayout       = "payout"
)

// Well-known setting keys read by business logic. Values are stored as strings and typed on read.
const (
	SettingKeyDefaultCurrency    = "default_currency"
	SettingKeyBasicPlanPercent   = "basic_plan_percent"
	SettingKeyStripeFeePercent   = "stripe_fee_percent" // basis points; 230 == 2.3%
	SettingKeyAgentResponseHours = "agent_response_hours"
	SettingKeyUserPaymentHours   = "user_payment_hours"
)
