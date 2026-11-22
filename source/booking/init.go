// Package booking seeds the full booking lifecycle with varied states (requested → completed,
// cancelled, refunded) plus travelers, offers, status-history timelines, documents, SLA events,
// custom-trip requests, disputes and document-access logs. This is the heart of the demo data.
package booking

import (
	"context"
	"time"

	bookingmodel "github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/common"
	"github.com/mo7amed4522/tayara_api/service/system"
	"gorm.io/gorm"
)

const initOrderBooking = system.InitOrderExternal + 6

type bookingInit struct{}

func init() {
	system.RegisterInit(initOrderBooking, &bookingInit{})
}

func (b *bookingInit) InitializerName() string { return "demo_bookings" }

func (b *bookingInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (b *bookingInit) TableCreated(ctx context.Context) bool                     { return true }

func (b *bookingInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var n, k int64
	db.Model(&bookingmodel.Booking{}).Count(&n)
	db.Model(&bookingmodel.Booking{}).Where("booking_code = ?", "BK-6001").Count(&k)
	return n > 0 && k > 0
}

func (b *bookingInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}

	uAhmed := userID(db, "ahmed.ali")
	uYoussef := userID(db, "youssef.k")
	uOmar := userID(db, "omar.h")
	aDesert := agentID(db, "hello@desertsafari.ae")
	aLuxury := agentID(db, "book@luxurytravel.ae")
	pIstanbul := packageID(db, "PK-1001")
	pDubai := packageID(db, "PK-1002")

	now := time.Now()
	bookings := []bookingmodel.Booking{
		{
			BookingCode: "BK-2001", UserID: uAhmed, AgentID: aDesert, PackageID: &pIstanbul, Type: "package",
			BookingStatus: "completed", PaymentStatus: "captured", KYCStatus: "approved",
			AmountGross: 320000, DiscountAmount: 12800, DiscountPercent: 400, CommissionPercent: 1000,
			CommissionAmount: 30720, StripeFee: 7360, AgentShare: 281920, Currency: "AED",
			DestinationCountryID: countryID(db, "TR"), StartDate: tAdd(now, 20*24), EndDate: tAdd(now, 25*24),
			TravelerCount: 2, HotelChoiceID: uintPtr(1),
			SelectedActivities: common.JSONMap{"items": []string{"Bosphorus Cruise", "Blue Mosque"}},
			RequestedAt:        tAdd(now, -40*24), PriceConfirmedAt: tAdd(now, -39*24),
			PaidAt: tAdd(now, -37*24), CapturedAt: tAdd(now, -36*24), ConfirmedAt: tAdd(now, -10*24),
		},
		{
			BookingCode: "BK-2002", UserID: uYoussef, AgentID: aDesert, PackageID: &pDubai, Type: "package",
			BookingStatus: "docs_pending", PaymentStatus: "captured", KYCStatus: "approved",
			AmountGross: 210000, DiscountAmount: 10500, DiscountPercent: 500, CommissionPercent: 1000,
			CommissionAmount: 19950, StripeFee: 4830, AgentShare: 185220, Currency: "AED",
			DestinationCountryID: countryID(db, "AE"), StartDate: tAdd(now, 15*24), EndDate: tAdd(now, 19*24),
			TravelerCount: 2,
			RequestedAt:   tAdd(now, -12*24), PriceConfirmedAt: tAdd(now, -11*24), PaidAt: tAdd(now, -10*24), CapturedAt: tAdd(now, -9*24),
		},
		{
			BookingCode: "BK-2003", UserID: uOmar, AgentID: aLuxury, PackageID: &pDubai, Type: "package",
			BookingStatus: "awaiting_payment", PaymentStatus: "authorized", KYCStatus: "not_submitted",
			AmountGross: 210000, DiscountAmount: 0, DiscountPercent: 0, CommissionPercent: 500,
			CommissionAmount: 0, StripeFee: 0, AgentShare: 0, Currency: "AED",
			DestinationCountryID: countryID(db, "AE"), StartDate: tAdd(now, 25*24), EndDate: tAdd(now, 29*24),
			TravelerCount: 1, HoldExpiresAt: tAdd(now, 3*24),
			RequestedAt: tAdd(now, -2*24), PriceConfirmedAt: tAdd(now, -20*time.Hour),
		},
		{
			BookingCode: "BK-2004", UserID: uYoussef, AgentID: aDesert, PackageID: &pIstanbul, Type: "package",
			BookingStatus: "agent_reviewing", PaymentStatus: "pending", KYCStatus: "not_submitted",
			AmountGross: 320000, DiscountAmount: 0, DiscountPercent: 0, CommissionPercent: 0,
			CommissionAmount: 0, StripeFee: 0, AgentShare: 0, Currency: "AED",
			DestinationCountryID: countryID(db, "TR"), StartDate: tAdd(now, 40*24), EndDate: tAdd(now, 45*24),
			TravelerCount: 1, RequestedAt: tAdd(now, -4*time.Hour),
		},
		{
			BookingCode: "BK-2005", UserID: uAhmed, AgentID: aDesert, PackageID: &pDubai, Type: "package",
			BookingStatus: "cancelled", PaymentStatus: "cancelled", KYCStatus: "not_submitted",
			AmountGross: 210000, DiscountAmount: 0, DiscountPercent: 0, CommissionPercent: 0,
			CommissionAmount: 0, StripeFee: 0, AgentShare: 0, Currency: "AED",
			DestinationCountryID: countryID(db, "AE"), TravelerCount: 1,
			CancellationReason: "Changed travel plans", CancelledBy: uintPtr(uAhmed),
			RequestedAt: tAdd(now, -30*24), CancelledAt: tAdd(now, -29*24),
		},
		{
			BookingCode: "BK-2006", UserID: uAhmed, AgentID: aDesert, PackageID: nil, Type: "custom_trip",
			BookingStatus: "refunded", PaymentStatus: "refunded", KYCStatus: "approved",
			AmountGross: 460000, DiscountAmount: 0, DiscountPercent: 0, CommissionPercent: 1000,
			CommissionAmount: 44000, StripeFee: 10580, AgentShare: 405420, Currency: "AED",
			DestinationCountryID: countryID(db, "TR"), StartDate: tAdd(now, 35*24), EndDate: tAdd(now, 42*24),
			TravelerCount: 1,
			RequestedAt:   tAdd(now, -20*24), PriceConfirmedAt: tAdd(now, -19*24),
			PaidAt: tAdd(now, -17*24), CapturedAt: tAdd(now, -16*24), CancelledAt: tAdd(now, -14*24),
			CancellationReason: "Customer requested cancellation",
		},
	}
	if err := db.Create(&bookings).Error; err != nil {
		return ctx, err
	}

	// ── User 6 (Khaled / Google) full dataset ─────────────────────
	if err := seedKhaledBookings(db, now); err != nil {
		return ctx, err
	}

	ct := bookingmodel.CustomTripRequest{
		RequestCode: "CT-3001", UserID: uAhmed, AgentID: aDesert,
		DestinationCountryID: countryID(db, "TR"), EstimatedCost: 460000,
		RequiredDocumentsSummary: common.JSONMap{"docs": []string{"passport_copy", "personal_photo"}},
		TravelersCount:           1, Status: "offer_submitted", SubmittedAt: tAdd(now, -20*24),
	}
	if err := db.Create(&ct).Error; err != nil {
		return ctx, err
	}
	bk6 := bookingByCode(db, "BK-2006")
	if err := db.Model(&bookingmodel.Booking{}).Where("id = ?", bk6).Update("custom_trip_request_id", ct.ID).Error; err != nil {
		return ctx, err
	}

	if err := seedKhaledCustomTrip(db, now); err != nil {
		return ctx, err
	}

	if err := seedBookingComponents(db, now); err != nil {
		return ctx, err
	}

	// ── User 6 components ─────────────────────────────────────────
	if err := seedKhaledTravelers(db); err != nil {
		return ctx, err
	}
	if err := seedKhaledTravelerDocs(db); err != nil {
		return ctx, err
	}
	if err := seedKhaledOffers(db, now); err != nil {
		return ctx, err
	}
	if err := seedKhaledHistory(db); err != nil {
		return ctx, err
	}
	if err := seedKhaledSLA(db, now); err != nil {
		return ctx, err
	}
	if err := seedKhaledBookingDocs(db); err != nil {
		return ctx, err
	}

	if err := seedDisputes(db, now); err != nil {
		return ctx, err
	}
	if err := seedDocAccessLogs(db); err != nil {
		return ctx, err
	}
	if err := seedKhaledDisputes(db, now); err != nil {
		return ctx, err
	}
	if err := seedKhaledDocLogs(db); err != nil {
		return ctx, err
	}

	// Fix user 6 profile to be consistent with booking-level KYC snapshots
	fixKhaledProfile(db, now)

	return ctx, nil
}

