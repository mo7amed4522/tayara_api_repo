package booking

import (
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// TravelerDocument a document uploaded for a specific traveler on a booking.
// All access to these rows must be audit-logged (document access compliance).
type TravelerDocument struct {
	ID                uint       `json:"id" gorm:"primarykey;comment:ID"`
	BookingTravelerID uint       `json:"bookingTravelerId" form:"bookingTravelerId" gorm:"index;comment:Traveler ID"`
	DocumentTypeID    uint       `json:"documentTypeId" form:"documentTypeId" gorm:"index;comment:Document type ID"`
	FileURL           string     `json:"fileUrl" form:"fileUrl" gorm:"size:500;comment:File URL"`
	OriginalName      string     `json:"originalName" form:"originalName" gorm:"size:191;comment:Original file name"`
	MIME              string     `json:"mime" form:"mime" gorm:"size:128;comment:MIME type"`
	DocStatus         string     `json:"docStatus" form:"docStatus" gorm:"size:32;index;default:missing;comment:missing|pending|approved|not_ok|request_correction"`
	ReviewNote        string     `json:"reviewNote" form:"reviewNote" gorm:"size:500;comment:Agent review note"`
	ReviewedBy        *uint      `json:"reviewedBy" form:"reviewedBy" gorm:"index;comment:Reviewing agent ID"`
	ReviewedAt        *time.Time `json:"reviewedAt" form:"reviewedAt" gorm:"comment:Reviewed at"`
	global.GVA_MODEL
}

func (TravelerDocument) TableName() string { return "traveler_documents" }
