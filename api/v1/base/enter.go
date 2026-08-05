package base

import (
	baseService "github.com/mo7amed4522/tayara_api/service/base"
)

// ApiGroup groups the base/lookup api handlers; embedded into api.ApiGroupApp.BaseApiGroup.
type ApiGroup struct {
	LookupApi
}

// baseServiceApp reaches the base services via the package var.
var baseServiceApp = baseService.BaseServiceApp
