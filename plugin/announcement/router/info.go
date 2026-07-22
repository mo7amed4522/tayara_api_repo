package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/middleware"
)

var Info = new(info)

type info struct{}

// Init Initialize announcement routing information
func (r *info) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		group := private.Group("info").Use(middleware.OperationRecord())
		group.POST("createInfo", apiInfo.CreateInfo)             // Create announcement
		group.DELETE("deleteInfo", apiInfo.DeleteInfo)           // Delete announcement
		group.DELETE("deleteInfoByIds", apiInfo.DeleteInfoByIds) // Batch delete announcement
		group.PUT("updateInfo", apiInfo.UpdateInfo)              // Update announcement
	}
	{
		group := private.Group("info")
		group.GET("findInfo", apiInfo.FindInfo)       // Get announcement by ID
		group.GET("getInfoList", apiInfo.GetInfoList) // Get announcement list
	}
	{
		group := public.Group("info")
		group.GET("getInfoDataSource", apiInfo.GetInfoDataSource) // Get announcement data source
		group.GET("getInfoPublic", apiInfo.GetInfoPublic)         // Get announcement list
	}
}
