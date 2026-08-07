package base

import (
	"context"
	"strconv"
	"strings"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/base"
	"github.com/mo7amed4522/tayara_api/service/system"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Seed initializer that populates base reference data (countries, document types, system settings)
// on first boot. Tables themselves are created by AutoMigrate in initialize/gorm_biz.go, so the
// initializer only seeds data. Registered via init() so it runs inside the system initializer chain
// invoked by initialize.InitializeData().
const initOrderBase = system.InitOrderExternal - 3

type baseInit struct{}

func init() {
	system.RegisterInit(initOrderBase, &baseInit{})
}

func (b *baseInit) InitializerName() string { return "base_reference_data" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (b *baseInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (b *baseInit) TableCreated(ctx context.Context) bool                     { return true }

func (b *baseInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var c int64
	if err := db.Model(&base.Country{}).Count(&c).Error; err != nil {
		return false
	}
	if c == 0 {
		return false
	}
	var s int64
	return db.Model(&base.SystemSetting{}).Count(&s).Error == nil && s > 0
}

func (b *baseInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	if err := seedCountries(db); err != nil {
		return ctx, err
	}
	if err := seedDocumentTypes(db); err != nil {
		return ctx, err
	}
	if err := seedSystemSettings(db); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// seedCountries inserts a representative set, idempotent on the country code.
func seedCountries(db *gorm.DB) error {
	countries := []base.Country{
		{Code: "AE", NameEn: "United Arab Emirates", NameAr: "الإمارات العربية المتحدة", RequiresVisaFromUAE: false, IsActive: true},
		{Code: "TR", NameEn: "Turkey", NameAr: "تركيا", RequiresVisaFromUAE: false, IsActive: true},
		{Code: "US", NameEn: "United States", NameAr: "الولايات المتحدة", RequiresVisaFromUAE: true, IsActive: true},
		{Code: "GB", NameEn: "United Kingdom", NameAr: "المملكة المتحدة", RequiresVisaFromUAE: true, IsActive: true},
		{Code: "IN", NameEn: "India", NameAr: "الهند", RequiresVisaFromUAE: true, IsActive: true},
		{Code: "EG", NameEn: "Egypt", NameAr: "مصر", RequiresVisaFromUAE: false, IsActive: true},
		{Code: "TH", NameEn: "Thailand", NameAr: "تايلاند", RequiresVisaFromUAE: false, IsActive: true},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).
		Create(&countries).Error
}

// seedDocumentTypes inserts the platform document library, idempotent on the machine code.
func seedDocumentTypes(db *gorm.DB) error {
	docs := []base.DocumentType{
		{Code: "passport_copy", Name: "Passport copy", IsMandatoryDefault: true, IsActive: true},
		{Code: "emirates_id", Name: "Emirates ID copy", IsMandatoryDefault: false, IsActive: true},
		{Code: "bank_statement", Name: "Bank statement", IsMandatoryDefault: false, IsActive: true},
		{Code: "personal_photo", Name: "Personal photo", IsMandatoryDefault: true, IsActive: true},
		{Code: "contact_details", Name: "Contact details", IsMandatoryDefault: true, IsActive: true},
		{Code: "birth_certificate", Name: "Birth certificate", IsMandatoryDefault: false, IsActive: true},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).
		Create(&docs).Error
}

// systemSetting builds one SystemSetting row.
func systemSetting(section, key, value, description string) base.SystemSetting {
	return base.SystemSetting{Section: section, Key: key, Value: value, Description: description, IsActive: true}
}

// seedSystemSettings seeds the runtime settings from the `travel` config section
// (business numbers live here, never hardcoded in Go).
func seedSystemSettings(db *gorm.DB) error {
	t := global.GVA_CONFIG.Travel
	settings := []base.SystemSetting{
		systemSetting("currency", "default_currency", t.DefaultCurrency, "Default/major currency"),
		systemSetting("currency", "supported_currencies", strings.Join(t.SupportedCurrencies, ","), "Supported currencies (comma-separated)"),
		systemSetting("commission", "basic_plan_percent", strconv.FormatInt(t.Commission.BasicPlanPercent, 10), "Basic plan commission (basis points)"),
		systemSetting("commission", "premium_plan_percent", strconv.FormatInt(t.Commission.PremiumPlanPercent, 10), "Premium plan commission (basis points)"),
		systemSetting("commission", "allow_manual_override", strconv.FormatBool(t.Commission.AllowManualOverride), "Admin may override commission per booking"),
		systemSetting("commission", "stripe_fee_percent", "230", "Stripe fee (basis points, 230 == 2.3%), shown admin-only"),
		systemSetting("sla", "agent_response_hours", strconv.Itoa(t.SLA.AgentResponseHours), "Agent response SLA hours"),
		systemSetting("sla", "user_payment_hours", strconv.Itoa(t.SLA.UserPaymentHours), "User payment pay-window hours"),
		systemSetting("sla", "warning_threshold_hours", strconv.Itoa(t.SLA.WarningThresholdHours), "SLA warning threshold hours"),
		systemSetting("sla", "auto_alert_on_breach", strconv.FormatBool(t.SLA.AutoAlertOnBreach), "Auto-alert on SLA breach"),
		systemSetting("payment", "hold_duration_hours", strconv.Itoa(t.Payment.HoldDurationHours), "Stripe authorization hold hours"),
		systemSetting("payment", "auto_capture", strconv.FormatBool(t.Payment.AutoCapture), "Auto-capture on hold expiry"),
		systemSetting("payment", "auto_refund_before_capture", strconv.FormatBool(t.Payment.AutoRefundBeforeCapture), "Refund before capture allowed"),
		systemSetting("payment", "partial_refund_enabled", strconv.FormatBool(t.Payment.PartialRefundEnabled), "Partial refunds enabled"),
		systemSetting("cancellation", "auto_cancel_price_pending_hours", strconv.Itoa(t.Cancellation.AutoCancelPricePendingHours), "Auto-cancel price-pending hours"),
		systemSetting("cancellation", "auto_cancel_kyc_pending_hours", strconv.Itoa(t.Cancellation.AutoCancelKYCPendingHours), "Auto-cancel KYC-pending hours"),
		systemSetting("cancellation", "allow_admin_force_cancel", strconv.FormatBool(t.Cancellation.AllowAdminForceCancel), "Admin may force-cancel"),
		systemSetting("payout", "manual_only", strconv.FormatBool(t.Payout.ManualOnly), "Payouts are manual only (MVP)"),
		systemSetting("payout", "require_admin_verification", strconv.FormatBool(t.Payout.RequireAdminVerification), "Admin verification required"),
		systemSetting("payout", "require_confirmation_file", strconv.FormatBool(t.Payout.RequireConfirmationFile), "Agent confirmation file required"),
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key"}}, DoNothing: true}).
		Create(&settings).Error
}
