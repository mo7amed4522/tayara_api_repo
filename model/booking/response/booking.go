package response

import "github.com/mo7amed4522/tayara_api/model/booking"

// BookingDetail a booking header plus its travelers, offers, documents, and status timeline.
type BookingDetail struct {
	booking.Booking
	Travelers     []booking.BookingTraveler      `json:"travelers"`
	Offers        []booking.BookingOffer         `json:"offers"`
	Documents     []booking.BookingDocument      `json:"documents"`
	StatusHistory []booking.BookingStatusHistory `json:"statusHistory"`
}
