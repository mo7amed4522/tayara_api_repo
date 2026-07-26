package router

import (
	"github.com/mo7amed4522/tayara_api/router/bxapi"
	"github.com/mo7amed4522/tayara_api/router/dxapi"
	"github.com/mo7amed4522/tayara_api/router/system"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	System system.RouterGroup
	BxApi  bxapi.RouterGroup
	Manage dxapi.RouterGroup
}
