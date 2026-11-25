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
	// RedirectURL is the provider-hosted approval page (PayPal). Open it to let the buyer approve.
	RedirectURL string `json:"redirectUrl,omitempty"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	// ChargeAmount/ChargeCurrency are what the provider actually charges (may differ from
	// Amount/Currency when a settlement currency is used, e.g. AED -> USD for PayPal).
	ChargeAmount   int64  `json:"chargeAmount"`
	ChargeCurrency string `json:"chargeCurrency"`
	FXRatePPM      int64  `json:"fxRatePpm,omitempty"`
}

// PaymentDetail a payment transaction + its refunds.
type PaymentDetail struct {
	payment.PaymentTransaction
	Refunds []payment.Refund `json:"refunds"`
}
