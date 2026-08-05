package dxapi

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the dxapi (agent) routers; embedded into router.RouterGroupApp.Manage.
type RouterGroup struct {
	AgentRouter
}

var (
	agentApi = api.ApiGroupApp.ManageApiGroup.AgentApi
)
