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
	if err := seedPackageExtras(db, packages); err != nil {
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
	nextDep1 := now.Add(14 * 24 * time.Hour)
	nextDep2 := now.Add(30 * 24 * time.Hour)
	nextDep3 := now.Add(7 * 24 * time.Hour)

	// Resolve destination city IDs by name (seeded by demo geography).
	cityDubai := cityIDByName(db, "Dubai")
	cityAbuDhabi := cityIDByName(db, "Abu Dhabi")
	cityIstanbul := cityIDByName(db, "Istanbul")
	cityAntalya := cityIDByName(db, "Antalya")
	cityBali := cityIDByName(db, "Bali")
	cityBangkok := cityIDByName(db, "Bangkok")
	cityPhuket := cityIDByName(db, "Phuket")
	cityCairo := cityIDByName(db, "Cairo")
	citySharm := cityIDByName(db, "Sharm El Sheikh")
	cityLondon := cityIDByName(db, "London")
	cityParis := cityIDByName(db, "Paris")
	cityNYC := cityIDByName(db, "New York")
	cityTokyo := cityIDByName(db, "Tokyo")
	cityZurich := cityIDByName(db, "Zurich")
	cityRAK := cityIDByName(db, "Ras Al Khaimah")
	citySharjah := cityIDByName(db, "Sharjah")
	cityAmman := cityIDByName(db, "Amman")
	cityMaldives := uint(0) // no city seed; leave nil

	packages := []catmodel.Package{
		// ── Existing six (kept) ──────────────────────────────────────────
		{
			PackageCode: "PK-1001", AgentID: agent, Type: "outbound",
			Name: "Istanbul Highlights", CoverImageURL: "https://files.tayara.test/pk/istanbul.jpg",
			Description:          "5 days in Istanbul: old city, Bosphorus cruise, culinary tour.",
			GalleryImages:        common.JSONMap{"images": []string{"https://files.tayara.test/pk/ist1.jpg", "https://files.tayara.test/pk/ist2.jpg"}},
			DestinationCountryID: countryByCode(db, "TR"), CityID: uintPtrIfSet(cityIstanbul),
			Location: "Istanbul", Category: "cultural",
			DaysDuration: 5, AvailabilityType: "open_calendar", DisplayPrice: 320000, EstimatedPrice: 290000,
			DiscountPercent: 400, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Cultural Tour", TimeDisplay: "5 Days / 4 Nights", Destination: "Istanbul City Tour",
			OriginalPrice: 320000, PromoTag: "Save 10%", NextDepartureDate: &nextDep1, AvailableSeats: 12,
		},
		{
			PackageCode: "PK-1002", AgentID: agent, Type: "inbound",
			Name: "Dubai City Break", CoverImageURL: "https://files.tayara.test/pk/dubai.jpg",
			Description:          "3 nights in Dubai with Burj Khalifa, desert safari and mall tour.",
			DestinationCountryID: countryByCode(db, "AE"), CityID: uintPtrIfSet(cityDubai),
			Location: "Dubai", Category: "beach",
			DaysDuration: 4, AvailabilityType: "fixed_dates", DisplayPrice: 210000, EstimatedPrice: 195000,
			DiscountPercent: 500, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "City Tour", TimeDisplay: "4 Days / 3 Nights", Destination: "Dubai to Abu Dhabi",
			OriginalPrice: 210000, PromoTag: "Limited Time", NextDepartureDate: &nextDep3, AvailableSeats: 16,
		},
		{
			PackageCode: "PK-1003", AgentID: agent2, Type: "outbound",
			Name: "Bali Adventure", CoverImageURL: "https://files.tayara.test/pk/bali.jpg",
			Description:          "Surf, volcanoes and temples across Bali in 7 days.",
			DestinationCountryID: countryByCode(db, "ID"), CityID: uintPtrIfSet(cityBali),
			Location: "Bali", Category: "adventure",
			DaysDuration: 7, AvailabilityType: "open_calendar", DisplayPrice: 450000, EstimatedPrice: 430000,
			DiscountPercent: 300, Status: "pending_approval", SubmittedAt: &now,
		},
		{
			PackageCode: "PK-1004", AgentID: agent, Type: "outbound",
			Name: "Maldives Honeymoon", CoverImageURL: "https://files.tayara.test/pk/maldives.jpg",
			Description:          "Private island villa with all-inclusive dining.",
			DestinationCountryID: countryByCode(db, "MV"), CityID: uintPtrIfSet(cityMaldives),
			Location: "Maldives", Category: "honeymoon",
			DaysDuration: 6, AvailabilityType: "fixed_dates", DisplayPrice: 890000, EstimatedPrice: 850000,
			DiscountPercent: 300, Status: "draft",
		},
		{
			PackageCode: "PK-1005", AgentID: agent, Type: "outbound",
			Name: "Bangkok Explorer", CoverImageURL: "https://files.tayara.test/pk/bangkok.jpg",
			Description:          "Temples, street food and night markets in Bangkok and Phuket.",
			DestinationCountryID: countryByCode(db, "TH"), CityID: uintPtrIfSet(cityBangkok),
			Location: "Bangkok", Category: "adventure",
			DaysDuration: 6, AvailabilityType: "open_calendar", DisplayPrice: 380000, EstimatedPrice: 360000,
			DiscountPercent: 400, Status: "flagged", FlagIssueType: "price_mismatch",
			FlagDetails: "Estimated price differs from display price by more than 10%.",
		},
		{
			PackageCode: "PK-1006", AgentID: agent, Type: "inbound",
			Name: "Abu Dhabi Desert Safari", CoverImageURL: "https://files.tayara.test/pk/abudhabi.jpg",
			Description:          "Desert camping, dune bashing and falconry in Abu Dhabi.",
			DestinationCountryID: countryByCode(db, "AE"), CityID: uintPtrIfSet(cityAbuDhabi),
			Location: "Abu Dhabi", Category: "adventure",
			DaysDuration: 2, AvailabilityType: "open_calendar", DisplayPrice: 145000, EstimatedPrice: 140000,
			DiscountPercent: 350, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			NextDepartureDate: &nextDep3, AvailableSeats: 8,
		},

		// ── New packages ────────────────────────────────────────────────
		{
			PackageCode: "PK-1007", AgentID: agent2, Type: "outbound",
			Name: "Paris Romance Escape", CoverImageURL: "https://files.tayara.test/pk/paris.jpg",
			Description:          "4 nights in Paris: Seine cruise, Eiffel Tower, Montmartre walking tour.",
			GalleryImages:        common.JSONMap{"images": []string{"https://files.tayara.test/pk/paris1.jpg", "https://files.tayara.test/pk/paris2.jpg"}},
			DestinationCountryID: countryByCode(db, "FR"), CityID: uintPtrIfSet(cityParis),
			Location: "Paris", Category: "honeymoon",
			DaysDuration: 4, AvailabilityType: "open_calendar", DisplayPrice: 620000, EstimatedPrice: 590000,
			DiscountPercent: 500, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "City Tour", TimeDisplay: "4 Days / 3 Nights", Destination: "Paris City Break",
			OriginalPrice: 620000, PromoTag: "Save 5%", NextDepartureDate: &nextDep2, AvailableSeats: 10,
		},
		{
			PackageCode: "PK-1008", AgentID: agent2, Type: "outbound",
			Name: "London City & Theatre", CoverImageURL: "https://files.tayara.test/pk/london.jpg",
			Description:          "5 nights in London: West End show, Tower of London, Thames cruise.",
			DestinationCountryID: countryByCode(db, "GB"), CityID: uintPtrIfSet(cityLondon),
			Location: "London", Category: "cultural",
			DaysDuration: 5, AvailabilityType: "open_calendar", DisplayPrice: 720000, EstimatedPrice: 690000,
			DiscountPercent: 300, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Cultural Tour", TimeDisplay: "5 Days / 4 Nights", Destination: "London City Tour",
			OriginalPrice: 720000, PromoTag: "Best Seller", NextDepartureDate: &nextDep1, AvailableSeats: 14,
		},
		{
			PackageCode: "PK-1009", AgentID: agent, Type: "outbound",
			Name: "New York City Explorer", CoverImageURL: "https://files.tayara.test/pk/nyc.jpg",
			Description:          "6 days in NYC: Statue of Liberty, Times Square, Brooklyn food tour.",
			DestinationCountryID: countryByCode(db, "US"), CityID: uintPtrIfSet(cityNYC),
			Location: "New York", Category: "city_break",
			DaysDuration: 6, AvailabilityType: "open_calendar", DisplayPrice: 810000, EstimatedPrice: 780000,
			DiscountPercent: 400, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "City Tour", TimeDisplay: "6 Days / 5 Nights", Destination: "New York City Break",
			OriginalPrice: 810000, PromoTag: "Save 8%", NextDepartureDate: &nextDep2, AvailableSeats: 20,
		},
		{
			PackageCode: "PK-1010", AgentID: agent, Type: "outbound",
			Name: "Tokyo Discovery", CoverImageURL: "https://files.tayara.test/pk/tokyo.jpg",
			Description:          "7 days in Tokyo & Kyoto: temples, sushi class, Mt. Fuji day trip.",
			DestinationCountryID: countryByCode(db, "JP"), CityID: uintPtrIfSet(cityTokyo),
			Location: "Tokyo", Category: "cultural",
			DaysDuration: 7, AvailabilityType: "fixed_dates", DisplayPrice: 990000, EstimatedPrice: 940000,
			DiscountPercent: 350, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Cultural Tour", TimeDisplay: "7 Days / 6 Nights", Destination: "Tokyo to Kyoto",
			OriginalPrice: 990000, PromoTag: "Limited Time", NextDepartureDate: &nextDep1, AvailableSeats: 12,
		},
		{
			PackageCode: "PK-1011", AgentID: agent2, Type: "outbound",
			Name: "Antalya All-Inclusive", CoverImageURL: "https://files.tayara.test/pk/antalya.jpg",
			Description:          "5 nights all-inclusive resort on the Turkish Riviera.",
			DestinationCountryID: countryByCode(db, "TR"), CityID: uintPtrIfSet(cityAntalya),
			Location: "Antalya", Category: "beach",
			DaysDuration: 5, AvailabilityType: "fixed_dates", DisplayPrice: 340000, EstimatedPrice: 320000,
			DiscountPercent: 450, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Beach & Resort", TimeDisplay: "5 Days / 4 Nights", Destination: "Antalya Riviera",
			OriginalPrice: 340000, PromoTag: "Save 12%", NextDepartureDate: &nextDep3, AvailableSeats: 30,
		},
		{
			PackageCode: "PK-1012", AgentID: agent, Type: "inbound",
			Name: "Dubai Luxury Weekend", CoverImageURL: "https://files.tayara.test/pk/dubai-lux.jpg",
			Description:          "3 nights luxury: Burj Al Arab, private yacht, Michelin-star dining.",
			DestinationCountryID: countryByCode(db, "AE"), CityID: uintPtrIfSet(cityDubai),
			Location: "Dubai", Category: "luxury",
			DaysDuration: 3, AvailabilityType: "fixed_dates", DisplayPrice: 1450000, EstimatedPrice: 1400000,
			DiscountPercent: 300, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Luxury Tour", TimeDisplay: "3 Days / 2 Nights", Destination: "Dubai Luxury",
			OriginalPrice: 1450000, PromoTag: "Premium", NextDepartureDate: &nextDep3, AvailableSeats: 6,
		},
		{
			PackageCode: "PK-1013", AgentID: agent, Type: "outbound",
			Name: "Egypt Nile Cruise", CoverImageURL: "https://files.tayara.test/pk/nile.jpg",
			Description:          "5 days Nile cruise from Luxor to Aswan with Egyptologist guide.",
			DestinationCountryID: countryByCode(db, "EG"), CityID: uintPtrIfSet(cityCairo),
			Location: "Luxor", Category: "cultural",
			DaysDuration: 5, AvailabilityType: "fixed_dates", DisplayPrice: 420000, EstimatedPrice: 400000,
			DiscountPercent: 400, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Cultural Tour", TimeDisplay: "5 Days / 4 Nights", Destination: "Luxor to Aswan",
			OriginalPrice: 420000, PromoTag: "Save 10%", NextDepartureDate: &nextDep2, AvailableSeats: 24,
		},
		{
			PackageCode: "PK-1014", AgentID: agent, Type: "outbound",
			Name: "Sharm El Sheikh Dive", CoverImageURL: "https://files.tayara.test/pk/sharm.jpg",
			Description:          "6 days Red Sea diving: Ras Mohammed, Thistlegorm wreck, Blue Hole.",
			DestinationCountryID: countryByCode(db, "EG"), CityID: uintPtrIfSet(citySharm),
			Location: "Sharm El Sheikh", Category: "adventure",
			DaysDuration: 6, AvailabilityType: "open_calendar", DisplayPrice: 380000, EstimatedPrice: 360000,
			DiscountPercent: 350, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Adventure Tour", TimeDisplay: "6 Days / 5 Nights", Destination: "Sharm Red Sea",
			OriginalPrice: 380000, PromoTag: "Diver Special", NextDepartureDate: &nextDep1, AvailableSeats: 10,
		},
		{
			PackageCode: "PK-1015", AgentID: agent2, Type: "outbound",
			Name: "Phuket Island Hop", CoverImageURL: "https://files.tayara.test/pk/phuket.jpg",
			Description:          "5 nights in Phuket: Phi Phi, James Bond Island, nightlife.",
			DestinationCountryID: countryByCode(db, "TH"), CityID: uintPtrIfSet(cityPhuket),
			Location: "Phuket", Category: "beach",
			DaysDuration: 5, AvailabilityType: "fixed_dates", DisplayPrice: 360000, EstimatedPrice: 340000,
			DiscountPercent: 400, Status: "pending_approval", SubmittedAt: &now,
		},
		{
			PackageCode: "PK-1016", AgentID: agent, Type: "outbound",
			Name: "Umrah Package – 10 Days", CoverImageURL: "https://files.tayara.test/pk/umrah.jpg",
			Description:          "10-day Umrah: Makkah, Madinah, guided ziyarat, visa included.",
			DestinationCountryID: countryByCode(db, "SA"),
			Location:             "Makkah", Category: "religious",
			DaysDuration: 10, AvailabilityType: "fixed_dates", DisplayPrice: 480000, EstimatedPrice: 460000,
			DiscountPercent: 300, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Pilgrimage", TimeDisplay: "10 Days / 9 Nights", Destination: "Makkah & Madinah",
			OriginalPrice: 480000, PromoTag: "Spiritual Journey", NextDepartureDate: &nextDep2, AvailableSeats: 40,
		},
		{
			PackageCode: "PK-1017", AgentID: agent2, Type: "outbound",
			Name: "Swiss Alps Winter", CoverImageURL: "https://files.tayara.test/pk/swiss.jpg",
			Description:          "6 days in Zermatt & Interlaken: skiing, Jungfraujoch, glacier express.",
			DestinationCountryID: countryByCode(db, "CH"), CityID: uintPtrIfSet(cityZurich),
			Location: "Zermatt", Category: "ski",
			DaysDuration: 6, AvailabilityType: "fixed_dates", DisplayPrice: 950000, EstimatedPrice: 900000,
			DiscountPercent: 350, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Winter Tour", TimeDisplay: "6 Days / 5 Nights", Destination: "Swiss Alps",
			OriginalPrice: 950000, PromoTag: "Winter Special",
		},
		{
			PackageCode: "PK-1018", AgentID: agent, Type: "inbound",
			Name: "Ras Al Khaimah Wellness Retreat", CoverImageURL: "https://files.tayara.test/pk/rak.jpg",
			Description:          "4 nights wellness: yoga, spa, mountain hikes in RAK.",
			DestinationCountryID: countryByCode(db, "AE"), CityID: uintPtrIfSet(cityRAK),
			Location: "Ras Al Khaimah", Category: "wellness",
			DaysDuration: 4, AvailabilityType: "open_calendar", DisplayPrice: 320000, EstimatedPrice: 300000,
			DiscountPercent: 400, Status: "draft",
		},
		{
			PackageCode: "PK-1019", AgentID: agent, Type: "outbound",
			Name: "Kenya Safari & Beach", CoverImageURL: "https://files.tayara.test/pk/kenya.jpg",
			Description:          "8 days: Masai Mara safari, Diani beach, wildlife photography.",
			DestinationCountryID: countryByCode(db, "KE"),
			Location:             "Masai Mara", Category: "safari",
			DaysDuration: 8, AvailabilityType: "fixed_dates", DisplayPrice: 1250000, EstimatedPrice: 1200000,
			DiscountPercent: 300, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Safari", TimeDisplay: "8 Days / 7 Nights", Destination: "Masai Mara & Diani",
			OriginalPrice: 1250000, PromoTag: "Safari Special", NextDepartureDate: &nextDep2, AvailableSeats: 12,
		},
		{
			PackageCode: "PK-1020", AgentID: agent2, Type: "outbound",
			Name: "Mediterranean Cruise", CoverImageURL: "https://files.tayara.test/pk/cruise.jpg",
			Description:          "7-night cruise from Barcelona: Rome, Naples, Santorini, Mykonos.",
			DestinationCountryID: countryByCode(db, "ES"),
			Location:             "Barcelona", Category: "cruise",
			DaysDuration: 8, AvailabilityType: "fixed_dates", DisplayPrice: 780000, EstimatedPrice: 750000,
			DiscountPercent: 350, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Cruise", TimeDisplay: "8 Days / 7 Nights", Destination: "Western Mediterranean",
			OriginalPrice: 780000, PromoTag: "Early Bird", NextDepartureDate: &nextDep1, AvailableSeats: 50,
		},
		{
			PackageCode: "PK-1021", AgentID: agent, Type: "outbound",
			Name: "Azerbaijan Discovery", CoverImageURL: "https://files.tayara.test/pk/baku.jpg",
			Description:          "5 days in Baku & Gabala: Old City, mud volcanoes, Caucasus.",
			DestinationCountryID: countryByCode(db, "AZ"),
			Location:             "Baku", Category: "cultural",
			DaysDuration: 5, AvailabilityType: "open_calendar", DisplayPrice: 290000, EstimatedPrice: 275000,
			DiscountPercent: 400, Status: "pending_approval", SubmittedAt: &now,
		},
		{
			PackageCode: "PK-1022", AgentID: agent2, Type: "outbound",
			Name: "Georgia Family Adventure", CoverImageURL: "https://files.tayara.test/pk/georgia.jpg",
			Description:          "6 days in Tbilisi, Kazbegi, Batumi — family friendly.",
			DestinationCountryID: countryByCode(db, "GE"),
			Location:             "Tbilisi", Category: "family",
			DaysDuration: 6, AvailabilityType: "open_calendar", DisplayPrice: 310000, EstimatedPrice: 295000,
			DiscountPercent: 400, Status: "draft",
		},
		{
			PackageCode: "PK-1023", AgentID: agent, Type: "outbound",
			Name: "Maldives Family Resort", CoverImageURL: "https://files.tayara.test/pk/maldives-fam.jpg",
			Description:          "5 nights family-friendly Maldives resort with kids club.",
			DestinationCountryID: countryByCode(db, "MV"),
			Location:             "Maldives", Category: "family",
			DaysDuration: 5, AvailabilityType: "fixed_dates", DisplayPrice: 640000, EstimatedPrice: 620000,
			DiscountPercent: 300, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Beach & Resort", TimeDisplay: "5 Days / 4 Nights", Destination: "Maldives",
			OriginalPrice: 640000, PromoTag: "Family Deal",
		},
		{
			PackageCode: "PK-1024", AgentID: agent, Type: "outbound",
			Name: "Moscow & St Petersburg", CoverImageURL: "https://files.tayara.test/pk/russia.jpg",
			Description:          "7 days: Red Square, Hermitage, Peterhof, metro tour.",
			DestinationCountryID: countryByCode(db, "RU"),
			Location:             "Moscow", Category: "cultural",
			DaysDuration: 7, AvailabilityType: "fixed_dates", DisplayPrice: 720000, EstimatedPrice: 690000,
			DiscountPercent: 350, Status: "flagged", FlagIssueType: "missing_documents",
			FlagDetails: "Package missing tour-operator registration document.",
		},
		{
			PackageCode: "PK-1025", AgentID: agent2, Type: "outbound",
			Name: "Spain & Portugal Grand Tour", CoverImageURL: "https://files.tayara.test/pk/iberia.jpg",
			Description:          "10 days across Madrid, Barcelona, Lisbon, Porto.",
			DestinationCountryID: countryByCode(db, "ES"),
			Location:             "Madrid", Category: "cultural",
			DaysDuration: 10, AvailabilityType: "fixed_dates", DisplayPrice: 1150000, EstimatedPrice: 1100000,
			DiscountPercent: 400, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Multi-City Tour", TimeDisplay: "10 Days / 9 Nights", Destination: "Spain & Portugal",
			OriginalPrice: 1150000, PromoTag: "Grand Tour",
		},
		{
			PackageCode: "PK-1026", AgentID: agent, Type: "inbound",
			Name: "Sharjah Cultural Tour", CoverImageURL: "https://files.tayara.test/pk/sharjah.jpg",
			Description:          "2 nights: Sharjah Museum, Heritage Area, desert sunset.",
			DestinationCountryID: countryByCode(db, "AE"), CityID: uintPtrIfSet(citySharjah),
			Location: "Sharjah", Category: "cultural",
			DaysDuration: 2, AvailabilityType: "open_calendar", DisplayPrice: 95000, EstimatedPrice: 90000,
			DiscountPercent: 300, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Cultural Tour", TimeDisplay: "2 Days / 1 Night", Destination: "Sharjah Heritage",
			OriginalPrice: 95000, PromoTag: "Save 3%",
		},
		{
			PackageCode: "PK-1027", AgentID: agent2, Type: "outbound",
			Name: "Maldives Honeymoon Deluxe", CoverImageURL: "https://files.tayara.test/pk/maldives-dlx.jpg",
			Description:          "7 nights overwater villa with private plunge pool and butler.",
			DestinationCountryID: countryByCode(db, "MV"),
			Location:             "Maldives", Category: "honeymoon",
			DaysDuration: 7, AvailabilityType: "fixed_dates", DisplayPrice: 1850000, EstimatedPrice: 1800000,
			DiscountPercent: 300, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Luxury Honeymoon", TimeDisplay: "7 Days / 6 Nights", Destination: "Maldives",
			OriginalPrice: 1850000, PromoTag: "Luxury",
		},
		{
			PackageCode: "PK-1028", AgentID: agent, Type: "outbound",
			Name: "Thailand Island Hopper", CoverImageURL: "https://files.tayara.test/pk/thailand-hop.jpg",
			Description:          "8 days: Bangkok, Koh Samui, Koh Phangan, Krabi.",
			DestinationCountryID: countryByCode(db, "TH"), CityID: uintPtrIfSet(cityBangkok),
			Location: "Bangkok", Category: "beach",
			DaysDuration: 8, AvailabilityType: "open_calendar", DisplayPrice: 490000, EstimatedPrice: 470000,
			DiscountPercent: 350, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Island Hopping", TimeDisplay: "8 Days / 7 Nights", Destination: "Thailand",
			OriginalPrice: 490000, PromoTag: "Save 5%", NextDepartureDate: &nextDep2, AvailableSeats: 18,
		},
		{
			PackageCode: "PK-1029", AgentID: agent2, Type: "outbound",
			Name: "Paris & Amsterdam", CoverImageURL: "https://files.tayara.test/pk/paris-ams.jpg",
			Description:          "7 days across Paris and Amsterdam with canal cruise.",
			DestinationCountryID: countryByCode(db, "FR"),
			Location:             "Paris", Category: "city_break",
			DaysDuration: 7, AvailabilityType: "fixed_dates", DisplayPrice: 850000, EstimatedPrice: 820000,
			DiscountPercent: 300, Status: "pending_approval", SubmittedAt: &now,
		},
		{
			PackageCode: "PK-1030", AgentID: agent, Type: "outbound",
			Name: "Jordan Petra & Wadi Rum", CoverImageURL: "https://files.tayara.test/pk/jordan.jpg",
			Description:          "5 days: Amman, Petra, Wadi Rum desert camp, Dead Sea.",
			DestinationCountryID: countryByCode(db, "JO"), CityID: uintPtrIfSet(cityAmman),
			Location: "Amman", Category: "adventure",
			DaysDuration: 5, AvailabilityType: "open_calendar", DisplayPrice: 380000, EstimatedPrice: 365000,
			DiscountPercent: 350, Status: "active", ApprovedBy: uintPtr(1), ApprovedAt: &submitted,
			StatusTour: "Adventure Tour", TimeDisplay: "5 Days / 4 Nights", Destination: "Jordan Highlights",
			OriginalPrice: 380000, PromoTag: "Save 5%", NextDepartureDate: &nextDep3, AvailableSeats: 14,
		},
	}

	if err := db.Create(&packages).Error; err != nil {
		return nil, err
	}
	return packages, nil
}
func uintPtrIfSet(v uint) *uint {
	if v == 0 {
		return nil
	}
	return &v
}

