package payment

import "github.com/gin-gonic/gin"

type PaymentRouter struct{}

// InitPaymentRouter registers payment endpoints:
//   - public  : create payment (buyer) + Stripe webhook (provider callback, no auth)
//   - manage  : agent payout + earnings views
//   - system  : capture, refund, payout prepare/pay + admin payout list (admin)
func (r PaymentRouter) InitPaymentRouter(publicGroup, manageGroup, systemGroup *gin.RouterGroup) {
	public := publicGroup.Group("api/v1/payment")
	manage := manageGroup.Group("api/v1/payment/payouts")
	system := systemGroup.Group("api/v1/payment")

	{
		public.POST("", paymentApi.CreatePayment)
		public.POST("webhook/stripe", paymentApi.StripeWebhook)
	}
	{
		manage.GET("manage", payoutApi.ListAgentPayouts)
		manage.GET("earnings", payoutApi.AgentEarnings)
	}
	{
		system.POST("capture", paymentApi.Capture)
		system.POST("refund", paymentApi.Refund)
		system.POST("payouts/prepare", payoutApi.PreparePayouts)
		system.POST("payouts/pay", payoutApi.PayAgent)
		system.GET("payouts", payoutApi.ListAdminPayouts)
	}
}
