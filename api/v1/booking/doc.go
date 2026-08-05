package booking

import (
	"context"

	"github.com/gin-gonic/gin"
	bookingReq "github.com/mo7amed4522/tayara_api/model/booking/request"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

type DocApi struct{}

// SubmitDocuments
// @Tags Booking
// @Summary User submits (or resubmits) traveler documents
// @Produce application/json
// @Param data body bookingReq.SubmitDocumentsReq true "documents payload"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/documents [post]
func (DocApi) SubmitDocuments(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	if userID == 0 {
		response.FailWithMessage("unauthorized user", c)
		return
	}
	var req bookingReq.SubmitDocumentsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.DocServices.SubmitDocuments(context.Background(), userID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("documents submitted", c)
}

// GetUserDocuments
// @Tags Booking
// @Summary User lists their own traveler documents (audited)
// @Produce application/json
// @Param id query int true "booking id"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/documents [get]
func (DocApi) GetUserDocuments(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := bookingService.DocServices.GetUserDocuments(context.Background(), userID, idReq.Uint())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ApproveDocument
// @Tags Booking
// @Summary Agent approves a traveler document or requests a correction
// @Produce application/json
// @Param data body bookingReq.ApproveDocumentReq true "review payload"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/documents/review [post]
func (DocApi) ApproveDocument(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req bookingReq.ApproveDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.DocServices.ApproveDocument(context.Background(), agentID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("document reviewed", c)
}

// AgentDocuments
// @Tags Booking
// @Summary Agent lists traveler + booking documents for one of their bookings (audited)
// @Produce application/json
// @Param id query int true "booking id"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/documents [get]
func (DocApi) AgentDocuments(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	travelerDocs, err := bookingService.DocServices.GetAgentTravelerDocuments(context.Background(), agentID, idReq.Uint())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	bookingDocs, _ := bookingService.DocServices.GetAgentBookingDocuments(context.Background(), agentID, idReq.Uint())
	response.OkWithData(gin.H{"travelerDocuments": travelerDocs, "bookingDocuments": bookingDocs}, c)
}

// UploadBookingDocument
// @Tags Booking
// @Summary Agent uploads a booking-level document (confirmation / issued ticket)
// @Produce application/json
// @Param data body bookingReq.BookingDocumentReq true "document payload"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/documents/upload [post]
func (DocApi) UploadBookingDocument(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var req bookingReq.BookingDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.DocServices.UploadBookingDocument(context.Background(), "agent", agentID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("document uploaded", c)
}

// MarkProcessing
// @Tags Booking
// @Summary Agent marks a booking processing (docs OK + KYC approved)
// @Produce application/json
// @Param id query int true "booking id"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/process [post]
func (DocApi) MarkProcessing(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.DocServices.MarkProcessing(context.Background(), agentID, idReq.Uint()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("booking processing", c)
}

// IssueBooking
// @Tags Booking
// @Summary Agent issues tickets / confirmation (issued_ready)
// @Produce application/json
// @Param id query int true "booking id"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/issue [post]
func (DocApi) IssueBooking(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.DocServices.IssueBooking(context.Background(), agentID, idReq.Uint()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("booking issued", c)
}

// CompleteBooking
// @Tags Booking
// @Summary Agent completes a booking (booking_confirmed -> completed)
// @Produce application/json
// @Param id query int true "booking id"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/manage/complete [post]
func (DocApi) CompleteBooking(c *gin.Context) {
	agentID := utils.GetAgentID(c)
	var idReq commonReq.GetById
	if err := c.ShouldBindQuery(&idReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.DocServices.CompleteBooking(context.Background(), agentID, idReq.Uint()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("booking completed", c)
}

// AdminReviewKYC
// @Tags Booking
// @Summary Admin approves or rejects user KYC
// @Produce application/json
// @Param data body bookingReq.KYCReviewReq true "kyc review payload"
// @Success 200 {object} response.Response
// @Router /api/v1/booking/admin/kyc [post]
func (DocApi) AdminReviewKYC(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req bookingReq.KYCReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := bookingService.DocServices.AdminReviewKYC(context.Background(), adminID, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("KYC reviewed", c)
}
