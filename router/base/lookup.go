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
		system.GET("settings", baseApi.Settings)
		system.GET("settings/keys", baseApi.ConfigKeys)
		system.GET("settings/:key", baseApi.GetSetting)
		system.PUT("settings", baseApi.UpdateSetting)
		system.DELETE("settings/:key", baseApi.DeleteSetting)
		system.POST("settings/reset", baseApi.ResetSection)
		system.POST("settings/history", baseApi.RecordSettingChange)
		system.GET("settings/history", baseApi.SettingHistory)
	}
}
