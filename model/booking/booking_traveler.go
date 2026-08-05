package booking

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// BookingTraveler one traveler on a booking (one may be flagged as the Booker).
type BookingTraveler struct {
	ID             uint       `json:"id" gorm:"primarykey;comment:ID"`
	BookingID      uint       `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	IsBooker       bool       `json:"isBooker" form:"isBooker" gorm:"default:false;comment:Is the booking contact"`
	FullName       string     `json:"fullName" form:"fullName" gorm:"size:191;comment:Name as on passport"`
	DOB            *time.Time `json:"dob" form:"dob" gorm:"comment:Date of birth"`
	NationalityID  *uint      `json:"nationalityId" form:"nationalityId" gorm:"index;comment:Nationality ID"`
	PassportNumber string     `json:"passportNumber" form:"passportNumber" gorm:"size:64;comment:Passport number"`
	PassportExpiry *time.Time `json:"passportExpiry" form:"passportExpiry" gorm:"comment:Passport expiry"`
	AdultOrChild   string     `json:"adultOrChild" form:"adultOrChild" gorm:"size:16;default:adult;comment:adult|child"`
	Email          string     `json:"email" form:"email" gorm:"size:191;comment:Booker email"`
	DocStatus      string     `json:"docStatus" form:"docStatus" gorm:"size:32;index;default:missing;comment:Document submission status"`
	global.GVA_MODEL
}

func (BookingTraveler) TableName() string { return "booking_travelers" }
