package bxapi

import (
	"github.com/gin-gonic/gin"
	api "github.com/mo7amed4522/tayara_api/api/v1"
)

// RouterGroup groups the bxapi routers; embedded into router.RouterGroupApp.BxApi.
type RouterGroup struct {
	UserRouter
	HomeRouter
	AgentDetailRouter
	PackageDetailRouter
	ExploreRouter
}

var (
	userApi          = api.ApiGroupApp.BxApiGroup.UserApi
	homeApi          = api.ApiGroupApp.BxApiGroup.HomeApi
	agentDetailApi   = api.ApiGroupApp.BxApiGroup.AgentDetailApi
	packageDetailApi = api.ApiGroupApp.BxApiGroup.PackageDetailApi
	exploreApi       = api.ApiGroupApp.BxApiGroup.ExploreApi
)

// InitExploreRouter registers explore topic routes
func (r RouterGroup) InitExploreRouter(publicGroup, bxGroup, manageGroup *gin.RouterGroup) {
	r.ExploreRouter.InitExploreRouter(publicGroup, bxGroup, manageGroup)
}
