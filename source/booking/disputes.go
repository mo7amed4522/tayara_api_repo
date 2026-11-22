package booking

import (
	"time"

	bookingmodel "github.com/mo7amed4522/tayara_api/model/booking"
	"gorm.io/gorm"
)

func seedDisputes(db *gorm.DB, now time.Time) error {
	uAhmed := userID(db, "ahmed.ali")
	b6 := bookingByCode(db, "BK-2006")
	disputes := []bookingmodel.Dispute{}
	if b6 != 0 && uAhmed != 0 {
		disputes = append(disputes, bookingmodel.Dispute{
			BookingID: b6, RaisedBy: "user", RaisedByID: uAhmed,
			Status: "resolved", Details: "Charged full amount but service cancelled",
			Resolution: "Full refund processed", ResolvedBy: uintPtr(1),
			ResolvedAt: tAdd(now, -13*24),
		})
	}
	if len(disputes) == 0 {
		return nil
	}
	return db.Create(&disputes).Error
}

func seedDocAccessLogs(db *gorm.DB) error {
	b2 := bookingByCode(db, "BK-2002")
	logs := []bookingmodel.DocumentAccessLog{}
	if b2 != 0 {
		logs = append(logs,
			bookingmodel.DocumentAccessLog{
				BookingID: b2, ActorType: "agent", ActorID: 1,
				Action: "view", Detail: "Agent viewed traveler documents",
			},
			bookingmodel.DocumentAccessLog{
				BookingID: b2, ActorType: "admin", ActorID: 1,
				Action: "download", Detail: "Admin downloaded KYC file for review",
			},
		)
	}
	if len(logs) == 0 {
		return nil
	}
	return db.Create(&logs).Error
}
