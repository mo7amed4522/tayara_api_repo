package bxapi

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the bxapi routers; embedded into router.RouterGroupApp.BxApi.
type RouterGroup struct {
	UserRouter
	HomeRouter
	AgentDetailRouter
}

var (
	userApi = api.ApiGroupApp.BxApiGroup.UserApi
	homeApi = api.ApiGroupApp.BxApiGroup.HomeApi
	agentDetailApi = api.ApiGroupApp.BxApiGroup.AgentDetailApi
)
