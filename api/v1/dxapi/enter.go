package dxapi

import "github.com/mo7amed4522/tayara_api/service"

type ApiGroup struct {
	DxAgentProfileApi
}

var (
	dxAgentProfileService = service.ServiceGroupApp.ManageServiceGroup
)
