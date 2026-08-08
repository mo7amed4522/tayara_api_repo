package ticket

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the ticket routers; embedded into router.RouterGroupApp.
type RouterGroup struct {
	TicketRouter
}

var (
	ticketApi = api.ApiGroupApp.TicketApiGroup.TicketApi
)