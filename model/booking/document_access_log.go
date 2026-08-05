package booking

import "github.com/mo7amed4522/tayara_api/global"

// DocumentAccessLog is the append-only audit trail for accesses to traveler/booking documents.
// PRD mandates that all document access be logged and no of these can be shown to anyone who
// did not have prior authorization — every read/download swings through this record.
type DocumentAccessLog struct {
	ID         uint   `json:"id" gorm:"primarykey;comment:ID"`
	BookingID  uint   `json:"bookingId" form:"bookingId" gorm:"index;comment:Booking ID"`
	DocumentID *uint  `json:"documentId" form:"documentId" gorm:"index;comment:Document (traveler or booking) ID"`
	ActorType  string `json:"actorType" form:"actorType" gorm:"size:16;index;comment:system|admin|agent|user"`
	ActorID    uint   `json:"actorId" form:"actorId" gorm:"index;comment:Acting user/agent/admin ID"`
	Action     string `json:"action" form:"action" gorm:"size:32;comment:view|download"`
	Detail     string `json:"detail" form:"detail" gorm:"size:1000;comment:Detail"`
	global.GVA_MODEL
}

func (DocumentAccessLog) TableName() string { return "document_access_log" }
