// Package dxapi seeds agent plans, agent profiles (varied approval states), verification
// documents and internal admin notes so the Manage (agent) dashboard has realistic data.
package dxapi

import (
	"context"
	"time"

	dxmodel "github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/service/system"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
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

func seedPlans(db *gorm.DB) error {
	plans := []dxmodel.AgentPlan{
		{Name: "basic", DefaultCommissionPercent: 500, Description: "Entry plan, 5% commission", IsActive: true},
		{Name: "premium", DefaultCommissionPercent: 1000, Description: "Preferred plan, 10% commission", IsActive: true},
		{Name: "custom", DefaultCommissionPercent: 800, Description: "Negotiated plan, 8% commission", IsActive: true},
	}
	return db.Create(&plans).Error
}

func planID(db *gorm.DB, name string) *uint {
	var p dxmodel.AgentPlan
	if err := db.Where("name = ?", name).First(&p).Error; err != nil {
		return nil
	}
	return &p.ID
}

func seedAgents(db *gorm.DB) ([]dxmodel.DxAgentProfile, error) {
	password := utils.BcryptHash("123456")
	approved := time.Now().Add(-720 * time.Hour)

	agents := []dxmodel.DxAgentProfile{
		{
			CompanyName: "Desert Safari Dubai", Password: password, ResponsibleName: "Khalid Al Mansouri",
			PhoneCountryCode: "+971", Phone: "0551010101", BusinessEmail: "hello@desertsafari.ae",
			Country: "UAE", CompanyAddress: "Al Quoz, Dubai", LicenseNumber: "DED-1001",
			Status: "active", PlanID: planID(db, "premium"), CommissionPercent: 1000, SLAHours: 2,
			ApprovedBy: uintPtr(1), ApprovedAt: &approved,
			ShortDescription: strPtr("Award-winning desert experiences"), LongDescription: strPtr("Full-day desert safaris, dune bashing and traditional camps."),
		},
		{
			CompanyName: "Luxury Travel Co", Password: password, ResponsibleName: "Sara Al Noori",
			PhoneCountryCode: "+971", Phone: "0552020202", BusinessEmail: "book@luxurytravel.ae",
			Country: "UAE", CompanyAddress: "Downtown, Dubai", LicenseNumber: "DED-1002",
			Status: "active", PlanID: planID(db, "basic"), CommissionPercent: 500, SLAHours: 4,
			ApprovedBy: uintPtr(1), ApprovedAt: &approved,
		},
		{
			CompanyName: "Skyline Tours", Password: password, ResponsibleName: "Mohammed Eid",
			PhoneCountryCode: "+971", Phone: "0553030303", BusinessEmail: "info@skyline.ae",
			Country: "UAE", CompanyAddress: "Deira, Dubai", LicenseNumber: "DED-1003",
			Status: "pending", PlanID: planID(db, "basic"), CommissionPercent: 0, SLAHours: 2,
		},
		{
			CompanyName: "Ocean Voyages", Password: password, ResponsibleName: "Nadia Haddad",
			PhoneCountryCode: "+971", Phone: "0554040404", BusinessEmail: "support@ocean.ae",
			Country: "UAE", CompanyAddress: "Marina, Dubai", LicenseNumber: "DED-1004",
			Status: "suspended", PlanID: planID(db, "custom"), CommissionPercent: 800, SLAHours: 6,
			ApprovedBy: uintPtr(1), ApprovedAt: &approved, RejectedReason: "Suspended for SLA breaches",
		},
		{
			CompanyName: "Global Getaways", Password: password, ResponsibleName: "Tariq Bashir",
			PhoneCountryCode: "+971", Phone: "0555050505", BusinessEmail: "sales@globalgetaways.ae",
			Country: "UAE", CompanyAddress: "JLT, Dubai", LicenseNumber: "DED-1005",
			Status: "rejected", RejectedReason: "License documents did not verify",
		},
	}
	if err := db.Create(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

func seedAgentDocuments(db *gorm.DB, agents []dxmodel.DxAgentProfile) error {
	// Reference the first seeded agent (Desert Safari Dubai).
	var agentID uint
	if len(agents) > 0 {
		agentID = agents[0].ID
	} else {
		return nil
	}
	docs := []dxmodel.AgentDocument{
		{AgentID: agentID, DocType: "trade_license", FileURL: "https://files.tayara.test/desertsafari/license.pdf", OriginalName: "trade_license.pdf", MIME: "application/pdf", Status: "approved"},
		{AgentID: agentID, DocType: "passport", FileURL: "https://files.tayara.test/desertsafari/passport.jpg", OriginalName: "passport.jpg", MIME: "image/jpeg", Status: "approved"},
	}
	return db.Create(&docs).Error
}

func seedAgentNotes(db *gorm.DB, agents []dxmodel.DxAgentProfile) error {
	var agentID uint
	if len(agents) > 0 {
		agentID = agents[0].ID
	} else {
		return nil
	}
	notes := []dxmodel.AgentNote{
		{AgentID: agentID, AdminID: 1, NoteText: "Verified trade license. Top performer for outbound Turkey."},
	}
	return db.Create(&notes).Error
}

func uintPtr(v uint) *uint          { return &v }
func strPtr(s string) *string       { return &s }
