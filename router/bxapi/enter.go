package bxapi

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the bxapi (user) routers; embedded into router.RouterGroupApp.BxApi.
type RouterGroup struct {
	UserRouter
}

var (
	userApi = api.ApiGroupApp.BxApiGroup.UserApi
)
