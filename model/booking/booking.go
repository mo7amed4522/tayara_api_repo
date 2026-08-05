package booking

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
)

// Booking the central entity of the marketplace, covering package and custom-trip bookings.
// Money is int64 minor units; percent is int64 basis points. Status axes are kept separate:
// booking_status (lifecycle), payment_status, kyc_status.
type Booking struct {
	ID                   uint           `json:"id" gorm:"primarykey;comment:ID"`
	BookingCode          string         `json:"bookingCode" form:"bookingCode" gorm:"size:64;uniqueIndex;comment:BK- prefixed code"`
	UserID               uint           `json:"userId" form:"userId" gorm:"index;comment:Booking user ID"`
	AgentID              uint           `json:"agentId" form:"agentId" gorm:"index;comment:Booking agent ID"`
	PackageID            *uint          `json:"packageId" form:"packageId" gorm:"index;comment:Package ID (package type)"`
	CustomTripRequestID  *uint          `json:"customTripRequestId" form:"customTripRequestId" gorm:"index;comment:Custom trip request ID"`
	Type                 string         `json:"type" form:"type" gorm:"size:32;index;comment:package|custom_trip"`
	BookingStatus        string         `json:"bookingStatus" form:"bookingStatus" gorm:"size:32;index;default:requested;comment:Lifecycle status"`
	PaymentStatus        string         `json:"paymentStatus" form:"paymentStatus" gorm:"size:32;index;default:pending;comment:Payment status"`
	KYCStatus            string         `json:"kycStatus" form:"kycStatus" gorm:"size:32;index;default:not_submitted;comment:User KYC status"`
	AmountGross          int64          `json:"amountGross" form:"amountGross" gorm:"default:0;comment:Gross amount, minor units"`
	DiscountAmount       int64          `json:"discountAmount" form:"discountAmount" gorm:"default:0;comment:Discount amount, minor units"`
	DiscountPercent      int64          `json:"discountPercent" form:"discountPercent" gorm:"default:0;comment:Discount, basis points"`
	CommissionPercent    int64          `json:"commissionPercent" form:"commissionPercent" gorm:"default:0;comment:Commission, basis points (snapshot)"`
	CommissionAmount     int64          `json:"commissionAmount" form:"commissionAmount" gorm:"default:0;comment:Platform commission, minor units (snapshot)"`
	StripeFee            int64          `json:"stripeFee" form:"stripeFee" gorm:"default:0;comment:Stripe fee, minor units"`
	AgentShare           int64          `json:"agentShare" form:"agentShare" gorm:"default:0;comment:Agent share, minor units"`
	Currency             string         `json:"currency" form:"currency" gorm:"size:8;default:AED;comment:Currency code"`
	DestinationCountryID uint           `json:"destinationCountryId" form:"destinationCountryId" gorm:"index;comment:Destination country ID"`
	StartDate            *time.Time     `json:"startDate" form:"startDate" gorm:"comment:Travel start date"`
	EndDate              *time.Time     `json:"endDate" form:"endDate" gorm:"comment:Travel end date"`
	TravelerCount        int            `json:"travelerCount" form:"travelerCount" gorm:"default:1;comment:Number of travelers"`
	HotelChoiceID        *uint          `json:"hotelChoiceId" form:"hotelChoiceId" gorm:"index;comment:Selected hotel option ID"`
	SelectedActivities   common.JSONMap `json:"selectedActivities" form:"selectedActivities" gorm:"comment:Selected activities"`
	Notes                string         `json:"notes" form:"notes" gorm:"size:1000;comment:User notes"`
	AdminNotes           string         `json:"adminNotes" form:"adminNotes" gorm:"size:1000;comment:Admin notes"`
	CancellationReason   string         `json:"cancellationReason" form:"cancellationReason" gorm:"size:1000;comment:Cancellation reason"`
	CancelledBy          *uint          `json:"cancelledBy" form:"cancelledBy" gorm:"index;comment:Cancelling actor ID"`
	RequestedAt          *time.Time     `json:"requestedAt" form:"requestedAt" gorm:"comment:Requested at"`
	PriceConfirmedAt     *time.Time     `json:"priceConfirmedAt" form:"priceConfirmedAt" gorm:"comment:Price confirmed at"`
	PaidAt               *time.Time     `json:"paidAt" form:"paidAt" gorm:"comment:Paid at"`
	CapturedAt           *time.Time     `json:"capturedAt" form:"capturedAt" gorm:"comment:Captured at"`
	ConfirmedAt          *time.Time     `json:"confirmedAt" form:"confirmedAt" gorm:"comment:Booking confirmed at"`
	CancelledAt          *time.Time     `json:"cancelledAt" form:"cancelledAt" gorm:"comment:Cancelled at"`
	global.GVA_MODEL
}

func (Booking) TableName() string { return "bookings" }
