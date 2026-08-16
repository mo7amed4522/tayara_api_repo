package bxapi

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the bxapi api handlers; embedded into api.ApiGroupApp.BxApiGroup.
type ApiGroup struct {
	UserApi
	HomeApi
	AgentDetailApi
}

// homeService reaches the bxapi services via the group var.
var homeService = service.ServiceGroupApp.BxApiServiceGroup

// userService reaches the bxapi services via the group var.
var userService = service.ServiceGroupApp.BxApiServiceGroup

// agentDetailService reaches the bxapi services via the group var.
var agentDetailService = service.ServiceGroupApp.BxApiServiceGroup
