package payment

import (
	"github.com/mo7amed4522/tayara_api/global"
)

// CommissionRecord the per-booking commission calculation + any admin override (audited). Written
// at capture and updated if admin overrides commission; overrides carry the reason + acting admin.
type CommissionRecord struct {
	ID                uint   `json:"id" gorm:"primarykey;comment:ID"`
	BookingID         uint   `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	AgentID           uint   `json:"agentId" form:"agentId" gorm:"index;comment:Agent ID"`
	PlanType          string `json:"planType" form:"planType" gorm:"size:16;index;comment:basic|premium|custom"`
	GrossAmount       int64  `json:"grossAmount" form:"grossAmount" gorm:"default:0;comment:Gross amount, minor units"`
	StripeFee         int64  `json:"stripeFee" form:"stripeFee" gorm:"default:0;comment:Stripe fee, minor units"`
	CommissionPercent int64  `json:"commissionPercent" form:"commissionPercent" gorm:"default:0;comment:Commission, basis points"`
	CommissionAmount  int64  `json:"commissionAmount" form:"commissionAmount" gorm:"default:0;comment:Commission amount, minor units"`
	AgentShare        int64  `json:"agentShare" form:"agentShare" gorm:"default:0;comment:Agent share, minor units"`
	IsOverride        bool   `json:"isOverride" form:"isOverride" gorm:"default:false;comment:Admin overrode commission"`
	OverrideReason    string `json:"overrideReason" form:"overrideReason" gorm:"size:1000;comment:Override reason"`
	OverrideBy        *uint  `json:"overrideBy" form:"overrideBy" gorm:"index;comment:Overriding admin ID"`
	Status            string `json:"status" form:"status" gorm:"size:16;index;default:captured;comment:captured|paid"`
	global.GVA_MODEL
}

func (CommissionRecord) TableName() string { return "commission_records" }
