package analytics

import "github.com/gin-gonic/gin"

type AnalyticsRouter struct{}

// InitAnalyticsRouter registers the analytics endpoints:
//   - manage (agent): own dashboard card (scoped by agent id from the token)
//   - system (admin): Users / Agents / Business aggregate cards
func (r AnalyticsRouter) InitAnalyticsRouter(manageGroup, systemGroup *gin.RouterGroup) {
	manage := manageGroup.Group("api/v1/analytics")
	system := systemGroup.Group("api/v1/analytics")

	{
		manage.GET("agent", analyticsApi.AgentOverview)
	}
	{
		system.GET("admin/users", analyticsApi.AdminUsers)
		system.GET("admin/agents", analyticsApi.AdminAgents)
		system.GET("admin/business", analyticsApi.AdminBusiness)
	}
}
