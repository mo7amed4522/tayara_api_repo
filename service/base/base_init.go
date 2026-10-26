package base

import (
	"context"
	"strconv"
	"strings"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/base"
	"github.com/mo7amed4522/tayara_api/model/common"
	"github.com/mo7amed4522/tayara_api/service/system"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const initOrderBase = system.InitOrderExternal - 3

type baseInit struct{}

func init() {
	system.RegisterInit(initOrderBase, &baseInit{})
}

func (b *baseInit) InitializerName() string { return "base_reference_data" }

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
		// GCC / Middle East
		{Code: "AE", NameEn: "United Arab Emirates", NameAr: "الإمارات العربية المتحدة", RequiresVisaFromUAE: false, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"UAE", "United Arab Emirates", "Emirates", "AE", "Dubai", "Abu Dhabi", "Sharjah"}}},
		{Code: "SA", NameEn: "Saudi Arabia", NameAr: "المملكة العربية السعودية", RequiresVisaFromUAE: false, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Saudi Arabia", "KSA", "SA", "Riyadh", "Jeddah", "Mecca", "Medina"}}},
		{Code: "QA", NameEn: "Qatar", NameAr: "قطر", RequiresVisaFromUAE: false, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Qatar", "QA", "Doha"}}},
		{Code: "KW", NameEn: "Kuwait", NameAr: "الكويت", RequiresVisaFromUAE: false, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Kuwait", "KW", "Kuwait City"}}},
		{Code: "BH", NameEn: "Bahrain", NameAr: "البحرين", RequiresVisaFromUAE: false, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Bahrain", "BH", "Manama"}}},
		{Code: "OM", NameEn: "Oman", NameAr: "عمان", RequiresVisaFromUAE: false, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Oman", "OM", "Muscat"}}},
		{Code: "JO", NameEn: "Jordan", NameAr: "الأردن", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Jordan", "JO", "Amman", "Petra"}}},
		{Code: "LB", NameEn: "Lebanon", NameAr: "لبنان", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Lebanon", "LB", "Beirut"}}},
		{Code: "IQ", NameEn: "Iraq", NameAr: "العراق", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Iraq", "IQ", "Baghdad", "Erbil"}}},
		{Code: "SY", NameEn: "Syria", NameAr: "سوريا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Syria", "SY", "Damascus"}}},
		{Code: "YE", NameEn: "Yemen", NameAr: "اليمن", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Yemen", "YE", "Sanaa", "Aden"}}},
		{Code: "PS", NameEn: "Palestine", NameAr: "فلسطين", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Palestine", "PS", "Ramallah", "Gaza"}}},

		// North Africa
		{Code: "EG", NameEn: "Egypt", NameAr: "مصر", RequiresVisaFromUAE: false, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Egypt", "EG", "Misr", "Cairo", "Alexandria", "Sharm El Sheikh"}}},
		{Code: "MA", NameEn: "Morocco", NameAr: "المغرب", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Morocco", "MA", "Marrakesh", "Casablanca", "Rabat"}}},
		{Code: "TN", NameEn: "Tunisia", NameAr: "تونس", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Tunisia", "TN", "Tunis"}}},
		{Code: "DZ", NameEn: "Algeria", NameAr: "الجزائر", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Algeria", "DZ", "Algiers"}}},
		{Code: "LY", NameEn: "Libya", NameAr: "ليبيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Libya", "LY", "Tripoli"}}},
		{Code: "SD", NameEn: "Sudan", NameAr: "السودان", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Sudan", "SD", "Khartoum"}}},

		// Europe
		{Code: "GB", NameEn: "United Kingdom", NameAr: "المملكة المتحدة", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"UK", "United Kingdom", "GB", "Britain", "England", "Scotland", "London", "Manchester"}}},
		{Code: "FR", NameEn: "France", NameAr: "فرنسا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"France", "FR", "Paris", "Nice", "Lyon"}}},
		{Code: "DE", NameEn: "Germany", NameAr: "ألمانيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Germany", "DE", "Berlin", "Munich", "Frankfurt"}}},
		{Code: "IT", NameEn: "Italy", NameAr: "إيطاليا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Italy", "IT", "Rome", "Milan", "Venice"}}},
		{Code: "ES", NameEn: "Spain", NameAr: "إسبانيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Spain", "ES", "Madrid", "Barcelona"}}},
		{Code: "NL", NameEn: "Netherlands", NameAr: "هولندا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Netherlands", "NL", "Holland", "Amsterdam"}}},
		{Code: "CH", NameEn: "Switzerland", NameAr: "سويسرا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Switzerland", "CH", "Zurich", "Geneva"}}},
		{Code: "AT", NameEn: "Austria", NameAr: "النمسا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Austria", "AT", "Vienna"}}},
		{Code: "BE", NameEn: "Belgium", NameAr: "بلجيكا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Belgium", "BE", "Brussels"}}},
		{Code: "SE", NameEn: "Sweden", NameAr: "السويد", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Sweden", "SE", "Stockholm"}}},
		{Code: "NO", NameEn: "Norway", NameAr: "النرويج", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Norway", "NO", "Oslo"}}},
		{Code: "DK", NameEn: "Denmark", NameAr: "الدنمارك", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Denmark", "DK", "Copenhagen"}}},
		{Code: "FI", NameEn: "Finland", NameAr: "فنلندا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Finland", "FI", "Helsinki"}}},
		{Code: "PT", NameEn: "Portugal", NameAr: "البرتغال", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Portugal", "PT", "Lisbon", "Porto"}}},
		{Code: "GR", NameEn: "Greece", NameAr: "اليونان", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Greece", "GR", "Athens", "Santorini"}}},
		{Code: "IE", NameEn: "Ireland", NameAr: "أيرلندا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Ireland", "IE", "Dublin"}}},
		{Code: "PL", NameEn: "Poland", NameAr: "بولندا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Poland", "PL", "Warsaw"}}},
		{Code: "CZ", NameEn: "Czech Republic", NameAr: "التشيك", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Czech Republic", "Czechia", "CZ", "Prague"}}},
		{Code: "HU", NameEn: "Hungary", NameAr: "المجر", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Hungary", "HU", "Budapest"}}},
		{Code: "RO", NameEn: "Romania", NameAr: "رومانيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Romania", "RO", "Bucharest"}}},
		{Code: "RU", NameEn: "Russia", NameAr: "روسيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Russia", "RU", "Moscow", "St Petersburg"}}},
		{Code: "UA", NameEn: "Ukraine", NameAr: "أوكرانيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Ukraine", "UA", "Kyiv"}}},
		{Code: "TR", NameEn: "Turkey", NameAr: "تركيا", RequiresVisaFromUAE: false, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Turkey", "TR", "Türkiye", "Istanbul", "Antalya", "Ankara"}}},

		// Asia
		{Code: "IN", NameEn: "India", NameAr: "الهند", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"India", "IN", "Bharat", "Delhi", "Mumbai", "Bangalore", "Kerala"}}},
		{Code: "PK", NameEn: "Pakistan", NameAr: "باكستان", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Pakistan", "PK", "Karachi", "Lahore", "Islamabad"}}},
		{Code: "BD", NameEn: "Bangladesh", NameAr: "بنغلاديش", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Bangladesh", "BD", "Dhaka"}}},
		{Code: "LK", NameEn: "Sri Lanka", NameAr: "سريلانكا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Sri Lanka", "LK", "Colombo"}}},
		{Code: "NP", NameEn: "Nepal", NameAr: "نيبال", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Nepal", "NP", "Kathmandu"}}},
		{Code: "PH", NameEn: "Philippines", NameAr: "الفلبين", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Philippines", "PH", "Manila", "Cebu"}}},
		{Code: "ID", NameEn: "Indonesia", NameAr: "إندونيسيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Indonesia", "ID", "Jakarta", "Bali"}}},
		{Code: "MY", NameEn: "Malaysia", NameAr: "ماليزيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Malaysia", "MY", "Kuala Lumpur"}}},
		{Code: "SG", NameEn: "Singapore", NameAr: "سنغافورة", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Singapore", "SG"}}},
		{Code: "TH", NameEn: "Thailand", NameAr: "تايلاند", RequiresVisaFromUAE: false, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Thailand", "TH", "Bangkok", "Phuket", "Chiang Mai"}}},
		{Code: "VN", NameEn: "Vietnam", NameAr: "فيتنام", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Vietnam", "VN", "Hanoi", "Ho Chi Minh"}}},
		{Code: "CN", NameEn: "China", NameAr: "الصين", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"China", "CN", "Beijing", "Shanghai"}}},
		{Code: "JP", NameEn: "Japan", NameAr: "اليابان", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Japan", "JP", "Tokyo", "Osaka", "Kyoto"}}},
		{Code: "KR", NameEn: "South Korea", NameAr: "كوريا الجنوبية", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"South Korea", "Korea", "KR", "Seoul"}}},
		{Code: "AZ", NameEn: "Azerbaijan", NameAr: "أذربيجان", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Azerbaijan", "AZ", "Baku"}}},
		{Code: "GE", NameEn: "Georgia", NameAr: "جورجيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Georgia", "GE", "Tbilisi"}}},
		{Code: "AM", NameEn: "Armenia", NameAr: "أرمينيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Armenia", "AM", "Yerevan"}}},
		{Code: "KZ", NameEn: "Kazakhstan", NameAr: "كازاخستان", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Kazakhstan", "KZ", "Almaty", "Astana"}}},
		{Code: "UZ", NameEn: "Uzbekistan", NameAr: "أوزبكستان", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Uzbekistan", "UZ", "Tashkent"}}},

		// Americas
		{Code: "US", NameEn: "United States", NameAr: "الولايات المتحدة", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"USA", "United States", "US", "America", "New York", "Los Angeles", "Chicago", "Miami"}}},
		{Code: "CA", NameEn: "Canada", NameAr: "كندا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Canada", "CA", "Toronto", "Vancouver", "Montreal"}}},
		{Code: "MX", NameEn: "Mexico", NameAr: "المكسيك", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Mexico", "MX", "Mexico City", "Cancun"}}},
		{Code: "BR", NameEn: "Brazil", NameAr: "البرازيل", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Brazil", "BR", "Sao Paulo", "Rio de Janeiro"}}},
		{Code: "AR", NameEn: "Argentina", NameAr: "الأرجنتين", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Argentina", "AR", "Buenos Aires"}}},

		// Oceania
		{Code: "AU", NameEn: "Australia", NameAr: "أستراليا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Australia", "AU", "Sydney", "Melbourne"}}},
		{Code: "NZ", NameEn: "New Zealand", NameAr: "نيوزيلندا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"New Zealand", "NZ", "Auckland"}}},

		// Africa (sub-Saharan)
		{Code: "ZA", NameEn: "South Africa", NameAr: "جنوب أفريقيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"South Africa", "ZA", "Cape Town", "Johannesburg"}}},
		{Code: "KE", NameEn: "Kenya", NameAr: "كينيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Kenya", "KE", "Nairobi"}}},
		{Code: "NG", NameEn: "Nigeria", NameAr: "نيجيريا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Nigeria", "NG", "Lagos", "Abuja"}}},
		{Code: "ET", NameEn: "Ethiopia", NameAr: "إثيوبيا", RequiresVisaFromUAE: true, IsActive: true,
			Keywords: common.JSONMap{"keywords": []string{"Ethiopia", "ET", "Addis Ababa"}}},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).
		Create(&countries).Error
}

