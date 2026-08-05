package payment

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// Refund a full/partial refund raised by an admin post-capture. Amount never exceeds captured.
type Refund struct {
	ID                   uint       `json:"id" gorm:"primarykey;comment:ID"`
	PaymentTransactionID uint       `json:"paymentTransactionId" form:"paymentTransactionId" gorm:"index;comment:Payment transaction ID"`
	BookingID            uint       `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	RefundType           string     `json:"refundType" form:"refundType" gorm:"size:16;comment:full|partial"`
	Amount               int64      `json:"amount" form:"amount" gorm:"default:0;comment:Refund amount, minor units"`
	Reason               string     `json:"reason" form:"reason" gorm:"size:1000;comment:Reason"`
	ProviderRefundID     string     `json:"providerRefundId" form:"providerRefundId" gorm:"size:128;comment:Provider refund ID"`
	Status               string     `json:"status" form:"status" gorm:"size:32;index;default:pending;comment:pending|succeeded|failed"`
	InitiatedBy          *uint      `json:"initiatedBy" form:"initiatedBy" gorm:"index;comment:Initiating admin ID"`
	InitiatedAt          *time.Time `json:"initiatedAt" form:"initiatedAt" gorm:"comment:Initiated at"`
	global.GVA_MODEL
}

func (Refund) TableName() string { return "refunds" }
