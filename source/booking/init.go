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

// MigrateTable/TableCreated are no-ops: tables migrate via gorm_biz AutoMigrate.
func (b *bookingInit) MigrateTable(ctx context.Context) (context.Context, error) { return ctx, nil }
func (b *bookingInit) TableCreated(ctx context.Context) bool                     { return true }

func (b *bookingInit) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return false
	}
	var n int64
	return db.Model(&bookingmodel.Booking{}).Count(&n).Error == nil && n > 0
}

func (b *bookingInit) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return ctx, system.ErrMissingDBContext
	}

	// Resolve parent IDs from already-seeded data.
	uAhmed := userID(db, "ahmed.ali")
	uYoussef := userID(db, "youssef.k")
	uOmar := userID(db, "omar.h")
	aDesert := agentID(db, "hello@desertsafari.ae")
	aLuxury := agentID(db, "book@luxurytravel.ae")
	pIstanbul := packageID(db, "PK-1001")
	pDubai := packageID(db, "PK-1002")

	now := time.Now()
	bookings := []bookingmodel.Booking{
		// 1) Fully completed trip (paid + captured + payout released).
		{
			BookingCode: "BK-2001", UserID: uAhmed, AgentID: aDesert, PackageID: &pIstanbul, Type: "package",
			BookingStatus: "completed", PaymentStatus: "captured", KYCStatus: "approved",
			AmountGross: 320000, DiscountAmount: 12800, DiscountPercent: 400, CommissionPercent: 1000,
			CommissionAmount: 30720, StripeFee: 7360, AgentShare: 281920, Currency: "AED",
			DestinationCountryID: countryID(db, "TR"), StartDate: tAdd(now, 20*24), EndDate: tAdd(now, 25*24),
			TravelerCount: 2, HotelChoiceID: uintPtr(1),
			SelectedActivities: common.JSONMap{"items": []string{"Bosphorus Cruise", "Blue Mosque"}},
			RequestedAt: tAdd(now, -40*24), PriceConfirmedAt: tAdd(now, -39*24),
			PaidAt: tAdd(now, -37*24), CapturedAt: tAdd(now, -36*24), ConfirmedAt: tAdd(now, -10*24),
		},
		// 2) Paid but waiting on traveler documents.
		{
			BookingCode: "BK-2002", UserID: uYoussef, AgentID: aDesert, PackageID: &pDubai, Type: "package",
			BookingStatus: "docs_pending", PaymentStatus: "captured", KYCStatus: "approved",
			AmountGross: 210000, DiscountAmount: 10500, DiscountPercent: 500, CommissionPercent: 1000,
			CommissionAmount: 19950, StripeFee: 4830, AgentShare: 185220, Currency: "AED",
			DestinationCountryID: countryID(db, "AE"), StartDate: tAdd(now, 15*24), EndDate: tAdd(now, 19*24),
			TravelerCount: 2,
			RequestedAt: tAdd(now, -12*24), PriceConfirmedAt: tAdd(now, -11*24), PaidAt: tAdd(now, -10*24), CapturedAt: tAdd(now, -9*24),
		},
		// 3) Price confirmed, awaiting user payment (live pay window + active offer).
		{
			BookingCode: "BK-2003", UserID: uOmar, AgentID: aLuxury, PackageID: &pDubai, Type: "package",
			BookingStatus: "awaiting_payment", PaymentStatus: "authorized", KYCStatus: "not_submitted",
			AmountGross: 210000, DiscountAmount: 0, DiscountPercent: 0, CommissionPercent: 500,
			CommissionAmount: 0, StripeFee: 0, AgentShare: 0, Currency: "AED",
			DestinationCountryID: countryID(db, "AE"), StartDate: tAdd(now, 25*24), EndDate: tAdd(now, 29*24),
			TravelerCount: 1, HoldExpiresAt: tAdd(now, 3*24),
			RequestedAt: tAdd(now, -2*24), PriceConfirmedAt: tAdd(now, -20*time.Hour),
		},
		// 4) Fresh request, agent reviewing (live agent-response SLA).
		{
			BookingCode: "BK-2004", UserID: uYoussef, AgentID: aDesert, PackageID: &pIstanbul, Type: "package",
			BookingStatus: "agent_reviewing", PaymentStatus: "pending", KYCStatus: "not_submitted",
			AmountGross: 320000, DiscountAmount: 0, DiscountPercent: 0, CommissionPercent: 0,
			CommissionAmount: 0, StripeFee: 0, AgentShare: 0, Currency: "AED",
			DestinationCountryID: countryID(db, "TR"), StartDate: tAdd(now, 40*24), EndDate: tAdd(now, 45*24),
			TravelerCount: 1, RequestedAt: tAdd(now, -4*time.Hour),
		},
		// 5) Cancelled by user before payment.
		{
			BookingCode: "BK-2005", UserID: uAhmed, AgentID: aDesert, PackageID: &pDubai, Type: "package",
			BookingStatus: "cancelled", PaymentStatus: "cancelled", KYCStatus: "not_submitted",
			AmountGross: 210000, DiscountAmount: 0, DiscountPercent: 0, CommissionPercent: 0,
			CommissionAmount: 0, StripeFee: 0, AgentShare: 0, Currency: "AED",
			DestinationCountryID: countryID(db, "AE"), TravelerCount: 1,
			CancellationReason: "Changed travel plans", CancelledBy: uintPtr(uAhmed),
			RequestedAt: tAdd(now, -30*24), CancelledAt: tAdd(now, -29*24),
		},
		// 6) Custom-trip booking, refunded after capture (full refund).
		{
			BookingCode: "BK-2006", UserID: uAhmed, AgentID: aDesert, PackageID: nil, Type: "custom_trip",
			BookingStatus: "refunded", PaymentStatus: "refunded", KYCStatus: "approved",
			AmountGross: 460000, DiscountAmount: 0, DiscountPercent: 0, CommissionPercent: 1000,
			CommissionAmount: 44000, StripeFee: 10580, AgentShare: 405420, Currency: "AED",
			DestinationCountryID: countryID(db, "TR"), StartDate: tAdd(now, 35*24), EndDate: tAdd(now, 42*24),
			TravelerCount: 1,
			RequestedAt: tAdd(now, -20*24), PriceConfirmedAt: tAdd(now, -19*24),
			PaidAt: tAdd(now, -17*24), CapturedAt: tAdd(now, -16*24), CancelledAt: tAdd(now, -14*24),
			CancellationReason: "Customer requested cancellation",
		},
	}
	if err := db.Create(&bookings).Error; err != nil {
		return ctx, err
	}

	// Custom trip request referenced by BK-2006.
	ct := bookingmodel.CustomTripRequest{
		RequestCode: "CT-3001", UserID: uAhmed, AgentID: aDesert,
		DestinationCountryID: countryID(db, "TR"), EstimatedCost: 460000,
		RequiredDocumentsSummary: common.JSONMap{"docs": []string{"passport_copy", "personal_photo"}},
		TravelersCount: 1, Status: "offer_submitted", SubmittedAt: tAdd(now, -20*24),
	}
	if err := db.Create(&ct).Error; err != nil {
		return ctx, err
	}
	bk6 := bookingByCode(db, "BK-2006")
	if err := db.Model(&bookingmodel.Booking{}).Where("id = ?", bk6).Update("custom_trip_request_id", ct.ID).Error; err != nil {
		return ctx, err
	}

	// Travelers, offers, status history, SLA events, booking docs, traveler docs per booking.
	if err := seedBookingComponents(db, now); err != nil {
		return ctx, err
	}

	// Disputes + document access logs.
	if err := seedDisputes(db, now); err != nil {
		return ctx, err
	}
	if err := seedDocAccessLogs(db); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// ---- id lookups ----

