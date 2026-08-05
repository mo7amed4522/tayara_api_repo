package analytics

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the analytics routers; embedded into router.RouterGroupApp.Analytics.
type RouterGroup struct {
	AnalyticsRouter
}

var (
	analyticsApi = api.ApiGroupApp.AnalyticsApiGroup.AnalyticsApi
)
