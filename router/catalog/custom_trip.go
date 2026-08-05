package catalog

import "github.com/gin-gonic/gin"

type CustomTripRouter struct{}

// InitCustomTripRouter registers custom-trip builder template endpoints (agent-only).
func (r CustomTripRouter) InitCustomTripRouter(manageGroup *gin.RouterGroup) {
	manage := manageGroup.Group("api/v1/catalog")

	{
		manage.POST("templates", customTripApi.CreateTemplate)
		manage.PUT("templates", customTripApi.UpdateTemplate)
		manage.DELETE("templates", customTripApi.DeleteTemplate)
		manage.GET("templates", customTripApi.ListOwnTemplates)
		manage.GET("templates/detail", customTripApi.GetOwnTemplate)
	}
}
