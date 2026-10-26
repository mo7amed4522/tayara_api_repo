// Package dxapi seeds agent plans, agent profiles (varied approval states), verification
// documents and internal admin notes so the Manage (agent) dashboard has realistic data.
package dxapi

import (
	"context"
	"time"

	basemodel "github.com/mo7amed4522/tayara_api/model/base"
	dxmodel "github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/service/system"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const initOrderAgents = system.InitOrderExternal + 3

type agentInit struct{}

func init() {
	system.RegisterInit(initOrderAgents, &agentInit{})
}

func (a *agentInit) InitializerName() string { return "demo_agents" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (a *agentInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (a *agentInit) TableCreated(ctx context.Context) bool                     { return true }

func (a *agentInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var c int64
	return db.Model(&dxmodel.DxAgentProfile{}).Count(&c).Error == nil && c > 0
}

func (a *agentInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	if err := seedPlans(db); err != nil {
		return ctx, err
	}
	agents, err := seedAgents(db)
	if err != nil {
		return ctx, err
	}
	if err := seedAgentDocuments(db, agents); err != nil {
		return ctx, err
	}
	if err := seedAgentNotes(db, agents); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// ── Plans ────────────────────────────────────────────────────────────

func seedPlans(db *gorm.DB) error {
	plans := []dxmodel.AgentPlan{
		{Name: "basic", DefaultCommissionPercent: 500, Description: "Entry plan, 5% commission", IsActive: true},
		{Name: "premium", DefaultCommissionPercent: 1000, Description: "Preferred plan, 10% commission", IsActive: true},
		{Name: "custom", DefaultCommissionPercent: 800, Description: "Negotiated plan, 8% commission", IsActive: true},
		{Name: "enterprise", DefaultCommissionPercent: 1200, Description: "Enterprise plan, 12% commission", IsActive: true},
		{Name: "starter", DefaultCommissionPercent: 400, Description: "Starter plan, 4% commission", IsActive: true},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
		Create(&plans).Error
}

func planID(db *gorm.DB, name string) *uint {
	var p dxmodel.AgentPlan
	if err := db.Where("name = ?", name).First(&p).Error; err != nil {
		return nil
	}
	return &p.ID
}

// countryIDByCode resolves base_country.id by ISO code (seeded by service/base).
func countryIDByCode(db *gorm.DB, code string) *uint {
	var c basemodel.Country
	if err := db.Select("id").Where("code = ?", code).First(&c).Error; err != nil {
		return nil
	}
	return &c.ID
}

// cityIDByName resolves base_city.id by name (seeded by demo geography).
func cityIDByName(db *gorm.DB, name string) *uint {
	var c basemodel.City
	if err := db.Select("id").Where("name = ?", name).First(&c).Error; err != nil {
		return nil
	}
	return &c.ID
}

// ── Agents ───────────────────────────────────────────────────────────

func seedAgents(db *gorm.DB) ([]dxmodel.DxAgentProfile, error) {
	password := utils.BcryptHash("123456")
	now := time.Now()
	approved := now.Add(-720 * time.Hour)
	approved2 := now.Add(-480 * time.Hour)
	approved3 := now.Add(-240 * time.Hour)
	lastLogin := now.Add(-6 * time.Hour)
	lastLogin2 := now.Add(-48 * time.Hour)

	agents := []dxmodel.DxAgentProfile{
		// 1. Active, premium — flagship demo agent (kept as hello@desertsafari.ae)
		{
			CompanyName:      "Desert Safari Dubai",
			Password:         password,
			ResponsibleName:  "Khalid Al Mansouri",
			HeaderImg:        "https://files.tayara.test/agents/desert-safari.png",
			BackgroundImage:  "https://files.tayara.test/agents/desert-safari-bg.jpg",
			PhoneCountryCode: "+971", Phone: "0551010101",
			BusinessEmail: "hello@desertsafari.ae",
			CountryID:     countryIDByCode(db, "AE"),
			CityID:        cityIDByName(db, "Dubai"),
			Country:       "UAE", City: "Dubai",
			CompanyAddress:        "Al Quoz Industrial Area 3, Warehouse 12, Dubai, UAE",
			LicenseNumber:         "DED-1001",
			Status:                "active",
			PlanID:                planID(db, "premium"),
			ShortDescription:      strPtr("Award-winning desert experiences"),
			LongDescription:       strPtr("Full-day desert safaris, dune bashing and traditional camps across Dubai and Abu Dhabi. Operating since 2012 with over 40,000 happy travellers."),
			TrustedTravelersCount: 4200,
			CommissionPercent:     1000,
			SLAHours:              2,
			ApprovedBy:            uintPtr(1),
			ApprovedAt:            &approved,
			LastLoginAt:           &lastLogin,
		},

		// 2. Active, basic — used as agent2 in catalog seed
		{
			CompanyName:      "Luxury Travel Co",
			Password:         password,
			ResponsibleName:  "Sara Al Noori",
			HeaderImg:        "https://files.tayara.test/agents/luxury-travel.png",
			BackgroundImage:  "https://files.tayara.test/agents/luxury-travel-bg.jpg",
			PhoneCountryCode: "+971", Phone: "0552020202",
			BusinessEmail: "book@luxurytravel.ae",
			CountryID:     countryIDByCode(db, "AE"),
			CityID:        cityIDByName(db, "Dubai"),
			Country:       "UAE", City: "Dubai",
			CompanyAddress:        "Burj Khalifa Blvd, Downtown Dubai, UAE",
			LicenseNumber:         "DED-1002",
			Status:                "active",
			PlanID:                planID(db, "basic"),
			ShortDescription:      strPtr("Luxury curated travel"),
			LongDescription:       strPtr("Bespoke luxury journeys: private jet charters, five-star resorts and concierge-grade itineraries across Europe and Asia."),
			TrustedTravelersCount: 1850,
			CommissionPercent:     500,
			SLAHours:              4,
			ApprovedBy:            uintPtr(1),
			ApprovedAt:            &approved2,
			LastLoginAt:           &lastLogin2,
		},

		// 3. Pending — awaiting first review
		{
			CompanyName:      "Skyline Tours",
			Password:         password,
			ResponsibleName:  "Mohammed Eid",
			HeaderImg:        "https://files.tayara.test/agents/skyline.png",
			BackgroundImage:  "https://files.tayara.test/agents/skyline-bg.jpg",
			PhoneCountryCode: "+971", Phone: "0553030303",
			BusinessEmail: "info@skyline.ae",
			CountryID:     countryIDByCode(db, "AE"),
			CityID:        cityIDByName(db, "Dubai"),
			Country:       "UAE", City: "Dubai",
			CompanyAddress:    "Deira Clocktower, Dubai, UAE",
			LicenseNumber:     "DED-1003",
			Status:            "pending",
			PlanID:            planID(db, "basic"),
			ShortDescription:  strPtr("City tours and hop-on-hop-off"),
			LongDescription:   strPtr("Daily city tours, hop-on-hop-off buses and attraction tickets across the UAE."),
			CommissionPercent: 0,
			SLAHours:          2,
		},

		// 4. Suspended — SLA breach, kept for admin UI demo
		{
			CompanyName:      "Ocean Voyages",
			Password:         password,
			ResponsibleName:  "Nadia Haddad",
			HeaderImg:        "https://files.tayara.test/agents/ocean.png",
			BackgroundImage:  "https://files.tayara.test/agents/ocean-bg.jpg",
			PhoneCountryCode: "+971", Phone: "0554040404",
			BusinessEmail: "support@ocean.ae",
			CountryID:     countryIDByCode(db, "AE"),
			CityID:        cityIDByName(db, "Dubai"),
			Country:       "UAE", City: "Dubai",
			CompanyAddress:    "Dubai Marina, Tower 3, Dubai, UAE",
			LicenseNumber:     "DED-1004",
			Status:            "suspended",
			PlanID:            planID(db, "custom"),
			ShortDescription:  strPtr("Cruises and sea escapes"),
			LongDescription:   strPtr("Mediterranean cruises, Red Sea dive trips and private yacht charters."),
			CommissionPercent: 800,
			SLAHours:          6,
			ApprovedBy:        uintPtr(1),
			ApprovedAt:        &approved2,
			RejectedReason:    "Suspended for repeated SLA breaches (3 in 30 days)",
		},

		// 5. Rejected — invalid license
		{
			CompanyName:      "Global Getaways",
			Password:         password,
			ResponsibleName:  "Tariq Bashir",
			HeaderImg:        "https://files.tayara.test/agents/global.png",
			BackgroundImage:  "https://files.tayara.test/agents/global-bg.jpg",
			PhoneCountryCode: "+971", Phone: "0555050505",
			BusinessEmail: "sales@globalgetaways.ae",
			CountryID:     countryIDByCode(db, "AE"),
			CityID:        cityIDByName(db, "Dubai"),
			Country:       "UAE", City: "Dubai",
			CompanyAddress:   "JLT Cluster C, Dubai, UAE",
			LicenseNumber:    "DED-1005",
			Status:           "rejected",
			PlanID:           nil,
			ShortDescription: strPtr("International holidays"),
			LongDescription:  strPtr("Package holidays to Europe, Asia and the Americas."),
			RejectedReason:   "License documents did not verify; trade license number mismatch with DED registry",
		},

		// 6. Active, enterprise — a large operator
		{
			CompanyName:      "Emirates Travel Group",
			Password:         password,
			ResponsibleName:  "Aisha Rahman",
			HeaderImg:        "https://files.tayara.test/agents/emirates-travel.png",
			BackgroundImage:  "https://files.tayara.test/agents/emirates-travel-bg.jpg",
			PhoneCountryCode: "+971", Phone: "0556060606",
			BusinessEmail: "partners@emiratestravel.ae",
			CountryID:     countryIDByCode(db, "AE"),
			CityID:        cityIDByName(db, "Abu Dhabi"),
			Country:       "UAE", City: "Abu Dhabi",
			CompanyAddress:        "Corniche Road, Tower 5, Abu Dhabi, UAE",
			LicenseNumber:         "DED-2001",
			Status:                "active",
			PlanID:                planID(db, "enterprise"),
			ShortDescription:      strPtr("UAE's largest outbound operator"),
			LongDescription:       strPtr("Enterprise-grade outbound operator with 200+ destination partners across 60 countries. Dedicated 24/7 concierge."),
			TrustedTravelersCount: 12800,
			CommissionPercent:     1200,
			SLAHours:              1,
			ApprovedBy:            uintPtr(1),
			ApprovedAt:            &approved,
			LastLoginAt:           &lastLogin,
		},

		// 7. Active, starter — a new agent that just got approved
		{
			CompanyName:      "Sunset Dhow Cruises",
			Password:         password,
			ResponsibleName:  "Rashid Al Falasi",
			HeaderImg:        "https://files.tayara.test/agents/sunset-dhow.png",
			BackgroundImage:  "https://files.tayara.test/agents/sunset-dhow-bg.jpg",
			PhoneCountryCode: "+971", Phone: "0557070707",
			BusinessEmail: "hello@sunsetdhow.ae",
			CountryID:     countryIDByCode(db, "AE"),
			CityID:        cityIDByName(db, "Dubai"),
			Country:       "UAE", City: "Dubai",
			CompanyAddress:        "Dubai Creek, Bur Dubai, UAE",
			LicenseNumber:         "DED-3001",
			Status:                "active",
			PlanID:                planID(db, "starter"),
			ShortDescription:      strPtr("Traditional dhow dinner cruises"),
			LongDescription:       strPtr("Sunset and dinner dhow cruises along Dubai Creek and Dubai Marina. Family-run since 2019."),
			TrustedTravelersCount: 320,
			CommissionPercent:     400,
			SLAHours:              3,
			ApprovedBy:            uintPtr(1),
			ApprovedAt:            &approved3,
			LastLoginAt:           &lastLogin2,
		},

		// 8. Active, premium — inbound specialist, Turkey
		{
			CompanyName:      "Anatolia Journeys",
			Password:         password,
			ResponsibleName:  "Yusuf Demir",
			HeaderImg:        "https://files.tayara.test/agents/anatolia.png",
			BackgroundImage:  "https://files.tayara.test/agents/anatolia-bg.jpg",
			PhoneCountryCode: "+90", Phone: "05321234567",
			BusinessEmail: "hello@anatoliajourneys.tr",
			CountryID:     countryIDByCode(db, "TR"),
			CityID:        cityIDByName(db, "Istanbul"),
			Country:       "Turkey", City: "Istanbul",
			CompanyAddress:        "Sultanahmet, Istanbul, Turkey",
			LicenseNumber:         "TR-IST-4021",
			Status:                "active",
			PlanID:                planID(db, "premium"),
			ShortDescription:      strPtr("Inbound Turkey specialist"),
			LongDescription:       strPtr("Inbound travel specialist for Gulf travellers visiting Turkey: Istanbul, Cappadocia, Antalya. Arabic-speaking guides."),
			TrustedTravelersCount: 2100,
			CommissionPercent:     1000,
			SLAHours:              2,
			ApprovedBy:            uintPtr(1),
			ApprovedAt:            &approved2,
			LastLoginAt:           &lastLogin,
		},

		// 9. Pending — new applicant, no plan yet
		{
			CompanyName:      "Red Sea Divers",
			Password:         password,
			ResponsibleName:  "Omar Farouk",
			HeaderImg:        "https://files.tayara.test/agents/red-sea.png",
			BackgroundImage:  "https://files.tayara.test/agents/red-sea-bg.jpg",
			PhoneCountryCode: "+20", Phone: "01098765432",
			BusinessEmail: "dive@redseadivers.eg",
			CountryID:     countryIDByCode(db, "EG"),
			CityID:        cityIDByName(db, "Sharm El Sheikh"),
			Country:       "Egypt", City: "Sharm El Sheikh",
			CompanyAddress:    "Naama Bay Promenade, Sharm El Sheikh, Egypt",
			LicenseNumber:     "EG-SSH-8801",
			Status:            "pending",
			PlanID:            nil,
			ShortDescription:  strPtr("PADI dive centre in Sharm"),
			LongDescription:   strPtr("PADI 5-star dive centre offering Red Sea diving packages including Ras Mohammed and the Thistlegorm wreck."),
			CommissionPercent: 0,
			SLAHours:          4,
		},

		// 10. Suspended — payment fraud investigation
		{
			CompanyName:      "FastFares Travel",
			Password:         password,
			ResponsibleName:  "Imran Sheikh",
			HeaderImg:        "https://files.tayara.test/agents/fastfares.png",
			BackgroundImage:  "https://files.tayara.test/agents/fastfares-bg.jpg",
			PhoneCountryCode: "+971", Phone: "0558080808",
			BusinessEmail: "ops@fastfares.ae",
			CountryID:     countryIDByCode(db, "AE"),
			CityID:        cityIDByName(db, "Sharjah"),
			Country:       "UAE", City: "Sharjah",
			CompanyAddress:    "Al Majaz, Sharjah, UAE",
			LicenseNumber:     "DED-1010",
			Status:            "suspended",
			PlanID:            planID(db, "basic"),
			ShortDescription:  strPtr("Discounted flight packages"),
			LongDescription:   strPtr("Discounted multi-city flight packages and visa assistance."),
			CommissionPercent: 500,
			SLAHours:          4,
			ApprovedBy:        uintPtr(1),
			ApprovedAt:        &approved3,
			RejectedReason:    "Suspended pending fraud investigation into chargeback pattern",
		},

		// 11. Active, custom — Saudi Umrah specialist
		{
			CompanyName:      "Haramain Pilgrim Services",
			Password:         password,
			ResponsibleName:  "Abdullah Al Otaibi",
			HeaderImg:        "https://files.tayara.test/agents/haramain.png",
			BackgroundImage:  "https://files.tayara.test/agents/haramain-bg.jpg",
			PhoneCountryCode: "+966", Phone: "0551234567",
			BusinessEmail: "pilgrim@haramain.sa",
			CountryID:     countryIDByCode(db, "SA"),
			CityID:        cityIDByName(db, "Jeddah"),
			Country:       "Saudi Arabia", City: "Jeddah",
			CompanyAddress:        "Al Balad, Jeddah, Saudi Arabia",
			LicenseNumber:         "SA-JED-9920",
			Status:                "active",
			PlanID:                planID(db, "custom"),
			ShortDescription:      strPtr("Licensed Umrah operator"),
			LongDescription:       strPtr("Ministry-licensed Umrah and Hajj operator offering full-package pilgrimage services from Jeddah, Makkah and Madinah."),
			TrustedTravelersCount: 9500,
			CommissionPercent:     800,
			SLAHours:              2,
			ApprovedBy:            uintPtr(1),
			ApprovedAt:            &approved,
			LastLoginAt:           &lastLogin2,
		},

		// 12. Rejected — inactive license
		{
			CompanyName:      "Visa Express",
			Password:         password,
			ResponsibleName:  "Hassan Ali",
			PhoneCountryCode: "+971", Phone: "0559090909",
			BusinessEmail: "apply@visaexpress.ae",
			CountryID:     countryIDByCode(db, "AE"),
			CityID:        cityIDByName(db, "Dubai"),
			Country:       "UAE", City: "Dubai",
			CompanyAddress:   "Business Bay, Dubai, UAE",
			LicenseNumber:    "DED-1011",
			Status:           "rejected",
			ShortDescription: strPtr("Visa application assistance"),
			LongDescription:  strPtr("Visa application assistance for Schengen, UK and US destinations."),
			RejectedReason:   "Trade license expired more than 90 days ago",
		},
	}
	if err := db.Create(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

// ── Documents ────────────────────────────────────────────────────────

func seedAgentDocuments(db *gorm.DB, agents []dxmodel.DxAgentProfile) error {
	// Map agents by business email so documents survive re-ordering.
	byEmail := make(map[string]uint, len(agents))
	for _, a := range agents {
		byEmail[a.BusinessEmail] = a.ID
	}

	docs := []dxmodel.AgentDocument{
		// Desert Safari Dubai — fully approved
		{AgentID: byEmail["hello@desertsafari.ae"], DocType: "trade_license", FileURL: "https://files.tayara.test/desertsafari/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "approved"},
		{AgentID: byEmail["hello@desertsafari.ae"], DocType: "passport", FileURL: "https://files.tayara.test/desertsafari/passport.jpg", OriginalName: "passport.jpg", MIME: "image/jpeg", Status: "approved"},
		{AgentID: byEmail["hello@desertsafari.ae"], DocType: "vat_certificate", FileURL: "https://files.tayara.test/desertsafari/vat.pdf", OriginalName: "vat_certificate.pdf", MIME: "application/pdf", Status: "approved"},

		// Luxury Travel Co — approved
		{AgentID: byEmail["book@luxurytravel.ae"], DocType: "trade_license", FileURL: "https://files.tayara.test/luxurytravel/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "approved"},
		{AgentID: byEmail["book@luxurytravel.ae"], DocType: "passport", FileURL: "https://files.tayara.test/luxurytravel/passport.jpg", OriginalName: "passport.jpg", MIME: "image/jpeg", Status: "approved"},

		// Skyline Tours — pending review
		{AgentID: byEmail["info@skyline.ae"], DocType: "trade_license", FileURL: "https://files.tayara.test/skyline/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "pending"},
		{AgentID: byEmail["info@skyline.ae"], DocType: "passport", FileURL: "https://files.tayara.test/skyline/passport.jpg", OriginalName: "passport.jpg", MIME: "image/jpeg", Status: "pending"},

		// Ocean Voyages — approved historically, one doc rejected
		{AgentID: byEmail["support@ocean.ae"], DocType: "trade_license", FileURL: "https://files.tayara.test/ocean/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "approved"},
		{AgentID: byEmail["support@ocean.ae"], DocType: "vat_certificate", FileURL: "https://files.tayara.test/ocean/vat.pdf", OriginalName: "vat_certificate.pdf", MIME: "application/pdf", Status: "rejected"},

		// Global Getaways — rejected
		{AgentID: byEmail["sales@globalgetaways.ae"], DocType: "trade_license", FileURL: "https://files.tayara.test/globalgetaways/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "rejected"},

		// Emirates Travel Group — fully approved (enterprise)
		{AgentID: byEmail["partners@emiratestravel.ae"], DocType: "trade_license", FileURL: "https://files.tayara.test/emiratestravel/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "approved"},
		{AgentID: byEmail["partners@emiratestravel.ae"], DocType: "passport", FileURL: "https://files.tayara.test/emiratestravel/passport.jpg", OriginalName: "passport.jpg", MIME: "image/jpeg", Status: "approved"},
		{AgentID: byEmail["partners@emiratestravel.ae"], DocType: "vat_certificate", FileURL: "https://files.tayara.test/emiratestravel/vat.pdf", OriginalName: "vat_certificate.pdf", MIME: "application/pdf", Status: "approved"},
		{AgentID: byEmail["partners@emiratestravel.ae"], DocType: "insurance", FileURL: "https://files.tayara.test/emiratestravel/insurance.pdf", OriginalName: "liability_insurance.pdf", MIME: "application/pdf", Status: "approved"},

		// Sunset Dhow Cruises — approved
		{AgentID: byEmail["hello@sunsetdhow.ae"], DocType: "trade_license", FileURL: "https://files.tayara.test/sunsetdhow/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "approved"},

		// Anatolia Journeys — approved
		{AgentID: byEmail["hello@anatoliajourneys.tr"], DocType: "trade_license", FileURL: "https://files.tayara.test/anatolia/license.pdf", OriginalName: "ticaret_sicil.pdf", MIME: "application/pdf", Status: "approved"},
		{AgentID: byEmail["hello@anatoliajourneys.tr"], DocType: "passport", FileURL: "https://files.tayara.test/anatolia/passport.jpg", OriginalName: "passport.jpg", MIME: "image/jpeg", Status: "approved"},

		// Red Sea Divers — pending
		{AgentID: byEmail["dive@redseadivers.eg"], DocType: "trade_license", FileURL: "https://files.tayara.test/redsea/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "pending"},

		// FastFares Travel — approved but suspended
		{AgentID: byEmail["ops@fastfares.ae"], DocType: "trade_license", FileURL: "https://files.tayara.test/fastfares/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "approved"},

		// Haramain Pilgrim Services — approved
		{AgentID: byEmail["pilgrim@haramain.sa"], DocType: "trade_license", FileURL: "https://files.tayara.test/haramain/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "approved"},
		{AgentID: byEmail["pilgrim@haramain.sa"], DocType: "ministry_permit", FileURL: "https://files.tayara.test/haramain/permit.pdf", OriginalName: "ministry_umrah_permit.pdf", MIME: "application/pdf", Status: "approved"},

		// Visa Express — rejected
		{AgentID: byEmail["apply@visaexpress.ae"], DocType: "trade_license", FileURL: "https://files.tayara.test/visaexpress/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "rejected"},
	}
	return db.Create(&docs).Error
}

// ── Notes ────────────────────────────────────────────────────────────

func seedAgentNotes(db *gorm.DB, agents []dxmodel.DxAgentProfile) error {
	byEmail := make(map[string]uint, len(agents))
	for _, a := range agents {
		byEmail[a.BusinessEmail] = a.ID
	}

	notes := []dxmodel.AgentNote{
		{AgentID: byEmail["hello@desertsafari.ae"], AdminID: 1, NoteText: "Verified trade license. Top performer for outbound Turkey."},
		{AgentID: byEmail["hello@desertsafari.ae"], AdminID: 1, NoteText: "Excellent SLA adherence — 1.8h avg response last 30 days."},
		{AgentID: byEmail["hello@desertsafari.ae"], AdminID: 2, NoteText: "Requested premium plan upgrade — approved 2024-08."},

		{AgentID: byEmail["book@luxurytravel.ae"], AdminID: 1, NoteText: "Premium listings maintained; strong customer ratings."},
		{AgentID: byEmail["book@luxurytravel.ae"], AdminID: 2, NoteText: "Minor SLA warnings twice in Q3 — coaching note sent."},

		{AgentID: byEmail["info@skyline.ae"], AdminID: 1, NoteText: "Awaiting trade license verification from DED registry."},

		{AgentID: byEmail["support@ocean.ae"], AdminID: 1, NoteText: "Suspended 2024-09 due to SLA breaches. Review scheduled 2024-12."},
		{AgentID: byEmail["support@ocean.ae"], AdminID: 2, NoteText: "VAT certificate mismatch — asked for corrected copy."},

		{AgentID: byEmail["sales@globalgetaways.ae"], AdminID: 1, NoteText: "Rejected — license number does not match DED public registry."},
		{AgentID: byEmail["sales@globalgetaways.ae"], AdminID: 2, NoteText: "Applicant may re-apply with updated documents."},

		{AgentID: byEmail["partners@emiratestravel.ae"], AdminID: 1, NoteText: "Enterprise onboarding completed. Dedicated account manager assigned."},
		{AgentID: byEmail["partners@emiratestravel.ae"], AdminID: 1, NoteText: "Quarterly review passed — all SLAs green."},

		{AgentID: byEmail["hello@sunsetdhow.ae"], AdminID: 2, NoteText: "New agent — monitor first 90 days for SLA compliance."},

		{AgentID: byEmail["hello@anatoliajourneys.tr"], AdminID: 1, NoteText: "Arabic-speaking guides confirmed. Good inbound Turkey coverage."},

		{AgentID: byEmail["dive@redseadivers.eg"], AdminID: 1, NoteText: "PADI certification needs to be uploaded before approval."},

		{AgentID: byEmail["ops@fastfares.ae"], AdminID: 1, NoteText: "Suspended pending fraud investigation — 4 chargebacks in 30 days."},
		{AgentID: byEmail["ops@fastfares.ae"], AdminID: 2, NoteText: "Escalated to risk team. Do not re-activate without sign-off."},

		{AgentID: byEmail["pilgrim@haramain.sa"], AdminID: 1, NoteText: "Ministry-issued Umrah permit on file. High-value partner."},

		{AgentID: byEmail["apply@visaexpress.ae"], AdminID: 2, NoteText: "Trade license expired 2024-06 — rejection sent."},
	}
	return db.Create(&notes).Error
}

// ── Helpers ──────────────────────────────────────────────────────────

func uintPtr(v uint) *uint    { return &v }
func strPtr(s string) *string { return &s }
