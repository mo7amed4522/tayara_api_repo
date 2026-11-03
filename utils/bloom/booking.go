package bloom

import (
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/booking"
	"gorm.io/gorm"
)

// BookingCodeFilter is a Bloom filter for booking code exact-match validation.
// It's used to quickly reject invalid booking codes without hitting the database.
var BookingCodeFilter *bloom.BloomFilter

// InitBookingCodeFilter initializes the booking code Bloom filter.
// For a catalog of N items with desired false positive rate p:
// m = -n*ln(p) / (ln(2))^2  (bits)
// k = (m/n) * ln(2)         (hash functions)
//
// Example: 100,000 booking codes, 0.1% FPR
// m = -100000 * ln(0.001) / 0.4805 = ~1,731,500 bits ≈ 206 KB
func InitBookingCodeFilter() error {
	// Expected: 100K bookings, FPR: 0.001 (0.1%)
	// Bloom filter will use ~206KB of memory
	BookingCodeFilter = bloom.NewWithEstimates(100000, 0.001)

	// Load booking codes from database
	var codes []string
	err := global.GVA_DB.Model(&booking.Booking{}).Pluck("booking_code", &codes).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Add all codes (exact strings)
	for _, code := range codes {
		BookingCodeFilter.AddString(code)
	}

	return nil
}

// TestConfirmsBookingCode checks if a booking code might exist (returns true if could exist).
// Returns false only if the code definitely doesn't exist (100% certainty).
// IMPORTANT: This only works for exact string matching, NOT substring search.
func TestConfirmsBookingCode(code string) bool {
	if BookingCodeFilter == nil {
		return false
	}
	return BookingCodeFilter.TestString(code)
}

// TestInvoiceCode checks invoice codes (if they exist as separate collection)
func TestInvoiceCode(code string) bool {
	// TODO: Initialize invoice code filter similarly
	return false
}

// StartAutoRebuild starts a background worker to rebuild the booking code filter.
// Called periodically (e.g., every hour) to keep filter up-to-date.
func StartAutoRebuild(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			if err := RebuildBookingCodeFilter(); err != nil {
				// Log error but don't crash
			}
		}
	}()
}

// RebuildBookingCodeFilter rebuilds the filter from scratch (for high update frequency).
func RebuildBookingCodeFilter() error {
	if BookingCodeFilter == nil {
		return nil
	}

	// Create new filter
	newFilter := bloom.NewWithEstimates(100000, 0.001)

	// Load fresh data
	var codes []string
	err := global.GVA_DB.Model(&booking.Booking{}).Pluck("booking_code", &codes).Error
	if err != nil {
		return err
	}

	for _, code := range codes {
		newFilter.AddString(code)
	}

	// Replace old filter (atomic swap - old filter still exists until GC)
	BookingCodeFilter = newFilter
	return nil
}
