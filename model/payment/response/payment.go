package response

import "github.com/mo7amed4522/tayara_api/model/payment"

// CreatePaymentResp returned to the client so it can complete checkout with the provider
// (Stripe client secret, or a BNPL session id). Never contains card data.
type CreatePaymentResp struct {
	PaymentID    string `json:"paymentId"`
	Provider     string `json:"provider"`
	Method       string `json:"method"`
	IntentID     string `json:"intentId"`
	ClientSecret string `json:"clientSecret"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
}

// PaymentDetail a payment transaction + its refunds.
type PaymentDetail struct {
	payment.PaymentTransaction
	Refunds []payment.Refund `json:"refunds"`
}
