package payment

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/model/common/response"
)

// PayPalWebhook
//
//	@Tags		Payment
//	@Summary	PayPal webhook (verified order approval -> payment authorized)
//	@Produce	application/json
//	@Success	200	{object}	response.Response
//	@Router		/api/v1/payment/webhook/paypal [post]
func (PaymentApi) PayPalWebhook(c *gin.Context) {
	// PayPal payloads are small; cap the body so an unauthenticated route cannot be used to
	// buffer arbitrary data.
	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		response.FailWithMessage("failed to read body", c)
		return
	}
	if err := paymentService.PaymentServices.HandlePayPalWebhook(payload, c.Request.Header); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "ok"})
}
