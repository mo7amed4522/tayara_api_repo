package providers

import (
	"context"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/utils/paypal"
)

// PayPalProvider handles PayPal wallet payments through Orders v2 with the AUTHORIZE intent:
// the buyer approves on PayPal, the verified webhook authorizes the order, and the admin captures
// later exactly like a manual-capture Stripe payment. No card or bank data reaches our server.
type PayPalProvider struct{}

func (PayPalProvider) Name() string  { return constants.PaymentProviderPayPal }
func (PayPalProvider) Enabled() bool { return paypal.Enabled() }

func (p PayPalProvider) CreatePayment(ctx context.Context, pr CreatePaymentParams) (*CreatePaymentResult, error) {
	order, err := paypal.CreateOrder(ctx, pr.BookingID, pr.BookingCode, pr.Amount, pr.Currency)
	if err != nil {
		return nil, err
	}
	return &CreatePaymentResult{
		Provider:       p.Name(),
		IntentID:       order.ID,
		RedirectURL:    order.ApprovalURL,
		ChargeAmount:   order.ChargeAmount,
		ChargeCurrency: order.ChargeCurrency,
		FXRatePPM:      order.FXRatePPM,
		Status:         order.Status,
	}, nil
}

func (PayPalProvider) Capture(ctx context.Context, intentID string, amount *int64) error {
	_, err := paypal.CaptureOrder(ctx, intentID)
	return err
}

func (PayPalProvider) Refund(ctx context.Context, intentID string, amount *int64, reason string) (string, error) {
	return paypal.RefundOrder(ctx, intentID, amount, reason)
}
