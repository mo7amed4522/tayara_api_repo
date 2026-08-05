package providers

import (
	"context"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/utils/stripe"
)

// StripeProvider handles card and wallet payments (Visa/Mastercard/Apple Pay/Google Pay) via
// PaymentIntents. Card numbers never reach our server: Stripe.js tokenises them client-side
// and we only ever hold a payment_intent id and client secret.
type StripeProvider struct{}

func (StripeProvider) Name() string  { return constants.PaymentProviderStripe }
func (StripeProvider) Enabled() bool { return stripe.Enabled() }

func (p StripeProvider) CreatePayment(ctx context.Context, pr CreatePaymentParams) (*CreatePaymentResult, error) {
	view, err := stripe.CreatePaymentIntentWithMethods(
		pr.BookingID, pr.BookingCode, pr.Amount, pr.Currency,
		global.GVA_CONFIG.Stripe.PaymentMethods,
	)
	if err != nil {
		return nil, err
	}
	return &CreatePaymentResult{
		Provider:     p.Name(),
		IntentID:     view.ID,
		ClientSecret: view.ClientSecret,
		Status:       view.Status,
	}, nil
}

func (StripeProvider) Capture(ctx context.Context, intentID string, amount *int64) error {
	_, err := stripe.CapturePaymentIntent(intentID, amount)
	return err
}

func (StripeProvider) Refund(ctx context.Context, intentID string, amount *int64, reason string) (string, error) {
	return stripe.RefundPaymentIntent(intentID, amount, reason)
}
