package stripe

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/refund"
	"github.com/stripe/stripe-go/v82/webhook"
)

type PaymentIntentView struct {
	ID               string // stripe payment intent id
	ClientSecret     string // stripe payment intent client secret
	Status           string // stripe payment intent status
	Amount           int64  // stripe payment intent amount
	Currency         string // stripe payment intent currency
	CaptureMethod    string // stripe payment intent capture method
	BookingID        uint   // booking id
	BookingCode      string // booking code
	AuthorizedAt     int64  // authorized at
	CapturedAt       int64  // captured at
	CancellationAt   int64  // cancellation at
	LastPaymentError string // last payment error
} // payment intent view

func Enabled() bool {
	return strings.TrimSpace(global.GVA_CONFIG.Stripe.SecretKey) != ""
}

func PublishableKey() string {
	return strings.TrimSpace(global.GVA_CONFIG.Stripe.PublishableKey)
}

func defaultCurrency() string {
	currency := strings.TrimSpace(global.GVA_CONFIG.Stripe.Currency)
	if currency == "" {
		return "AED"
	}
	return strings.ToUpper(currency)
}

func captureMethod() string {
	if global.GVA_CONFIG.Stripe.AutoCapture {
		return string(stripe.PaymentIntentCaptureMethodAutomatic)
	}
	method := strings.TrimSpace(global.GVA_CONFIG.Stripe.CaptureMethod)
	if method == "" {
		return string(stripe.PaymentIntentCaptureMethodManual)
	}
	return method
}

func configure() {
	stripe.Key = strings.TrimSpace(global.GVA_CONFIG.Stripe.SecretKey)
}

func CreatePaymentIntent(bookingID uint, bookingCode string, amount int64, currency string) (*PaymentIntentView, error) {
	if !Enabled() {
		return nil, errors.New("stripe is not configured")
	}
	if amount <= 0 {
		return nil, errors.New("payment amount must be greater than zero")
	}
	configure()
	if currency == "" {
		currency = defaultCurrency()
	}
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amount),
		Currency:      stripe.String(strings.ToLower(currency)),
		CaptureMethod: stripe.String(captureMethod()),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: map[string]string{
			"booking_id":   strconv.FormatUint(uint64(bookingID), 10),
			"booking_code": bookingCode,
		},
	}
	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}
	return toPaymentIntentView(intent), nil
}

// CreatePaymentIntentWithMethods creates a PaymentIntent restricted to the given client
// card/wallet methods (card, apple_pay, google_pay). Card numbers never touch our server —
// Stripe.js/Elements tokenise the payment method client-side; we only receive method ids.
func CreatePaymentIntentWithMethods(bookingID uint, bookingCode string, amount int64, currency string, methods []string) (*PaymentIntentView, error) {
	if !Enabled() {
		return nil, errors.New("stripe is not configured")
	}
	if amount <= 0 {
		return nil, errors.New("payment amount must be greater than zero")
	}
	configure()
	if currency == "" {
		currency = defaultCurrency()
	}
	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(amount),
		Currency:           stripe.String(strings.ToLower(currency)),
		CaptureMethod:      stripe.String(captureMethod()),
		PaymentMethodTypes: paymentMethodTypes(methods),
		Metadata: map[string]string{
			"booking_id":   strconv.FormatUint(uint64(bookingID), 10),
			"booking_code": bookingCode,
		},
	}
	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}
	return toPaymentIntentView(intent), nil
}

// paymentMethodTypes maps our logical methods to Stripe PaymentMethodType string values.
// card/visa/mastercard -> "card"; apple_pay -> "apple_pay"; google_pay -> "google_pay".
func paymentMethodTypes(methods []string) []*string {
	if len(methods) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var out []*string
	for _, m := range methods {
		switch strings.ToLower(strings.TrimSpace(m)) {
		case "card", "visa", "mastercard":
			if !seen["card"] {
				seen["card"] = true
				out = append(out, stripe.String(string(stripe.PaymentMethodTypeCard)))
			}
		case "apple_pay":
			if !seen["apple_pay"] {
				seen["apple_pay"] = true
				out = append(out, stripe.String("apple_pay"))
			}
		case "google_pay":
			if !seen["google_pay"] {
				seen["google_pay"] = true
				out = append(out, stripe.String("google_pay"))
			}
		}
	}
	return out
}

