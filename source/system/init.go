// Package system seeds platform admin accounts, password-reset tokens, and default
// Casbin RBAC policies so a test app has admins to log into the System dashboard
// (SystemJWTAuth) and an authority matrix ready for route-guard enforcement.
package system

import (
	"context"
	"strconv"
	"time"

	adapter "github.com/casbin/gorm-adapter/v3"
	basemodel "github.com/mo7amed4522/tayara_api/model/base"
	sysmodel "github.com/mo7amed4522/tayara_api/model/system"
	"github.com/mo7amed4522/tayara_api/service/system"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

// initOrderAdmin runs after geography seed (initOrderGeography+1).
const initOrderAdmin = system.InitOrderExternal + 2

// initOrderCasbin runs after admin seed so the authority-id→route matrix is
// populated once the platform is bootstrapped.
const initOrderCasbin = system.InitOrderExternal + 3

type adminInit struct{}

func init() {
	system.RegisterInit(initOrderAdmin, &adminInit{})
	system.RegisterInit(initOrderCasbin, &casbinInit{})
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

// ─── Casbin RBAC seed ────────────────────────────────────────────────────────

type casbinInit struct{}

func (c *casbinInit) InitializerName() string { return "demo_casbin_default_policies" }

func (c *casbinInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (c *casbinInit) TableCreated(ctx context.Context) bool                     { return true }

func (c *casbinInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var count int64
	return db.Model(&adapter.CasbinRule{}).Where("ptype = ?", "p").Count(&count).Error == nil && count > 0
}

// InitializeData seeds the default authority→route matrix for the three built-in
// roles (1=admin, 2=agent, 3=user). Each rule is a p-type policy:
//
//
//	p <authority_id> <path> <method>
//
// The paths mirror the registered routers in router/* so the Casbin enforcer
// can validate them once CasbinHandler is wired to the relevant groups.
func (c *casbinInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}

	// authorityID → list of (path, method) pairs.
	policies := map[uint][][2]string{
		// Authority 1: platform admin — full access.
		1: {
			// System admin auth & profile.
			{"/api/v1/system/register", "POST"}, {"/api/v1/system/register", "GET"},
			{"/api/v1/system/login", "POST"},
			{"/api/v1/system/forgot-password", "POST"},
			{"/api/v1/system/reset-password", "POST"},
			{"/api/v1/system/changePassword", "POST"},
			{"/api/v1/system/jobs", "GET"},
			// Casbin policy management.
			{"/api/v1/system/casbin/policies", "GET"},
			{"/api/v1/system/casbin/policies", "PUT"},
			{"/api/v1/system/casbin/policy", "PUT"},
			// Agent lifecycle (admin side).
			{"/api/v1/agent/admin", "GET"},
			{"/api/v1/agent/admin/detail", "GET"},
			{"/api/v1/agent/admin/approve", "POST"},
			{"/api/v1/agent/admin/reject", "POST"},
			{"/api/v1/agent/admin/suspend", "POST"},
			{"/api/v1/agent/admin/note", "POST"},
			// User lifecycle (admin side).
			{"/api/v1/user/admin", "GET"},
			{"/api/v1/user/admin/block", "POST"},
			// Catalog moderation.
			{"/api/v1/catalog/admin/packages", "GET"},
			{"/api/v1/catalog/admin/packages/detail", "GET"},
			{"/api/v1/catalog/admin/packages/status", "PUT"},
			// Booking admin.
			{"/api/v1/booking/admin", "GET"},
			{"/api/v1/booking/admin/force-cancel", "POST"},
			{"/api/v1/booking/admin/kyc", "POST"},
			{"/api/v1/booking/admin/disputes", "GET"},
			{"/api/v1/booking/admin/disputes/resolve", "POST"},
			{"/api/v1/booking/admin/documents/audit", "GET"},
			// Payment admin.
			{"/api/v1/payment/capture", "POST"},
			{"/api/v1/payment/refund", "POST"},
			{"/api/v1/payment/payouts/prepare", "POST"},
			{"/api/v1/payment/payouts/pay", "POST"},
			{"/api/v1/payment/payouts", "GET"},
			{"/api/v1/payment/commission/override", "POST"},
			{"/api/v1/payment/commission", "GET"},
			// Engage admin.
			{"/api/v1/engage/admin/reviews", "GET"},
			{"/api/v1/engage/admin/reviews/status", "PUT"},
			{"/api/v1/engage/ads", "POST"},
			{"/api/v1/engage/ads", "PUT"},
			{"/api/v1/engage/ads/status", "PUT"},
			{"/api/v1/engage/ads", "DELETE"},
			{"/api/v1/engage/admin/ads", "GET"},
			{"/api/v1/engage/admin/ads/analytics", "GET"},
			{"/api/v1/engage/admin/notifications", "GET"},
			{"/api/v1/engage/admin/notifications/resend", "POST"},
			{"/api/v1/engage/admin/notification-templates", "POST"},
			{"/api/v1/engage/admin/notification-templates", "GET"},
			{"/api/v1/engage/admin/notification-templates", "DELETE"},
			{"/api/v1/engage/admin/coupons", "POST"},
			{"/api/v1/engage/admin/coupons", "GET"},
			// Analytics admin.
			{"/api/v1/analytics/admin/users", "GET"},
			{"/api/v1/analytics/admin/agents", "GET"},
			{"/api/v1/analytics/admin/business", "GET"},
			// Export / search / tickets.
			{"/api/v1/admin/export/users", "GET"},
			{"/api/v1/admin/export/agents", "GET"},
			{"/api/v1/admin/export/bookings", "GET"},
			{"/api/v1/admin/export/payments", "GET"},
			{"/api/v1/admin/search", "GET"},
			{"/api/v1/admin/tickets/:id/pdf", "GET"},
			{"/api/v1/admin/tickets/:id/send", "POST"},
			// Base lookups.
			{"/api/v1/base/categories", "GET"},
		},
		// Authority 2: travel-agent — own dashboard + booking/catalog/payment management.
		2: {
			// Agent auth & profile.
			{"/api/v1/agent/register", "POST"},
			{"/api/v1/agent/login", "POST"},
			{"/api/v1/agent/profile", "GET"},
			{"/api/v1/agent/change-password", "POST"},
			// Catalog (own packages & templates).
			{"/api/v1/catalog/packages", "POST"},
			{"/api/v1/catalog/packages", "PUT"},
			{"/api/v1/catalog/packages", "DELETE"},
			{"/api/v1/catalog/packages/submit", "POST"},
			{"/api/v1/catalog/packages", "GET"},
			{"/api/v1/catalog/packages/detail", "GET"},
			{"/api/v1/catalog/templates", "POST"},
			{"/api/v1/catalog/templates", "PUT"},
			{"/api/v1/catalog/templates", "DELETE"},
			{"/api/v1/catalog/templates", "GET"},
			{"/api/v1/catalog/templates/detail", "GET"},
			{"/api/v1/catalog/templates/requirements", "POST"},
			{"/api/v1/catalog/templates/requirements", "GET"},
			// Booking (agent side).
			{"/api/v1/booking/manage", "GET"},
			{"/api/v1/booking/manage/detail", "GET"},
			{"/api/v1/booking/manage/review", "POST"},
			{"/api/v1/booking/manage/offer", "POST"},
			{"/api/v1/booking/manage/reject", "POST"},
			{"/api/v1/booking/manage/documents", "GET"},
			{"/api/v1/booking/manage/documents/upload", "POST"},
			{"/api/v1/booking/manage/documents/review", "POST"},
			{"/api/v1/booking/manage/process", "POST"},
			{"/api/v1/booking/manage/issue", "POST"},
			{"/api/v1/booking/manage/complete", "POST"},
			{"/api/v1/booking/manage/hold", "POST"},
			{"/api/v1/booking/manage/hold/status", "GET"},
			// Payment (agent payouts & earnings).
			{"/api/v1/payment/payouts/manage", "GET"},
			{"/api/v1/payment/payouts/earnings", "GET"},
			{"/api/v1/payment/commission/manage", "GET"},
			// Analytics (agent dashboard).
			{"/api/v1/analytics/agent", "GET"},
			// Engage (agent-side reviews — if any in future).
		},
		// Authority 3: end-user — browse, book, pay, engage.
		3: {
			// User auth & profile.
			{"/api/v1/user/register", "POST"},
			{"/api/v1/user/login", "POST"},
			{"/api/v1/user/firebase-login", "POST"},
			{"/api/v1/user/profile", "GET"},
			{"/api/v1/user/change-password", "POST"},
			// Customer-facing catalog browsing.
			{"/api/v1/catalog/public/packages", "GET"},
			{"/api/v1/catalog/public/packages/detail", "GET"},
			{"/api/v1/catalog/public/packages/detail/reviews", "GET"},
			{"/api/v1/customer/packages/detail", "GET"},
			{"/api/v1/customer/packages/detail/reviews", "GET"},
			{"/api/v1/customer/agent/detail", "GET"},
			{"/api/v1/customer/search", "GET"},
			{"/api/v1/customer/explore/trips", "GET"},
			{"/api/v1/customer/home/destinations", "GET"},
			{"/api/v1/customer/home/deals", "GET"},
			{"/api/v1/customer/home/agents", "GET"},
			{"/api/v1/customer/home/ads", "GET"},
			{"/api/v1/customer/home/recommended", "GET"},
			// Booking (user side).
			{"/api/v1/booking", "POST"},
			{"/api/v1/booking", "GET"},
			{"/api/v1/booking/detail", "GET"},
			{"/api/v1/booking/cancel", "POST"},
			{"/api/v1/booking/documents", "POST"},
			{"/api/v1/booking/documents", "GET"},
			{"/api/v1/booking/disputes", "POST"},
			{"/api/v1/booking/disputes", "GET"},
			// Payment (user side).
			{"/api/v1/payment", "POST"},
			// Engage (user side).
			{"/api/v1/engage/reviews", "GET"},
			{"/api/v1/engage/ads", "GET"},
			{"/api/v1/engage/ads/impression", "POST"},
			{"/api/v1/engage/ads/click", "POST"},
			{"/api/v1/engage/reviews", "POST"},
			{"/api/v1/engage/wishlist", "POST"},
			{"/api/v1/engage/wishlist", "GET"},
			{"/api/v1/engage/share", "POST"},
			{"/api/v1/engage/coupons", "GET"},
			{"/api/v1/engage/coupons/apply", "POST"},
			{"/api/v1/engage/notifications", "GET"},
			{"/api/v1/engage/devices", "POST"},
			{"/api/v1/engage/devices", "GET"},
			{"/api/v1/engage/devices/remove", "POST"},
		},
	}

	var rules []adapter.CasbinRule
	for authorityID, pairs := range policies {
		for _, p := range pairs {
			rules = append(rules, adapter.CasbinRule{
				Ptype: "p",
				V0:    strconv.FormatUint(uint64(authorityID), 10),
				V1:    p[0],
				V2:    p[1],
			})
		}
	}
	if len(rules) > 0 {
		if err := db.Create(&rules).Error; err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
