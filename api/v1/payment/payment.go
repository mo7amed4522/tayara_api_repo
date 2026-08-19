package payment

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	paymentReq "github.com/mo7amed4522/tayara_api/model/payment/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

type PaymentApi struct{}

// CreatePayment
//	@Tags		Payment
//	@Summary	User opens a payment for a booking (returns client secret / BNPL session)
//	@Produce	application/json
//	@Param		data	body		paymentReq.CreatePaymentReq	true	"payment payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/payment [post]
func (PaymentApi) CreatePayment(c *gin.Context) {
	userID := utils.GetBxUserID(c)
	if userID == 0 {
		response.FailWithMessage("unauthorized user", c)
		return
	}
	var req paymentReq.CreatePaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	resp, err := paymentService.PaymentServices.CreatePayment(context.Background(), userID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(resp, c)
}

// Capture
//	@Tags		Payment
//	@Summary	Admin/system captures an authorized payment
//	@Produce	application/json
//	@Param		data	body		paymentReq.CaptureReq	true	"capture payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/payment/capture [post]
func (PaymentApi) Capture(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req paymentReq.CaptureReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := paymentService.PaymentServices.Capture(context.Background(), adminID, req.BookingID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// Refund
//	@Tags		Payment
//	@Summary	Admin refunds a captured payment (full/partial)
//	@Produce	application/json
//	@Param		data	body		paymentReq.RefundReq	true	"refund payload"
//	@Success	200		{object}	response.Response
//	@Router		/api/v1/payment/refund [post]
func (PaymentApi) Refund(c *gin.Context) {
	adminID := utils.GetUserID(c)
	var req paymentReq.RefundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	detail, err := paymentService.PaymentServices.Refund(context.Background(), adminID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(detail, c)
}

// StripeWebhook
//	@Tags		Payment
//	@Summary	Stripe webhook (payment authorization events)
//	@Produce	application/json
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/payment/webhook/stripe [post]
func (PaymentApi) StripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.FailWithMessage("failed to read body", c)
		return
	}
	sig := c.GetHeader("Stripe-Signature")
	if err := paymentService.PaymentServices.HandleStripeWebhook(payload, sig); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "ok"})
}
