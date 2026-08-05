package catalog

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the catalog routers; embedded into router.RouterGroupApp.Catalog.
type RouterGroup struct {
	CatalogRouter
	CustomTripRouter
}

var (
	catalogApi     = api.ApiGroupApp.CatalogApiGroup.CatalogApi
	customTripApi  = api.ApiGroupApp.CatalogApiGroup.CustomTripApi
	requirementApi = api.ApiGroupApp.CatalogApiGroup.RequirementApi
)
