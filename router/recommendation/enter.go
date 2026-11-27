package recommendation

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the recommendation routers; embedded into router.RouterGroupApp.Recommendation.
type RouterGroup struct {
	RecommendationRouter
}

var (
	recommendationApi       = api.ApiGroupApp.RecommendationApiGroup.RecommendationApi
	recommendationPublicApi = api.ApiGroupApp.RecommendationApiGroup.RecommendationPublicApi
)
