package payment

import "github.com/gin-gonic/gin"

type PaymentRouter struct{}

// InitPaymentRouter registers payment endpoints:
//   - public  : create payment (buyer) + Stripe webhook (provider callback, no auth)
//   - system  : capture + refund (admin)
func (r PaymentRouter) InitPaymentRouter(publicGroup, systemGroup *gin.RouterGroup) {
	public := publicGroup.Group("api/v1/payment")
	system := systemGroup.Group("api/v1/payment")

	{
		public.POST("", paymentApi.CreatePayment)
		public.POST("webhook/stripe", paymentApi.StripeWebhook)
	}
	{
		system.POST("capture", paymentApi.Capture)
		system.POST("refund", paymentApi.Refund)
	}
}
