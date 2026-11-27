// Package recommendation seeds the per-placement recommendation settings (with the versioned
// prompt template) and the Casbin permissions for the admin recommendation endpoints.
package recommendation

import (
	"context"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/model/common"
	recmodel "github.com/mo7amed4522/tayara_api/model/recommendation"
	recommendationSvc "github.com/mo7amed4522/tayara_api/service/recommendation"
	"github.com/mo7amed4522/tayara_api/service/system"
	"gorm.io/gorm"
)

// initOrderRecommendationSetting runs after catalog, booking and engage so the placement keys
// (home_top, search, package_detail, agency_page) resolve against the seeded marketplace data.
const initOrderRecommendationSetting = system.InitOrderExternal + 20

type settingInit struct{}

func init() {
	system.RegisterInit(initOrderRecommendationSetting, &settingInit{})
}

func (s *settingInit) InitializerName() string { return "demo_recommendation_settings" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (s *settingInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (s *settingInit) TableCreated(ctx context.Context) bool                     { return true }

func (s *settingInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var n int64
	return db.Model(&recmodel.RecommendationSetting{}).Count(&n).Error == nil && n > 0
}

func (s *settingInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	if err := db.Create(settings()).Error; err != nil {
		return ctx, err
	}
	return ctx, nil
}

// settings the four default placements. Only home_top starts enabled so the first cron pass
// exercises the engine; admins turn the rest on from the recommendation settings screen.
func settings() []recmodel.RecommendationSetting {
	allTypes := func() common.JSONMap {
		return common.JSONMap{"targetTypes": []string{
			constants.RecommendationTargetTypePackage,
			constants.RecommendationTargetTypeTrip,
			constants.RecommendationTargetTypeAgency,
		}}
	}
	return []recmodel.RecommendationSetting{
		{
			Placement: "home_top", Enabled: true, MaxItems: 8, AutoApprove: false,
			AutoApproveThreshold: 0.85, CandidateLimit: 60, TargetTypes: allTypes(),
			PromptTemplate: recommendationSvc.DefaultPromptTemplate, PromptVersion: "v1",
			RefreshCron: "6h", ExpireAfterHours: 24,
		},
		{
			Placement: "search", Enabled: false, MaxItems: 10, AutoApprove: false,
			AutoApproveThreshold: 0.85, CandidateLimit: 60,
			TargetTypes:    common.JSONMap{"targetTypes": []string{constants.RecommendationTargetTypePackage, constants.RecommendationTargetTypeTrip}},
			PromptTemplate: recommendationSvc.DefaultPromptTemplate, PromptVersion: "v1",
			RefreshCron: "6h", ExpireAfterHours: 24,
		},
		{
			Placement: "package_detail", Enabled: false, MaxItems: 6, AutoApprove: false,
			AutoApproveThreshold: 0.85, CandidateLimit: 40,
			TargetTypes:    common.JSONMap{"targetTypes": []string{constants.RecommendationTargetTypePackage}},
			PromptTemplate: recommendationSvc.DefaultPromptTemplate, PromptVersion: "v1",
			RefreshCron: "12h", ExpireAfterHours: 48,
		},
		{
			Placement: "agency_page", Enabled: false, MaxItems: 6, AutoApprove: false,
			AutoApproveThreshold: 0.85, CandidateLimit: 40,
			TargetTypes:    common.JSONMap{"targetTypes": []string{constants.RecommendationTargetTypeAgency, constants.RecommendationTargetTypePackage}},
			PromptTemplate: recommendationSvc.DefaultPromptTemplate, PromptVersion: "v1",
			RefreshCron: "12h", ExpireAfterHours: 48,
		},
	}
}