func userID(db *gorm.DB, username string) uint {
	var r struct{ ID uint }
	if err := db.Table("bx_user_profile").Select("id").Where("username = ?", username).Scan(&r).Error; err != nil {
		return 0
	}
	return r.ID
}
func agentID(db *gorm.DB, email string) uint {
	var r struct{ ID uint }
	if err := db.Table("dx_agent_profile").Select("id").Where("business_email = ?", email).Scan(&r).Error; err != nil {
		return 0
	}
	return r.ID
}
func packageID(db *gorm.DB, code string) uint {
	var r struct{ ID uint }
	if err := db.Table("catalog_package").Select("id").Where("package_code = ?", code).Scan(&r).Error; err != nil {
		return 0
	}
	return r.ID
}
func bookingByCode(db *gorm.DB, code string) uint {
	var r struct{ ID uint }
	if err := db.Table("bookings").Select("id").Where("booking_code = ?", code).Scan(&r).Error; err != nil {
		return 0
	}
	return r.ID
}
func countryID(db *gorm.DB, code string) uint {
	var r struct{ ID uint }
	if err := db.Table("base_country").Select("id").Where("code = ?", code).Scan(&r).Error; err != nil {
		return 0
	}
	return r.ID
}
func docTypeID(db *gorm.DB, code string) uint {
	var r struct{ ID uint }
	if err := db.Table("base_document_type").Select("id").Where("code = ?", code).Scan(&r).Error; err != nil {
		return 0
	}
	return r.ID
}

