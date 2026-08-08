package ticket

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the ticket api handlers; embedded into api.ApiGroupApp.
type ApiGroup struct {
	TicketApi
}

// ticketService reaches the ticket services via the service group var.
var ticketService = service.ServiceGroupApp.TicketServiceGroup