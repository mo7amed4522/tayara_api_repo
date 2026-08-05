package engage

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the engage api handlers; embedded into api.ApiGroupApp.EngageApiGroup.
type ApiGroup struct {
	ReviewApi
	AdApi
	NotificationApi
	UserEngageApi
}

// engageService reaches the engage services via the group var.
var engageService = service.ServiceGroupApp.EngageServiceGroup
