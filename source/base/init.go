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
	return db.Model(&basemodel.Nationality{}).Count(&n).Error == nil && n > 0
}

func (g *geographyInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	// Countries referenced below must exist (seeded by service/base). If they are missing we
	// still seed cities/nationalities and let GORM map countryId to whatever is present.
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

func seedCities(db *gorm.DB) error {
	// City rows reference countries by their numeric ID (resolved from code).
	cities := []basemodel.City{
		{CountryID: countryByCode(db, "AE"), Name: "Dubai", IsActive: true},
		{CountryID: countryByCode(db, "AE"), Name: "Abu Dhabi", IsActive: true},
		{CountryID: countryByCode(db, "AE"), Name: "Sharjah", IsActive: true},
		{CountryID: countryByCode(db, "TR"), Name: "Istanbul", IsActive: true},
		{CountryID: countryByCode(db, "TR"), Name: "Antalya", IsActive: true},
		{CountryID: countryByCode(db, "GB"), Name: "London", IsActive: true},
		{CountryID: countryByCode(db, "US"), Name: "New York", IsActive: true},
		{CountryID: countryByCode(db, "TH"), Name: "Bangkok", IsActive: true},
		{CountryID: countryByCode(db, "TH"), Name: "Phuket", IsActive: true},
		{CountryID: countryByCode(db, "EG"), Name: "Cairo", IsActive: true},
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
		Create(&cities).Error
}

func seedNationalities(db *gorm.DB) error {
	nationalities := []basemodel.Nationality{
		{Code: "AE", Name: "Emirati", IsActive: true},
		{Code: "IN", Name: "Indian", IsActive: true},
		{Code: "EG", Name: "Egyptian", IsActive: true},
		{Code: "PK", Name: "Pakistani", IsActive: true},
		{Code: "GB", Name: "British", IsActive: true},
		{Code: "US", Name: "American", IsActive: true},
		{Code: "PH", Name: "Filipino", IsActive: true},
		{Code: "NP", Name: "Nepali", IsActive: true},
		{Code: "TR", Name: "Turkish", IsActive: true},
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
	}
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).
		Create(&categories).Error
}
