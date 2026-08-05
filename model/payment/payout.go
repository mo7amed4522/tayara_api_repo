package payment

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// Payout a manual admin payout for an eligible, captured booking. The agent share derives from
// the payment transaction (net − fee − commission). MVP method is manual bank transfer; payoff
// is locked behind eligibility gates until admin confirms.
type Payout struct {
	ID              uint       `json:"id" gorm:"primarykey;comment:ID"`
	BookingID       uint       `json:"bookingId" form:"bookingId" gorm:"uniqueIndex;comment:Booking ID"`
	AgentID         uint       `json:"agentId" form:"agentId" gorm:"index;comment:Agent ID"`
	AgentShare      int64      `json:"agentShare" form:"agentShare" gorm:"default:0;comment:Agent share, minor units"`
	PayoutStatus    string     `json:"payoutStatus" form:"payoutStatus" gorm:"size:32;index;default:not_eligible;comment:not_eligible|pending|paid|failed"`
	Method          string     `json:"method" form:"method" gorm:"size:32;default:manual;comment:manual|stripe_transfer"`
	ReferenceNumber string     `json:"referenceNumber" form:"referenceNumber" gorm:"size:128;comment:Bank/transfer reference"`
	AdminNotes      string     `json:"adminNotes" form:"adminNotes" gorm:"size:1000;comment:Admin notes"`
	RequestedAt     *time.Time `json:"requestedAt" form:"requestedAt" gorm:"comment:Requested at"`
	PaidAt          *time.Time `json:"paidAt" form:"paidAt" gorm:"comment:Paid at"`
	PaidBy          *uint      `json:"paidBy" form:"paidBy" gorm:"index;comment:Acting admin ID"`
	global.GVA_MODEL
}

func (Payout) TableName() string { return "payouts" }
