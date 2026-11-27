package recommendation

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the recommendation handlers; embedded into api.ApiGroupApp.RecommendationApiGroup.
type ApiGroup struct {
	RecommendationApi
	RecommendationPublicApi
}

// recommendationSvc reaches the recommendation services via the group var.
var recommendationSvc = &service.ServiceGroupApp.RecommendationServiceGroup.RecommendationServices
