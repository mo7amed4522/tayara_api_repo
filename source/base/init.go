// Package base seeds the geographic reference data that the core `service/base` initializer
// does not already cover (countries/document-types/system-settings are seeded there). This
// package adds cities, nationalities and package categories so a test app has a full, varied
// reference dictionary to drive packages, custom-trip builders, ads and KYC.
//
// Tables are created by AutoMigrate (initialize/gorm_biz.go); this initializer only seeds rows.
package base

import (
	"context"

	basemodel "github.com/mo7amed4522/tayara_api/model/base"
	"github.com/mo7amed4522/tayara_api/model/common"
	"github.com/mo7amed4522/tayara_api/service/system"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// initOrderGeography runs right after the core base_reference_data seed (order InitOrderExternal-3).
const initOrderGeography = system.InitOrderExternal + 1

type geographyInit struct{}

func init() {
	system.RegisterInit(initOrderGeography, &geographyInit{})
}

func (g *geographyInit) InitializerName() string { return "demo_geography_reference_data" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (g *geographyInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (g *geographyInit) TableCreated(ctx context.Context) bool                     { return true }

func (g *geographyInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var c int64
	if err := db.Model(&basemodel.City{}).Count(&c).Error; err != nil {
		return false
	}
	if c == 0 {
		return false
	}
	var n int64
	if err := db.Model(&basemodel.Nationality{}).Count(&n).Error; err != nil || n == 0 {
		return false
	}
	var p int64
	return db.Model(&basemodel.PackageCategory{}).Count(&p).Error == nil && p > 0
}

func (g *geographyInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	// Countries referenced below must exist (seeded by service/base). If they are missing we
	// still seed cities/nationalities and let GORM map countryId to whatever is present.
	if err := seedCountryKeywords(db); err != nil {
		return ctx, err
	}
	if err := seedCities(db); err != nil {
		return ctx, err
	}
	if err := seedNationalities(db); err != nil {
		return ctx, err
	}
	if err := seedPackageCategories(db); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// countryByCode resolves the numeric ID of a country row (seeded by service/base).
func countryByCode(db *gorm.DB, code string) uint {
	var c basemodel.Country
	if err := db.Where("code = ?", code).First(&c).Error; err != nil {
		return 0
	}
	return c.ID
}

// seedCountryKeywords backfills search keywords/aliases on countries created by service/base.
// Uses Updates so it is safe to re-run; only touches the keywords column.
func seedCountryKeywords(db *gorm.DB) error {
	type kw struct {
		code string
		list []string
	}
	updates := []kw{
		{"AE", []string{"UAE", "United Arab Emirates", "Emirates", "AE", "Dubai", "Abu Dhabi", "Sharjah"}},
		{"SA", []string{"Saudi Arabia", "KSA", "SA", "Riyadh", "Jeddah", "Mecca", "Medina"}},
		{"QA", []string{"Qatar", "QA", "Doha"}},
		{"KW", []string{"Kuwait", "KW", "Kuwait City"}},
		{"BH", []string{"Bahrain", "BH", "Manama"}},
		{"OM", []string{"Oman", "OM", "Muscat"}},
		{"JO", []string{"Jordan", "JO", "Amman", "Petra"}},
		{"LB", []string{"Lebanon", "LB", "Beirut"}},
		{"IQ", []string{"Iraq", "IQ", "Baghdad", "Erbil"}},
		{"SY", []string{"Syria", "SY", "Damascus"}},
		{"YE", []string{"Yemen", "YE", "Sanaa"}},
		{"PS", []string{"Palestine", "PS", "Ramallah", "Gaza"}},
		{"EG", []string{"Egypt", "EG", "Misr", "Cairo", "Alexandria", "Sharm El Sheikh"}},
		{"MA", []string{"Morocco", "MA", "Marrakesh", "Casablanca"}},
		{"TN", []string{"Tunisia", "TN", "Tunis"}},
		{"DZ", []string{"Algeria", "DZ", "Algiers"}},
		{"LY", []string{"Libya", "LY", "Tripoli"}},
		{"SD", []string{"Sudan", "SD", "Khartoum"}},
		{"GB", []string{"UK", "United Kingdom", "GB", "Britain", "England", "London"}},
		{"FR", []string{"France", "FR", "Paris", "Nice"}},
		{"DE", []string{"Germany", "DE", "Berlin", "Munich", "Frankfurt"}},
		{"IT", []string{"Italy", "IT", "Rome", "Milan", "Venice"}},
		{"ES", []string{"Spain", "ES", "Madrid", "Barcelona"}},
		{"NL", []string{"Netherlands", "NL", "Holland", "Amsterdam"}},
		{"CH", []string{"Switzerland", "CH", "Zurich", "Geneva"}},
		{"AT", []string{"Austria", "AT", "Vienna"}},
		{"BE", []string{"Belgium", "BE", "Brussels"}},
		{"SE", []string{"Sweden", "SE", "Stockholm"}},
		{"NO", []string{"Norway", "NO", "Oslo"}},
		{"DK", []string{"Denmark", "DK", "Copenhagen"}},
		{"FI", []string{"Finland", "FI", "Helsinki"}},
		{"PT", []string{"Portugal", "PT", "Lisbon", "Porto"}},
		{"GR", []string{"Greece", "GR", "Athens", "Santorini"}},
		{"IE", []string{"Ireland", "IE", "Dublin"}},
		{"PL", []string{"Poland", "PL", "Warsaw"}},
		{"CZ", []string{"Czech Republic", "Czechia", "CZ", "Prague"}},
		{"HU", []string{"Hungary", "HU", "Budapest"}},
		{"RO", []string{"Romania", "RO", "Bucharest"}},
		{"RU", []string{"Russia", "RU", "Moscow", "St Petersburg"}},
		{"UA", []string{"Ukraine", "UA", "Kyiv"}},
		{"TR", []string{"Turkey", "TR", "Türkiye", "Istanbul", "Antalya", "Ankara"}},
		{"IN", []string{"India", "IN", "Bharat", "Delhi", "Mumbai", "Bangalore"}},
		{"PK", []string{"Pakistan", "PK", "Karachi", "Lahore", "Islamabad"}},
		{"BD", []string{"Bangladesh", "BD", "Dhaka"}},
		{"LK", []string{"Sri Lanka", "LK", "Colombo"}},
		{"NP", []string{"Nepal", "NP", "Kathmandu"}},
		{"PH", []string{"Philippines", "PH", "Manila", "Cebu"}},
		{"ID", []string{"Indonesia", "ID", "Jakarta", "Bali"}},
		{"MY", []string{"Malaysia", "MY", "Kuala Lumpur"}},
		{"SG", []string{"Singapore", "SG"}},
		{"TH", []string{"Thailand", "TH", "Bangkok", "Phuket"}},
		{"VN", []string{"Vietnam", "VN", "Hanoi", "Ho Chi Minh"}},
		{"CN", []string{"China", "CN", "Beijing", "Shanghai"}},
		{"JP", []string{"Japan", "JP", "Tokyo", "Osaka", "Kyoto"}},
		{"KR", []string{"South Korea", "Korea", "KR", "Seoul"}},
		{"AZ", []string{"Azerbaijan", "AZ", "Baku"}},
		{"GE", []string{"Georgia", "GE", "Tbilisi"}},
		{"AM", []string{"Armenia", "AM", "Yerevan"}},
		{"KZ", []string{"Kazakhstan", "KZ", "Almaty", "Astana"}},
		{"UZ", []string{"Uzbekistan", "UZ", "Tashkent", "Samarkand"}},
		{"US", []string{"USA", "United States", "US", "America", "New York", "Los Angeles", "Chicago", "Miami"}},
		{"CA", []string{"Canada", "CA", "Toronto", "Vancouver"}},
		{"MX", []string{"Mexico", "MX", "Mexico City", "Cancun"}},
		{"BR", []string{"Brazil", "BR", "Sao Paulo", "Rio de Janeiro"}},
		{"AR", []string{"Argentina", "AR", "Buenos Aires"}},
		{"AU", []string{"Australia", "AU", "Sydney", "Melbourne"}},
		{"NZ", []string{"New Zealand", "NZ", "Auckland"}},
		{"ZA", []string{"South Africa", "ZA", "Cape Town", "Johannesburg"}},
		{"KE", []string{"Kenya", "KE", "Nairobi"}},
		{"NG", []string{"Nigeria", "NG", "Lagos", "Abuja"}},
		{"ET", []string{"Ethiopia", "ET", "Addis Ababa"}},
	}

	for _, u := range updates {
		payload := basemodel.Country{
			Keywords: common.JSONMap{"keywords": u.list},
		}
		if err := db.Model(&basemodel.Country{}).
			Where("code = ?", u.code).
			Updates(payload).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedCities(db *gorm.DB) error {
	// City rows reference countries by their numeric ID (resolved from code),
	// and carry search keyword aliases used by autocomplete and match logic.
	type citySeed struct {
		country  string
		name     string
		keywords []string
	}

	seeds := []citySeed{
		// UAE
		{"AE", "Dubai", []string{"Dubai", "DXB", "Dubayy"}},
		{"AE", "Abu Dhabi", []string{"Abu Dhabi", "AUH", "Abu Zabi"}},
		{"AE", "Sharjah", []string{"Sharjah", "SHJ"}},
		{"AE", "Ajman", []string{"Ajman", "AJM"}},
		{"AE", "Ras Al Khaimah", []string{"Ras Al Khaimah", "RAK"}},
		{"AE", "Fujairah", []string{"Fujairah", "FJR"}},
		{"AE", "Umm Al Quwain", []string{"Umm Al Quwain", "UAQ"}},
		{"AE", "Al Ain", []string{"Al Ain", "AAN"}},

		// Saudi Arabia
		{"SA", "Riyadh", []string{"Riyadh", "RUH"}},
		{"SA", "Jeddah", []string{"Jeddah", "JED"}},
		{"SA", "Mecca", []string{"Mecca", "Makkah", "MKX"}},
		{"SA", "Medina", []string{"Medina", "Madinah", "MED"}},
		{"SA", "Dammam", []string{"Dammam", "DMM"}},
		{"SA", "Taif", []string{"Taif", "TIF"}},
		{"SA", "Abha", []string{"Abha", "AHB"}},
		{"SA", "Al Khobar", []string{"Al Khobar", "Khobar", "DHA"}},

		// Qatar / Kuwait / Bahrain / Oman
		{"QA", "Doha", []string{"Doha", "DOH"}},
		{"QA", "Al Wakrah", []string{"Al Wakrah", "Wakra"}},
		{"KW", "Kuwait City", []string{"Kuwait City", "KWI"}},
		{"KW", "Hawalli", []string{"Hawalli"}},
		{"BH", "Manama", []string{"Manama", "BAH"}},
		{"BH", "Muharraq", []string{"Muharraq"}},
		{"OM", "Muscat", []string{"Muscat", "MCT"}},
		{"OM", "Salalah", []string{"Salalah", "SLL"}},

		// Levant / Iraq
		{"JO", "Amman", []string{"Amman", "AMM"}},
		{"JO", "Aqaba", []string{"Aqaba", "AQJ"}},
		{"JO", "Petra", []string{"Petra", "Wadi Musa"}},
		{"LB", "Beirut", []string{"Beirut", "BEY"}},
		{"LB", "Tripoli", []string{"Tripoli", "Tarabulus"}},
		{"IQ", "Baghdad", []string{"Baghdad", "BGW"}},
		{"IQ", "Erbil", []string{"Erbil", "Arbil", "EBL"}},
		{"IQ", "Basra", []string{"Basra", "BSR"}},
		{"SY", "Damascus", []string{"Damascus", "DAM"}},
		{"SY", "Aleppo", []string{"Aleppo", "ALP"}},
		{"PS", "Ramallah", []string{"Ramallah"}},
		{"PS", "Bethlehem", []string{"Bethlehem"}},

		// North Africa
		{"EG", "Cairo", []string{"Cairo", "CAI", "Qahira"}},
		{"EG", "Alexandria", []string{"Alexandria", "ALY"}},
		{"EG", "Sharm El Sheikh", []string{"Sharm El Sheikh", "Sharm", "SSH"}},
		{"EG", "Hurghada", []string{"Hurghada", "HRG"}},
		{"EG", "Luxor", []string{"Luxor", "LXR"}},
		{"EG", "Aswan", []string{"Aswan", "ASW"}},
		{"EG", "Marsa Alam", []string{"Marsa Alam", "RMF"}},
		{"MA", "Marrakesh", []string{"Marrakesh", "Marrakech", "RAK"}},
		{"MA", "Casablanca", []string{"Casablanca", "CMN"}},
		{"MA", "Rabat", []string{"Rabat", "RBA"}},
		{"MA", "Fes", []string{"Fes", "Fez", "FEZ"}},
		{"MA", "Tangier", []string{"Tangier", "TNG"}},
		{"TN", "Tunis", []string{"Tunis", "TUN"}},
		{"TN", "Sousse", []string{"Sousse", "Susa"}},
		{"TN", "Djerba", []string{"Djerba", "DJE"}},
		{"DZ", "Algiers", []string{"Algiers", "ALG"}},
		{"DZ", "Oran", []string{"Oran", "ORN"}},
		{"LY", "Tripoli", []string{"Tripoli", "TIP"}},
		{"LY", "Benghazi", []string{"Benghazi", "BEN"}},
		{"SD", "Khartoum", []string{"Khartoum", "KRT"}},

		// Turkey
		{"TR", "Istanbul", []string{"Istanbul", "IST", "Constantinople"}},
		{"TR", "Antalya", []string{"Antalya", "AYT"}},
		{"TR", "Ankara", []string{"Ankara", "ESB"}},
		{"TR", "Izmir", []string{"Izmir", "ADB"}},
		{"TR", "Bodrum", []string{"Bodrum", "BJV"}},
		{"TR", "Cappadocia", []string{"Cappadocia", "Kapadokya"}},
		{"TR", "Trabzon", []string{"Trabzon", "TZX"}},
		{"TR", "Bursa", []string{"Bursa"}},

		// Europe
		{"GB", "London", []string{"London", "LON", "LHR", "LGW"}},
		{"GB", "Manchester", []string{"Manchester", "MAN"}},
		{"GB", "Edinburgh", []string{"Edinburgh", "EDI"}},
		{"GB", "Birmingham", []string{"Birmingham", "BHX"}},
		{"GB", "Glasgow", []string{"Glasgow", "GLA"}},
		{"FR", "Paris", []string{"Paris", "PAR", "CDG"}},
		{"FR", "Nice", []string{"Nice", "NCE"}},
		{"FR", "Lyon", []string{"Lyon", "LYS"}},
		{"FR", "Marseille", []string{"Marseille", "MRS"}},
		{"DE", "Berlin", []string{"Berlin", "BER"}},
		{"DE", "Munich", []string{"Munich", "Munchen", "MUC"}},
		{"DE", "Frankfurt", []string{"Frankfurt", "FRA"}},
		{"DE", "Hamburg", []string{"Hamburg", "HAM"}},
		{"IT", "Rome", []string{"Rome", "Roma", "ROM", "FCO"}},
		{"IT", "Milan", []string{"Milan", "Milano", "MIL", "MXP"}},
		{"IT", "Venice", []string{"Venice", "Venezia", "VCE"}},
		{"IT", "Florence", []string{"Florence", "Firenze", "FLR"}},
		{"IT", "Naples", []string{"Naples", "Napoli", "NAP"}},
		{"ES", "Madrid", []string{"Madrid", "MAD"}},
		{"ES", "Barcelona", []string{"Barcelona", "BCN"}},
		{"ES", "Seville", []string{"Seville", "Sevilla", "SVQ"}},
		{"ES", "Valencia", []string{"Valencia", "VLC"}},
		{"NL", "Amsterdam", []string{"Amsterdam", "AMS"}},
		{"NL", "Rotterdam", []string{"Rotterdam", "RTM"}},
		{"CH", "Zurich", []string{"Zurich", "Zürich", "ZRH"}},
		{"CH", "Geneva", []string{"Geneva", "Geneve", "GVA"}},
		{"CH", "Lucerne", []string{"Lucerne", "Luzern"}},
		{"AT", "Vienna", []string{"Vienna", "Wien", "VIE"}},
		{"AT", "Salzburg", []string{"Salzburg", "SZG"}},
		{"BE", "Brussels", []string{"Brussels", "Bruxelles", "BRU"}},
		{"BE", "Antwerp", []string{"Antwerp", "Antwerpen", "ANR"}},
		{"SE", "Stockholm", []string{"Stockholm", "ARN"}},
		{"NO", "Oslo", []string{"Oslo", "OSL"}},
		{"DK", "Copenhagen", []string{"Copenhagen", "Kobenhavn", "CPH"}},
		{"FI", "Helsinki", []string{"Helsinki", "HEL"}},
		{"PT", "Lisbon", []string{"Lisbon", "Lisboa", "LIS"}},
		{"PT", "Porto", []string{"Porto", "OPO"}},
		{"GR", "Athens", []string{"Athens", "Athina", "ATH"}},
		{"GR", "Santorini", []string{"Santorini", "Thira", "JTR"}},
		{"GR", "Mykonos", []string{"Mykonos", "JMK"}},
		{"IE", "Dublin", []string{"Dublin", "DUB"}},
		{"PL", "Warsaw", []string{"Warsaw", "Warszawa", "WAW"}},
		{"PL", "Krakow", []string{"Krakow", "Cracow", "KRK"}},
		{"CZ", "Prague", []string{"Prague", "Praha", "PRG"}},
		{"HU", "Budapest", []string{"Budapest", "BUD"}},
		{"RO", "Bucharest", []string{"Bucharest", "Bucuresti", "OTP"}},
		{"RU", "Moscow", []string{"Moscow", "Moskva", "SVO"}},
		{"RU", "St Petersburg", []string{"St Petersburg", "Saint Petersburg", "LED"}},
		{"UA", "Kyiv", []string{"Kyiv", "Kiev", "KBP"}},

		// Asia
		{"IN", "Mumbai", []string{"Mumbai", "Bombay", "BOM"}},
		{"IN", "Delhi", []string{"Delhi", "New Delhi", "DEL"}},
		{"IN", "Bangalore", []string{"Bangalore", "Bengaluru", "BLR"}},
		{"IN", "Chennai", []string{"Chennai", "Madras", "MAA"}},
		{"IN", "Kolkata", []string{"Kolkata", "Calcutta", "CCU"}},
		{"IN", "Hyderabad", []string{"Hyderabad", "HYD"}},
		{"IN", "Goa", []string{"Goa", "GOI"}},
		{"IN", "Kerala", []string{"Kerala", "Cochin", "Kochi", "COK"}},
		{"PK", "Karachi", []string{"Karachi", "KHI"}},
		{"PK", "Lahore", []string{"Lahore", "LHE"}},
		{"PK", "Islamabad", []string{"Islamabad", "ISB"}},
		{"BD", "Dhaka", []string{"Dhaka", "DAC"}},
		{"BD", "Chittagong", []string{"Chittagong", "Chattogram", "CGP"}},
		{"LK", "Colombo", []string{"Colombo", "CMB"}},
		{"NP", "Kathmandu", []string{"Kathmandu", "KTM"}},
		{"PH", "Manila", []string{"Manila", "MNL"}},
		{"PH", "Cebu", []string{"Cebu", "CEB"}},
		{"ID", "Jakarta", []string{"Jakarta", "CGK"}},
		{"ID", "Bali", []string{"Bali", "Denpasar", "DPS"}},
		{"ID", "Surabaya", []string{"Surabaya", "SUB"}},
		{"MY", "Kuala Lumpur", []string{"Kuala Lumpur", "KL", "KUL"}},
		{"MY", "Penang", []string{"Penang", "PEN"}},
		{"SG", "Singapore", []string{"Singapore", "SIN"}},
		{"TH", "Bangkok", []string{"Bangkok", "BKK"}},
		{"TH", "Phuket", []string{"Phuket", "HKT"}},
		{"TH", "Chiang Mai", []string{"Chiang Mai", "CNX"}},
		{"TH", "Pattaya", []string{"Pattaya", "UTP"}},
		{"TH", "Krabi", []string{"Krabi", "KBV"}},
		{"VN", "Hanoi", []string{"Hanoi", "HAN"}},
		{"VN", "Ho Chi Minh City", []string{"Ho Chi Minh City", "Saigon", "SGN"}},
		{"VN", "Da Nang", []string{"Da Nang", "Danang", "DAD"}},
		{"CN", "Beijing", []string{"Beijing", "Peking", "PEK"}},
		{"CN", "Shanghai", []string{"Shanghai", "PVG"}},
		{"CN", "Guangzhou", []string{"Guangzhou", "Canton", "CAN"}},
		{"JP", "Tokyo", []string{"Tokyo", "TYO", "NRT", "HND"}},
		{"JP", "Osaka", []string{"Osaka", "KIX"}},
		{"JP", "Kyoto", []string{"Kyoto"}},
		{"KR", "Seoul", []string{"Seoul", "SEL", "ICN"}},
		{"KR", "Busan", []string{"Busan", "PUS"}},
		{"AZ", "Baku", []string{"Baku", "GYD"}},
		{"GE", "Tbilisi", []string{"Tbilisi", "TBS"}},
		{"GE", "Batumi", []string{"Batumi", "BUS"}},
		{"AM", "Yerevan", []string{"Yerevan", "EVN"}},
		{"KZ", "Almaty", []string{"Almaty", "ALA"}},
		{"KZ", "Astana", []string{"Astana", "Nur-Sultan", "NQZ"}},
		{"UZ", "Tashkent", []string{"Tashkent", "TAS"}},
		{"UZ", "Samarkand", []string{"Samarkand", "SKD"}},

		// Americas
		{"US", "New York", []string{"New York", "NYC", "JFK", "Manhattan"}},
		{"US", "Los Angeles", []string{"Los Angeles", "LA", "LAX"}},
		{"US", "Chicago", []string{"Chicago", "ORD"}},
		{"US", "Miami", []string{"Miami", "MIA"}},
		{"US", "San Francisco", []string{"San Francisco", "SF", "SFO"}},
		{"US", "Washington DC", []string{"Washington DC", "DC", "IAD"}},
		{"US", "Las Vegas", []string{"Las Vegas", "Vegas", "LAS"}},
		{"US", "Orlando", []string{"Orlando", "MCO"}},
		{"CA", "Toronto", []string{"Toronto", "YYZ"}},
		{"CA", "Vancouver", []string{"Vancouver", "YVR"}},
		{"CA", "Montreal", []string{"Montreal", "YUL"}},
		{"MX", "Mexico City", []string{"Mexico City", "CDMX", "MEX"}},
		{"MX", "Cancun", []string{"Cancun", "CUN"}},
		{"BR", "Sao Paulo", []string{"Sao Paulo", "GRU"}},
		{"BR", "Rio de Janeiro", []string{"Rio de Janeiro", "Rio", "GIG"}},
		{"AR", "Buenos Aires", []string{"Buenos Aires", "EZE"}},

		// Oceania
		{"AU", "Sydney", []string{"Sydney", "SYD"}},
		{"AU", "Melbourne", []string{"Melbourne", "MEL"}},
		{"AU", "Brisbane", []string{"Brisbane", "BNE"}},
		{"AU", "Perth", []string{"Perth", "PER"}},
		{"NZ", "Auckland", []string{"Auckland", "AKL"}},
		{"NZ", "Wellington", []string{"Wellington", "WLG"}},

		// Africa (sub-Saharan)
		{"ZA", "Cape Town", []string{"Cape Town", "CPT"}},
		{"ZA", "Johannesburg", []string{"Johannesburg", "Joburg", "JNB"}},
		{"KE", "Nairobi", []string{"Nairobi", "NBO"}},
		{"KE", "Mombasa", []string{"Mombasa", "MBA"}},
		{"NG", "Lagos", []string{"Lagos", "LOS"}},
		{"NG", "Abuja", []string{"Abuja", "ABV"}},
		{"ET", "Addis Ababa", []string{"Addis Ababa", "Addis", "ADD"}},
	}

	cities := make([]basemodel.City, 0, len(seeds))
	for _, s := range seeds {
		cid := countryByCode(db, s.country)
		if cid == 0 {
			// Country not present — skip rather than insert a dangling row.
			continue
		}
		cities = append(cities, basemodel.City{
			CountryID: cid,
			Name:      s.name,
			IsActive:  true,
			Keywords:  common.JSONMap{"keywords": s.keywords},
		})
	}

	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
		Create(&cities).Error
}

func seedNationalities(db *gorm.DB) error {
	nationalities := []basemodel.Nationality{
		// GCC / Middle East
		{Code: "AE", Name: "Emirati", IsActive: true},
		{Code: "SA", Name: "Saudi", IsActive: true},
		{Code: "QA", Name: "Qatari", IsActive: true},
		{Code: "KW", Name: "Kuwaiti", IsActive: true},
		{Code: "BH", Name: "Bahraini", IsActive: true},
		{Code: "OM", Name: "Omani", IsActive: true},
		{Code: "JO", Name: "Jordanian", IsActive: true},
		{Code: "LB", Name: "Lebanese", IsActive: true},
		{Code: "IQ", Name: "Iraqi", IsActive: true},
		{Code: "SY", Name: "Syrian", IsActive: true},
		{Code: "YE", Name: "Yemeni", IsActive: true},
		{Code: "PS", Name: "Palestinian", IsActive: true},

		// North Africa
		{Code: "EG", Name: "Egyptian", IsActive: true},
		{Code: "MA", Name: "Moroccan", IsActive: true},
		{Code: "TN", Name: "Tunisian", IsActive: true},
		{Code: "DZ", Name: "Algerian", IsActive: true},
		{Code: "LY", Name: "Libyan", IsActive: true},
		{Code: "SD", Name: "Sudanese", IsActive: true},

		// Europe
		{Code: "GB", Name: "British", IsActive: true},
		{Code: "FR", Name: "French", IsActive: true},
		{Code: "DE", Name: "German", IsActive: true},
		{Code: "IT", Name: "Italian", IsActive: true},
		{Code: "ES", Name: "Spanish", IsActive: true},
		{Code: "NL", Name: "Dutch", IsActive: true},
		{Code: "CH", Name: "Swiss", IsActive: true},
		{Code: "AT", Name: "Austrian", IsActive: true},
		{Code: "BE", Name: "Belgian", IsActive: true},
		{Code: "SE", Name: "Swedish", IsActive: true},
		{Code: "NO", Name: "Norwegian", IsActive: true},
		{Code: "DK", Name: "Danish", IsActive: true},
		{Code: "FI", Name: "Finnish", IsActive: true},
		{Code: "PT", Name: "Portuguese", IsActive: true},
		{Code: "GR", Name: "Greek", IsActive: true},
		{Code: "IE", Name: "Irish", IsActive: true},
		{Code: "PL", Name: "Polish", IsActive: true},
		{Code: "CZ", Name: "Czech", IsActive: true},
		{Code: "HU", Name: "Hungarian", IsActive: true},
		{Code: "RO", Name: "Romanian", IsActive: true},
		{Code: "RU", Name: "Russian", IsActive: true},
		{Code: "UA", Name: "Ukrainian", IsActive: true},
		{Code: "TR", Name: "Turkish", IsActive: true},

		// Asia
		{Code: "IN", Name: "Indian", IsActive: true},
		{Code: "PK", Name: "Pakistani", IsActive: true},
		{Code: "BD", Name: "Bangladeshi", IsActive: true},
		{Code: "LK", Name: "Sri Lankan", IsActive: true},
		{Code: "NP", Name: "Nepali", IsActive: true},
		{Code: "PH", Name: "Filipino", IsActive: true},
		{Code: "ID", Name: "Indonesian", IsActive: true},
		{Code: "MY", Name: "Malaysian", IsActive: true},
		{Code: "SG", Name: "Singaporean", IsActive: true},
		{Code: "TH", Name: "Thai", IsActive: true},
		{Code: "VN", Name: "Vietnamese", IsActive: true},
		{Code: "CN", Name: "Chinese", IsActive: true},
		{Code: "JP", Name: "Japanese", IsActive: true},
		{Code: "KR", Name: "South Korean", IsActive: true},
		{Code: "AZ", Name: "Azerbaijani", IsActive: true},
		{Code: "GE", Name: "Georgian", IsActive: true},
		{Code: "AM", Name: "Armenian", IsActive: true},
		{Code: "KZ", Name: "Kazakh", IsActive: true},
		{Code: "UZ", Name: "Uzbek", IsActive: true},

		// Americas
		{Code: "US", Name: "American", IsActive: true},
		{Code: "CA", Name: "Canadian", IsActive: true},
		{Code: "MX", Name: "Mexican", IsActive: true},
		{Code: "BR", Name: "Brazilian", IsActive: true},
		{Code: "AR", Name: "Argentine", IsActive: true},

		// Oceania
		{Code: "AU", Name: "Australian", IsActive: true},
		{Code: "NZ", Name: "New Zealander", IsActive: true},

		// Africa (sub-Saharan)
		{Code: "ZA", Name: "South African", IsActive: true},
		{Code: "KE", Name: "Kenyan", IsActive: true},
		{Code: "NG", Name: "Nigerian", IsActive: true},
		{Code: "ET", Name: "Ethiopian", IsActive: true},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).
		Create(&nationalities).Error
}

func seedPackageCategories(db *gorm.DB) error {
	categories := []basemodel.PackageCategory{
		{Code: "honeymoon", Name: "Honeymoon", Description: "Romantic escapes for couples", SortOrder: 1, IsActive: true},
		{Code: "adventure", Name: "Adventure", Description: "Adrenaline and outdoor trips", SortOrder: 2, IsActive: true},
		{Code: "family", Name: "Family", Description: "Family-friendly packages", SortOrder: 3, IsActive: true},
		{Code: "beach", Name: "Beach & Relaxation", Description: "Coastal and resort stays", SortOrder: 4, IsActive: true},
		{Code: "cultural", Name: "Cultural", Description: "Heritage and city tours", SortOrder: 5, IsActive: true},
		{Code: "luxury", Name: "Luxury", Description: "Premium, high-end travel", SortOrder: 6, IsActive: true},
		{Code: "budget", Name: "Budget", Description: "Value-focused getaways", SortOrder: 7, IsActive: true},
		{Code: "business", Name: "Business", Description: "Corporate and business travel", SortOrder: 8, IsActive: true},
		{Code: "wellness", Name: "Wellness & Spa", Description: "Health, spa and retreat packages", SortOrder: 9, IsActive: true},
		{Code: "religious", Name: "Religious & Pilgrimage", Description: "Umrah, Hajj and spiritual trips", SortOrder: 10, IsActive: true},
		{Code: "safari", Name: "Safari & Wildlife", Description: "Wildlife and nature expeditions", SortOrder: 11, IsActive: true},
		{Code: "city_break", Name: "City Break", Description: "Short city getaways", SortOrder: 12, IsActive: true},
		{Code: "cruise", Name: "Cruise", Description: "Ocean and river cruises", SortOrder: 13, IsActive: true},
		{Code: "ski", Name: "Ski & Winter", Description: "Snow and winter sports", SortOrder: 14, IsActive: true},
		{Code: "road_trip", Name: "Road Trip", Description: "Self-drive and overland trips", SortOrder: 15, IsActive: true},
		{Code: "educational", Name: "Educational", Description: "Study trips and student tours", SortOrder: 16, IsActive: true},
		{Code: "medical", Name: "Medical Tourism", Description: "Treatment and recovery travel", SortOrder: 17, IsActive: true},
		{Code: "sports", Name: "Sports & Events", Description: "Matches, races and events", SortOrder: 18, IsActive: true},
		{Code: "weekend", Name: "Weekend Escape", Description: "Short domestic weekend trips", SortOrder: 19, IsActive: true},
		{Code: "group", Name: "Group Travel", Description: "Large group and incentive trips", SortOrder: 20, IsActive: true},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).
		Create(&categories).Error
}
