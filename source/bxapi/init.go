// Package bxapi seeds end-user profiles with varied KYC/status states (approved, pending,
// rejected, blocked) and realistic auth providers so the user app and admin dashboards have data.
package bxapi

import (
	"context"
	"time"

	bxmodel "github.com/mo7amed4522/tayara_api/model/bxapi"
	"github.com/mo7amed4522/tayara_api/service/system"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

const initOrderUsers = system.InitOrderExternal + 4

type userInit struct{}

func init() {
	system.RegisterInit(initOrderUsers, &userInit{})
}

func (u *userInit) InitializerName() string { return "demo_users" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (u *userInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (u *userInit) TableCreated(ctx context.Context) bool                     { return true }

func (u *userInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var c int64
	return db.Model(&bxmodel.BxsUserProfile{}).Count(&c).Error == nil && c > 0
}

func (u *userInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	password := utils.BcryptHash("123456")
	now := time.Now()
	lastLogin := now.Add(-6 * time.Hour)

	users := []bxmodel.BxsUserProfile{
		{
			FullName: "Ahmed Ali", Username: "ahmed.ali", Password: password, NickName: "Ahmed",
			PhoneCountryCode: "+971", Phone: "0561000001", Email: "ahmed.ali@mail.test",
			Nationality: "Emirati", KYCStatus: "approved", TotalSpend: 1500000, Status: "active",
			LastLoginAt: &lastLogin, FirebaseUID: "firebase-uid-001", AuthProvider: "email",
		},
		{
			FullName: "Fatima Zahra", Username: "fatima.z", Password: password, NickName: "Fatima",
			PhoneCountryCode: "+971", Phone: "0561000002", Email: "fatima.z@mail.test",
			Nationality: "Indian", KYCStatus: "pending", TotalSpend: 0, Status: "active",
			AuthProvider: "google",
		},
		{
			FullName: "Omar Hassan", Username: "omar.h", Password: password, NickName: "Omar",
			PhoneCountryCode: "+971", Phone: "0561000003", Email: "omar.h@mail.test",
			Nationality: "Egyptian", KYCStatus: "not_submitted", TotalSpend: 0, Status: "active",
			AuthProvider: "phone",
		},
		{
			FullName: "Layla Mohamed", Username: "layla.m", Password: password, NickName: "Layla",
			PhoneCountryCode: "+971", Phone: "0561000004", Email: "layla.m@mail.test",
			Nationality: "Pakistani", KYCStatus: "rejected", TotalSpend: 0, Status: "blocked",
			BlockReason: "Suspected chargeback abuse", AuthProvider: "apple",
		},
		{
			FullName: "Youssef Karim", Username: "youssef.k", Password: password, NickName: "Youssef",
			PhoneCountryCode: "+971", Phone: "0561000005", Email: "youssef.k@mail.test",
			Nationality: "Filipino", KYCStatus: "approved", TotalSpend: 3200000, Status: "active",
			LastLoginAt: &now, AuthProvider: "email",
		},
	}
	if err := db.Create(&users).Error; err != nil {
		return ctx, err
	}
	// Dictionary entries for mobile app pickers (bx_dict).
	if err := db.Create(&[]bxmodel.BxDict{
		{Value: "honeymoon", Label: "Honeymoon", Type: "trip_category", Description: "Trip category", Sort: 1, Remarks: "category"},
		{Value: "adventure", Label: "Adventure", Type: "trip_category", Description: "Trip category", Sort: 2, Remarks: "category"},
		{Value: "family", Label: "Family", Type: "trip_category", Description: "Trip category", Sort: 3, Remarks: "category"},
		{Value: "AED", Label: "AED", Type: "currency", Description: "Currency", Sort: 1, Remarks: "currency"},
		{Value: "USD", Label: "USD", Type: "currency", Description: "Currency", Sort: 2, Remarks: "currency"},
	}).Error; err != nil {
		return ctx, err
	}
	return ctx, nil
}
