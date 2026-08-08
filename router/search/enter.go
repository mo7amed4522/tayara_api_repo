package search

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the search routers; embedded into router.RouterGroupApp.
type RouterGroup struct {
	SearchRouter
}

var (
	searchApi = api.ApiGroupApp.SearchApiGroup.SearchApi
)