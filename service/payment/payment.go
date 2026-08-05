package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	paymentModel "github.com/mo7amed4522/tayara_api/model/payment"
	paymentMongo "github.com/mo7amed4522/tayara_api/model/payment/mongo"
	"github.com/mo7amed4522/tayara_api/model/payment/request"
	"github.com/mo7amed4522/tayara_api/model/payment/response"
	"github.com/mo7amed4522/tayara_api/service/booking"
	"github.com/mo7amed4522/tayara_api/service/engage"
	"github.com/mo7amed4522/tayara_api/service/payment/providers"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/mo7amed4522/tayara_api/utils/stripe"
	"gorm.io/gorm"
)

type PaymentServices struct{}

var (
	paymentRepoCopy = paymentMongo.LedgerRepo{}
	bookingSvc      = booking.BookingServices{}
)

// --- lifecycle lock: the payment lifecycle is the most sensitive structure. Every mutation of
// the Mongo ledger and MySQL payment rows happens under this per-booking red lock (Redis-backed,
// falls back to in-process). A second concurrent op on the same booking is rejected — no lost
// lifecycle event, no double capture, no missed refund.

func lockPayment(ctx context.Context, bookingID uint) (release func(), err error) {
	release, ok := utils.TryLock(ctx, fmt.Sprintf("payment:booking:%d", bookingID), time.Minute)
	if !ok {
		return nil, errors.New("payment lifecycle is locked by another operation; retry")
	}
	return release, nil
}

// CreatePayment opens a payment for a payable booking. Cards never touch our server: for Stripe
// the client completes the PaymentIntent with Stripe.js/Elements; we return only the client secret.
func (p *PaymentServices) CreatePayment(ctx context.Context, userID uint, req request.CreatePaymentReq) (*response.CreatePaymentResp, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	prov := providers.Get(req.PaymentMethod)
	if prov == nil {
		return nil, errors.New("unsupported payment method")
	}
	if !prov.Enabled() {
		return nil, errors.New(req.PaymentMethod + " provider is not enabled")
	}
	var bk bookingModel.Booking
	if err := global.GVA_DB.Where("id = ? AND user_id = ?", req.BookingID, userID).First(&bk).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	if bk.BookingStatus != constants.BookingStatusPriceConfirmed && bk.BookingStatus != constants.BookingStatusAwaitingPayment {
		return nil, errors.New("booking is not payable in its current state")
	}
	if _, err := p.paymentByBooking(req.BookingID); err == nil {
		return nil, errors.New("a payment already exists for this booking")
	}

	amount := netAmount(bk)
	if amount <= 0 {
		return nil, errors.New("payable amount must be greater than zero")
	}

	release, err := lockPayment(ctx, bk.ID)
	if err != nil {
		return nil, err
	}
	defer release()

	result, err := prov.CreatePayment(ctx, providers.CreatePaymentParams{
		BookingID: bk.ID, BookingCode: bk.BookingCode, Amount: amount, Currency: bk.Currency,
	})
	if err != nil {
		return nil, err
	}
	// Mongo ledger first (source of truth), then the authoritative MySQL row.
	if err := p.ledger(ctx, result.IntentID, constants.PaymentEventCreated, prov.Name(),
		req.PaymentMethod, amount, 0, bk.Currency, constants.TimelineActorUser, "payment created"); err != nil {
		return nil, err
	}
	tx := &paymentModel.PaymentTransaction{
		BookingID: bk.ID, UserID: userID, AgentID: bk.AgentID,
		Provider: prov.Name(), PaymentMethod: req.PaymentMethod,
		IntentID: result.IntentID, ClientSecret: result.ClientSecret,
		Amount: amount, CommissionPercent: bk.CommissionPercent, Currency: bk.Currency,
		Status: constants.PaymentLifecycleCreated,
	}
	if err := global.GVA_DB.Create(tx).Error; err != nil {
		return nil, err
	}
	return &response.CreatePaymentResp{
		PaymentID: result.IntentID, Provider: prov.Name(), Method: req.PaymentMethod,
		IntentID: result.IntentID, ClientSecret: result.ClientSecret,
		Amount: amount, Currency: bk.Currency,
	}, nil
}

