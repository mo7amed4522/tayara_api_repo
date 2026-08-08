package export

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the export api handlers; embedded into api.ApiGroupApp.
type ApiGroup struct {
	ExportApi
}

// exportService reaches the export services via the service group var.
var exportService = service.ServiceGroupApp.ExportServiceGroup