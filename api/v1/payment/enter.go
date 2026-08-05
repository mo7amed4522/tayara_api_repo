package payment

import "github.com/mo7amed4522/tayara_api/service"

// ApiGroup groups the payment api handlers; embedded into api.ApiGroupApp.PaymentApiGroup.
type ApiGroup struct {
	PaymentApi
	PayoutApi
	CommissionApi
}

// paymentService reaches the payment services via the group var.
var paymentService = service.ServiceGroupApp.PaymentServiceGroup
