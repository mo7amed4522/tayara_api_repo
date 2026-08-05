package booking

import "github.com/mo7amed4522/tayara_api/global"

// BookingStatusHistory an append-only timeline of every lifecycle transition on a booking.
// Written by the service on each status change; never mutated in place.
type BookingStatusHistory struct {
	ID         uint   `json:"id" gorm:"primarykey;comment:ID"`
	BookingID  uint   `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	ActorType  string `json:"actorType" form:"actorType" gorm:"size:16;index;comment:system|admin|agent|user"`
	ActorID    uint   `json:"actorId" form:"actorId" gorm:"index;comment:Actor ID"`
	FromStatus string `json:"fromStatus" form:"fromStatus" gorm:"size:32;comment:Previous status"`
	ToStatus   string `json:"toStatus" form:"toStatus" gorm:"size:32;comment:New status"`
	Notes      string `json:"notes" form:"notes" gorm:"size:1000;comment:Transition notes"`
	global.GVA_MODEL
}

func (BookingStatusHistory) TableName() string { return "booking_status_history" }
