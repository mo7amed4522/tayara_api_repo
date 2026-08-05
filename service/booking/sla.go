package booking

import (
	"strconv"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/base"
	"github.com/mo7amed4522/tayara_api/model/booking"
	"gorm.io/gorm"
)

// slaHours reads a runtime System Setting (seeded in Phase 0) and falls back to a sane default.
// Keys live in base_system_setting; never hardcode business numbers.
func slaHours(key string, fallback int) int {
	if global.GVA_DB == nil {
		return fallback
	}
	var s base.SystemSetting
	if err := global.GVA_DB.Where("`key` = ? AND is_active = ?", key, true).First(&s).Error; err != nil {
		return fallback
	}
	v, err := strconv.Atoi(s.Value)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

// startSLAFor opens the SLA timer a booking needs when it enters `status` (single active timer).
func startSLAFor(tx *gorm.DB, bookingID uint, status string) error {
	var slaType, key string
	var hours int
	switch status {
	case constants.BookingStatusRequested, constants.BookingStatusAgentReviewing:
		slaType = constants.SLATypeAgentResponse
		key = "agent_response_hours"
		hours = slaHours(key, 2)
	case constants.BookingStatusPriceConfirmed:
		slaType = constants.SLATypeUserPayment
		key = "user_payment_hours"
		hours = slaHours(key, 2)
	default:
		return nil // no timer needed in this state
	}
	return tx.Create(&booking.SLAEvent{
		BookingID:  bookingID,
		SLAType:    slaType,
		DeadlineAt: utilsDeadlineAfter(hours),
		Active:     true,
	}).Error
}

// clearActiveSLA deactivates all live timers for a booking (moving state cancels the old timer).
func clearActiveSLA(tx *gorm.DB, bookingID uint) error {
	return tx.Model(&booking.SLAEvent{}).
		Where("booking_id = ? AND active = ?", bookingID, true).
		Update("active", false).Error
}

// utilsDeadlineAfter is a tiny local alias so SLA math stays pure and testable.
func utilsDeadlineAfter(hours int) time.Time {
	if hours <= 0 {
		return time.Now()
	}
	return time.Now().Add(time.Duration(hours) * time.Hour)
}
