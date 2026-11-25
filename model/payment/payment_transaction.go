package payment

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// PaymentTransaction the authoritative MySQL record of a payment lifecycle, written as the
// provider + Mongo ledger advance. Card details never appear here — only provider tokens.
type PaymentTransaction struct {
	ID                uint   `json:"id" gorm:"primarykey;comment:ID"`
	BookingID         uint   `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	UserID            uint   `json:"userId" form:"userId" gorm:"index;comment:Paying user ID"`
	AgentID           uint   `json:"agentId" form:"agentId" gorm:"index;comment:Agent ID"`
	Provider          string `json:"provider" form:"provider" gorm:"size:16;index;comment:stripe|tabby|tamara|paypal"`
	PaymentMethod     string `json:"paymentMethod" form:"paymentMethod" gorm:"size:32;index;comment:card|apple_pay|google_pay|tabby|tamara|paypal"`
	IntentID          string `json:"intentId" form:"intentId" gorm:"size:128;uniqueIndex;comment:Provider intent/session ID"`
	ClientSecret      string `json:"clientSecret" form:"clientSecret" gorm:"size:500;comment:Stripe client secret (session only)"`
	RedirectURL       string `json:"redirectUrl" form:"redirectUrl" gorm:"size:1000;comment:Provider approval/redirect URL (PayPal)"`
	Amount            int64  `json:"amount" form:"amount" gorm:"default:0;comment:Captured amount, minor units"`
	FeeBPS            int64  `json:"feeBPS" form:"feeBPS" gorm:"default:0;comment:Platform fee, basis points (snapshot)"`
	FeeAmount         int64  `json:"feeAmount" form:"feeAmount" gorm:"default:0;comment:Fee amount, minor units"`
	CommissionPercent int64  `json:"commissionPercent" form:"commissionPercent" gorm:"default:0;comment:Agent commission, basis points (snapshot)"`
	CommissionAmount  int64  `json:"commissionAmount" form:"commissionAmount" gorm:"default:0;comment:Commission amount, minor units"`
	AgentShare        int64  `json:"agentShare" form:"agentShare" gorm:"default:0;comment:Agent share, minor units"`
	Currency          string `json:"currency" form:"currency" gorm:"size:8;default:AED;comment:Booking (ledger) currency"`
	// ChargeAmount/ChargeCurrency are what the provider was asked to charge. They differ from
	// Amount/Currency only when the provider needs a settlement currency (PayPal + AED).
	ChargeAmount   int64  `json:"chargeAmount" form:"chargeAmount" gorm:"default:0;comment:Provider charge amount, minor units"`
	ChargeCurrency string `json:"chargeCurrency" form:"chargeCurrency" gorm:"size:8;comment:Provider charge currency"`
	// FXRatePPM is the booking->charge rate snapshot in parts per million (1e6 == 1:1).
	FXRatePPM  int64      `json:"fxRatePpm" form:"fxRatePpm" gorm:"default:0;comment:Booking->charge FX rate, ppm"`
	Status     string     `json:"status" form:"status" gorm:"size:32;index;default:authorized;comment:authorized|captured|refunded|partially_refunded|failed"`
	CapturedAt *time.Time `json:"capturedAt" form:"capturedAt" gorm:"comment:Captured at"`
	RefundedAt *time.Time `json:"refundedAt" form:"refundedAt" gorm:"comment:Fully refunded at"`
	global.GVA_MODEL
}

func (PaymentTransaction) TableName() string { return "payment_transactions" }
