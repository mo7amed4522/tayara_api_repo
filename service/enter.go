package service

import (
	"github.com/mo7amed4522/tayara_api/service/bxapi"
	"github.com/mo7amed4522/tayara_api/service/dxapi"
	"github.com/mo7amed4522/tayara_api/service/system"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup system.ServiceGroup
	BxApiServiceGroup  bxapi.ServiceGroup
	ManageServiceGroup dxapi.ServiceGroup
}
