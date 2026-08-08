package ticket

import "github.com/gin-gonic/gin"

type TicketRouter struct{}

// InitTicketRouter registers the e-ticket endpoints on the system (admin) guard group.
func (r TicketRouter) InitTicketRouter(systemGroup *gin.RouterGroup) {
	system := systemGroup.Group("api/v1/admin/tickets")
	{
		system.GET(":id/pdf", ticketApi.Download)
		system.POST(":id/send", ticketApi.Send)
	}
}