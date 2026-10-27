package bxapi

import (
	"github.com/gin-gonic/gin"
	api "github.com/mo7amed4522/tayara_api/api/v1"
)

// ExploreRouter handles explore topic routes for public/agent APIs
type ExploreRouter struct{}

// InitExploreRouter registers explore topic endpoints
func (r ExploreRouter) InitExploreRouter(publicGroup, bxGroup, manageGroup *gin.RouterGroup) {
	// Public: explore topics
	public := publicGroup.Group("api/v1/customer/explore")
	{
		public.GET("topics", api.ApiGroupApp.BxApiGroup.ExploreApi.ListExploreTopics)
		public.GET("topics/:code", api.ApiGroupApp.BxApiGroup.ExploreApi.GetExploreTopicDetail)
		public.GET("topics/:code/packages", api.ApiGroupApp.BxApiGroup.ExploreApi.GetExploreTopicPackages)
	}

	// Agent: manage package-topic associations
	agent := manageGroup.Group("api/v1/agent/explore")
	{
		agent.GET("topics", api.ApiGroupApp.BxApiGroup.ExploreApi.AgentGetMyTopics)
		agent.POST("packages/:id/topics", api.ApiGroupApp.BxApiGroup.ExploreApi.AgentAddPackageToTopics)
		agent.DELETE("packages/:id/topics/:topicId", api.ApiGroupApp.BxApiGroup.ExploreApi.AgentRemovePackageFromTopic)
		agent.GET("packages/:id/topics", api.ApiGroupApp.BxApiGroup.ExploreApi.AgentGetPackageTopics)
	}
}