func seedBookingComponents(db *gorm.DB, now time.Time) error {
	// Resolve the actors referenced in status-history timelines (by username).
	uYoussef := userID(db, "youssef.k")
	uOmar := userID(db, "omar.h")

	// ---- Booking 1: completed ----
	b1 := bookingByCode(db, "BK-2001")
	travelers1 := []bookingmodel.BookingTraveler{
		{BookingID: b1, IsBooker: true, FullName: "Ahmed Ali", AdultOrChild: "adult", Email: "ahmed.ali@mail.test", DocStatus: "approved"},
		{BookingID: b1, IsBooker: false, FullName: "Sara Ali", AdultOrChild: "adult", DocStatus: "approved"},
	}
	if err := db.Create(&travelers1).Error; err != nil {
		return err
	}
	// Approved traveler documents for the completed booking (all doc_status approved).
	if err := db.Create(&[]bookingmodel.TravelerDocument{
		{BookingTravelerID: travelers1[0].ID, DocumentTypeID: docTypeID(db, "passport_copy"), FileURL: "https://files.tayara.test/bk1/t1_passport.jpg", OriginalName: "passport.jpg", MIME: "image/jpeg", DocStatus: "approved", ReviewNote: "OK", ReviewedBy: uintPtr(1)},
		{BookingTravelerID: travelers1[1].ID, DocumentTypeID: docTypeID(db, "personal_photo"), FileURL: "https://files.tayara.test/bk1/t2_photo.jpg", OriginalName: "photo.jpg", MIME: "image/jpeg", DocStatus: "approved", ReviewNote: "OK", ReviewedBy: uintPtr(1)},
	}).Error; err != nil {
		return err
	}
	if err := db.Create(&[]bookingmodel.BookingOffer{{
		BookingID: b1, AgentID: 1, FinalPrice: 320000, Status: "accepted", OfferedAt: tAdd(now, -39*24), RespondedAt: tAdd(now, -38*24),
	}}).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b1, "system", 0, "requested", "agent_reviewing", "Request submitted")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b1, "agent", 1, "agent_reviewing", "price_confirmed", "Offer accepted")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b1, "user", 1, "price_confirmed", "payment_authorized", "Payment authorized")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b1, "system", 0, "payment_authorized", "docs_pending", "Payment captured")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b1, "agent", 1, "docs_pending", "docs_submitted", "Docs submitted by traveler")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b1, "agent", 1, "docs_submitted", "booking_confirmed", "Docs verified")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b1, "agent", 1, "booking_confirmed", "processing", "Processing tickets")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b1, "agent", 1, "processing", "issued_ready", "Tickets issued")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b1, "system", 0, "issued_ready", "completed", "Trip completed")).Error; err != nil {
		return err
	}
	if err := db.Create(&[]bookingmodel.BookingDocument{
		{BookingID: b1, DocRole: "issued_ticket", FileURL: "https://files.tayara.test/bk1/tickets.pdf", OriginalName: "tickets.pdf", UploaderRole: "agent", UploadedBy: 1, Status: "verified"},
	}).Error; err != nil {
		return err
	}

	// ---- Booking 2: docs pending ----
	b2 := bookingByCode(db, "BK-2002")
	travelers2 := []bookingmodel.BookingTraveler{
		{BookingID: b2, IsBooker: true, FullName: "Youssef Karim", AdultOrChild: "adult", Email: "youssef.k@mail.test", DocStatus: "missing"},
		{BookingID: b2, IsBooker: false, FullName: "Mona Karim", AdultOrChild: "adult", DocStatus: "missing"},
	}
	if err := db.Create(&travelers2).Error; err != nil {
		return err
	}
	if err := db.Create(&[]bookingmodel.BookingOffer{{
		BookingID: b2, AgentID: 1, FinalPrice: 210000, Status: "accepted", OfferedAt: tAdd(now, -11*24), RespondedAt: tAdd(now, -10*24),
	}}).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b2, "user", uYoussef, "requested", "price_confirmed", "Offer accepted")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b2, "system", 0, "price_confirmed", "payment_authorized", "Payment authorized")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b2, "system", 0, "payment_authorized", "docs_pending", "Payment captured")).Error; err != nil {
		return err
	}
	if err := db.Create(&[]bookingmodel.SLAEvent{
		{BookingID: b2, SLAType: "user_payment", DeadlineAt: now.Add(30 * 24 * time.Hour), Breached: false, Active: true},
	}).Error; err != nil {
		return err
	}

	// ---- Booking 3: awaiting payment ----
	b3 := bookingByCode(db, "BK-2003")
	if err := db.Create(&[]bookingmodel.BookingTraveler{
		{BookingID: b3, IsBooker: true, FullName: "Omar Hassan", AdultOrChild: "adult", Email: "omar.h@mail.test", DocStatus: "missing"},
	}).Error; err != nil {
		return err
	}
	if err := db.Create(&[]bookingmodel.BookingOffer{{
		BookingID: b3, AgentID: 2, FinalPrice: 210000, Status: "active",
		AvailableDates: common.JSONMap{"from": now.Add(25 * 24 * time.Hour).Format("2006-01-02")},
		OfferedAt: tAdd(now, -20*time.Hour), ExpiresAt: tAdd(now, 52*time.Hour),
	}}).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b3, "user", uOmar, "requested", "price_confirmed", "Offer received")).Error; err != nil {
		return err
	}
	if err := db.Create(&[]bookingmodel.SLAEvent{
		{BookingID: b3, SLAType: "user_payment", DeadlineAt: now.Add(52 * time.Hour), Breached: false, Active: true},
	}).Error; err != nil {
		return err
	}

	// ---- Booking 4: agent reviewing ----
	b4 := bookingByCode(db, "BK-2004")
	if err := db.Create(&[]bookingmodel.BookingTraveler{
		{BookingID: b4, IsBooker: true, FullName: "Youssef Karim", AdultOrChild: "adult", Email: "youssef.k@mail.test", DocStatus: "missing"},
	}).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b4, "user", 2, "", "requested", "New booking request")).Error; err != nil {
		return err
	}
	if err := db.Create(&[]bookingmodel.SLAEvent{
		{BookingID: b4, SLAType: "agent_response", DeadlineAt: now.Add(2 * time.Hour), Breached: false, Active: true},
	}).Error; err != nil {
		return err
	}

	// ---- Booking 5: cancelled ----
	b5 := bookingByCode(db, "BK-2005")
	if err := db.Create(&[]bookingmodel.BookingTraveler{
		{BookingID: b5, IsBooker: true, FullName: "Ahmed Ali", AdultOrChild: "adult", Email: "ahmed.ali@mail.test", DocStatus: "missing"},
	}).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b5, "user", 1, "requested", "cancelled", "Cancelled by user")).Error; err != nil {
		return err
	}

	// ---- Booking 6: refunded ----
	b6 := bookingByCode(db, "BK-2006")
	if err := db.Create(&[]bookingmodel.BookingTraveler{
		{BookingID: b6, IsBooker: true, FullName: "Ahmed Ali", AdultOrChild: "adult", Email: "ahmed.ali@mail.test", DocStatus: "approved"},
	}).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b6, "system", 0, "requested", "price_confirmed", "Custom trip priced")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b6, "system", 0, "payment_authorized", "paid_confirmed", "Payment captured")).Error; err != nil {
		return err
	}
	if err := db.Create(bookingHistory(b6, "admin", 1, "paid_confirmed", "refunded", "Full refund issued")).Error; err != nil {
		return err
	}

	return nil
}

