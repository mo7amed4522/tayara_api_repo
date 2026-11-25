package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	paymentModel "github.com/mo7amed4522/tayara_api/model/payment"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/paypal"
)

// paypalEventOrderApproved is the only PayPal event that advances our lifecycle: the buyer has
// approved the order, so we authorize the funds (hold) and wait for the admin to capture.
const paypalEventOrderApproved = "CHECKOUT.ORDER.APPROVED"

// HandlePayPalWebhook processes a PayPal webhook. The signature is verified with PayPal before the
// payload is trusted; the order amount/currency must match the stored payment; the event id is
// processed once; and the state change happens under the per-booking payment red lock.
func (p *PaymentServices) HandlePayPalWebhook(payload []byte, headers http.Header) error {
	if !paypal.Enabled() {
		return errors.New("paypal is not configured")
	}
	if err := paypal.VerifyWebhook(headers, payload); err != nil {
		return err
	}
	var ev struct {
		ID        string          `json:"id"`
		EventType string          `json:"event_type"`
		Resource  json.RawMessage `json:"resource"`
	}
	if err := json.Unmarshal(payload, &ev); err != nil {
		return errors.New("malformed paypal webhook payload")
	}
	if ev.EventType != paypalEventOrderApproved {
		return nil
	}
	if strings.TrimSpace(ev.ID) == "" {
		return errors.New("paypal webhook has no event id")
	}
	// One event id is applied once. A failure releases the key so PayPal's retry is processed.
	release, ok := utils.TryLock(context.Background(), "paypal:event:"+ev.ID, 24*time.Hour)
	if !ok {
		return nil
	}
	if err := p.authorizePayPalOrder(context.Background(), ev.Resource); err != nil {
		release()
		return err
	}
	return nil
}

// chargeOf returns the amount and currency the provider was asked to charge, falling back to the
// ledger amount for rows created before the charge snapshot existed.
func chargeOf(tx *paymentModel.PaymentTransaction) (int64, string) {
	amount, currency := tx.ChargeAmount, strings.ToUpper(strings.TrimSpace(tx.ChargeCurrency))
	if amount <= 0 {
		amount = tx.Amount
	}
	if currency == "" {
		currency = strings.ToUpper(tx.Currency)
	}
	return amount, currency
}

// authorizePayPalOrder authorizes a buyer-approved PayPal order and records the authorized state
// in the ledger, the MySQL transaction and the booking lifecycle.
func (p *PaymentServices) authorizePayPalOrder(ctx context.Context, resource json.RawMessage) error {
	var res struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		PurchaseUnits []struct {
			Amount struct {
				CurrencyCode string `json:"currency_code"`
				Value        string `json:"value"`
			} `json:"amount"`
		} `json:"purchase_units"`
	}
	if err := json.Unmarshal(resource, &res); err != nil || strings.TrimSpace(res.ID) == "" {
		return errors.New("malformed paypal order payload")
	}

	tx, err := p.paymentByIntent(res.ID)
	if err != nil {
		return errors.New("payment not found for paypal order")
	}
	if tx.Provider != constants.PaymentProviderPayPal {
		return errors.New("paypal order does not belong to a paypal payment")
	}
	if tx.Status != constants.PaymentLifecycleCreated {
		return nil // already authorized/captured/refunded: nothing to do
	}
	if len(res.PurchaseUnits) != 1 {
		return errors.New("paypal order has an unexpected purchase unit count")
	}
	currency := strings.ToUpper(strings.TrimSpace(res.PurchaseUnits[0].Amount.CurrencyCode))
	amount, err := paypal.ParseAmount(res.PurchaseUnits[0].Amount.Value, currency)
	if err != nil {
		return errors.New("malformed paypal order amount")
	}
	// Compare against the snapshotted charge (settlement currency when the booking currency is
	// not PayPal-supported), never a freshly converted amount.
	expectedAmount, expectedCurrency := chargeOf(tx)
	if currency != expectedCurrency || amount != expectedAmount {
		return errors.New("paypal order amount does not match the payment")
	}

	release, err := lockPayment(ctx, tx.BookingID)
	if err != nil {
		return err
	}
	defer release()

	// re-read under the lock: another path may have advanced the payment meanwhile
	tx, err = p.paymentByIntent(res.ID)
	if err != nil {
		return errors.New("payment not found for paypal order")
	}
	if tx.Status != constants.PaymentLifecycleCreated {
		return nil
	}

	order, err := paypal.AuthorizeOrder(ctx, tx.IntentID)
	if err != nil {
		// PayPal rejects a second authorize for the same order; an existing authorization is success.
		existing, getErr := paypal.GetOrder(ctx, tx.IntentID)
		if getErr != nil || existing.AuthorizationID == "" {
			return err
		}
		order = existing
	}
	if order.AuthorizationID == "" {
		return errors.New("paypal order was not authorized")
	}
	expectedAmount, expectedCurrency = chargeOf(tx)
	if order.AmountMinor > 0 && (order.AmountMinor != expectedAmount || !strings.EqualFold(order.Currency, expectedCurrency)) {
		return errors.New("paypal authorized amount does not match the payment")
	}

	now := time.Now()
	if err := p.ledger(ctx, tx.IntentID, constants.PaymentEventAuthorized, tx.Provider,
		tx.PaymentMethod, tx.Amount, 0, tx.Currency, constants.TimelineActorSystem, "payment authorized"); err != nil {
		return err
	}
	if err := global.GVA_DB.Model(&paymentModel.PaymentTransaction{}).Where("id = ?", tx.ID).
		Updates(map[string]interface{}{
			"status": constants.PaymentLifecycleAuthorized, "updated_at": now,
		}).Error; err != nil {
		return err
	}
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).Where("id = ?", tx.BookingID).
		Updates(map[string]interface{}{"payment_status": constants.PaymentStatusAuthorized, "paid_at": &now}).Error
	return bookingSvc.Advance(ctx, tx.BookingID, constants.BookingStatusPaymentAuthorized, "payment authorized")
}
