package booking

import (
	"time"

	bookingmodel "github.com/mo7amed4522/tayara_api/model/booking"
	"gorm.io/gorm"
)

func seedBookingComponents(db *gorm.DB, now time.Time) error {
	if err := seedOrigTravelers(db); err != nil {
		return err
	}
	if err := seedOrigTravelerDocs(db); err != nil {
		return err
	}
	if err := seedOrigOffers(db, now); err != nil {
		return err
	}
	if err := seedOrigHistory(db); err != nil {
		return err
	}
	if err := seedOrigSLA(db, now); err != nil {
		return err
	}
	if err := seedOrigBookingDocs(db); err != nil {
		return err
	}
	return nil
}

func seedOrigTravelers(db *gorm.DB) error {
	type t struct{ code, name, email, doc string }
	rows := []t{
		{"BK-2001", "Ahmed Ali", "ahmed.ali@example.com", "approved"},
		{"BK-2001", "Fatima Ali", "", "approved"},
		{"BK-2002", "Youssef K", "youssef.k@example.com", "missing"},
		{"BK-2002", "Nadia K", "", "missing"},
		{"BK-2003", "Omar H", "omar.h@example.com", "missing"},
		{"BK-2004", "Youssef K", "youssef.k@example.com", "missing"},
		{"BK-2005", "Ahmed Ali", "ahmed.ali@example.com", "missing"},
		{"BK-2006", "Ahmed Ali", "ahmed.ali@example.com", "approved"},
	}
	travelers := make([]bookingmodel.BookingTraveler, 0, len(rows))
	for _, r := range rows {
		bid := bookingByCode(db, r.code)
		if bid == 0 {
			continue
		}
		travelers = append(travelers, bookingmodel.BookingTraveler{
			BookingID: bid, IsBooker: r.email != "",
			FullName: r.name, AdultOrChild: "adult", Email: r.email, DocStatus: r.doc,
		})
	}
	if len(travelers) == 0 {
		return nil
	}
	return db.Create(&travelers).Error
}

