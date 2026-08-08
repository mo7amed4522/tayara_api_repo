package export

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the export routers; embedded into router.RouterGroupApp.
type RouterGroup struct {
	ExportRouter
}

var (
	exportApi = api.ApiGroupApp.ExportApiGroup.ExportApi
)