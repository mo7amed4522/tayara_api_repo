// Package system seeds platform admin accounts and password-reset tokens so a test app has
// admins to log into the System dashboard (SystemJWTAuth).
package system

import (
	"context"
	"time"

	basemodel "github.com/mo7amed4522/tayara_api/model/base"
	sysmodel "github.com/mo7amed4522/tayara_api/model/system"
	"github.com/mo7amed4522/tayara_api/service/system"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

// initOrderAdmin runs after geography seed (initOrderGeography+1).
const initOrderAdmin = system.InitOrderExternal + 2

type adminInit struct{}

func init() {
	system.RegisterInit(initOrderAdmin, &adminInit{})
}

func (a *adminInit) InitializerName() string { return "demo_system_admins" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (a *adminInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (a *adminInit) TableCreated(ctx context.Context) bool                     { return true }

func (a *adminInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var c int64
	return db.Model(&sysmodel.SysAdminProfile{}).Count(&c).Error == nil && c > 0
}

func (a *adminInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	password := utils.BcryptHash("123456") // demo password for every seeded admin

	admins := []sysmodel.SysAdminProfile{
		{
			Username: "admin", NickName: "Platform Admin", Password: password,
			PhoneCountryCode: "+971", Phone: "0501111111", Email: "admin@tayara.test",
			Department: "Operations", JobTitle: "Super Admin",
		},
		{
			Username: "moderator", NickName: "Content Moderator", Password: password,
			PhoneCountryCode: "+971", Phone: "0502222222", Email: "moderator@tayara.test",
			Department: "Content", JobTitle: "Package Moderator",
		},
		{
			Username: "finance", NickName: "Finance Officer", Password: password,
			PhoneCountryCode: "+971", Phone: "0503333333", Email: "finance@tayara.test",
			Department: "Finance", JobTitle: "Refunds & Payouts",
		},
	}
	if err := db.Create(&admins).Error; err != nil {
		return ctx, err
	}

	// A single-use password-reset token for the admin (email-recovery flow demo).
	now := time.Now().Add(48 * time.Hour)
	reset := []sysmodel.PasswordReset{
		{
			Email: "admin@tayara.test", Role: "admin",
			TokenHash: utils.BcryptHash("reset-token-demo-0001"), Used: false, ExpiresAt: &now,
		},
	}
	if err := db.Create(&reset).Error; err != nil {
		return ctx, err
	}

	// One audited settings change (base_setting_change_history) for the admin who ran it.
	adminID := uint(1)
	if err := db.Create(&[]basemodel.SettingChangeHistory{
		{Section: "commission", Key: "basic_plan_percent", OldValue: "500", NewValue: "550", ChangedBy: &adminID},
	}).Error; err != nil {
		return ctx, err
	}
	return ctx, nil
}
