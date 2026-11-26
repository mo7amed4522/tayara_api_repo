package support

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the support routers; embedded into router.RouterGroupApp.Support.
type RouterGroup struct {
	SupportRouter
}

var (
	ticketApi     = api.ApiGroupApp.SupportApiGroup.TicketApi
	messageApi    = api.ApiGroupApp.SupportApiGroup.MessageApi
	attachmentApi = api.ApiGroupApp.SupportApiGroup.AttachmentApi
	presenceApi   = api.ApiGroupApp.SupportApiGroup.PresenceApi
	faqApi        = api.ApiGroupApp.SupportApiGroup.FaqApi
	adminApi      = api.ApiGroupApp.SupportApiGroup.AdminApi
	sseClientApi  = api.ApiGroupApp.SupportApiGroup.SseClientApi
	pushApi       = api.ApiGroupApp.SupportApiGroup.PushApi
)
