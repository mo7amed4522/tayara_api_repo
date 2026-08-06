package booking

import (
	"context"
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	"go.uber.org/zap"
)

// BookingServices reservation-hold methods. The "Secure Price Now" reservation window is the time
// between price confirmation and the payment deadline — the offer's ExpiresAt is mirrored here on
// the booking as HoldExpiresAt so the dashboard can show "hold expires at" and a cron can release
// expired holds automatically.

// SetHold opens/extends the reservation hold on a booking (called when price is confirmed or the
// hold is renewed). Returns the new hold expiry.
func (b *BookingServices) SetHold(bookingID uint, hours int) (*time.Time, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if hours <= 0 {
		return nil, errors.New("hold hours must be positive")
	}
	var bk bookingModel.Booking
	if err := global.GVA_DB.First(&bk, bookingID).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	if bk.BookingStatus != constants.BookingStatusPriceConfirmed &&
		bk.BookingStatus != constants.BookingStatusAwaitingPayment {
		return nil, errors.New("hold can only be set while awaiting payment")
	}
	expires := time.Now().Add(time.Duration(hours) * time.Hour)
	if err := global.GVA_DB.Model(&bookingModel.Booking{}).Where("id = ?", bookingID).
		Update("hold_expires_at", &expires).Error; err != nil {
		return nil, err
	}
	return &expires, nil
}

// HoldStatus reports whether the reservation hold on a booking is live or expired.
func (b *BookingServices) HoldStatus(bookingID uint) (live bool, expiresAt *time.Time, err error) {
	if global.GVA_DB == nil {
		return false, nil, errors.New("db not init")
	}
	var bk bookingModel.Booking
	if err := global.GVA_DB.First(&bk, bookingID).Error; err != nil {
		return false, nil, errors.New("booking not found")
	}
	if bk.HoldExpiresAt == nil {
		return false, nil, nil
	}
	return time.Now().Before(*bk.HoldExpiresAt), bk.HoldExpiresAt, nil
}

// ReleaseExpiredHolds releases reservation holds that have expired: it cancels bookings still in
// awaiting_payment whose hold window has passed. Called by the timer cron each tick.
func (b *BookingServices) ReleaseExpiredHolds() error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	now := time.Now()
	var bookings []bookingModel.Booking
	if err := global.GVA_DB.Where("booking_status = ? AND hold_expires_at IS NOT NULL AND hold_expires_at <= ?",
		constants.BookingStatusAwaitingPayment, now).Find(&bookings).Error; err != nil {
		return err
	}
	for _, bk := range bookings {
		// move to price_expired (offer removed) so the user must re-apply
		if err := b.Advance(context.Background(), bk.ID, constants.BookingStatusPriceExpired, "reservation hold expired"); err != nil {
			global.GVA_LOG.Warn("failed to release expired hold", zap.Uint("booking_id", bk.ID), zap.Error(err))
		}
	}
	return nil
}