// ---- user 6 bookings ----

func seedKhaledBookings(db *gorm.DB, now time.Time) error {
	uKhaled := userID(db, "mo7amed4522@gmail.com")
	if uKhaled == 0 {
		return nil
	}
	aDesert := agentID(db, "hello@desertsafari.ae")
	aLuxury := agentID(db, "book@luxurytravel.ae")
	pIst := packageID(db, "PK-1001")
	pDub := packageID(db, "PK-1002")
	pBali := packageID(db, "PK-1003")
	pBang := packageID(db, "PK-1005")

	bookings := []bookingmodel.Booking{
		{
			BookingCode: "BK-6001", UserID: uKhaled, AgentID: aDesert, PackageID: &pIst, Type: "package",
			BookingStatus: "completed", PaymentStatus: "captured", KYCStatus: "approved",
			AmountGross: 320000, DiscountAmount: 12800, DiscountPercent: 400, CommissionPercent: 1000,
			CommissionAmount: 30720, StripeFee: 7360, AgentShare: 281920, Currency: "AED",
			DestinationCountryID: countryID(db, "TR"),
			StartDate:            tAdd(now, -30*24), EndDate: tAdd(now, -25*24),
			TravelerCount: 2, HotelChoiceID: uintPtr(1),
			SelectedActivities: common.JSONMap{"items": []string{"Bosphorus Cruise", "Blue Mosque"}},
			RequestedAt:        tAdd(now, -60*24), PriceConfirmedAt: tAdd(now, -59*24),
			PaidAt: tAdd(now, -57*24), CapturedAt: tAdd(now, -56*24), ConfirmedAt: tAdd(now, -20*24),
		},
		{
			BookingCode: "BK-6002", UserID: uKhaled, AgentID: aDesert, PackageID: &pDub, Type: "package",
			BookingStatus: "docs_pending", PaymentStatus: "captured", KYCStatus: "approved",
			AmountGross: 210000, DiscountAmount: 10500, DiscountPercent: 500, CommissionPercent: 1000,
			CommissionAmount: 19950, StripeFee: 4830, AgentShare: 185220, Currency: "AED",
			DestinationCountryID: countryID(db, "AE"),
			StartDate:            tAdd(now, 15*24), EndDate: tAdd(now, 19*24), TravelerCount: 2,
			RequestedAt: tAdd(now, -10*24), PriceConfirmedAt: tAdd(now, -9*24),
			PaidAt: tAdd(now, -8*24), CapturedAt: tAdd(now, -8*24),
		},
		{
			BookingCode: "BK-6003", UserID: uKhaled, AgentID: aLuxury, PackageID: &pDub, Type: "package",
			BookingStatus: "awaiting_payment", PaymentStatus: "authorized", KYCStatus: "not_submitted",
			AmountGross: 210000, CommissionPercent: 500, Currency: "AED",
			DestinationCountryID: countryID(db, "AE"),
			StartDate:            tAdd(now, 25*24), EndDate: tAdd(now, 29*24),
			TravelerCount: 1, HoldExpiresAt: tAdd(now, 3*24),
			RequestedAt: tAdd(now, -3*24), PriceConfirmedAt: tAdd(now, -20*time.Hour),
		},
		{
			BookingCode: "BK-6004", UserID: uKhaled, AgentID: aDesert, PackageID: &pIst, Type: "package",
			BookingStatus: "agent_reviewing", PaymentStatus: "pending", KYCStatus: "not_submitted",
			AmountGross: 320000, Currency: "AED",
			DestinationCountryID: countryID(db, "TR"),
			StartDate:            tAdd(now, 40*24), EndDate: tAdd(now, 45*24),
			TravelerCount: 1, RequestedAt: tAdd(now, -6*time.Hour),
		},
		{
			BookingCode: "BK-6005", UserID: uKhaled, AgentID: aDesert, PackageID: &pDub, Type: "package",
			BookingStatus: "cancelled", PaymentStatus: "cancelled", KYCStatus: "not_submitted",
			AmountGross: 210000, Currency: "AED",
			DestinationCountryID: countryID(db, "AE"), TravelerCount: 1,
			CancellationReason: "Changed travel plans", CancelledBy: uintPtr(uKhaled),
			RequestedAt: tAdd(now, -28*24), CancelledAt: tAdd(now, -25*24),
		},
		{
			BookingCode: "BK-6006", UserID: uKhaled, AgentID: aDesert, PackageID: nil, Type: "custom_trip",
			BookingStatus: "refunded", PaymentStatus: "refunded", KYCStatus: "approved",
			AmountGross: 460000, CommissionPercent: 1000, CommissionAmount: 44000,
			StripeFee: 10580, AgentShare: 405420, Currency: "AED",
			DestinationCountryID: countryID(db, "TR"),
			StartDate:            tAdd(now, 35*24), EndDate: tAdd(now, 42*24), TravelerCount: 1,
			RequestedAt: tAdd(now, -18*24), PriceConfirmedAt: tAdd(now, -17*24),
			PaidAt: tAdd(now, -15*24), CapturedAt: tAdd(now, -14*24), CancelledAt: tAdd(now, -12*24),
			CancellationReason: "Customer requested cancellation",
		},
		{
			BookingCode: "BK-6007", UserID: uKhaled, AgentID: aLuxury, PackageID: &pBali, Type: "package",
			BookingStatus: "price_confirmed", PaymentStatus: "pending", KYCStatus: "pending",
			AmountGross: 450000, DiscountPercent: 300, CommissionPercent: 700, Currency: "AED",
			DestinationCountryID: countryID(db, "ID"),
			StartDate:            tAdd(now, 20*24), EndDate: tAdd(now, 27*24), TravelerCount: 3,
			RequestedAt: tAdd(now, -4*24), PriceConfirmedAt: tAdd(now, -2*24),
		},
		{
			BookingCode: "BK-6008", UserID: uKhaled, AgentID: aDesert, PackageID: &pBang, Type: "package",
			BookingStatus: "processing", PaymentStatus: "captured", KYCStatus: "approved",
			AmountGross: 380000, DiscountAmount: 15200, DiscountPercent: 400, CommissionPercent: 1000,
			CommissionAmount: 36480, StripeFee: 8740, AgentShare: 334780, Currency: "AED",
			DestinationCountryID: countryID(db, "TH"),
			StartDate:            tAdd(now, 10*24), EndDate: tAdd(now, 16*24), TravelerCount: 2,
			RequestedAt: tAdd(now, -12*24), PriceConfirmedAt: tAdd(now, -11*24),
			PaidAt: tAdd(now, -9*24), CapturedAt: tAdd(now, -9*24), ConfirmedAt: tAdd(now, -3*24),
		},
	}
	return db.Create(&bookings).Error
}

