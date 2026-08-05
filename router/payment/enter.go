package payment

import api "github.com/mo7amed4522/tayara_api/api/v1"

// RouterGroup groups the payment routers; embedded into router.RouterGroupApp.Payment.
type RouterGroup struct {
	PaymentRouter
}

var (
	paymentApi    = api.ApiGroupApp.PaymentApiGroup.PaymentApi
	payoutApi     = api.ApiGroupApp.PaymentApiGroup.PayoutApi
	commissionApi = api.ApiGroupApp.PaymentApiGroup.CommissionApi
)
