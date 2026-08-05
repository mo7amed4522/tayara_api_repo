package utils

import (
	"github.com/mo7amed4522/tayara_api/model/booking"
	"gorm.io/gorm"
)

// LogAccess writes an append-only document-access audit record. Every read/download of a
// traveler or booking document routes here so access is always traceable (PRD compliance).
func LogAccess(db *gorm.DB, bookingID uint, docID *uint, actorType string, actorID uint, action, detail string) {
	if db == nil {
		return
	}
	_ = db.Create(&booking.DocumentAccessLog{
		BookingID: bookingID, DocumentID: docID, ActorType: actorType,
		ActorID: actorID, Action: action, Detail: detail,
	}).Error
}
