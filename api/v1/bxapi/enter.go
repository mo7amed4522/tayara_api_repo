package bxapi

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the bxapi api handlers; embedded into api.ApiGroupApp.BxApiGroup.
type ApiGroup struct {
	UserApi
	HomeApi
	AgentDetailApi
	PackageDetailApi
}

// bxapiSvc reaches the bxapi services via the group var.
var bxapiSvc = service.ServiceGroupApp.BxApiServiceGroup
