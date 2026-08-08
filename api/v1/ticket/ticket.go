package ticket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type TicketApi struct{}

// Download
// @Tags Ticket
// @Summary Generate + download a booking e-ticket PDF
// @Produce application/pdf
// @Param id path uint true "booking id"
// @Router /api/v1/admin/tickets/{id}/pdf [get]
func (TicketApi) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("valid booking id required", c)
		return
	}
	data, err := ticketService.TicketServices.GenerateBookingTicket(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Writer.Header().Set("Content-Type", "application/pdf")
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=ticket_"+c.Param("id")+".pdf")
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write(data)
}

// Send
// @Tags Ticket
// @Summary Admin emails a booking e-ticket PDF to the buyer
// @Produce application/json
// @Param id path uint true "booking id"
// @Router /api/v1/admin/tickets/{id}/send [post]
func (TicketApi) Send(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("valid booking id required", c)
		return
	}
	status, err := ticketService.TicketServices.SendBookingTicket(utils.GetUserID(c), uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"delivery": status}, c)
}