// HandleStripeWebhook processes a signed Stripe webhook. On payment_intent.requires_capture it
// advances the payment to authorized (ledger + MySQL) and moves the booking lifecycle forward.
func (p *PaymentServices) HandleStripeWebhook(payload []byte, signature string) error {
	ev, err := stripe.ConstructWebhookEvent(payload, signature)
	if err != nil {
		return errors.New("invalid webhook signature")
	}
	if ev.Type != "payment_intent.requires_capture" {
		return nil
	}
	var it struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(ev.Data.Raw, &it); err != nil || it.ID == "" {
		return errors.New("malformed webhook payload")
	}
	tx, err := p.paymentByIntent(it.ID)
	if err != nil {
		return errors.New("payment intent not found")
	}
	release, err := lockPayment(context.Background(), tx.BookingID)
	if err != nil {
		return err
	}
	defer release()

	now := time.Now()
	if err := p.ledger(context.Background(), tx.IntentID, constants.PaymentEventAuthorized, tx.Provider,
		tx.PaymentMethod, tx.Amount, 0, tx.Currency, constants.TimelineActorSystem, "payment authorized"); err != nil {
		return err
	}
	if err := global.GVA_DB.Model(&paymentModel.PaymentTransaction{}).Where("id = ?", tx.ID).
		Updates(map[string]interface{}{"status": constants.PaymentLifecycleAuthorized, "updated_at": now}).Error; err != nil {
		return err
	}
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).Where("id = ?", tx.BookingID).
		Updates(map[string]interface{}{"payment_status": constants.PaymentStatusAuthorized, "paid_at": &now}).Error
	return bookingSvc.Advance(context.Background(), tx.BookingID, constants.BookingStatusPaymentAuthorized, "payment authorized")
}

// Capture captures an authorized payment, computes fees/commission/agent share, and records all
// of it to the ledger + MySQL. Bookkeeping (fees) is never skipped — it is part of the lifecycle.
func (p *PaymentServices) Capture(ctx context.Context, adminID, bookingID uint) (*response.PaymentDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	tx, err := p.paymentByBooking(bookingID)
	if err != nil {
		return nil, errors.New("payment not found")
	}
	if tx.Status != constants.PaymentLifecycleAuthorized {
		return nil, errors.New("payment must be authorized before capture")
	}
	release, err := lockPayment(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	defer release()

	prov := providers.Get(tx.PaymentMethod)
	if prov == nil {
		return nil, errors.New("unsupported payment method")
	}
	if err := prov.Capture(ctx, tx.IntentID, nil); err != nil {
		return nil, err
	}
	// reconcile the bookkeeping for this payment
	var bk bookingModel.Booking
	_ = global.GVA_DB.First(&bk, bookingID).Error
	net := tx.Amount
	feeBPS := int64(global.GVA_CONFIG.Stripe.FeeBPS)
	if feeBPS <= 0 {
		feeBPS = 0
	}
	feeAmount := utils.MoneyPercentOf(net, feeBPS)
	commissionAmount := utils.MoneyPercentOf(net, tx.CommissionPercent)
	agentShare := net - feeAmount - commissionAmount
	now := time.Now()

	if err := p.ledger(ctx, tx.IntentID, constants.PaymentEventCaptured, tx.Provider,
		tx.PaymentMethod, net, feeAmount, tx.Currency, constants.TimelineActorAdmin, "payment captured"); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"status": constants.PaymentLifecycleCaptured, "fee_bps": feeBPS, "fee_amount": feeAmount,
		"commission_amount": commissionAmount, "agent_share": agentShare, "captured_at": &now, "updated_at": now,
	}
	if err := global.GVA_DB.Model(&paymentModel.PaymentTransaction{}).Where("id = ?", tx.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).Where("id = ?", bookingID).
		Updates(map[string]interface{}{
			"payment_status": constants.PaymentStatusCaptured, "stripe_fee": feeAmount,
			"commission_amount": commissionAmount, "agent_share": agentShare, "captured_at": &now,
		}).Error
	// reflect the captured amount on the paying user's lifetime spend (identity aggregate)
	_ = global.GVA_DB.Model(&bxapi.BxsUserProfile{}).Where("id = ?", tx.UserID).
		UpdateColumn("total_spend", gorm.Expr("total_spend + ?", net)).Error
	// audit-complete commission snapshot (financial record per booking)
	tx.FeeAmount = feeAmount
	tx.CommissionAmount = commissionAmount
	tx.AgentShare = agentShare
	_ = p.createCommissionRecord(tx, constants.CommissionPlanCustom)

	engage.NotificationServices{}.Emit(constants.NotifEventPriceConfirmed, constants.NotifRecipientUser,
		tx.UserID, constants.NotifChannelPush, "Payment received", "Your payment was captured")
	_ = bookingSvc.Advance(ctx, bookingID, constants.BookingStatusPaidConfirmed, "payment captured")
	return p.paymentDetail(tx.BookingID, tx.ID)
}

