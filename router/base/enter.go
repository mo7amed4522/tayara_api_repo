package base

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the base routers; embedded into router.RouterGroupApp.Base.
type RouterGroup struct {
	LookupRouter
}

var (
	baseApi = api.ApiGroupApp.BaseApiGroup.LookupApi
)
