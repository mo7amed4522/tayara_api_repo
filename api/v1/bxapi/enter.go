package bxapi

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the bxapi (user) api handlers; embedded into api.ApiGroupApp.BxApiGroup.
type ApiGroup struct {
	UserApi
}

// userService reaches the bxapi services via the group var.
var userService = service.ServiceGroupApp.BxApiServiceGroup