// Refund refunds a captured payment (full or partial) and moves the bookkeeping/lifecycle to refunded.
func (p *PaymentServices) Refund(ctx context.Context, adminID uint, req request.RefundReq) (*response.PaymentDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	tx, err := p.paymentByBooking(req.BookingID)
	if err != nil {
		return nil, errors.New("payment not found")
	}
	if tx.Status != constants.PaymentLifecycleCaptured {
		return nil, errors.New("only captured payments can be refunded")
	}
	refundAmount := tx.Amount
	if req.RefundType == constants.RefundTypePartial {
		if req.Amount <= 0 || req.Amount > tx.Amount {
			return nil, errors.New("partial refund amount must be between 0 and the captured amount")
		}
		refundAmount = req.Amount
	}
	release, err := lockPayment(ctx, req.BookingID)
	if err != nil {
		return nil, err
	}
	defer release()

	prov := providers.Get(tx.PaymentMethod)
	if prov == nil {
		return nil, errors.New("unsupported payment method")
	}
	providerRefundID, err := prov.Refund(ctx, tx.IntentID, &refundAmount, req.Reason)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	refundRow := &paymentModel.Refund{
		PaymentTransactionID: tx.ID, BookingID: req.BookingID, RefundType: req.RefundType,
		Amount: refundAmount, Reason: req.Reason, ProviderRefundID: providerRefundID,
		Status: constants.RefundStatusRefunded, InitiatedBy: &adminID, InitiatedAt: &now,
	}
	newStatus := constants.PaymentLifecycleRefunded
	if refundAmount < tx.Amount {
		newStatus = constants.PaymentLifecyclePartiallyRefunded
	}

	if err := p.ledger(ctx, tx.IntentID, constants.PaymentEventRefunded, tx.Provider,
		tx.PaymentMethod, refundAmount, 0, tx.Currency, constants.TimelineActorAdmin, "refund issued"); err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Create(&refundRow).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"status": newStatus, "refunded_at": &now, "updated_at": now}
	if err := global.GVA_DB.Model(&paymentModel.PaymentTransaction{}).Where("id = ?", tx.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).Where("id = ?", req.BookingID).
		Updates(map[string]interface{}{"payment_status": newStatus}).Error
	engage.NotificationServices{}.Emit(constants.NotifEventCancellation, constants.NotifRecipientUser,
		tx.UserID, constants.NotifChannelPush, "Refund issued", "A refund was processed")
	if newStatus == constants.PaymentLifecycleRefunded {
		_ = bookingSvc.Advance(ctx, req.BookingID, constants.BookingStatusRefunded, "payment refunded")
	}
	return p.paymentDetail(req.BookingID, tx.ID)
}

// --- read side ---

func (p *PaymentServices) paymentByBooking(bookingID uint) (*paymentModel.PaymentTransaction, error) {
	var tx paymentModel.PaymentTransaction
	err := global.GVA_DB.Where("booking_id = ?", bookingID).First(&tx).Error
	return &tx, err
}

func (p *PaymentServices) paymentByIntent(intentID string) (*paymentModel.PaymentTransaction, error) {
	var tx paymentModel.PaymentTransaction
	err := global.GVA_DB.Where("intent_id = ?", intentID).First(&tx).Error
	return &tx, err
}

func (p *PaymentServices) paymentDetail(bookingID, txID uint) (*response.PaymentDetail, error) {
	var tx paymentModel.PaymentTransaction
	if err := global.GVA_DB.First(&tx, txID).Error; err != nil {
		return nil, err
	}
	var refunds []paymentModel.Refund
	_ = global.GVA_DB.Where("payment_transaction_id = ?", txID).Order("created_at desc").Find(&refunds).Error
	return &response.PaymentDetail{PaymentTransaction: tx, Refunds: refunds}, nil
}

// ledger records an event to the Mongo payment event log and bumps the current state. It is the
// source of truth for the lifecycle; the MySQL row is the authoritative persistence alongside it.
func (p *PaymentServices) ledger(ctx context.Context, intentID, evType, provider, method string, amount, fee int64, currency, actor, note string) error {
	if err := paymentRepoCopy.AppendEvent(ctx, &paymentMongo.LedgerEvent{
		EventID: utils.NextCode(constants.CodePrefixPayment), PaymentID: intentID, Type: evType,
		Provider: provider, Method: method, Amount: amount, FeeAmount: fee, Currency: currency,
		Actor: actor, Note: note, At: time.Now(),
	}); err != nil {
		return err
	}
	st := &paymentMongo.PaymentState{
		PaymentID: intentID, Provider: provider, Method: method,
		Status: evType, Amount: amount, FeeAmount: fee, Currency: currency, UpdatedAt: time.Now(),
	}
	return paymentRepoCopy.UpsertState(ctx, st)
}

// netAmount is the payable amount for a booking.
func netAmount(bk bookingModel.Booking) int64 {
	return bk.AmountGross - bk.DiscountAmount
}