func cityIDByName(db *gorm.DB, name string) uint {
	var c struct{ ID uint }
	if err := db.Table("base_city").Select("id").Where("name = ?", name).Scan(&c).Error; err != nil {
		return 0
	}
	return c.ID
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
	ids := func(code string) uint { return packageIDByCode(db, code) }
	ist := ids("PK-1001")
	dub := ids("PK-1002")
	bali := ids("PK-1003")
	mald := ids("PK-1004")
	bang := ids("PK-1005")
	adh := ids("PK-1006")
	paris := ids("PK-1007")
	london := ids("PK-1008")
	nyc := ids("PK-1009")
	tokyo := ids("PK-1010")
	antalya := ids("PK-1011")
	dubLux := ids("PK-1012")
	nile := ids("PK-1013")
	sharm := ids("PK-1014")
	phuket := ids("PK-1015")
	umrah := ids("PK-1016")
	swiss := ids("PK-1017")
	rak := ids("PK-1018")
	kenya := ids("PK-1019")
	cruise := ids("PK-1020")
	baku := ids("PK-1021")
	georgia := ids("PK-1022")
	maldFam := ids("PK-1023")
	russia := ids("PK-1024")
	iberia := ids("PK-1025")
	sharjah := ids("PK-1026")
	maldDlx := ids("PK-1027")
	thaiHop := ids("PK-1028")
	parisAms := ids("PK-1029")
	jordan := ids("PK-1030")
	now := time.Now()

	// ── Availability ────────────────────────────────────────────────
	avail := []catmodel.PackageAvailability{
		// Original
		{PackageID: ist, StartDate: &now, EndDate: tAdd(now, 90*24), IsFixedGroup: false, Notes: "Open calendar"},
		{PackageID: dub, StartDate: &now, EndDate: tAdd(now, 60*24), IsFixedGroup: true, MaxSeats: intPtr(16), Notes: "Fixed groups weekly"},
		{PackageID: bali, StartDate: tAdd(now, 20*24), EndDate: tAdd(now, 120*24), IsFixedGroup: false},
		{PackageID: bang, StartDate: &now, EndDate: tAdd(now, 100*24), IsFixedGroup: false},

		// New
		{PackageID: paris, StartDate: tAdd(now, 10*24), EndDate: tAdd(now, 200*24), IsFixedGroup: false, Notes: "Daily departures"},
		{PackageID: london, StartDate: &now, EndDate: tAdd(now, 180*24), IsFixedGroup: false, Notes: "Weekdays + weekends"},
		{PackageID: nyc, StartDate: &now, EndDate: tAdd(now, 180*24), IsFixedGroup: false},
		{PackageID: tokyo, StartDate: tAdd(now, 30*24), EndDate: tAdd(now, 270*24), IsFixedGroup: true, MaxSeats: intPtr(20), Notes: "Monthly fixed group"},
		{PackageID: antalya, StartDate: tAdd(now, 5*24), EndDate: tAdd(now, 240*24), IsFixedGroup: false},
		{PackageID: dubLux, StartDate: tAdd(now, 7*24), EndDate: tAdd(now, 120*24), IsFixedGroup: true, MaxSeats: intPtr(8)},
		{PackageID: nile, StartDate: tAdd(now, 14*24), EndDate: tAdd(now, 220*24), IsFixedGroup: true, MaxSeats: intPtr(30), Notes: "Weekly Friday departures"},
		{PackageID: sharm, StartDate: &now, EndDate: tAdd(now, 200*24), IsFixedGroup: false},
		{PackageID: phuket, StartDate: tAdd(now, 15*24), EndDate: tAdd(now, 210*24), IsFixedGroup: false},
		{PackageID: umrah, StartDate: tAdd(now, 30*24), EndDate: tAdd(now, 300*24), IsFixedGroup: true, MaxSeats: intPtr(40), Notes: "Monthly group"},
		{PackageID: swiss, StartDate: tAdd(now, 60*24), EndDate: tAdd(now, 240*24), IsFixedGroup: true, MaxSeats: intPtr(20)},
		{PackageID: rak, StartDate: &now, EndDate: tAdd(now, 365*24), IsFixedGroup: false},
		{PackageID: kenya, StartDate: tAdd(now, 30*24), EndDate: tAdd(now, 300*24), IsFixedGroup: true, MaxSeats: intPtr(12)},
		{PackageID: cruise, StartDate: tAdd(now, 20*24), EndDate: tAdd(now, 250*24), IsFixedGroup: true, MaxSeats: intPtr(50), Notes: "Monthly departure"},
		{PackageID: baku, StartDate: &now, EndDate: tAdd(now, 180*24), IsFixedGroup: false},
		{PackageID: georgia, StartDate: &now, EndDate: tAdd(now, 200*24), IsFixedGroup: false},
		{PackageID: maldFam, StartDate: tAdd(now, 10*24), EndDate: tAdd(now, 260*24), IsFixedGroup: false},
		{PackageID: russia, StartDate: tAdd(now, 25*24), EndDate: tAdd(now, 220*24), IsFixedGroup: false},
		{PackageID: iberia, StartDate: tAdd(now, 30*24), EndDate: tAdd(now, 260*24), IsFixedGroup: true, MaxSeats: intPtr(25)},
		{PackageID: sharjah, StartDate: &now, EndDate: tAdd(now, 365*24), IsFixedGroup: false},
		{PackageID: maldDlx, StartDate: tAdd(now, 20*24), EndDate: tAdd(now, 300*24), IsFixedGroup: true, MaxSeats: intPtr(10)},
		{PackageID: thaiHop, StartDate: &now, EndDate: tAdd(now, 240*24), IsFixedGroup: false},
		{PackageID: parisAms, StartDate: tAdd(now, 15*24), EndDate: tAdd(now, 200*24), IsFixedGroup: false},
		{PackageID: jordan, StartDate: &now, EndDate: tAdd(now, 200*24), IsFixedGroup: false},
	}
	if err := db.Create(&avail).Error; err != nil {
		return err
	}

	// ── Hotels ──────────────────────────────────────────────────────
	hotels := []catmodel.PackageHotel{
		// Original
		{PackageID: ist, HotelStars: 5, RoomType: "double", OptionName: "Hagia Sophia Suites", PricePerNightEst: 25000, IncludeInPackage: true, Nights: 4, SortOrder: 1},
		{PackageID: ist, HotelStars: 4, RoomType: "single", OptionName: "Galata View", PricePerNightEst: 18000, IncludeInPackage: false, Nights: 4, SortOrder: 2},
		{PackageID: dub, HotelStars: 5, RoomType: "double", OptionName: "Downtown Grand", PricePerNightEst: 30000, IncludeInPackage: true, Nights: 3, SortOrder: 1},
		{PackageID: mald, HotelStars: 5, RoomType: "double", OptionName: "Overwater Villa", PricePerNightEst: 60000, IncludeInPackage: true, Nights: 5, SortOrder: 1},

		// New
		{PackageID: paris, HotelStars: 5, RoomType: "double", OptionName: "Le Marais Boutique", PricePerNightEst: 35000, IncludeInPackage: true, Nights: 3, SortOrder: 1},
		{PackageID: london, HotelStars: 4, RoomType: "double", OptionName: "Westminster Inn", PricePerNightEst: 30000, IncludeInPackage: true, Nights: 4, SortOrder: 1},
		{PackageID: nyc, HotelStars: 5, RoomType: "double", OptionName: "Midtown Manhattan", PricePerNightEst: 40000, IncludeInPackage: true, Nights: 5, SortOrder: 1},
		{PackageID: tokyo, HotelStars: 5, RoomType: "double", OptionName: "Shinjuku Grand", PricePerNightEst: 42000, IncludeInPackage: true, Nights: 6, SortOrder: 1},
		{PackageID: antalya, HotelStars: 5, RoomType: "double", OptionName: "Belek All-Inclusive", PricePerNightEst: 22000, IncludeInPackage: true, Nights: 4, SortOrder: 1},
		{PackageID: dubLux, HotelStars: 5, RoomType: "suite", OptionName: "Burj Al Arab Suite", PricePerNightEst: 250000, IncludeInPackage: true, Nights: 2, SortOrder: 1},
		{PackageID: nile, HotelStars: 5, RoomType: "double", OptionName: "Nile Cruise Ship – Deluxe", PricePerNightEst: 26000, IncludeInPackage: true, Nights: 4, SortOrder: 1},
		{PackageID: sharm, HotelStars: 4, RoomType: "double", OptionName: "Naama Bay Resort", PricePerNightEst: 18000, IncludeInPackage: true, Nights: 5, SortOrder: 1},
		{PackageID: phuket, HotelStars: 4, RoomType: "double", OptionName: "Patong Beach Hotel", PricePerNightEst: 20000, IncludeInPackage: true, Nights: 4, SortOrder: 1},
		{PackageID: umrah, HotelStars: 5, RoomType: "double", OptionName: "Makkah Clock Tower", PricePerNightEst: 28000, IncludeInPackage: true, Nights: 9, SortOrder: 1},
		{PackageID: swiss, HotelStars: 4, RoomType: "double", OptionName: "Zermatt Alpine Lodge", PricePerNightEst: 48000, IncludeInPackage: true, Nights: 5, SortOrder: 1},
		{PackageID: rak, HotelStars: 5, RoomType: "double", OptionName: "RAK Wellness Resort", PricePerNightEst: 32000, IncludeInPackage: true, Nights: 3, SortOrder: 1},
		{PackageID: kenya, HotelStars: 5, RoomType: "double", OptionName: "Mara Safari Lodge", PricePerNightEst: 55000, IncludeInPackage: true, Nights: 7, SortOrder: 1},
		{PackageID: cruise, HotelStars: 5, RoomType: "double", OptionName: "Balcony Cabin – Western Med", PricePerNightEst: 38000, IncludeInPackage: true, Nights: 7, SortOrder: 1},
		{PackageID: baku, HotelStars: 5, RoomType: "double", OptionName: "Flame Towers Hotel", PricePerNightEst: 22000, IncludeInPackage: true, Nights: 4, SortOrder: 1},
		{PackageID: georgia, HotelStars: 4, RoomType: "double", OptionName: "Tbilisi Old Town", PricePerNightEst: 18000, IncludeInPackage: true, Nights: 5, SortOrder: 1},
		{PackageID: maldFam, HotelStars: 5, RoomType: "family", OptionName: "Beach Family Villa", PricePerNightEst: 52000, IncludeInPackage: true, Nights: 4, SortOrder: 1},
		{PackageID: russia, HotelStars: 5, RoomType: "double", OptionName: "Moscow Metropol", PricePerNightEst: 26000, IncludeInPackage: true, Nights: 6, SortOrder: 1},
		{PackageID: iberia, HotelStars: 4, RoomType: "double", OptionName: "Madrid Gran Via", PricePerNightEst: 28000, IncludeInPackage: true, Nights: 9, SortOrder: 1},
		{PackageID: sharjah, HotelStars: 4, RoomType: "double", OptionName: "Sharjah Heritage Hotel", PricePerNightEst: 15000, IncludeInPackage: true, Nights: 1, SortOrder: 1},
		{PackageID: maldDlx, HotelStars: 5, RoomType: "villa", OptionName: "Overwater Pool Villa", PricePerNightEst: 120000, IncludeInPackage: true, Nights: 6, SortOrder: 1},
		{PackageID: thaiHop, HotelStars: 4, RoomType: "double", OptionName: "Koh Samui Beachfront", PricePerNightEst: 22000, IncludeInPackage: true, Nights: 7, SortOrder: 1},
		{PackageID: parisAms, HotelStars: 4, RoomType: "double", OptionName: "Amsterdam Canal View", PricePerNightEst: 30000, IncludeInPackage: true, Nights: 6, SortOrder: 1},
		{PackageID: jordan, HotelStars: 4, RoomType: "double", OptionName: "Amman Boulevard Hotel", PricePerNightEst: 16000, IncludeInPackage: true, Nights: 4, SortOrder: 1},
	}
	if err := db.Create(&hotels).Error; err != nil {
		return err
	}

	// ── Activities ──────────────────────────────────────────────────
	activities := []catmodel.PackageActivity{
		// Original
		{PackageID: ist, Title: "Bosphorus Cruise", EstimatedCost: 15000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: ist, Title: "Blue Mosque & Hagia Sophia", EstimatedCost: 8000, IncludeInPackage: true, SortOrder: 2},
		{PackageID: dub, Title: "Burj Khalifa At The Top", EstimatedCost: 12000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: dub, Title: "Desert Safari", EstimatedCost: 18000, IncludeInPackage: true, SortOrder: 2},
		{PackageID: adh, Title: "Dune Bashing", EstimatedCost: 14000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: adh, Title: "Falconry Show", EstimatedCost: 9000, IncludeInPackage: false, SortOrder: 2},

		// New
		{PackageID: paris, Title: "Seine River Cruise", EstimatedCost: 9000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: paris, Title: "Eiffel Tower Summit", EstimatedCost: 15000, IncludeInPackage: true, SortOrder: 2},
		{PackageID: london, Title: "West End Show", EstimatedCost: 22000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: nyc, Title: "Statue of Liberty & Ellis Island", EstimatedCost: 10000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: tokyo, Title: "Mt. Fuji Day Trip", EstimatedCost: 20000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: antalya, Title: "Boat Trip", EstimatedCost: 8000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: dubLux, Title: "Private Yacht Charter", EstimatedCost: 120000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: nile, Title: "Karnak & Luxor Temples", EstimatedCost: 12000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: sharm, Title: "Ras Mohammed Diving", EstimatedCost: 25000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: phuket, Title: "Phi Phi Island Tour", EstimatedCost: 16000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: umrah, Title: "Ziyarat Tour", EstimatedCost: 10000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: swiss, Title: "Jungfraujoch Excursion", EstimatedCost: 45000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: rak, Title: "Sunrise Mountain Hike", EstimatedCost: 8000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: kenya, Title: "Masai Mara Game Drive", EstimatedCost: 40000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: cruise, Title: "Shore Excursions (4 ports)", EstimatedCost: 60000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: baku, Title: "Gobustan & Mud Volcanoes", EstimatedCost: 12000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: georgia, Title: "Kazbegi Mountain Trip", EstimatedCost: 14000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: maldFam, Title: "Snorkeling Safari", EstimatedCost: 18000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: russia, Title: "Hermitage Museum Tour", EstimatedCost: 9000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: iberia, Title: "Sagrada Familia & Alhambra", EstimatedCost: 30000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: sharjah, Title: "Heritage Area Walking Tour", EstimatedCost: 5000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: maldDlx, Title: "Private Sunset Cruise", EstimatedCost: 60000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: thaiHop, Title: "Full Moon Party Transfer", EstimatedCost: 12000, IncludeInPackage: false, SortOrder: 1},
		{PackageID: parisAms, Title: "Amsterdam Canal Cruise", EstimatedCost: 9000, IncludeInPackage: true, SortOrder: 1},
		{PackageID: jordan, Title: "Petra Guided Tour", EstimatedCost: 12000, IncludeInPackage: true, SortOrder: 1},
	}
	if err := db.Create(&activities).Error; err != nil {
		return err
	}

	// ── Add-ons ─────────────────────────────────────────────────────
	addons := []catmodel.PackageAddon{
		// Original
		{PackageID: ist, AddonType: "visa_issue_help", EstimatedPrice: 25000, PriceModel: "per_traveler", IncludeInPackage: false, Active: true},
		{PackageID: ist, AddonType: "travel_insurance", EstimatedPrice: 12000, PriceModel: "per_traveler", IncludeInPackage: false, Active: true},
		{PackageID: dub, AddonType: "tour_guide", EstimatedPrice: 30000, PriceModel: "fixed", IncludeInPackage: false, Active: true},
		{PackageID: mald, AddonType: "photographer", EstimatedPrice: 40000, PriceModel: "fixed", IncludeInPackage: true, Active: true},

		// New
		{PackageID: paris, AddonType: "travel_insurance", EstimatedPrice: 15000, PriceModel: "per_traveler", Active: true},
		{PackageID: london, AddonType: "airport_transfer", EstimatedPrice: 20000, PriceModel: "fixed", Active: true},
		{PackageID: nyc, AddonType: "travel_insurance", EstimatedPrice: 18000, PriceModel: "per_traveler", Active: true},
		{PackageID: tokyo, AddonType: "pocket_wifi", EstimatedPrice: 6000, PriceModel: "per_traveler", Active: true},
		{PackageID: antalya, AddonType: "spa_package", EstimatedPrice: 22000, PriceModel: "per_traveler", Active: true},
		{PackageID: dubLux, AddonType: "photographer", EstimatedPrice: 80000, PriceModel: "fixed", Active: true},
		{PackageID: nile, AddonType: "abu_simbel_tour", EstimatedPrice: 30000, PriceModel: "per_traveler", Active: true},
		{PackageID: sharm, AddonType: "padi_certification", EstimatedPrice: 90000, PriceModel: "per_traveler", Active: true},
		{PackageID: phuket, AddonType: "massage_package", EstimatedPrice: 18000, PriceModel: "per_traveler", Active: true},
		{PackageID: umrah, AddonType: "vip_transfer", EstimatedPrice: 25000, PriceModel: "per_traveler", Active: true},
		{PackageID: swiss, AddonType: "ski_equipment_rental", EstimatedPrice: 30000, PriceModel: "per_traveler", Active: true},
		{PackageID: kenya, AddonType: "balloon_safari", EstimatedPrice: 120000, PriceModel: "per_traveler", Active: true},
		{PackageID: cruise, AddonType: "drink_package", EstimatedPrice: 45000, PriceModel: "per_traveler", Active: true},
		{PackageID: baku, AddonType: "travel_insurance", EstimatedPrice: 10000, PriceModel: "per_traveler", Active: true},
		{PackageID: maldDlx, AddonType: "private_chef", EstimatedPrice: 100000, PriceModel: "fixed", Active: true},
	}
	if err := db.Create(&addons).Error; err != nil {
		return err
	}

	// ── Required fields ────────────────────────────────────────────
	fields := []catmodel.PackageRequiredField{
		// Original
		{PackageID: ist, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: ist, FieldCode: "dob", IsAgentSelectable: true, IsAdminLocked: false, IsRequired: true, SortOrder: 2},
		{PackageID: dub, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: dub, FieldCode: "emirates_id", IsAgentSelectable: false, IsAdminLocked: true, IsRequired: false, SortOrder: 2},

		// New
		{PackageID: paris, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: london, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: nyc, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: nyc, FieldCode: "bank_statement", IsAgentSelectable: true, IsAdminLocked: false, IsRequired: true, SortOrder: 2},
		{PackageID: tokyo, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: antalya, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: dubLux, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: nile, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: sharm, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: sharm, FieldCode: "medical_report", IsAgentSelectable: true, IsAdminLocked: false, IsRequired: true, SortOrder: 2},
		{PackageID: phuket, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: umrah, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: umrah, FieldCode: "vaccination_certificate", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 2},
		{PackageID: swiss, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: swiss, FieldCode: "travel_insurance", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 2},
		{PackageID: rak, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: kenya, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: kenya, FieldCode: "vaccination_certificate", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 2},
		{PackageID: cruise, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: baku, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: georgia, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: maldFam, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: maldFam, FieldCode: "birth_certificate", IsAgentSelectable: true, IsAdminLocked: false, IsRequired: false, SortOrder: 2},
		{PackageID: russia, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: iberia, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: sharjah, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: maldDlx, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: thaiHop, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: parisAms, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
		{PackageID: jordan, FieldCode: "passport_copy", IsAgentSelectable: true, IsAdminLocked: true, IsRequired: true, SortOrder: 1},
	}
	if err := db.Create(&fields).Error; err != nil {
		return err
	}

	// ── Document requirements ──────────────────────────────────────
	docReqs := []catmodel.PackageDocumentRequirement{
		// Original
		{PackageID: ist, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment", IsLocked: false, Notes: "Scanned copy of passport"},
		{PackageID: ist, DocumentTypeID: docTypeID(db, "personal_photo"), IsRequired: true, IsOptional: false, RequiredAfter: "payment", IsLocked: false},
		{PackageID: dub, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "booking", IsLocked: true},
		{PackageID: dub, DocumentTypeID: docTypeID(db, "emirates_id"), IsRequired: false, IsOptional: true, RequiredAfter: "payment", IsLocked: false},
		{PackageID: bali, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment", IsLocked: false},

		// New
		{PackageID: paris, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: london, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: nyc, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: nyc, DocumentTypeID: docTypeID(db, "bank_statement"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: tokyo, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: antalya, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: dubLux, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "booking"},
		{PackageID: nile, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: sharm, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: sharm, DocumentTypeID: docTypeID(db, "medical_report"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: phuket, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: umrah, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: umrah, DocumentTypeID: docTypeID(db, "vaccination_certificate"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: swiss, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: swiss, DocumentTypeID: docTypeID(db, "travel_insurance"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: rak, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: kenya, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: kenya, DocumentTypeID: docTypeID(db, "vaccination_certificate"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: cruise, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: baku, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: georgia, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: maldFam, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: maldFam, DocumentTypeID: docTypeID(db, "birth_certificate"), IsRequired: false, IsOptional: true, RequiredAfter: "payment"},
		{PackageID: russia, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: iberia, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: sharjah, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "booking"},
		{PackageID: maldDlx, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: thaiHop, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: parisAms, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
		{PackageID: jordan, DocumentTypeID: docTypeID(db, "passport_copy"), IsRequired: true, IsOptional: false, RequiredAfter: "payment"},
	}
	if err := db.Create(&docReqs).Error; err != nil {
		return err
	}

	// ── Views (analytics) ──────────────────────────────────────────
	views := []catmodel.PackageView{
		{PackageID: ist, ViewDate: &now, IP: "192.168.1.10", Session: "web"},
		{PackageID: ist, ViewDate: tAdd(now, -2*24), IP: "192.168.1.11", Session: "mobile"},
		{PackageID: dub, ViewDate: &now, IP: "192.168.1.12", Session: "mobile"},
		{PackageID: mald, ViewDate: tAdd(now, -5*24), IP: "192.168.1.13", Session: "web"},
		{PackageID: bang, ViewDate: tAdd(now, -1*24), IP: "192.168.1.14", Session: "mobile"},
		// New
		{PackageID: paris, ViewDate: &now, IP: "192.168.1.20", Session: "web"},
		{PackageID: london, ViewDate: &now, IP: "192.168.1.21", Session: "mobile"},
		{PackageID: nyc, ViewDate: tAdd(now, -3*24), IP: "192.168.1.22", Session: "web"},
		{PackageID: tokyo, ViewDate: tAdd(now, -1*24), IP: "192.168.1.23", Session: "mobile"},
		{PackageID: antalya, ViewDate: &now, IP: "192.168.1.24", Session: "web"},
		{PackageID: dubLux, ViewDate: &now, IP: "192.168.1.25", Session: "mobile"},
		{PackageID: nile, ViewDate: tAdd(now, -2*24), IP: "192.168.1.26", Session: "web"},
		{PackageID: sharm, ViewDate: &now, IP: "192.168.1.27", Session: "mobile"},
		{PackageID: phuket, ViewDate: tAdd(now, -4*24), IP: "192.168.1.28", Session: "web"},
		{PackageID: umrah, ViewDate: &now, IP: "192.168.1.29", Session: "mobile"},
		{PackageID: swiss, ViewDate: tAdd(now, -6*24), IP: "192.168.1.30", Session: "web"},
		{PackageID: rak, ViewDate: &now, IP: "192.168.1.31", Session: "mobile"},
		{PackageID: kenya, ViewDate: &now, IP: "192.168.1.32", Session: "web"},
		{PackageID: cruise, ViewDate: tAdd(now, -2*24), IP: "192.168.1.33", Session: "mobile"},
		{PackageID: baku, ViewDate: &now, IP: "192.168.1.34", Session: "web"},
		{PackageID: georgia, ViewDate: &now, IP: "192.168.1.35", Session: "mobile"},
		{PackageID: maldFam, ViewDate: &now, IP: "192.168.1.36", Session: "web"},
		{PackageID: russia, ViewDate: tAdd(now, -1*24), IP: "192.168.1.37", Session: "mobile"},
		{PackageID: iberia, ViewDate: &now, IP: "192.168.1.38", Session: "web"},
		{PackageID: sharjah, ViewDate: &now, IP: "192.168.1.39", Session: "mobile"},
		{PackageID: maldDlx, ViewDate: &now, IP: "192.168.1.40", Session: "web"},
		{PackageID: thaiHop, ViewDate: tAdd(now, -3*24), IP: "192.168.1.41", Session: "mobile"},
		{PackageID: parisAms, ViewDate: &now, IP: "192.168.1.42", Session: "web"},
		{PackageID: jordan, ViewDate: &now, IP: "192.168.1.43", Session: "mobile"},
	}
	return db.Create(&views).Error
}

func seedPackageExtras(db *gorm.DB, _ []catmodel.Package) error {
	ids := func(code string) uint { return packageIDByCode(db, code) }
	ist := ids("PK-1001")
	dub := ids("PK-1002")
	paris := ids("PK-1007")
	london := ids("PK-1008")
	nyc := ids("PK-1009")
	tokyo := ids("PK-1010")
	antalya := ids("PK-1011")
	dubLux := ids("PK-1012")
	nile := ids("PK-1013")
	sharm := ids("PK-1014")
	phuket := ids("PK-1015")
	umrah := ids("PK-1016")
	swiss := ids("PK-1017")
	kenya := ids("PK-1019")
	cruise := ids("PK-1020")
	jordan := ids("PK-1030")

	// ── Itinerary days ─────────────────────────────────────────────
	itinDays := []catmodel.ItineraryDay{
		// Original
		{PackageID: ist, DayNumber: 1, Title: "Day 1: Istanbul Arrival & Old City", Description: "Arrival and guided tour of the city's most iconic landmarks", SortOrder: 1},
		{PackageID: ist, DayNumber: 2, Title: "Day 2: Bosphorus Cruise & Culinary Tour", Description: "Scenic cruise along the Bosphorus followed by a local food experience", SortOrder: 2},
		{PackageID: dub, DayNumber: 1, Title: "Day 1: Dubai City Arrival", Description: "Check-in and evening Dubai Marina walk", SortOrder: 1},
		{PackageID: dub, DayNumber: 2, Title: "Day 2: Burj Khalifa & Desert Safari", Description: "Morning at the world's tallest building, afternoon desert adventure", SortOrder: 2},

		// New
		{PackageID: paris, DayNumber: 1, Title: "Day 1: Paris Arrival & Seine Cruise", Description: "Check-in and evening Seine river cruise", SortOrder: 1},
		{PackageID: paris, DayNumber: 2, Title: "Day 2: Eiffel Tower & Montmartre", Description: "Eiffel Tower summit and Montmartre walking tour", SortOrder: 2},
		{PackageID: london, DayNumber: 1, Title: "Day 1: London Arrival", Description: "Arrival and evening walk by the Thames", SortOrder: 1},
		{PackageID: london, DayNumber: 2, Title: "Day 2: Tower of London & West End", Description: "Morning at Tower of London, evening West End show", SortOrder: 2},
		{PackageID: nyc, DayNumber: 1, Title: "Day 1: NYC Arrival & Times Square", Description: "Hotel check-in and Times Square evening walk", SortOrder: 1},
		{PackageID: nyc, DayNumber: 2, Title: "Day 2: Statue of Liberty & Brooklyn", Description: "Ferry to Statue of Liberty and Brooklyn food tour", SortOrder: 2},
		{PackageID: tokyo, DayNumber: 1, Title: "Day 1: Tokyo Arrival & Shinjuku", Description: "Check-in and Shinjuku evening exploration", SortOrder: 1},
		{PackageID: tokyo, DayNumber: 2, Title: "Day 2: Mt. Fuji Day Trip", Description: "Full-day excursion to Mt. Fuji and Hakone", SortOrder: 2},
		{PackageID: antalya, DayNumber: 1, Title: "Day 1: Antalya Arrival", Description: "Transfer to all-inclusive resort", SortOrder: 1},
		{PackageID: antalya, DayNumber: 2, Title: "Day 2: Old Town & Boat Trip", Description: "Antalya Old Town tour and boat trip", SortOrder: 2},
		{PackageID: dubLux, DayNumber: 1, Title: "Day 1: Burj Al Arab Check-in", Description: "VIP check-in and dinner at Al Mahara", SortOrder: 1},
		{PackageID: dubLux, DayNumber: 2, Title: "Day 2: Private Yacht Day", Description: "Full-day private yacht charter", SortOrder: 2},
		{PackageID: nile, DayNumber: 1, Title: "Day 1: Luxor Embarkation", Description: "Board cruise ship and visit Luxor Temple", SortOrder: 1},
		{PackageID: nile, DayNumber: 2, Title: "Day 2: Valley of the Kings", Description: "West Bank tour including Valley of the Kings", SortOrder: 2},
		{PackageID: sharm, DayNumber: 1, Title: "Day 1: Sharm Arrival", Description: "Check-in at Naama Bay resort", SortOrder: 1},
		{PackageID: sharm, DayNumber: 2, Title: "Day 2: Ras Mohammed Diving", Description: "Two dives at Ras Mohammed National Park", SortOrder: 2},
		{PackageID: phuket, DayNumber: 1, Title: "Day 1: Phuket Arrival", Description: "Transfer to Patong Beach hotel", SortOrder: 1},
		{PackageID: phuket, DayNumber: 2, Title: "Day 2: Phi Phi Island Tour", Description: "Speedboat tour to Phi Phi islands", SortOrder: 2},
		{PackageID: umrah, DayNumber: 1, Title: "Day 1: Arrival in Jeddah & Transfer to Makkah", Description: "Transfer to Makkah hotel and first Umrah", SortOrder: 1},
		{PackageID: umrah, DayNumber: 2, Title: "Day 2: Makkah Ziyarat", Description: "Guided ziyarat of Makkah holy sites", SortOrder: 2},
		{PackageID: swiss, DayNumber: 1, Title: "Day 1: Zurich Arrival", Description: "Transfer to Zermatt by panoramic train", SortOrder: 1},
		{PackageID: swiss, DayNumber: 2, Title: "Day 2: Ski Day in Zermatt", Description: "Full-day skiing with instructor", SortOrder: 2},
		{PackageID: kenya, DayNumber: 1, Title: "Day 1: Nairobi Arrival", Description: "Transfer to Masai Mara safari lodge", SortOrder: 1},
		{PackageID: kenya, DayNumber: 2, Title: "Day 2: Morning & Evening Game Drives", Description: "Two game drives with Maasai guide", SortOrder: 2},
		{PackageID: cruise, DayNumber: 1, Title: "Day 1: Barcelona Embarkation", Description: "Board cruise ship and evening sail", SortOrder: 1},
		{PackageID: cruise, DayNumber: 2, Title: "Day 2: Rome (Civitavecchia)", Description: "Full-day shore excursion to Rome", SortOrder: 2},
		{PackageID: jordan, DayNumber: 1, Title: "Day 1: Amman Arrival", Description: "Hotel check-in and Amman city tour", SortOrder: 1},
		{PackageID: jordan, DayNumber: 2, Title: "Day 2: Petra Full Day", Description: "Guided tour of the Rose City", SortOrder: 2},
	}
	if err := db.Create(&itinDays).Error; err != nil {
		return err
	}

	// ── Itinerary items ────────────────────────────────────────────
	// Original items reference itinDays[0..3]; new items append day rows.
	items := []catmodel.ItineraryItem{
		// Original
		{ItineraryDayID: itinDays[0].ID, Time: "14:00", Title: "Airport Transfer", Description: "Private VIP transfer to hotel", SortOrder: 1},
		{ItineraryDayID: itinDays[0].ID, Time: "16:00", Title: "Hagia Sophia Visit", Description: "Guided tour of the historic Hagia Sophia", SortOrder: 2},
		{ItineraryDayID: itinDays[0].ID, Time: "18:00", Title: "Blue Mosque", Description: "Exterior and guided visit to the Blue Mosque", SortOrder: 3},
		{ItineraryDayID: itinDays[0].ID, Time: "20:00", Title: "Welcome Dinner", Description: "Traditional Turkish dinner at a local restaurant", SortOrder: 4},
		{ItineraryDayID: itinDays[1].ID, Time: "09:00", Title: "Topkapi Palace", Description: "Guided tour of the Ottoman sultan's palace", SortOrder: 1},
		{ItineraryDayID: itinDays[1].ID, Time: "12:00", Title: "Spice Bazaar", Description: "Free time to explore the Grand Bazaar area", SortOrder: 2},
		{ItineraryDayID: itinDays[1].ID, Time: "15:00", Title: "Bosphorus Cruise", Description: "2-hour scenic cruise between Europe and Asia", SortOrder: 3},
		{ItineraryDayID: itinDays[1].ID, Time: "19:00", Title: "Food Tasting Tour", Description: "Street food tasting across Sultanahmet", SortOrder: 4},
		{ItineraryDayID: itinDays[2].ID, Time: "15:00", Title: "Hotel Check-in", Description: "Drop bags and freshen up at the hotel", SortOrder: 1},
		{ItineraryDayID: itinDays[2].ID, Time: "19:00", Title: "Dubai Marina Walk", Description: "Evening stroll along the Marina promenade", SortOrder: 2},
		{ItineraryDayID: itinDays[3].ID, Time: "09:00", Title: "Burj Khalifa - At The Top", Description: "Skip-the-line tickets to the 124th floor", SortOrder: 1},
		{ItineraryDayID: itinDays[3].ID, Time: "15:00", Title: "Desert Safari", Description: "Dune bashing, camel ride, and BBQ dinner", SortOrder: 2},

		// New — Paris day 1 (index 4) and day 2 (index 5)
		{ItineraryDayID: itinDays[4].ID, Time: "15:00", Title: "Hotel Check-in", Description: "Check-in at Le Marais Boutique", SortOrder: 1},
		{ItineraryDayID: itinDays[4].ID, Time: "20:00", Title: "Seine Dinner Cruise", Description: "Evening dinner cruise on the Seine", SortOrder: 2},
		{ItineraryDayID: itinDays[5].ID, Time: "09:00", Title: "Eiffel Tower Summit", Description: "Skip-the-line summit access", SortOrder: 1},
		{ItineraryDayID: itinDays[5].ID, Time: "14:00", Title: "Montmartre Walking Tour", Description: "Guided tour of Montmartre and Sacré-Cœur", SortOrder: 2},

		// New — London day 1 (6) / day 2 (7)
		{ItineraryDayID: itinDays[6].ID, Time: "14:00", Title: "Thames Walk", Description: "Evening walk along the Thames and Tower Bridge", SortOrder: 1},
		{ItineraryDayID: itinDays[7].ID, Time: "10:00", Title: "Tower of London", Description: "Guided tour of the Tower of London", SortOrder: 1},
		{ItineraryDayID: itinDays[7].ID, Time: "19:30", Title: "West End Show", Description: "Premium seats at a West End show", SortOrder: 2},

		// New — NYC day 1 (8) / day 2 (9)
		{ItineraryDayID: itinDays[8].ID, Time: "16:00", Title: "Times Square Walk", Description: "Evening exploration of Times Square", SortOrder: 1},
		{ItineraryDayID: itinDays[9].ID, Time: "08:30", Title: "Statue of Liberty Ferry", Description: "Ferry to Liberty Island and Ellis Island", SortOrder: 1},
		{ItineraryDayID: itinDays[9].ID, Time: "15:00", Title: "Brooklyn Food Tour", Description: "Guided food tour in Brooklyn", SortOrder: 2},

		// New — Tokyo day 1 (10) / day 2 (11)
		{ItineraryDayID: itinDays[10].ID, Time: "17:00", Title: "Shinjuku Evening Walk", Description: "Explore Shinjuku nightlife and Tokyo Metropolitan view", SortOrder: 1},
		{ItineraryDayID: itinDays[11].ID, Time: "07:30", Title: "Mt. Fuji Day Trip", Description: "Full-day excursion to Mt. Fuji 5th Station and Hakone", SortOrder: 1},

		// New — Antalya day 1 (12) / day 2 (13)
		{ItineraryDayID: itinDays[12].ID, Time: "14:00", Title: "Resort Check-in", Description: "All-inclusive resort check-in", SortOrder: 1},
		{ItineraryDayID: itinDays[13].ID, Time: "10:00", Title: "Old Town Tour", Description: "Kaleiçi Old Town guided tour", SortOrder: 1},
		{ItineraryDayID: itinDays[13].ID, Time: "14:00", Title: "Boat Trip", Description: "Afternoon boat trip along Antalya coast", SortOrder: 2},

		// New — Dubai Luxury day 1 (14) / day 2 (15)
		{ItineraryDayID: itinDays[14].ID, Time: "15:00", Title: "VIP Check-in", Description: "Suite check-in at Burj Al Arab", SortOrder: 1},
		{ItineraryDayID: itinDays[14].ID, Time: "20:00", Title: "Al Mahara Dinner", Description: "Dinner at Burj Al Arab's underwater restaurant", SortOrder: 2},
		{ItineraryDayID: itinDays[15].ID, Time: "10:00", Title: "Private Yacht Charter", Description: "Full-day private yacht around Dubai coastline", SortOrder: 1},

		// New — Nile day 1 (16) / day 2 (17)
		{ItineraryDayID: itinDays[16].ID, Time: "11:00", Title: "Luxor Temple", Description: "Guided tour of Luxor Temple", SortOrder: 1},
		{ItineraryDayID: itinDays[17].ID, Time: "08:00", Title: "Valley of the Kings", Description: "Guided tour of the Valley of the Kings", SortOrder: 1},
		{ItineraryDayID: itinDays[17].ID, Time: "15:00", Title: "Hatshepsut Temple", Description: "Guided tour of Hatshepsut Temple", SortOrder: 2},

		// New — Sharm day 1 (18) / day 2 (19)
		{ItineraryDayID: itinDays[18].ID, Time: "15:00", Title: "Resort Check-in", Description: "Check-in at Naama Bay resort", SortOrder: 1},
		{ItineraryDayID: itinDays[19].ID, Time: "08:00", Title: "Ras Mohammed Dives", Description: "Two guided dives at Ras Mohammed", SortOrder: 1},

		// New — Phuket day 1 (20) / day 2 (21)
		{ItineraryDayID: itinDays[20].ID, Time: "14:00", Title: "Hotel Check-in", Description: "Check-in at Patong Beach hotel", SortOrder: 1},
		{ItineraryDayID: itinDays[21].ID, Time: "08:00", Title: "Phi Phi Tour", Description: "Full-day speedboat tour to Phi Phi", SortOrder: 1},

		// New — Umrah day 1 (22) / day 2 (23)
		{ItineraryDayID: itinDays[22].ID, Time: "16:00", Title: "First Umrah", Description: "First Umrah with group guide", SortOrder: 1},
		{ItineraryDayID: itinDays[23].ID, Time: "08:00", Title: "Ziyarat Tour", Description: "Guided ziyarat of Makkah holy sites", SortOrder: 1},

		// New — Swiss day 1 (24) / day 2 (25)
		{ItineraryDayID: itinDays[24].ID, Time: "10:00", Title: "Panoramic Train to Zermatt", Description: "Glacier Express to Zermatt", SortOrder: 1},
		{ItineraryDayID: itinDays[25].ID, Time: "09:00", Title: "Ski Day", Description: "Ski pass, equipment rental, instructor", SortOrder: 1},

		// New — Kenya day 1 (26) / day 2 (27)
		{ItineraryDayID: itinDays[26].ID, Time: "11:00", Title: "Transfer to Mara", Description: "Transfer to safari lodge in Masai Mara", SortOrder: 1},
		{ItineraryDayID: itinDays[27].ID, Time: "06:30", Title: "Morning Game Drive", Description: "Early morning game drive", SortOrder: 1},
		{ItineraryDayID: itinDays[27].ID, Time: "16:30", Title: "Evening Game Drive", Description: "Sunset game drive", SortOrder: 2},

		// New — Cruise day 1 (28) / day 2 (29)
		{ItineraryDayID: itinDays[28].ID, Time: "14:00", Title: "Embarkation", Description: "Board cruise ship in Barcelona", SortOrder: 1},
		{ItineraryDayID: itinDays[29].ID, Time: "08:00", Title: "Rome Shore Excursion", Description: "Full-day shore excursion to Rome", SortOrder: 1},

		// New — Jordan day 1 (30) / day 2 (31)
		{ItineraryDayID: itinDays[30].ID, Time: "15:00", Title: "Amman City Tour", Description: "Amman Citadel and Roman Theatre", SortOrder: 1},
		{ItineraryDayID: itinDays[31].ID, Time: "08:00", Title: "Petra Tour", Description: "Full-day guided tour of Petra", SortOrder: 1},
	}
	if err := db.Create(&items).Error; err != nil {
		return err
	}

	// ── Highlights ─────────────────────────────────────────────────
	highlights := []catmodel.PackageHighlight{
		// Original
		{PackageID: ist, Name: "Expert Guide", SortOrder: 1},
		{PackageID: ist, Name: "City Center Hotels", SortOrder: 2},
		{PackageID: ist, Name: "Cruise Included", SortOrder: 3},
		{PackageID: dub, Name: "5-Star Accommodation", SortOrder: 1},
		{PackageID: dub, Name: "All Meals Included", SortOrder: 2},
		{PackageID: dub, Name: "VIP Transfers", SortOrder: 3},

		// New
		{PackageID: paris, Name: "Central Hotel", SortOrder: 1},
		{PackageID: paris, Name: "Seine Cruise Included", SortOrder: 2},
		{PackageID: london, Name: "West End Show", SortOrder: 1},
		{PackageID: nyc, Name: "Statue of Liberty", SortOrder: 1},
		{PackageID: nyc, Name: "Brooklyn Food Tour", SortOrder: 2},
		{PackageID: tokyo, Name: "Mt. Fuji Day Trip", SortOrder: 1},
		{PackageID: antalya, Name: "All-Inclusive", SortOrder: 1},
		{PackageID: dubLux, Name: "Burj Al Arab Stay", SortOrder: 1},
		{PackageID: dubLux, Name: "Private Yacht", SortOrder: 2},
		{PackageID: nile, Name: "Nile Cruise", SortOrder: 1},
		{PackageID: sharm, Name: "Red Sea Diving", SortOrder: 1},
		{PackageID: phuket, Name: "Phi Phi Islands", SortOrder: 1},
		{PackageID: umrah, Name: "Guided Ziyarat", SortOrder: 1},
		{PackageID: swiss, Name: "Ski Instructor", SortOrder: 1},
		{PackageID: kenya, Name: "Big Five Safari", SortOrder: 1},
		{PackageID: cruise, Name: "4 Ports of Call", SortOrder: 1},
		{PackageID: jordan, Name: "Petra & Wadi Rum", SortOrder: 1},
	}
	if err := db.Create(&highlights).Error; err != nil {
		return err
	}

	// ── Includes ───────────────────────────────────────────────────
	includes := []catmodel.PackageInclusion{
		// Original
		{PackageID: ist, Category: "Accommodation", ItemText: "4 nights in a 4-star hotel in Sultanahmet", SortOrder: 1},
		{PackageID: ist, Category: "Meals", ItemText: "Daily breakfast buffet at the hotel", SortOrder: 2},
		{PackageID: ist, Category: "Tours", ItemText: "Guided old city tour with licensed Egyptologist", SortOrder: 3},
		{PackageID: ist, Category: "Tours", ItemText: "Bosphorus sunset cruise (2 hours)", SortOrder: 4},
		{PackageID: dub, Category: "Accommodation", ItemText: "3 nights in a 5-star downtown hotel", SortOrder: 1},
		{PackageID: dub, Category: "Meals", ItemText: "All meals: 3 breakfasts, 2 lunches, 3 dinners", SortOrder: 2},
		{PackageID: dub, Category: "Tours", ItemText: "Burj Khalifa At The Top tickets", SortOrder: 3},
		{PackageID: dub, Category: "Tours", ItemText: "Desert safari with BBQ dinner and show", SortOrder: 4},

		// New (a couple per package)
		{PackageID: paris, Category: "Accommodation", ItemText: "3 nights at Le Marais Boutique", SortOrder: 1},
		{PackageID: paris, Category: "Tours", ItemText: "Eiffel Tower summit access", SortOrder: 2},
		{PackageID: london, Category: "Accommodation", ItemText: "4 nights at Westminster Inn", SortOrder: 1},
		{PackageID: london, Category: "Tours", ItemText: "Tower of London entry", SortOrder: 2},
		{PackageID: nyc, Category: "Accommodation", ItemText: "5 nights in Midtown Manhattan", SortOrder: 1},
		{PackageID: nyc, Category: "Tours", ItemText: "Statue of Liberty ferry tickets", SortOrder: 2},
		{PackageID: tokyo, Category: "Accommodation", ItemText: "6 nights at Shinjuku Grand", SortOrder: 1},
		{PackageID: tokyo, Category: "Tours", ItemText: "Mt. Fuji & Hakone full-day tour", SortOrder: 2},
		{PackageID: antalya, Category: "Accommodation", ItemText: "4 nights all-inclusive at Belek", SortOrder: 1},
		{PackageID: antalya, Category: "Meals", ItemText: "All meals and local drinks", SortOrder: 2},
		{PackageID: dubLux, Category: "Accommodation", ItemText: "2 nights suite at Burj Al Arab", SortOrder: 1},
		{PackageID: dubLux, Category: "Tours", ItemText: "Private yacht full-day charter", SortOrder: 2},
		{PackageID: nile, Category: "Accommodation", ItemText: "4 nights Nile cruise ship", SortOrder: 1},
		{PackageID: nile, Category: "Tours", ItemText: "All guided temple visits", SortOrder: 2},
		{PackageID: sharm, Category: "Accommodation", ItemText: "5 nights at Naama Bay Resort", SortOrder: 1},
		{PackageID: sharm, Category: "Tours", ItemText: "2 guided dives at Ras Mohammed", SortOrder: 2},
		{PackageID: phuket, Category: "Accommodation", ItemText: "4 nights at Patong Beach hotel", SortOrder: 1},
		{PackageID: phuket, Category: "Tours", ItemText: "Phi Phi speedboat tour", SortOrder: 2},
		{PackageID: umrah, Category: "Accommodation", ItemText: "9 nights Makkah + Madinah hotels", SortOrder: 1},
		{PackageID: umrah, Category: "Tours", ItemText: "Guided ziyarat tours", SortOrder: 2},
		{PackageID: swiss, Category: "Accommodation", ItemText: "5 nights Zermatt Alpine Lodge", SortOrder: 1},
		{PackageID: swiss, Category: "Tours", ItemText: "Jungfraujoch excursion", SortOrder: 2},
		{PackageID: kenya, Category: "Accommodation", ItemText: "7 nights safari lodge & beach", SortOrder: 1},
		{PackageID: kenya, Category: "Tours", ItemText: "Unlimited game drives", SortOrder: 2},
		{PackageID: cruise, Category: "Accommodation", ItemText: "7-night balcony cabin", SortOrder: 1},
		{PackageID: cruise, Category: "Meals", ItemText: "All onboard meals", SortOrder: 2},
		{PackageID: jordan, Category: "Accommodation", ItemText: "4 nights 4-star hotels", SortOrder: 1},
		{PackageID: jordan, Category: "Tours", ItemText: "Petra & Wadi Rum guided tours", SortOrder: 2},
	}
	if err := db.Create(&includes).Error; err != nil {
		return err
	}

	// ── Excludes ───────────────────────────────────────────────────
	excludes := []catmodel.PackageExclusion{
		// Original
		{PackageID: ist, ItemText: "International flights to/from Istanbul", SortOrder: 1},
		{PackageID: ist, ItemText: "Alcoholic beverages", SortOrder: 2},
		{PackageID: ist, ItemText: "Personal expenses and tips", SortOrder: 3},
		{PackageID: dub, ItemText: "International flights to/from Dubai", SortOrder: 1},
		{PackageID: dub, ItemText: "Private rooms for single occupancy", SortOrder: 2},
		{PackageID: dub, ItemText: "Spa and wellness treatments", SortOrder: 3},

		// New
		{PackageID: paris, ItemText: "International flights", SortOrder: 1},
		{PackageID: london, ItemText: "International flights", SortOrder: 1},
		{PackageID: nyc, ItemText: "International flights & ESTA fee", SortOrder: 1},
		{PackageID: tokyo, ItemText: "International flights", SortOrder: 1},
		{PackageID: antalya, ItemText: "International flights", SortOrder: 1},
		{PackageID: dubLux, ItemText: "International flights", SortOrder: 1},
		{PackageID: nile, ItemText: "Domestic flights to Luxor", SortOrder: 1},
		{PackageID: sharm, ItemText: "Diving equipment rental", SortOrder: 1},
		{PackageID: phuket, ItemText: "National park fees", SortOrder: 1},
		{PackageID: umrah, ItemText: "Visa fees (add-on)", SortOrder: 1},
		{PackageID: swiss, ItemText: "Ski pass & equipment", SortOrder: 1},
		{PackageID: kenya, ItemText: "International flights", SortOrder: 1},
		{PackageID: cruise, ItemText: "Gratuities & specialty dining", SortOrder: 1},
		{PackageID: jordan, ItemText: "International flights", SortOrder: 1},
	}
	return db.Create(&excludes).Error
}

func seedCustomTripTemplates(db *gorm.DB) error {
	agent := agentIDByEmail(db, "hello@desertsafari.ae")
	agent2 := agentIDByEmail(db, "book@luxurytravel.ae")

	templates := []catmodel.CustomTripTemplate{
		{AgentID: agent, Name: "Custom Middle East", DurationMinDays: 3, DurationMaxDays: 14,
			TransportTypes:    common.JSONMap{"options": []string{"car", "flight", "yacht"}},
			HotelLevels:       common.JSONMap{"levels": []int{3, 4, 5}},
			Activities:        common.JSONMap{"items": []string{"desert", "city", "sea"}},
			VisaAddonsEnabled: true, PricingLogicHidden: true, IsActive: true},
		{AgentID: agent, Name: "Custom Europe Escape", DurationMinDays: 5, DurationMaxDays: 21,
			TransportTypes:    common.JSONMap{"options": []string{"flight", "train"}},
			HotelLevels:       common.JSONMap{"levels": []int{4, 5}},
			Activities:        common.JSONMap{"items": []string{"culture", "food", "nature"}},
			VisaAddonsEnabled: true, PricingLogicHidden: true, IsActive: true},
		{AgentID: agent, Name: "Custom Asia Discovery", DurationMinDays: 5, DurationMaxDays: 21,
			TransportTypes:    common.JSONMap{"options": []string{"flight", "train", "ferry"}},
			HotelLevels:       common.JSONMap{"levels": []int{3, 4, 5}},
			Activities:        common.JSONMap{"items": []string{"culture", "beach", "food"}},
			VisaAddonsEnabled: true, PricingLogicHidden: true, IsActive: true},
		{AgentID: agent2, Name: "Custom Americas", DurationMinDays: 6, DurationMaxDays: 21,
			TransportTypes:    common.JSONMap{"options": []string{"flight", "car"}},
			HotelLevels:       common.JSONMap{"levels": []int{4, 5}},
			Activities:        common.JSONMap{"items": []string{"city", "nature", "culture"}},
			VisaAddonsEnabled: true, PricingLogicHidden: true, IsActive: true},
		{AgentID: agent2, Name: "Custom Africa Safari", DurationMinDays: 5, DurationMaxDays: 14,
			TransportTypes:    common.JSONMap{"options": []string{"flight", "4x4"}},
			HotelLevels:       common.JSONMap{"levels": []int{3, 4, 5}},
			Activities:        common.JSONMap{"items": []string{"safari", "beach", "culture"}},
			VisaAddonsEnabled: true, PricingLogicHidden: true, IsActive: true},
		{AgentID: agent, Name: "Custom Religious Journey", DurationMinDays: 5, DurationMaxDays: 21,
			TransportTypes:    common.JSONMap{"options": []string{"flight", "bus"}},
			HotelLevels:       common.JSONMap{"levels": []int{3, 4, 5}},
			Activities:        common.JSONMap{"items": []string{"pilgrimage", "city"}},
			VisaAddonsEnabled: true, PricingLogicHidden: true, IsActive: true},
		{AgentID: agent2, Name: "Custom Honeymoon", DurationMinDays: 4, DurationMaxDays: 14,
			TransportTypes:    common.JSONMap{"options": []string{"flight", "car", "yacht"}},
			HotelLevels:       common.JSONMap{"levels": []int{4, 5}},
			Activities:        common.JSONMap{"items": []string{"romance", "beach", "spa"}},
			VisaAddonsEnabled: true, PricingLogicHidden: true, IsActive: true},
	}
	if err := db.Create(&templates).Error; err != nil {
		return err
	}

	// Country join rows per template (indices into templates slice)
	countryMap := map[int][]string{
		0: {"AE", "SA", "QA", "KW", "BH", "OM"},
		1: {"GB", "FR", "DE", "IT", "ES", "NL", "CH", "TR"},
		2: {"TH", "MY", "SG", "ID", "VN", "JP", "KR", "CN"},
		3: {"US", "CA", "MX", "BR", "AR"},
		4: {"ZA", "KE", "NG", "ET", "EG", "MA"},
		5: {"SA", "JO", "PS"},
		6: {"MV", "TH", "ID", "TR", "AE"},
	}
	countries := []catmodel.CustomTripTemplateCountry{}
	for idx, codes := range countryMap {
		for _, code := range codes {
			if cid := countryByCode(db, code); cid != 0 {
				countries = append(countries, catmodel.CustomTripTemplateCountry{
					TemplateID: templates[idx].ID, CountryID: cid,
				})
			}
		}
	}
	if err := db.Create(&countries).Error; err != nil {
		return err
	}

	// Requirement profiles for the first two templates
	first := templates[0].ID
	second := templates[1].ID
	profiles := []catmodel.CustomTripRequirementProfile{
		{TemplateID: first, CountryID: countryByCode(db, "TR"), DocumentTypeID: docTypeID(db, "passport_copy"), TravelerField: "passport_copy", IsRequired: true, Phase: "before"},
		{TemplateID: first, CountryID: countryByCode(db, "TR"), DocumentTypeID: docTypeID(db, "personal_photo"), TravelerField: "personal_photo", IsRequired: true, Phase: "after"},
		{TemplateID: first, CountryID: countryByCode(db, "SA"), DocumentTypeID: docTypeID(db, "passport_copy"), TravelerField: "passport_copy", IsRequired: true, Phase: "before"},
		{TemplateID: first, CountryID: countryByCode(db, "SA"), DocumentTypeID: docTypeID(db, "vaccination_certificate"), TravelerField: "vaccination_certificate", IsRequired: true, Phase: "after"},
		{TemplateID: second, CountryID: countryByCode(db, "GB"), DocumentTypeID: docTypeID(db, "passport_copy"), TravelerField: "passport_copy", IsRequired: true, Phase: "before"},
		{TemplateID: second, CountryID: countryByCode(db, "FR"), DocumentTypeID: docTypeID(db, "passport_copy"), TravelerField: "passport_copy", IsRequired: true, Phase: "before"},
		{TemplateID: second, CountryID: countryByCode(db, "FR"), DocumentTypeID: docTypeID(db, "travel_insurance"), TravelerField: "travel_insurance", IsRequired: true, Phase: "before"},
	}
	return db.Create(&profiles).Error
}

// small helpers (package-scoped)
func uintPtr(v uint) *uint { return &v }
func intPtr(v int) *int    { return &v }
func tAdd(t time.Time, h int) *time.Time {
	n := t.Add(time.Duration(h) * time.Hour)
	return &n
}
