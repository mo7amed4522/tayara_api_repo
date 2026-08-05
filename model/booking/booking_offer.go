package booking

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
)

// BookingOffer a final price offer the agent submits for a booking. The latest active offer
// drives the booking to price_confirmed; expired/declined offers are tracked for audit.
type BookingOffer struct {
	ID                        uint           `json:"id" gorm:"primarykey;comment:ID"`
	BookingID                 uint           `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	AgentID                   uint           `json:"agentId" form:"agentId" gorm:"index;comment:Submitting agent ID"`
	FinalPrice                int64          `json:"finalPrice" form:"finalPrice" gorm:"default:0;comment:Final price, minor units"`
	AvailableDates            common.JSONMap `json:"availableDates" form:"availableDates" gorm:"comment:Available dates (range/JSON)"`
	Notes                     string         `json:"notes" form:"notes" gorm:"size:1000;comment:Agent notes"`
	Status                    string         `json:"status" form:"status" gorm:"size:32;index;default:active;comment:active|accepted|expired|cancelled"`
	RejectedReason            string         `json:"rejectedReason" form:"rejectedReason" gorm:"size:1000;comment:Rejection reason"`
	AlternativeDatesSuggested common.JSONMap `json:"alternativeDatesSuggested" form:"alternativeDatesSuggested" gorm:"comment:Alternative dates suggested"`
	OfferedAt                 *time.Time     `json:"offeredAt" form:"offeredAt" gorm:"comment:Offered at"`
	RespondedAt               *time.Time     `json:"respondedAt" form:"respondedAt" gorm:"comment:Responded at"`
	ExpiresAt                 *time.Time     `json:"expiresAt" form:"expiresAt" gorm:"comment:Offer expiry (user pay window)"`
	global.GVA_MODEL
}

func (BookingOffer) TableName() string { return "booking_offers" }
