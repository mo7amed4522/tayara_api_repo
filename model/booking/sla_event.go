package booking

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// SLAEvent tracks an active SLA timer for a booking (agent response or user payment window).
// The cron job finds breached events and applies the configured auto action (auto_cancel,
// offer_expiry, reminder). Created on entry to a state, cleared/expired on exit.
type SLAEvent struct {
	ID              uint       `json:"id" gorm:"primarykey;comment:ID"`
	BookingID       uint       `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	SLAType         string     `json:"slaType" form:"slaType" gorm:"size:32;index;comment:agent_response|user_payment"`
	DeadlineAt      time.Time  `json:"deadlineAt" form:"deadlineAt" gorm:"comment:SLA deadline"`
	Breached        bool       `json:"breached" form:"breached" gorm:"default:false;comment:Marked breached"`
	AutoActionTaken string     `json:"autoActionTaken" form:"autoActionTaken" gorm:"size:32;comment:auto_cancel|offer_expiry|reminder"`
	TriggeredAt     *time.Time `json:"triggeredAt" form:"triggeredAt" gorm:"comment:When the auto action ran"`
	Active          bool       `json:"active" form:"active" gorm:"default:true;comment:Timer still live"`
	global.GVA_MODEL
}

func (SLAEvent) TableName() string { return "sla_events" }
