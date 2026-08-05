package engage

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the engage routers; embedded into router.RouterGroupApp.Engage.
type RouterGroup struct {
	EngageRouter
}

var (
	reviewApi = api.ApiGroupApp.EngageApiGroup.ReviewApi
	adApi     = api.ApiGroupApp.EngageApiGroup.AdApi
	notifApi  = api.ApiGroupApp.EngageApiGroup.NotificationApi
	userApi   = api.ApiGroupApp.EngageApiGroup.UserEngageApi
)
