package recommendation

import (
	"context"
	"strconv"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/mo7amed4522/tayara_api/service/system"
	"gorm.io/gorm"
)

// initOrderRecommendationAPI runs after the setting seed so the Casbin authority→route matrix gets
// the recommendation endpoints once the module's own tables exist.
const initOrderRecommendationAPI = initOrderRecommendationSetting + 1

type apiInit struct{}

func init() {
	system.RegisterInit(initOrderRecommendationAPI, &apiInit{})
}

func (a *apiInit) InitializerName() string { return "demo_recommendation_permissions" }

// MigrateTable/TableCreated are no-ops: Casbin rules live in the existing casbin_rule table.
func (a *apiInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (a *apiInit) TableCreated(ctx context.Context) bool                     { return true }

func (a *apiInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var n int64
	return db.Model(&adapter.CasbinRule{}).
		Where("ptype = ? AND v1 = ?", "p", "/api/v1/recommendation/generate").
		Count(&n).Error == nil && n > 0
}

// InitializeData appends the admin (authority 1) and customer (authority 3) policies. The system
// seed only writes policies when the table is empty, so appending here keeps existing installs working.
func (a *apiInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}

	policies := map[uint][][2]string{
		// Platform admin: full recommendation management.
		1: {
			{"/api/v1/recommendation/createRecommendation", "POST"},
			{"/api/v1/recommendation/updateRecommendation", "PUT"},
			{"/api/v1/recommendation/deleteRecommendation", "DELETE"},
			{"/api/v1/recommendation/deleteRecommendationByIds", "DELETE"},
			{"/api/v1/recommendation/findRecommendation", "GET"},
			{"/api/v1/recommendation/getRecommendationList", "GET"},
			{"/api/v1/recommendation/review", "POST"},
			{"/api/v1/recommendation/batchReview", "POST"},
			{"/api/v1/recommendation/generate", "POST"},
			{"/api/v1/recommendation/getRunList", "GET"},
			{"/api/v1/recommendation/findRun", "GET"},
			{"/api/v1/recommendation/getStats", "GET"},
			{"/api/v1/recommendation/getSetting", "GET"},
			{"/api/v1/recommendation/updateSetting", "PUT"},
		},
		// End-user: read approved picks + record engagement.
		3: {
			{"/api/v1/recommendation/public/list", "GET"},
			{"/api/v1/recommendation/public/event", "POST"},
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
	if len(rules) == 0 {
		return ctx, nil
	}
	if err := db.Create(&rules).Error; err != nil {
		return ctx, err
	}
	return ctx, nil
}