// seedDocumentTypes inserts the platform document library, idempotent on the machine code.
func seedDocumentTypes(db *gorm.DB) error {
	docs := []base.DocumentType{
		{Code: "passport_copy", Name: "Passport copy", IsMandatoryDefault: true, IsActive: true},
		{Code: "passport_front", Name: "Passport front page", IsMandatoryDefault: false, IsActive: true},
		{Code: "passport_back", Name: "Passport back page", IsMandatoryDefault: false, IsActive: true},
		{Code: "emirates_id", Name: "Emirates ID copy", IsMandatoryDefault: false, IsActive: true},
		{Code: "emirates_id_front", Name: "Emirates ID front", IsMandatoryDefault: false, IsActive: true},
		{Code: "emirates_id_back", Name: "Emirates ID back", IsMandatoryDefault: false, IsActive: true},
		{Code: "national_id", Name: "National ID copy", IsMandatoryDefault: false, IsActive: true},
		{Code: "residence_visa", Name: "Residence visa copy", IsMandatoryDefault: false, IsActive: true},
		{Code: "personal_photo", Name: "Personal photo", IsMandatoryDefault: true, IsActive: true},
		{Code: "contact_details", Name: "Contact details", IsMandatoryDefault: true, IsActive: true},
		{Code: "bank_statement", Name: "Bank statement", IsMandatoryDefault: false, IsActive: true},
		{Code: "salary_certificate", Name: "Salary certificate", IsMandatoryDefault: false, IsActive: true},
		{Code: "noc_letter", Name: "No-objection certificate (NOC)", IsMandatoryDefault: false, IsActive: true},
		{Code: "birth_certificate", Name: "Birth certificate", IsMandatoryDefault: false, IsActive: true},
		{Code: "marriage_certificate", Name: "Marriage certificate", IsMandatoryDefault: false, IsActive: true},
		{Code: "invitation_letter", Name: "Invitation letter", IsMandatoryDefault: false, IsActive: true},
		{Code: "hotel_booking", Name: "Hotel booking confirmation", IsMandatoryDefault: false, IsActive: true},
		{Code: "flight_ticket", Name: "Flight ticket / itinerary", IsMandatoryDefault: false, IsActive: true},
		{Code: "travel_insurance", Name: "Travel insurance", IsMandatoryDefault: false, IsActive: true},
		{Code: "vaccination_certificate", Name: "Vaccination certificate", IsMandatoryDefault: false, IsActive: true},
		{Code: "medical_report", Name: "Medical report", IsMandatoryDefault: false, IsActive: true},
		{Code: "police_clearance", Name: "Police clearance certificate", IsMandatoryDefault: false, IsActive: true},
		{Code: "employment_letter", Name: "Employment letter", IsMandatoryDefault: false, IsActive: true},
		{Code: "student_letter", Name: "Student enrollment letter", IsMandatoryDefault: false, IsActive: true},
		{Code: "proof_of_funds", Name: "Proof of funds", IsMandatoryDefault: false, IsActive: true},
		{Code: "previous_visa", Name: "Previous visa copy", IsMandatoryDefault: false, IsActive: true},
		{Code: "utility_bill", Name: "Utility bill (proof of address)", IsMandatoryDefault: false, IsActive: true},
		{Code: "tenancy_contract", Name: "Tenancy contract", IsMandatoryDefault: false, IsActive: true},
		{Code: "business_license", Name: "Business license", IsMandatoryDefault: false, IsActive: true},
		{Code: "company_letterhead", Name: "Company letterhead letter", IsMandatoryDefault: false, IsActive: true},
		{Code: "consent_form", Name: "Consent form", IsMandatoryDefault: false, IsActive: true},
		{Code: "minors_consent", Name: "Parental consent (minors)", IsMandatoryDefault: false, IsActive: true},
		{Code: "payout_receipt", Name: "Payout confirmation / receipt", IsMandatoryDefault: false, IsActive: true},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).
		Create(&docs).Error
}

