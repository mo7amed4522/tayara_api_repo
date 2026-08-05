package base

import "github.com/gin-gonic/gin"

type LookupRouter struct{}

// InitLookupRouter registers the base lookup endpoints:
//   - public  : list package categories
//   - system  : upsert/delete categories + settings change history (audit)
func (r LookupRouter) InitLookupRouter(publicGroup, systemGroup *gin.RouterGroup) {
	public := publicGroup.Group("api/v1/base")
	system := systemGroup.Group("api/v1/base")

	{
		public.GET("categories", baseApi.ListCategories)
	}
	{
		system.POST("categories", baseApi.UpsertCategory)
		system.DELETE("categories", baseApi.DeleteCategory)
		system.POST("settings/history", baseApi.RecordSettingChange)
		system.GET("settings/history", baseApi.SettingHistory)
	}
}
