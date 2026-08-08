// Package payment seeds payment transactions (across the capture/refund lifecycle), refunds
// (full + partial), payouts and per-booking commission records, all tied to the seeded bookings.
package payment

import (
	"context"
	"time"

	paymodel "github.com/mo7amed4522/tayara_api/model/payment"
	"github.com/mo7amed4522/tayara_api/service/system"
	"gorm.io/gorm"
)

const initOrderPayment = system.InitOrderExternal + 7

type paymentInit struct{}

func init() {
	system.RegisterInit(initOrderPayment, &paymentInit{})
}

func (p *paymentInit) InitializerName() string { return "demo_payment" }

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (p *paymentInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (p *paymentInit) TableCreated(ctx context.Context) bool                     { return true }

func (p *paymentInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var n int64
	return db.Model(&paymodel.PaymentTransaction{}).Count(&n).Error == nil && n > 0
}

func (p *paymentInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}
	now := time.Now()

	txs := []paymodel.PaymentTransaction{
		// Completed booking BK-2001 -> captured.
		{
			BookingID: bookingID(db, "BK-2001"), UserID: bookingUser(db, "BK-2001"), AgentID: bookingAgent(db, "BK-2001"),
			Provider: "stripe", PaymentMethod: "card", IntentID: "pi_1001_stripe", Amount: 307200,
			FeeBPS: 230, FeeAmount: 7360, CommissionPercent: 1000, CommissionAmount: 30720, AgentShare: 281920,
			Currency: "AED", Status: "captured", CapturedAt: tAdd(now, -36*24),
		},
		// Docs-pending booking BK-2002 -> captured.
		{
			BookingID: bookingID(db, "BK-2002"), UserID: bookingUser(db, "BK-2002"), AgentID: bookingAgent(db, "BK-2002"),
			Provider: "stripe", PaymentMethod: "apple_pay", IntentID: "pi_1002_stripe", Amount: 199500,
			FeeBPS: 230, FeeAmount: 4830, CommissionPercent: 1000, CommissionAmount: 19950, AgentShare: 185220,
			Currency: "AED", Status: "captured", CapturedAt: tAdd(now, -9*24),
		},
		// Awaiting-payment booking BK-2003 -> authorized only.
		{
			BookingID: bookingID(db, "BK-2003"), UserID: bookingUser(db, "BK-2003"), AgentID: bookingAgent(db, "BK-2003"),
			Provider: "stripe", PaymentMethod: "card", IntentID: "pi_1003_stripe", Amount: 210000,
			FeeBPS: 230, FeeAmount: 0, CommissionPercent: 500, CommissionAmount: 0, AgentShare: 0,
			Currency: "AED", Status: "authorized",
		},
		// Refunded booking BK-2006 -> refunded.
		{
			BookingID: bookingID(db, "BK-2006"), UserID: bookingUser(db, "BK-2006"), AgentID: bookingAgent(db, "BK-2006"),
			Provider: "stripe", PaymentMethod: "google_pay", IntentID: "pi_1006_stripe", Amount: 460000,
			FeeBPS: 230, FeeAmount: 10580, CommissionPercent: 1000, CommissionAmount: 44000, AgentShare: 405420,
			Currency: "AED", Status: "refunded", CapturedAt: tAdd(now, -16*24), RefundedAt: tAdd(now, -14*24),
		},
	}
	if err := db.Create(&txs).Error; err != nil {
		return ctx, err
	}

	// Refund rows (full on BK-2006, partial on BK-2002 to vary the data).
	tx2 := txByIntent(db, "pi_1002_stripe")
	refunds := []paymodel.Refund{
		{
			PaymentTransactionID: txByIntent(db, "pi_1006_stripe"), BookingID: bookingID(db, "BK-2006"),
			RefundType: "full", Amount: 460000, Reason: "Customer requested cancellation", ProviderRefundID: "re_6001",
			Status: "refunded", InitiatedBy: uintPtr(1), InitiatedAt: tAdd(now, -14*24),
		},
		{
			PaymentTransactionID: tx2, BookingID: bookingID(db, "BK-2002"),
			RefundType: "partial", Amount: 30000, Reason: "Partial accommodation adjustment", ProviderRefundID: "re_6002",
			Status: "pending", InitiatedBy: uintPtr(1), InitiatedAt: tAdd(now, -2*24),
		},
	}
	if err := db.Create(&refunds).Error; err != nil {
		return ctx, err
	}

	// Payouts: paid on BK-2001, pending on BK-2002.
	payouts := []paymodel.Payout{
		{
			BookingID: bookingID(db, "BK-2001"), AgentID: bookingAgent(db, "BK-2001"), AgentShare: 281920,
			PayoutStatus: "paid", Method: "manual", ReferenceNumber: "TRX-0001", AdminNotes: "Verified confirmation file",
			RequestedAt: tAdd(now, -20*24), PaidAt: tAdd(now, -18*24), PaidBy: uintPtr(1),
		},
		{
			BookingID: bookingID(db, "BK-2002"), AgentID: bookingAgent(db, "BK-2002"), AgentShare: 185220,
			PayoutStatus: "pending", Method: "manual", RequestedAt: tAdd(now, -4*24),
		},
	}
	if err := db.Create(&payouts).Error; err != nil {
		return ctx, err
	}

	// Commission records for captured bookings.
	commissions := []paymodel.CommissionRecord{
		{
			BookingID: bookingID(db, "BK-2001"), AgentID: bookingAgent(db, "BK-2001"), PlanType: "premium",
			GrossAmount: 320000, StripeFee: 7360, CommissionPercent: 1000, CommissionAmount: 30720, AgentShare: 281920,
			IsOverride: false, Status: "paid",
		},
		{
			BookingID: bookingID(db, "BK-2002"), AgentID: bookingAgent(db, "BK-2002"), PlanType: "premium",
			GrossAmount: 210000, StripeFee: 4830, CommissionPercent: 1000, CommissionAmount: 19950, AgentShare: 185220,
			IsOverride: false, Status: "captured",
		},
		{
			BookingID: bookingID(db, "BK-2006"), AgentID: bookingAgent(db, "BK-2006"), PlanType: "premium",
			GrossAmount: 460000, StripeFee: 10580, CommissionPercent: 1000, CommissionAmount: 44000, AgentShare: 405420,
			IsOverride: true, OverrideReason: "Adjusted on refund", OverrideBy: uintPtr(1), Status: "captured",
		},
	}
	if err := db.Create(&commissions).Error; err != nil {
		return ctx, err
	}
	return ctx, nil
}

// ---- id lookups ----

type bookingRef struct {
	ID      uint
	UserID  uint
	AgentID uint
}

func bookingRefByCode(db *gorm.DB, code string) bookingRef {
	var b bookingRef
	_ = db.Table("bookings").Select("id, user_id, agent_id").Where("booking_code = ?", code).Scan(&b).Error
	return b
}
func bookingID(db *gorm.DB, code string) uint    { return bookingRefByCode(db, code).ID }
func bookingUser(db *gorm.DB, code string) uint   { return bookingRefByCode(db, code).UserID }
func bookingAgent(db *gorm.DB, code string) uint  { return bookingRefByCode(db, code).AgentID }

func txByIntent(db *gorm.DB, intent string) uint {
	var r struct{ ID uint }
	_ = db.Table("payment_transactions").Select("id").Where("intent_id = ?", intent).Scan(&r).Error
	return r.ID
}

// small helpers (package-scoped)
func uintPtr(v uint) *uint { return &v }
func tAdd(t time.Time, d time.Duration) *time.Time {
	n := t.Add(d)
	return &n
}