func seedKhaledCustomTrip(db *gorm.DB, now time.Time) error {
	uKhaled := userID(db, "mo7amed4522@gmail.com")
	if uKhaled == 0 {
		return nil
	}
	ct := bookingmodel.CustomTripRequest{
		RequestCode:          "CT-6001",
		UserID:               uKhaled,
		AgentID:              agentID(db, "hello@desertsafari.ae"),
		DestinationCountryID: countryID(db, "TR"),
		EstimatedCost:        460000,
		RequiredDocumentsSummary: common.JSONMap{
			"docs": []string{"passport_copy", "personal_photo"},
		},
		TravelersCount: 1,
		Status:         "offer_submitted",
		SubmittedAt:    tAdd(now, -18*24),
	}
	if err := db.Create(&ct).Error; err != nil {
		return err
	}
	return db.Model(&bookingmodel.Booking{}).
		Where("booking_code = ?", "BK-6006").
		Update("custom_trip_request_id", ct.ID).Error
}

func seedKhaledTravelers(db *gorm.DB) error {
	type t struct{ code, name, email, doc string }
	rows := []t{
		{"BK-6001", "Khaled Mohamed", "mo7amed4522@gmail.com", "approved"},
		{"BK-6001", "Mariam Khaled", "", "approved"},
		{"BK-6002", "Khaled Mohamed", "mo7amed4522@gmail.com", "missing"},
		{"BK-6002", "Sara Khaled", "", "missing"},
		{"BK-6003", "Khaled Mohamed", "mo7amed4522@gmail.com", "missing"},
		{"BK-6004", "Khaled Mohamed", "mo7amed4522@gmail.com", "missing"},
		{"BK-6005", "Khaled Mohamed", "mo7amed4522@gmail.com", "missing"},
		{"BK-6006", "Khaled Mohamed", "mo7amed4522@gmail.com", "approved"},
		{"BK-6007", "Khaled Mohamed", "mo7amed4522@gmail.com", "pending"},
		{"BK-6007", "Omar Khaled", "", "pending"},
		{"BK-6007", "Lina Khaled", "", "pending"},
		{"BK-6008", "Khaled Mohamed", "mo7amed4522@gmail.com", "approved"},
		{"BK-6008", "Noor Khaled", "", "approved"},
	}
	travelers := make([]bookingmodel.BookingTraveler, 0, len(rows))
	for _, r := range rows {
		bid := bookingByCode(db, r.code)
		if bid == 0 {
			continue
		}
		travelers = append(travelers, bookingmodel.BookingTraveler{
			BookingID: bid, IsBooker: r.name == "Khaled Mohamed",
			FullName: r.name, AdultOrChild: "adult", Email: r.email, DocStatus: r.doc,
		})
	}
	if len(travelers) == 0 {
		return nil
	}
	return db.Create(&travelers).Error
}

