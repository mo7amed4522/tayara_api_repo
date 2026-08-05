package request

import (
	"github.com/mo7amed4522/tayara_api/model/common"
	commonReq "github.com/mo7amed4522/tayara_api/model/common/request"
)

// TravelerReq a traveler entry when creating a booking.
type TravelerReq struct {
	FullName       string `json:"fullName"`
	DOB            string `json:"dob"` // layout 2006-01-02
	NationalityID  uint   `json:"nationalityId"`
	PassportNumber string `json:"passportNumber"`
	PassportExpiry string `json:"passportExpiry"` // layout 2006-01-02
	AdultOrChild   string `json:"adultOrChild"`
	Email          string `json:"email"`
}

// CreateBookingReq user creates a package or custom-trip booking.
type CreateBookingReq struct {
	PackageID           *uint          `json:"packageId"`
	CustomTripRequestID *uint          `json:"customTripRequestId"`
	Type                string         `json:"type"` // package | custom_trip
	StartDate           string         `json:"startDate"`
	EndDate             string         `json:"endDate"`
	TravelerCount       int            `json:"travelerCount"`
	HotelChoiceID       *uint          `json:"hotelChoiceId"`
	SelectedActivities  common.JSONMap `json:"selectedActivities"`
	Notes               string         `json:"notes"`
	Travelers           []TravelerReq  `json:"travelers"`
}

// SubmitOfferReq agent submits the final price offer for a booking.
type SubmitOfferReq struct {
	BookingID      uint           `json:"bookingId"`
	FinalPrice     int64          `json:"finalPrice"` // minor units
	AvailableDates common.JSONMap `json:"availableDates"`
	Notes          string         `json:"notes"`
}

// RejectBookingReq agent rejects/modifies a booking with alternative dates + explanation.
type RejectBookingReq struct {
	BookingID                 uint           `json:"bookingId"`
	Reason                    string         `json:"reason"`
	AlternativeDatesSuggested common.JSONMap `json:"alternativeDatesSuggested"`
}

// CancelBookingReq cancels a booking (user or admin force-cancel) with a reason.
type CancelBookingReq struct {
	BookingID uint   `json:"bookingId"`
	Reason    string `json:"reason"`
}

// BookingPageReq paginated listing with optional status/type filter.
type BookingPageReq struct {
	commonReq.PageInfo
	Status string `json:"status" form:"status"`
	Type   string `json:"type" form:"type"`
}
