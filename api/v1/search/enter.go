package search

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the search api handlers; embedded into api.ApiGroupApp.
type ApiGroup struct {
	SearchApi
}

// searchService reaches the search services via the service group var.
var searchService = service.ServiceGroupApp.SearchServiceGroup