// Package catalog seeds packages (with their full child entities) and custom-trip templates so
// the public catalog, agent Manage dashboard and admin moderation all have varied data.
package catalog

import (
	"context"
	"time"

	catmodel "github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/common"
	"github.com/mo7amed4522/tayara_api/service/system"
	"gorm.io/gorm"
)

const initOrderCatalog = system.InitOrderExternal + 5

type catalogInit struct{}

func init() {
	system.RegisterInit(initOrderCatalog, &catalogInit{})
}

func (c *catalogInit) InitializerName() string { return "demo_catalog" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (c *catalogInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (c *catalogInit) TableCreated(ctx context.Context) bool                     { return true }

func (c *catalogInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var n int64
	return db.Model(&catmodel.Package{}).Count(&n).Error == nil && n > 0
}

func (c *catalogInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	packages, err := seedPackages(db)
	if err != nil {
		return ctx, err
	}
	if err := seedPackageChildren(db, packages); err != nil {
		return ctx, err
	}
	if err := seedCustomTripTemplates(db); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// agentIDByEmail resolves the numeric agent ID for a business email (seeded by source/dxapi).
func agentIDByEmail(db *gorm.DB, email string) uint {
	var a struct {
		ID uint
	}
	if err := db.Table("dx_agent_profile").Select("id").Where("business_email = ?", email).Scan(&a).Error; err != nil {
		return 0
	}
	return a.ID
}

// countryByCode resolves the numeric country ID (seeded by service/base).
func countryByCode(db *gorm.DB, code string) uint {
	var c struct {
		ID uint
	}
	if err := db.Table("base_country").Select("id").Where("code = ?", code).Scan(&c).Error; err != nil {
		return 0
	}
	return c.ID
}

// docTypeID resolves a document library row by machine code (seeded by service/base).
func docTypeID(db *gorm.DB, code string) uint {
	var d struct {
		ID uint
	}
	if err := db.Table("base_document_type").Select("id").Where("code = ?", code).Scan(&d).Error; err != nil {
		return 0
	}
	return d.ID
}

func seedPackages(db *gorm.DB) ([]catmodel.Package, error) {
	agent := agentIDByEmail(db, "hello@desertsafari.ae")
	agent2 := agentIDByEmail(db, "book@luxurytravel.ae")
	now := time.Now()
	submitted := now.Add(-240 * time.Hour)

	packages := []catmodel.Package{
		{
			PackageCode: "PK-1001", AgentID: agent, Type: "outbound",
			Name: "Istanbul Highlights", CoverImageURL: "https://files.tayara.test/pk/istanbul.jpg",
			Description: "5 days in Istanbul: old city, Bosphorus cruise, culinary tour.",
			GalleryImages: common.JSONMap{"images": []string{"https://files.tayara.test/pk/ist1.jpg", "https://files.tayara.test/pk/ist2.jpg"}},
			DestinationCountryID: countryByCode(db, "TR"), Location: "Istanbul", Category: "cultural",
			DaysDuration: 5, AvailabilityType: "open_calendar", DisplayPrice: 320000, EstimatedPrice: 290000,
			DiscountPercent: 400, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
		},
		{
			PackageCode: "PK-1002", AgentID: agent, Type: "inbound",
			Name: "Dubai City Break", CoverImageURL: "https://files.tayara.test/pk/dubai.jpg",
			Description: "3 nights in Dubai with Burj Khalifa, desert safari and mall tour.",
			DestinationCountryID: countryByCode(db, "AE"), Location: "Dubai", Category: "beach",
			DaysDuration: 4, AvailabilityType: "fixed_dates", DisplayPrice: 210000, EstimatedPrice: 195000,
			DiscountPercent: 500, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
		},
		{
			PackageCode: "PK-1003", AgentID: agent2, Type: "outbound",
			Name: "Bali Adventure", CoverImageURL: "https://files.tayara.test/pk/bali.jpg",
			Description: "Surf, volcanoes and temples across Bali in 7 days.",
			DestinationCountryID: countryByCode(db, "TH"), Location: "Bali", Category: "adventure",
			DaysDuration: 7, AvailabilityType: "open_calendar", DisplayPrice: 450000, EstimatedPrice: 430000,
			DiscountPercent: 300, Status: "pending_approval", SubmittedAt: &now,
		},
		{
			PackageCode: "PK-1004", AgentID: agent, Type: "outbound",
			Name: "Maldives Honeymoon", CoverImageURL: "https://files.tayara.test/pk/maldives.jpg",
			Description: "Private island villa with all-inclusive dining.",
			DestinationCountryID: countryByCode(db, "AE"), Location: "Maldives", Category: "honeymoon",
			DaysDuration: 6, AvailabilityType: "fixed_dates", DisplayPrice: 890000, EstimatedPrice: 850000,
			DiscountPercent: 300, Status: "draft",
		},
		{
			PackageCode: "PK-1005", AgentID: agent, Type: "outbound",
			Name: "Bangkok Explorer", CoverImageURL: "https://files.tayara.test/pk/bangkok.jpg",
			Description: "Temples, street food and night markets in Bangkok and Phuket.",
			DestinationCountryID: countryByCode(db, "TH"), Location: "Bangkok", Category: "adventure",
			DaysDuration: 6, AvailabilityType: "open_calendar", DisplayPrice: 380000, EstimatedPrice: 360000,
			DiscountPercent: 400, Status: "flagged", FlagIssueType: "price_mismatch",
			FlagDetails: "Estimated price differs from display price by more than 10%.",
		},
		{
			PackageCode: "PK-1006", AgentID: agent, Type: "inbound",
			Name: "Abu Dhabi Desert Safari", CoverImageURL: "https://files.tayara.test/pk/abudhabi.jpg",
			Description: "Desert camping, dune bashing and falconry in Abu Dhabi.",
			DestinationCountryID: countryByCode(db, "AE"), Location: "Abu Dhabi", Category: "adventure",
			DaysDuration: 2, AvailabilityType: "open_calendar", DisplayPrice: 145000, EstimatedPrice: 140000,
			DiscountPercent: 350, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
		},
	}
	if err := db.Create(&packages).Error; err != nil {
		return nil, err
	}
	return packages, nil
}

// packageIDByCode resolves a package row by its unique code.
func packageIDByCode(db *gorm.DB, code string) uint {
	var p struct {
		ID uint
	}
	if err := db.Table("catalog_package").Select("id").Where("package_code = ?", code).Scan(&p).Error; err != nil {
		return 0
	}
	return p.ID
}

func seedPackageChildren(db *gorm.DB, _ []catmodel.Package) error {
	ist := packageIDByCode(db, "PK-1001")
	dub := packageIDByCode(db, "PK-1002")
	bali := packageIDByCode(db, "PK-1003")
	mald := packageIDByCode(db, "PK-1004")
	bang := packageIDByCode(db, "PK-1005")
	adh := packageIDByCode(db, "PK-1006")
	now := time.Now()

	// Availability
	avail := []catmodel.PackageAvailability{
		{PackageID: ist, StartDate: &now, EndDate: tAdd(now, 90*24), IsFixedGroup: false, Notes: "Open calendar"},
		{PackageID: dub, StartDate: &now, EndDate: tAdd(now, 60*24), IsFixedGroup: true, MaxSeats: intPtr(16), Notes: "Fixed groups weekly"},
		{PackageID: bali, StartDate: tAdd(now, 20*24), EndDate: tAdd(now, 120*24), IsFixedGroup: false, MaxSeats: nil},
		{PackageID: bang, StartDate: &now, EndDate: tAdd(now, 100*24), IsFixedGroup: false},
	}
	if err := db.Create(&avail).Error; err != nil {
		return err
	}

	// Hotels
	hotels := []catmodel.PackageHotel{
		{PackageID: ist, HotelStars: 5, RoomType: "double", OptionName: "Hagia Sophia Suites", PricePerNightEst: 25000, IncludeInPackage: true, Nights: 4, SortOrder: 1},
		{PackageID: ist, HotelStars: 4, RoomType: "single", OptionName: "Galata View", PricePerNightEst: 18000, IncludeInPackage: false, Nights: 4, SortOrder: 2},
		{PackageID: dub, HotelStars: 5, RoomType: "double", OptionName: "Downtown Grand", PricePerNightEst: 30000, IncludeInPackage: true, Nights: 3, SortOrder: 1},
		{PackageID: mald, HotelStars: 5, RoomType: "double", OptionName: "Overwater Villa", PricePerNightEst: 60000, IncludeInPackage: true, Nights: 5, SortOrder: 1},
	}
	if err := db.Create(&hotels).Error; err != nil {
		return err
	}

	// Activities
	activities := []catmodel.PackageActivity{
		{PackageID: ist, Title: "Bosphorus Cruise", EstimatedCost: 15000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: ist, Title: "Blue Mosque & Hagia Sophia", EstimatedCost: 8000, IncludeInPackage: true, SortOrder: 2},
		{PackageID: dub, Title: "Burj Khalifa At The Top", EstimatedCost: 12000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: dub, Title: "Desert Safari", EstimatedCost: 18000, IncludeInPackage: true, SortOrder: 2},
		{PackageID: adh, Title: "Dune Bashing", EstimatedCost: 14000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: adh, Title: "Falconry Show", EstimatedCost: 9000, IncludeInPackage: false, SortOrder: 2},
	}
	if err := db.Create(&activities).Error; err != nil {
		return err
	}

	// Add-ons
	addons := []catmodel.PackageAddon{
		{PackageID: ist, AddonType: "visa_issue_help", EstimatedPrice: 25000, PriceModel: "per_traveler", IncludeInPackage: false, Active: true},
		{PackageID: ist, AddonType: "travel_insurance", EstimatedPrice: 12000, PriceModel: "per_traveler", IncludeInPackage: false, Active: true},
		{PackageID: dub, AddonType: "tour_guide", EstimatedPrice: 30000, PriceModel: "fixed", IncludeInPackage: false, Active: true},
		{PackageID: mald, AddonType: "photographer", EstimatedPrice: 40000, PriceModel: "fixed", IncludeInPackage: true, Active: true},
	}
	if err := db.Create(&addons).Error; err != nil {
		return err
	}

	// Required fields
	fields := []catmodel.PackageRequiredField{
		{PackageID: ist, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: ist, FieldCode: "dob", IsAgentSelectable: true, IsAdminLocked: false, IsRequired: true, SortOrder: 2},
		{PackageID: dub, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: dub, FieldCode: "emirates_id", IsAgentSelectable: false, IsAdminLocked: true, IsRequired: false, SortOrder: 2},
	}
	if err := db.Create(&fields).Error; err != nil {
		return err
	}

	// Document requirements (map library docs to packages)
	docReqs := []catmodel.PackageDocumentRequirement{
		{PackageID: ist, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment", IsLocked: false, Notes: "Scanned copy of passport"},
		{PackageID: ist, DocumentTypeID: docTypeID(db, "personal_photo"), IsRequired: true, IsOptional: false, RequiredAfter: "payment", IsLocked: false},
		{PackageID: dub, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "booking", IsLocked: true},
		{PackageID: dub, DocumentTypeID: docTypeID(db, "emirates_id"), IsRequired: false, IsOptional: true, RequiredAfter: "payment", IsLocked: false},
		{PackageID: bali, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment", IsLocked: false},
	}
	if err := db.Create(&docReqs).Error; err != nil {
		return err
	}

	// Package views (drive the agent "Package Views" metric)
	views := []catmodel.PackageView{
		{PackageID: ist, ViewDate: &now, IP: "192.168.1.10", Session: "web"},
		{PackageID: ist, ViewDate: tAdd(now, -2*24), IP: "192.168.1.11", Session: "mobile"},
		{PackageID: dub, ViewDate: &now, IP: "192.168.1.12", Session: "mobile"},
		{PackageID: mald, ViewDate: tAdd(now, -5*24), IP: "192.168.1.13", Session: "web"},
		{PackageID: bang, ViewDate: tAdd(now, -1*24), IP: "192.168.1.14", Session: "mobile"},
	}
	return db.Create(&views).Error
}

func seedCustomTripTemplates(db *gorm.DB) error {
	agent := agentIDByEmail(db, "hello@desertsafari.ae")
	templates := []catmodel.CustomTripTemplate{
		{
			AgentID: agent, Name: "Custom Middle East", DurationMinDays: 3, DurationMaxDays: 14,
			TransportTypes: common.JSONMap{"options": []string{"car", "flight", "yacht"}},
			HotelLevels:    common.JSONMap{"levels": []int{3, 4, 5}},
			Activities:     common.JSONMap{"items": []string{"desert", "city", "sea"}},
			VisaAddonsEnabled: true, PricingLogicHidden: true, IsActive: true,
		},
		{
			AgentID: agent, Name: "Custom Europe Escape", DurationMinDays: 5, DurationMaxDays: 21,
			TransportTypes: common.JSONMap{"options": []string{"flight", "train"}},
			HotelLevels:    common.JSONMap{"levels": []int{4, 5}},
			Activities:     common.JSONMap{"items": []string{"culture", "food", "nature"}},
			VisaAddonsEnabled: true, PricingLogicHidden: true, IsActive: true,
		},
	}
	if err := db.Create(&templates).Error; err != nil {
		return err
	}
	var first, second uint
	if len(templates) > 0 {
		first = templates[0].ID
	}
	if len(templates) > 1 {
		second = templates[1].ID
	}

	// Template countries join
	countries := []catmodel.CustomTripTemplateCountry{
		{TemplateID: first, CountryID: countryByCode(db, "TR")},
		{TemplateID: first, CountryID: countryByCode(db, "AE")},
		{TemplateID: second, CountryID: countryByCode(db, "GB")},
		{TemplateID: second, CountryID: countryByCode(db, "US")},
	}
	if err := db.Create(&countries).Error; err != nil {
		return err
	}

	// Requirement profiles (dynamic per-country required documents)
	profiles := []catmodel.CustomTripRequirementProfile{
		{TemplateID: first, CountryID: countryByCode(db, "TR"), DocumentTypeID: docTypeID(db, "passport_copy"), TravelerField: "passport_copy", IsRequired: true, Phase: "before"},
		{TemplateID: first, CountryID: countryByCode(db, "TR"), DocumentTypeID: docTypeID(db, "personal_photo"), TravelerField: "personal_photo", IsRequired: true, Phase: "after"},
		{TemplateID: second, CountryID: countryByCode(db, "GB"), DocumentTypeID: docTypeID(db, "passport_copy"), TravelerField: "passport_copy", IsRequired: true, Phase: "before"},
	}
	return db.Create(&profiles).Error
}

// small helpers (package-scoped)
func uintPtr(v uint) *uint    { return &v }
func intPtr(v int) *int       { return &v }
func tAdd(t time.Time, h int) *time.Time {
	n := t.Add(time.Duration(h) * time.Hour)
	return &n
}