func seedOrigTravelerDocs(db *gorm.DB) error {
	codes := []string{"BK-2001", "BK-2006"}
	ids := make([]uint, 0, len(codes))
	for _, c := range codes {
		if id := bookingByCode(db, c); id != 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	var travelers []bookingmodel.BookingTraveler
	if err := db.Where("booking_id IN ? AND doc_status = ?", ids, "approved").Find(&travelers).Error; err != nil {
		return err
	}
	docs := []bookingmodel.TravelerDocument{}
	for _, tr := range travelers {
		docs = append(docs, bookingmodel.TravelerDocument{
			BookingTravelerID: tr.ID,
			DocumentTypeID:    docTypeID(db, "passport_copy"),
			FileURL:           "https://files.tayara.test/orig/" + tr.FullName + "_passport.jpg",
			OriginalName:      "passport.jpg", MIME: "image/jpeg",
			DocStatus: "approved", ReviewNote: "OK", ReviewedBy: uintPtr(1),
		})
	}
	if len(docs) == 0 {
		return nil
	}
	return db.Create(&docs).Error
}

func seedOrigOffers(db *gorm.DB, now time.Time) error {
	mk := func(code string, agent uint, price int64, status string, off, exp *time.Time) (bookingmodel.BookingOffer, bool) {
		bid := bookingByCode(db, code)
		if bid == 0 {
			return bookingmodel.BookingOffer{}, false
		}
		return bookingmodel.BookingOffer{
			BookingID: bid, AgentID: agent,
			FinalPrice: price, Status: status, OfferedAt: off, ExpiresAt: exp,
		}, true
	}
	raw := []struct {
		code   string
		agent  uint
		price  int64
		status string
		off    *time.Time
		exp    *time.Time
	}{
		{"BK-2001", 1, 320000, "accepted", tAdd(now, -39*24), tAdd(now, -38*24)},
		{"BK-2002", 1, 210000, "accepted", tAdd(now, -11*24), tAdd(now, -10*24)},
		{"BK-2003", 2, 210000, "active", tAdd(now, -20*time.Hour), tAdd(now, 52*time.Hour)},
		{"BK-2005", 1, 210000, "declined", tAdd(now, -30*24), tAdd(now, -29*24)},
		{"BK-2006", 1, 460000, "accepted", tAdd(now, -19*24), tAdd(now, -18*24)},
	}
	offers := []bookingmodel.BookingOffer{}
	for _, r := range raw {
		o, ok := mk(r.code, r.agent, r.price, r.status, r.off, r.exp)
		if ok {
			offers = append(offers, o)
		}
	}
	if len(offers) == 0 {
		return nil
	}
	return db.Create(&offers).Error
}

func seedOrigHistory(db *gorm.DB) error {
	uAhmed := userID(db, "ahmed.ali")
	uYoussef := userID(db, "youssef.k")
	uOmar := userID(db, "omar.h")

	steps := map[string][][4]string{
		"BK-2001": {
			{"system", "requested", "agent_reviewing", "Request submitted"},
			{"agent", "agent_reviewing", "price_confirmed", "Offer accepted"},
			{"user", "price_confirmed", "payment_authorized", "Payment authorized"},
			{"system", "payment_authorized", "docs_pending", "Payment captured"},
			{"user", "docs_pending", "docs_submitted", "Docs submitted"},
			{"agent", "docs_submitted", "booking_confirmed", "Docs verified"},
			{"agent", "booking_confirmed", "processing", "Processing tickets"},
			{"agent", "processing", "issued_ready", "Tickets issued"},
			{"system", "issued_ready", "completed", "Trip completed"},
		},
		"BK-2002": {
			{"system", "requested", "agent_reviewing", "Request submitted"},
			{"agent", "agent_reviewing", "price_confirmed", "Offer accepted"},
			{"user", "price_confirmed", "payment_authorized", "Payment authorized"},
			{"system", "payment_authorized", "docs_pending", "Payment captured"},
		},
		"BK-2003": {
			{"user", "requested", "agent_reviewing", "Request submitted"},
			{"agent", "agent_reviewing", "price_confirmed", "Offer received"},
		},
		"BK-2004": {{"user", "", "requested", "New booking request"}},
		"BK-2005": {{"user", "requested", "cancelled", "Cancelled by user"}},
		"BK-2006": {
			{"system", "requested", "price_confirmed", "Custom trip priced"},
			{"user", "price_confirmed", "payment_authorized", "Payment authorized"},
			{"system", "payment_authorized", "paid_confirmed", "Payment captured"},
			{"admin", "paid_confirmed", "refunded", "Full refund issued"},
		},
	}

	userForBooking := map[string]uint{
		"BK-2001": uAhmed, "BK-2002": uYoussef, "BK-2003": uOmar,
		"BK-2004": uYoussef, "BK-2005": uAhmed, "BK-2006": uAhmed,
	}

	var all []bookingmodel.BookingStatusHistory
	for code, chain := range steps {
		bid := bookingByCode(db, code)
		if bid == 0 {
			continue
		}
		u := userForBooking[code]
		for _, s := range chain {
			var at string
			var aid uint
			switch s[0] {
			case "user":
				at, aid = "user", u
			case "agent":
				at, aid = "agent", 1
			case "admin":
				at, aid = "admin", 1
			default:
				at, aid = "system", 0
			}
			all = append(all, bookingmodel.BookingStatusHistory{
				BookingID: bid, ActorType: at, ActorID: aid,
				FromStatus: s[1], ToStatus: s[2], Notes: s[3],
			})
		}
	}
	if len(all) == 0 {
		return nil
	}
	return db.Create(&all).Error
}

func seedOrigSLA(db *gorm.DB, now time.Time) error {
	type row struct {
		code, slatype string
		deadline      time.Time
		breached      bool
		active        bool
	}
	rows := []row{
		{"BK-2002", "user_docs", now.Add(30 * 24 * time.Hour), false, true},
		{"BK-2003", "user_payment", now.Add(52 * time.Hour), false, true},
		{"BK-2004", "agent_response", now.Add(2 * time.Hour), false, true},
		{"BK-2005", "agent_response", now.Add(-29 * 24 * time.Hour), true, false},
		{"BK-2001", "agent_response", now.Add(-38 * 24 * time.Hour), false, false},
		{"BK-2001", "user_payment", now.Add(-36 * 24 * time.Hour), false, false},
	}
	events := []bookingmodel.SLAEvent{}
	for _, r := range rows {
		bid := bookingByCode(db, r.code)
		if bid == 0 {
			continue
		}
		events = append(events, bookingmodel.SLAEvent{
			BookingID: bid, SLAType: r.slatype,
			DeadlineAt: r.deadline, Breached: r.breached, Active: r.active,
		})
	}
	if len(events) == 0 {
		return nil
	}
	return db.Create(&events).Error
}

func seedOrigBookingDocs(db *gorm.DB) error {
	type row struct {
		code, role, url, name, status string
	}
	rows := []row{
		{"BK-2001", "issued_ticket", "https://files.tayara.test/orig/bk2001_tickets.pdf", "tickets.pdf", "verified"},
		{"BK-2001", "voucher", "https://files.tayara.test/orig/bk2001_voucher.pdf", "voucher.pdf", "verified"},
	}
	docs := []bookingmodel.BookingDocument{}
	for _, r := range rows {
		bid := bookingByCode(db, r.code)
		if bid == 0 {
			continue
		}
		docs = append(docs, bookingmodel.BookingDocument{
			BookingID: bid, DocRole: r.role, FileURL: r.url,
			OriginalName: r.name, UploaderRole: "agent", UploadedBy: 1, Status: r.status,
		})
	}
	if len(docs) == 0 {
		return nil
	}
	return db.Create(&docs).Error
}