func seedKhaledTravelerDocs(db *gorm.DB) error {
	codes := []string{"BK-6001", "BK-6002", "BK-6006", "BK-6007", "BK-6008"}
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
	if err := db.Where("booking_id IN ?", ids).Find(&travelers).Error; err != nil {
		return err
	}
	docs := []bookingmodel.TravelerDocument{}
	for _, tr := range travelers {
		switch tr.DocStatus {
		case "approved":
			docs = append(docs, bookingmodel.TravelerDocument{
				BookingTravelerID: tr.ID,
				DocumentTypeID:    docTypeID(db, "passport_copy"),
				FileURL:           "https://files.tayara.test/u6/" + tr.FullName + "_passport.jpg",
				OriginalName:      "passport.jpg", MIME: "image/jpeg",
				DocStatus: "approved", ReviewNote: "OK", ReviewedBy: uintPtr(1),
			})
		case "pending":
			docs = append(docs, bookingmodel.TravelerDocument{
				BookingTravelerID: tr.ID,
				DocumentTypeID:    docTypeID(db, "passport_copy"),
				FileURL:           "https://files.tayara.test/u6/" + tr.FullName + "_passport_pending.jpg",
				OriginalName:      "passport.jpg", MIME: "image/jpeg", DocStatus: "pending",
			})
		}
	}
	if len(docs) == 0 {
		return nil
	}
	return db.Create(&docs).Error
}

