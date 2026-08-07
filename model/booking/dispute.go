package booking

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// Dispute a booking dispute (raised by user/agent/admin, resolved by admin).
type Dispute struct {
	ID         uint       `json:"id" gorm:"primarykey;comment:ID"`
	BookingID  uint       `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	RaisedBy   string     `json:"raisedBy" form:"raisedBy" gorm:"size:16;index;comment:user|agent|admin"`
	RaisedByID uint       `json:"raisedById" form:"raisedById" gorm:"index;comment:Raiser ID"`
	Status     string     `json:"status" form:"status" gorm:"size:32;index;default:open;comment:open|resolved|closed"`
	Details    string     `json:"details" form:"details" gorm:"size:2000;comment:Dispute details"`
	Resolution string     `json:"resolution" form:"resolution" gorm:"size:2000;comment:Resolution note"`
	ResolvedBy *uint      `json:"resolvedBy" form:"resolvedBy" gorm:"index;comment:Resolving admin ID"`
	ResolvedAt *time.Time `json:"resolvedAt" form:"resolvedAt" gorm:"comment:Resolved at"`
	global.GVA_MODEL
}

func (Dispute) TableName() string { return "disputes" }
