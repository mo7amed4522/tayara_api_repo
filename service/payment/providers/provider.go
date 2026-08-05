package providers

import (
	"context"

	"github.com/mo7amed4522/tayara_api/constants"
)

// CreatePaymentParams inputs for opening a payment with a provider.
type CreatePaymentParams struct {
	BookingID   uint
	BookingCode string
	Amount      int64 // minor units
	Currency    string
}

// CreatePaymentResult the provider's payment handle returned to the client.
type CreatePaymentResult struct {
	Provider     string
	IntentID     string // payment intent / checkout session id
	ClientSecret string // Stripe client secret (never a card value)
	Status       string
}

// Provider abstracts a payment processor. Stripe handles cards/wallets; Tabby and Tamara are
// BNPL aggregators. Providers are interchangeable behind this interface.
type Provider interface {
	Name() string
	Enabled() bool
	CreatePayment(ctx context.Context, p CreatePaymentParams) (*CreatePaymentResult, error)
	Capture(ctx context.Context, intentID string, amount *int64) error
	Refund(ctx context.Context, intentID string, amount *int64, reason string) (string, error)
}

// Get resolves the provider for a payment method, or nil if unknown.
func Get(method string) Provider {
	switch method {
	case constants.PaymentMethodCard, constants.PaymentMethodApplePay, constants.PaymentMethodGooglePay:
		return StripeProvider{}
	case constants.PaymentMethodTabby:
		return TabbyProvider()
	case constants.PaymentMethodTamara:
		return TamaraProvider()
	}
	return nil
}