func seedKhaledOffers(db *gorm.DB, now time.Time) error {
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
		{"BK-6001", 1, 320000, "accepted", tAdd(now, -59*24), tAdd(now, -58*24)},
		{"BK-6002", 1, 210000, "accepted", tAdd(now, -9*24), tAdd(now, -8*24)},
		{"BK-6003", 2, 210000, "active", tAdd(now, -20*time.Hour), tAdd(now, 52*time.Hour)},
		{"BK-6005", 1, 210000, "declined", tAdd(now, -28*24), tAdd(now, -27*24)},
		{"BK-6006", 1, 460000, "accepted", tAdd(now, -17*24), tAdd(now, -16*24)},
		{"BK-6007", 2, 450000, "accepted", tAdd(now, -3*24), tAdd(now, -2*24)},
		{"BK-6008", 1, 380000, "accepted", tAdd(now, -11*24), tAdd(now, -10*24)},
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

func seedKhaledHistory(db *gorm.DB) error {
	u := userID(db, "mo7amed4522@gmail.com")
	if u == 0 {
		return nil
	}
	steps := map[string][][4]string{
		"BK-6001": {
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
		"BK-6002": {
			{"system", "requested", "agent_reviewing", "Request submitted"},
			{"agent", "agent_reviewing", "price_confirmed", "Offer accepted"},
			{"user", "price_confirmed", "payment_authorized", "Payment authorized"},
			{"system", "payment_authorized", "docs_pending", "Payment captured"},
		},
		"BK-6003": {
			{"user", "requested", "agent_reviewing", "Request submitted"},
			{"agent", "agent_reviewing", "price_confirmed", "Offer received"},
		},
		"BK-6004": {{"user", "", "requested", "New booking request"}},
		"BK-6005": {{"user", "requested", "cancelled", "Cancelled by user"}},
		"BK-6006": {
			{"system", "requested", "price_confirmed", "Custom trip priced"},
			{"user", "price_confirmed", "payment_authorized", "Payment authorized"},
			{"system", "payment_authorized", "paid_confirmed", "Payment captured"},
			{"admin", "paid_confirmed", "refunded", "Full refund issued"},
		},
		"BK-6007": {
			{"user", "requested", "agent_reviewing", "Request submitted"},
			{"agent", "agent_reviewing", "price_confirmed", "Offer accepted"},
		},
		"BK-6008": {
			{"user", "requested", "agent_reviewing", "Request submitted"},
			{"agent", "agent_reviewing", "price_confirmed", "Offer accepted"},
			{"user", "price_confirmed", "payment_authorized", "Payment authorized"},
			{"system", "payment_authorized", "docs_pending", "Payment captured"},
			{"user", "docs_pending", "docs_submitted", "Docs submitted"},
			{"agent", "docs_submitted", "booking_confirmed", "Docs verified"},
			{"agent", "booking_confirmed", "processing", "Processing tickets"},
		},
	}
	actor := func(t string) (string, uint) {
		switch t {
		case "user":
			return "user", u
		case "agent":
			return "agent", 1
		case "admin":
			return "admin", 1
		default:
			return "system", 0
		}
	}
	var all []bookingmodel.BookingStatusHistory
	for code, chain := range steps {
		bid := bookingByCode(db, code)
		if bid == 0 {
			continue
		}
		for _, s := range chain {
			at, aid := actor(s[0])
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

func seedKhaledSLA(db *gorm.DB, now time.Time) error {
	type row struct {
		code, slatype string
		deadline      time.Time
		breached      bool
		active        bool
	}
	rows := []row{
		{"BK-6002", "user_docs", now.Add(30 * 24 * time.Hour), false, true},
		{"BK-6003", "user_payment", now.Add(52 * time.Hour), false, true},
		{"BK-6004", "agent_response", now.Add(2 * time.Hour), false, true},
		{"BK-6005", "agent_response", now.Add(-27 * 24 * time.Hour), true, false},
		{"BK-6001", "agent_response", now.Add(-58 * 24 * time.Hour), false, false},
		{"BK-6001", "user_payment", now.Add(-56 * 24 * time.Hour), false, false},
		{"BK-6007", "user_payment", now.Add(48 * time.Hour), false, true},
		{"BK-6008", "agent_processing", now.Add(5 * 24 * time.Hour), false, true},
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

func seedKhaledBookingDocs(db *gorm.DB) error {
	type row struct {
		code, role, url, name, status string
	}
	rows := []row{
		{"BK-6001", "issued_ticket", "https://files.tayara.test/u6/bk6001_tickets.pdf", "tickets.pdf", "verified"},
		{"BK-6001", "voucher", "https://files.tayara.test/u6/bk6001_voucher.pdf", "voucher.pdf", "verified"},
		{"BK-6008", "issued_ticket", "https://files.tayara.test/u6/bk6008_tickets.pdf", "tickets.pdf", "pending"},
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

func seedKhaledDisputes(db *gorm.DB, now time.Time) error {
	u := userID(db, "mo7amed4522@gmail.com")
	if u == 0 {
		return nil
	}
	b6 := bookingByCode(db, "BK-6006")
	b3 := bookingByCode(db, "BK-6003")
	disputes := []bookingmodel.Dispute{}
	if b6 != 0 {
		disputes = append(disputes, bookingmodel.Dispute{
			BookingID: b6, RaisedBy: "user", RaisedByID: u,
			Status: "resolved", Details: "Charged full amount but service cancelled",
			Resolution: "Full refund processed", ResolvedBy: uintPtr(1),
			ResolvedAt: tAdd(now, -11*24),
		})
	}
	if b3 != 0 {
		disputes = append(disputes, bookingmodel.Dispute{
			BookingID: b3, RaisedBy: "agent", RaisedByID: 2,
			Status: "open", Details: "Customer disputes package price after confirmation",
		})
	}
	if len(disputes) == 0 {
		return nil
	}
	return db.Create(&disputes).Error
}

func seedKhaledDocLogs(db *gorm.DB) error {
	b2 := bookingByCode(db, "BK-6002")
	b7 := bookingByCode(db, "BK-6007")
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
	if b7 != 0 {
		logs = append(logs, bookingmodel.DocumentAccessLog{
			BookingID: b7, ActorType: "agent", ActorID: 2,
			Action: "view", Detail: "Agent reviewed pending passport",
		})
	}
	if len(logs) == 0 {
		return nil
	}
	return db.Create(&logs).Error
}

// fixKhaledProfile updates user 6's profile to be consistent with the seeded bookings.
// Bookings BK-6001/6002/6006/6008 have KYCStatus "approved", so the user-level KYC
// must also be approved. Also sets auth_provider, phone, nationality, last_login_at.
func fixKhaledProfile(db *gorm.DB, now time.Time) {
	u := userID(db, "mo7amed4522@gmail.com")
	if u == 0 {
		return
	}
	db.Table("bx_user_profile").Where("id = ?", u).Updates(map[string]interface{}{
		"auth_provider":      "google",
		"phone_country_code": "+971",
		"phone":              "0561000006",
		"nationality":        "Emirati",
		"kyc_status":         "approved",
		"last_login_at":      now,
	})
}
