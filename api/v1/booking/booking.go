package booking

import (
	"github.com/gin-gonic/gin"
	bookingReq "github.com/mo7amed4522/tayara_api/model/booking/request"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type BookingApi struct{}

// CreateBooking
// @Tags Booking
// @Summary User creates a booking (package or custom trip)
// @Produce application/json
// @Param data body bookingReq.CreateBookingReq true "booking payload"
// @Success 200 {object} response.Response
// @Router /api/v1/booking [post]
func (BookingApi) CreateBooking(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	if userID == 0 {
		response.FailWithMessage("unauthorized user", c)
		return
	}
	var req bookingReq.CreateBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	bk, err := bookingService.BookingServices.CreateBooking(userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(bk, c)
}

// GetOwnBooking
// @Tags Booking
// @Summary User gets one of their own bookings (with detail)
// @Produce application/json
// @Param id query int true "booking id"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/detail [get]
func (BookingApi) GetOwnBooking(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := bookingService.BookingServices.GetOwnBooking(userID, idReq.Uint())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// ListOwnBookings
// @Tags Booking
// @Summary User lists their own bookings (paginated)
// @Produce application/json
// @Param data query bookingReq.BookingPageReq true "page filter"
// @Success 200 {object} response.Response
// @Router /api/v1/booking [get]
func (BookingApi) ListOwnBookings(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req bookingReq.BookingPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := bookingService.BookingServices.ListOwnBookings(userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// CancelBooking
// @Tags Booking
// @Summary User cancels their own booking (before confirmation)
// @Produce application/json
// @Param data body bookingReq.CancelBookingReq true "cancel payload"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/cancel [post]
func (BookingApi) CancelBooking(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var req bookingReq.CancelBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.BookingServices.CancelBooking(userID, req.BookingID, req.Reason); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("booking cancelled", c)
}

// ListAgentBookings
// @Tags Booking
// @Summary Agent lists their bookings (paginated)
// @Produce application/json
// @Param data query bookingReq.BookingPageReq true "page filter"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage [get]
func (BookingApi) ListAgentBookings(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req bookingReq.BookingPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := bookingService.BookingServices.ListAgentBookings(agentID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// GetAgentBooking
// @Tags Booking
// @Summary Agent gets one of their bookings
// @Produce application/json
// @Param id query int true "booking id"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/detail [get]
func (BookingApi) GetAgentBooking(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := bookingService.BookingServices.GetAgentBooking(agentID, idReq.Uint())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// ReviewBooking
// @Tags Booking
// @Summary Agent starts reviewing a booking
// @Produce application/json
// @Param data body bookingReq.CancelBookingReq true "booking id + note"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/review [post]
func (BookingApi) ReviewBooking(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req bookingReq.CancelBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.BookingServices.ReviewBooking(agentID, req.BookingID, req.Reason); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("booking under review", c)
}

// SubmitOffer
// @Tags Booking
// @Summary Agent submits the final price offer
// @Produce application/json
// @Param data body bookingReq.SubmitOfferReq true "offer payload"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/offer [post]
func (BookingApi) SubmitOffer(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req bookingReq.SubmitOfferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	offer, err := bookingService.BookingServices.SubmitOffer(agentID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(offer, c)
}

// RejectBooking
// @Tags Booking
// @Summary Agent rejects/modifies a booking with alternative dates
// @Produce application/json
// @Param data body bookingReq.RejectBookingReq true "reject payload"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/reject [post]
func (BookingApi) RejectBooking(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req bookingReq.RejectBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.BookingServices.RejectBooking(agentID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("booking rejected", c)
}

// ListAllBookings
// @Tags Booking
// @Summary Admin lists all bookings
// @Produce application/json
// @Param data query bookingReq.BookingPageReq true "page filter"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/admin [get]
func (BookingApi) ListAllBookings(c *gin.Context) {
	var req bookingReq.BookingPageReq
	_ = c.ShouldBindQuery(&req)
	list, total, err := bookingService.BookingServices.ListAllBookings(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "ok", c)
}

// ForceCancel
// @Tags Booking
// @Summary Admin force-cancels a booking
// @Produce application/json
// @Param data body bookingReq.CancelBookingReq true "cancel payload"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/admin/force-cancel [post]
func (BookingApi) ForceCancel(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req bookingReq.CancelBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.BookingServices.ForceCancel(adminID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("booking force-cancelled", c)
}