func bookingHistory(bookingID uint, actorType string, actorID uint, from, to, notes string) []bookingmodel.BookingStatusHistory {
	return []bookingmodel.BookingStatusHistory{
		{BookingID: bookingID, ActorType: actorType, ActorID: actorID, FromStatus: from, ToStatus: to, Notes: notes},
	}
}

func seedDisputes(db *gorm.DB, now time.Time) error {
	b6 := bookingByCode(db, "BK-2006")
	b3 := bookingByCode(db, "BK-2003")
	resolvedAt := tAdd(now, -13*24)
	disputes := []bookingmodel.Dispute{
		{BookingID: b6, RaisedBy: "user", RaisedByID: 1, Status: "resolved", Details: "Charged full amount but service cancelled", Resolution: "Full refund processed", ResolvedBy: uintPtr(1), ResolvedAt: resolvedAt},
		{BookingID: b3, RaisedBy: "agent", RaisedByID: 2, Status: "open", Details: "Customer disputes package price after confirmation"},
	}
	return db.Create(&disputes).Error
}

func seedDocAccessLogs(db *gorm.DB) error {
	b2 := bookingByCode(db, "BK-2002")
	logs := []bookingmodel.DocumentAccessLog{
		{BookingID: b2, ActorType: "agent", ActorID: 1, Action: "view", Detail: "Agent viewed traveler documents"},
		{BookingID: b2, ActorType: "admin", ActorID: 1, Action: "download", Detail: "Admin downloaded KYC file for review"},
	}
	return db.Create(&logs).Error
}

// small helpers (package-scoped)
func uintPtr(v uint) *uint { return &v }
func tAdd(t time.Time, d time.Duration) *time.Time {
	n := t.Add(d)
	return &n
}
