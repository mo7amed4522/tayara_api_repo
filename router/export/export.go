package export

import "github.com/gin-gonic/gin"

type ExportRouter struct{}

// InitExportRouter registers the admin CSV export endpoints on the system (admin) guard group.
func (r ExportRouter) InitExportRouter(systemGroup *gin.RouterGroup) {
	system := systemGroup.Group("api/v1/admin/export")
	{
		system.GET("users", exportApi.Users)
		system.GET("agents", exportApi.Agents)
		system.GET("bookings", exportApi.Bookings)
		system.GET("payments", exportApi.Payments)
	}
}