func GetPaymentIntent(paymentIntentID string) (*PaymentIntentView, error) {
	if !Enabled() {
		return nil, errors.New("stripe is not configured")
	}
	if strings.TrimSpace(paymentIntentID) == "" {
		return nil, errors.New("payment intent id is required")
	}
	configure()
	intent, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}
	return toPaymentIntentView(intent), nil
}

func CapturePaymentIntent(paymentIntentID string, amount *int64) (*PaymentIntentView, error) {
	if !Enabled() {
		return nil, errors.New("stripe is not configured")
	}
	configure()
	params := &stripe.PaymentIntentCaptureParams{}
	if amount != nil && *amount > 0 {
		params.AmountToCapture = stripe.Int64(*amount)
	}
	intent, err := paymentintent.Capture(paymentIntentID, params)
	if err != nil {
		return nil, err
	}
	return toPaymentIntentView(intent), nil
}

func CancelPaymentIntent(paymentIntentID string, reason string) (*PaymentIntentView, error) {
	if !Enabled() {
		return nil, errors.New("stripe is not configured")
	}
	configure()
	params := &stripe.PaymentIntentCancelParams{}
	if reason != "" {
		params.CancellationReason = stripe.String(reason)
	}
	intent, err := paymentintent.Cancel(paymentIntentID, params)
	if err != nil {
		return nil, err
	}
	return toPaymentIntentView(intent), nil
}

func RefundPaymentIntent(paymentIntentID string, amount *int64, reason string) (string, error) {
	if !Enabled() {
		return "", errors.New("stripe is not configured")
	}
	configure()
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}
	if amount != nil && *amount > 0 {
		params.Amount = stripe.Int64(*amount)
	}
	if reason != "" {
		params.Reason = stripe.String(reason)
	}
	ref, err := refund.New(params)
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}

func ConstructWebhookEvent(payload []byte, signature string) (stripe.Event, error) {
	secret := strings.TrimSpace(global.GVA_CONFIG.Stripe.WebhookSecret)
	if secret == "" {
		return stripe.Event{}, errors.New("stripe webhook secret is not configured")
	}
	return webhook.ConstructEvent(payload, signature, secret)
}

func IsAuthorizedStatus(status string) bool {
	return status == string(stripe.PaymentIntentStatusRequiresCapture)
}

func IsCapturedStatus(status string) bool {
	return status == string(stripe.PaymentIntentStatusSucceeded)
}

func IsFailedStatus(status string) bool {
	return status == string(stripe.PaymentIntentStatusCanceled) ||
		status == string(stripe.PaymentIntentStatusRequiresPaymentMethod)
}

func toPaymentIntentView(intent *stripe.PaymentIntent) *PaymentIntentView {
	if intent == nil {
		return nil
	}
	view := &PaymentIntentView{
		ID:            intent.ID,
		ClientSecret:  intent.ClientSecret,
		Status:        string(intent.Status),
		Amount:        intent.Amount,
		Currency:      strings.ToUpper(string(intent.Currency)),
		CaptureMethod: string(intent.CaptureMethod),
	}
	if intent.Metadata != nil {
		if bookingID, err := strconv.ParseUint(intent.Metadata["booking_id"], 10, 64); err == nil {
			view.BookingID = uint(bookingID)
		}
		view.BookingCode = intent.Metadata["booking_code"]
	}
	if intent.LatestCharge != nil {
		if intent.LatestCharge.Created > 0 {
			view.AuthorizedAt = intent.LatestCharge.Created
		}
	}
	if intent.LastPaymentError != nil {
		view.LastPaymentError = intent.LastPaymentError.Msg
	}
	return view
}

func FormatStripeError(err error) string {
	if err == nil {
		return ""
	}
	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) {
		return fmt.Sprintf("%s: %s", stripeErr.Code, stripeErr.Msg)
	}
	return err.Error()
}
