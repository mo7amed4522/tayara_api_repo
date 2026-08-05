package booking

import "github.com/mo7amed4522/tayara_api/global"

// BookingDocument a booking-level file: KYC file, agent confirmation, traveler doc reference,
// or issued ticket. The agent confirmation file is a payout-eligibility gate.
type BookingDocument struct {
	ID           uint   `json:"id" gorm:"primarykey;comment:ID"`
	BookingID    uint   `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	DocRole      string `json:"docRole" form:"docRole" gorm:"size:32;index;comment:kyc_file|agent_confirmation|traveler_doc|issued_ticket"`
	FileURL      string `json:"fileUrl" form:"fileUrl" gorm:"size:500;comment:File URL"`
	OriginalName string `json:"originalName" form:"originalName" gorm:"size:191;comment:Original file name"`
	UploaderRole string `json:"uploaderRole" form:"uploaderRole" gorm:"size:16;comment:user|agent|admin"`
	UploadedBy   uint   `json:"uploadedBy" form:"uploadedBy" gorm:"index;comment:Uploader ID"`
	Status       string `json:"status" form:"status" gorm:"size:32;default:uploaded;comment:uploaded|verified|rejected"`
	global.GVA_MODEL
}

func (BookingDocument) TableName() string { return "booking_documents" }
