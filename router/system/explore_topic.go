package system

import (
	"github.com/gin-gonic/gin"
	api "github.com/mo7amed4522/tayara_api/api/v1"
)

// ExploreTopicRouter handles admin routes for explore topics
type ExploreTopicRouter struct{}

// InitExploreTopicRouter registers admin explore topic endpoints
func (r ExploreTopicRouter) InitExploreTopicRouter(systemGroup *gin.RouterGroup) {
	admin := systemGroup.Group("api/v1/admin/explore")
	{
		admin.GET("topics", api.ApiGroupApp.SystemApiGroup.ExploreTopicApi.AdminListTopics)
		admin.GET("topics/:id", api.ApiGroupApp.SystemApiGroup.ExploreTopicApi.AdminGetTopic)
		admin.POST("topics", api.ApiGroupApp.SystemApiGroup.ExploreTopicApi.AdminCreateTopic)
		admin.PUT("topics/:id", api.ApiGroupApp.SystemApiGroup.ExploreTopicApi.AdminUpdateTopic)
		admin.DELETE("topics/:id", api.ApiGroupApp.SystemApiGroup.ExploreTopicApi.AdminDeleteTopic)
		admin.PUT("topics/reorder", api.ApiGroupApp.SystemApiGroup.ExploreTopicApi.AdminReorderTopics)
		admin.POST("topics/:id/packages", api.ApiGroupApp.SystemApiGroup.ExploreTopicApi.AdminManageTopicPackages)
		admin.GET("topics/:id/packages", api.ApiGroupApp.SystemApiGroup.ExploreTopicApi.AdminGetTopicPackages)
	}
}