func systemSetting(section, key, value, description string) base.SystemSetting {
	return base.SystemSetting{Section: section, Key: key, Value: value, Description: description, IsActive: true}
}

// seedSystemSettings seeds the runtime settings from the `travel` config section.
func seedSystemSettings(db *gorm.DB) error {
	t := global.GVA_CONFIG.Travel
	settings := []base.SystemSetting{
		// Currency
		systemSetting("currency", "default_currency", t.DefaultCurrency, "Default/major currency"),
		systemSetting("currency", "supported_currencies", strings.Join(t.SupportedCurrencies, ","), "Supported currencies (comma-separated)"),
		systemSetting("currency", "display_decimals", "2", "Currency display decimal places"),
		systemSetting("currency", "exchange_rate_refresh_hours", "6", "FX rate refresh cadence (hours)"),

		// Commission
		systemSetting("commission", "basic_plan_percent", strconv.FormatInt(t.Commission.BasicPlanPercent, 10), "Basic plan commission (basis points)"),
		systemSetting("commission", "premium_plan_percent", strconv.FormatInt(t.Commission.PremiumPlanPercent, 10), "Premium plan commission (basis points)"),
		systemSetting("commission", "allow_manual_override", strconv.FormatBool(t.Commission.AllowManualOverride), "Admin may override commission per booking"),
		systemSetting("commission", "stripe_fee_percent", "230", "Stripe fee (basis points, 230 == 2.3%), shown admin-only"),
		systemSetting("commission", "min_commission_amount", "0", "Minimum commission floor (minor units)"),
		systemSetting("commission", "rounding_mode", "nearest", "Commission rounding: nearest|up|down"),

		// SLA
		systemSetting("sla", "agent_response_hours", strconv.Itoa(t.SLA.AgentResponseHours), "Agent response SLA hours"),
		systemSetting("sla", "user_payment_hours", strconv.Itoa(t.SLA.UserPaymentHours), "User payment pay-window hours"),
		systemSetting("sla", "warning_threshold_hours", strconv.Itoa(t.SLA.WarningThresholdHours), "SLA warning threshold hours"),
		systemSetting("sla", "auto_alert_on_breach", strconv.FormatBool(t.SLA.AutoAlertOnBreach), "Auto-alert on SLA breach"),
		systemSetting("sla", "agent_kyc_review_hours", "24", "Agent KYC review SLA hours"),
		systemSetting("sla", "quote_response_hours", "12", "Agent quote response SLA hours"),

		// Payment
		systemSetting("payment", "hold_duration_hours", strconv.Itoa(t.Payment.HoldDurationHours), "Stripe authorization hold hours"),
		systemSetting("payment", "auto_capture", strconv.FormatBool(t.Payment.AutoCapture), "Auto-capture on hold expiry"),
		systemSetting("payment", "auto_refund_before_capture", strconv.FormatBool(t.Payment.AutoRefundBeforeCapture), "Refund before capture allowed"),
		systemSetting("payment", "partial_refund_enabled", strconv.FormatBool(t.Payment.PartialRefundEnabled), "Partial refunds enabled"),
		systemSetting("payment", "min_amount", "100", "Minimum charge amount (minor units)"),
		systemSetting("payment", "max_amount", "5000000", "Maximum charge amount (minor units)"),
		systemSetting("payment", "allowed_methods", "card,apple_pay,google_pay", "Allowed payment methods"),
		systemSetting("payment", "three_ds_enabled", "true", "3-D Secure required"),
		systemSetting("payment", "retry_attempts", "3", "Max payment retry attempts"),

		// Cancellation
		systemSetting("cancellation", "auto_cancel_price_pending_hours", strconv.Itoa(t.Cancellation.AutoCancelPricePendingHours), "Auto-cancel price-pending hours"),
		systemSetting("cancellation", "auto_cancel_kyc_pending_hours", strconv.Itoa(t.Cancellation.AutoCancelKYCPendingHours), "Auto-cancel KYC-pending hours"),
		systemSetting("cancellation", "allow_admin_force_cancel", strconv.FormatBool(t.Cancellation.AllowAdminForceCancel), "Admin may force-cancel"),
		systemSetting("cancellation", "user_cancel_window_hours", "48", "User self-cancel window before departure (hours)"),
		systemSetting("cancellation", "fee_percent_before_window", "0", "Cancellation fee % before window"),
		systemSetting("cancellation", "fee_percent_after_window", "1000", "Cancellation fee % after window (basis points)"),

		// Payout
		systemSetting("payout", "manual_only", strconv.FormatBool(t.Payout.ManualOnly), "Payouts are manual only (MVP)"),
		systemSetting("payout", "require_admin_verification", strconv.FormatBool(t.Payout.RequireAdminVerification), "Admin verification required"),
		systemSetting("payout", "require_confirmation_file", strconv.FormatBool(t.Payout.RequireConfirmationFile), "Agent confirmation file required"),
		systemSetting("payout", "min_payout_amount", "10000", "Minimum payout amount (minor units)"),
		systemSetting("payout", "payout_schedule", "weekly", "Payout schedule: daily|weekly|monthly"),
		systemSetting("payout", "settlement_days", "7", "Days after trip to release payout"),

		// Booking
		systemSetting("booking", "min_notice_hours", "24", "Minimum lead time to book (hours)"),
		systemSetting("booking", "max_future_days", "365", "Max days ahead a booking can be made"),
		systemSetting("booking", "allow_multiple_agents_quote", "true", "Allow multiple agents to quote one request"),
		systemSetting("booking", "quote_expiry_minutes", "60", "Quote validity window (minutes)"),
		systemSetting("booking", "max_passengers_per_booking", "20", "Max passengers per single booking"),
		systemSetting("booking", "require_kyc_to_book", "true", "User must complete KYC to book"),

		// KYC
		systemSetting("kyc", "user_required", "true", "Users must complete KYC"),
		systemSetting("kyc", "agent_required", "true", "Agents must complete KYC"),
		systemSetting("kyc", "doc_expiry_warning_days", "30", "Warn when ID docs expire within N days"),
		systemSetting("kyc", "auto_approve_threshold", "0", "Score threshold for auto-approval (0=off)"),
		systemSetting("kyc", "liveness_check_enabled", "true", "Enable liveness check"),

		// Notification
		systemSetting("notification", "email_enabled", "true", "Email notifications enabled"),
		systemSetting("notification", "sms_enabled", "false", "SMS notifications enabled"),
		systemSetting("notification", "push_enabled", "true", "Push notifications enabled"),
		systemSetting("notification", "quiet_hours_start", "22", "Quiet hours start (local hour)"),
		systemSetting("notification", "quiet_hours_end", "7", "Quiet hours end (local hour)"),

		// Security
		systemSetting("security", "max_login_attempts", "5", "Max failed login attempts"),
		systemSetting("security", "lockout_minutes", "30", "Account lockout duration (minutes)"),
		systemSetting("security", "session_ttl_minutes", "1440", "Session lifetime (minutes)"),
		systemSetting("security", "password_min_length", "8", "Minimum password length"),
		systemSetting("security", "require_2fa_admins", "true", "Admins must use 2FA"),

		// System / feature flags
		systemSetting("system", "maintenance_mode", "false", "Global maintenance mode"),
		systemSetting("system", "support_email", "support@tayara.example", "Support contact email"),
		systemSetting("system", "support_phone", "", "Support contact phone"),
		systemSetting("system", "terms_version", "1.0", "Terms & conditions version"),
		systemSetting("system", "privacy_version", "1.0", "Privacy policy version"),
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key"}}, DoNothing: true}).
		Create(&settings).Error
}
