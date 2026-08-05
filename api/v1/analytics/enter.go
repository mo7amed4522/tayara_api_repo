package analytics

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the analytics api handlers; embedded into api.ApiGroupApp.AnalyticsApiGroup.
type ApiGroup struct {
	AnalyticsApi
}

// analyticsService reaches the analytics services via the group var.
var analyticsService = service.ServiceGroupApp.AnalyticsServiceGroup
