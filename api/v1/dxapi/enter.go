package dxapi

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the dxapi (agent) api handlers; embedded into api.ApiGroupApp.ManageApiGroup.
type ApiGroup struct {
	AgentApi
}

// agentService reaches the dxapi services via the group var.
var agentService = service.ServiceGroupApp.ManageServiceGroup
