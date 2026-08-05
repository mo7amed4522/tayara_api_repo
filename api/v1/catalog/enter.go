package catalog

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the catalog api handlers; embedded into api.ApiGroupApp.CatalogApiGroup.
type ApiGroup struct {
	CatalogApi
	CustomTripApi
}

// catalogService reaches the catalog services via the group var (promotes embedded methods).
var catalogService = service.ServiceGroupApp.CatalogServiceGroup
