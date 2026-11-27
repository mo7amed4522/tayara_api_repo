package recommendation

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/middleware"
)

type RecommendationRouter struct{}

// InitRecommendationRouter registers the admin endpoints on the system (admin token) group and the
// customer-facing endpoints on the public group, per the repo's route convention.
func (r RecommendationRouter) InitRecommendationRouter(publicGroup, systemGroup *gin.RouterGroup) {
	admin := systemGroup.Group("api/v1/recommendation")
	public := publicGroup.Group("api/v1/recommendation/public")

	{
		admin.POST("createRecommendation", recommendationApi.CreateRecommendation)
		admin.PUT("updateRecommendation", recommendationApi.UpdateRecommendation)
		admin.DELETE("deleteRecommendation", recommendationApi.DeleteRecommendation)
		admin.DELETE("deleteRecommendationByIds", recommendationApi.DeleteRecommendationByIds)
		admin.GET("findRecommendation", recommendationApi.FindRecommendation)
		admin.GET("getRecommendationList", recommendationApi.GetRecommendationList)
		admin.POST("review", recommendationApi.Review)
		admin.POST("batchReview", recommendationApi.BatchReview)
		admin.POST("generate", recommendationApi.Generate)
		admin.GET("getRunList", recommendationApi.GetRunList)
		admin.GET("findRun", recommendationApi.FindRun)
		admin.GET("getStats", recommendationApi.GetStats)
		admin.GET("getSetting", recommendationApi.GetSetting)
		admin.PUT("updateSetting", recommendationApi.UpdateSetting)
	}
	{
		public.GET("list", recommendationPublicApi.List)
		public.POST("event",
			middleware.SupportRateLimit(constants.RecommendationRateLimitEvent, 120, time.Minute),
			recommendationPublicApi.Event)
	}
}
