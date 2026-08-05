package booking_test

// End-to-end smoke test for the core booking lifecycle: seed -> book -> offer -> pay(authorize)
// -> docs -> KYC -> processing -> issued -> completed -> payout eligibility.
//
// It runs the REAL service code against an in-memory sqlite DB (no MySQL/Mongo/Redis needed:
// the payment red lock and SLA timers both fall back to in-process/db defaults). Payment is
// simulated as an authorization+capture since no Stripe key is available in a unit test.
//
// This is the Phase 8 "end-to-end smoke" exit check, kept deterministic and DB-local.

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	bkReq "github.com/mo7amed4522/tayara_api/model/booking/request"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/model/dxapi"
	"github.com/mo7amed4522/tayara_api/model/engage"
	paymentModel "github.com/mo7amed4522/tayara_api/model/payment"
	"github.com/mo7amed4522/tayara_api/service/booking"
	"github.com/mo7amed4522/tayara_api/service/payment"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// setupDB opens an in-memory sqlite DB, migrates the models the lifecycle touches, and seeds
// an agent, a user, and an active package. Sets global.GVA_DB for the service layer.
func setupDB(t *testing.T) (*gorm.DB, uint, uint, uint) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	models := []interface{}{
		&dxapi.DxAgentProfile{}, &bxapi.BxsUserProfile{}, &catalog.Package{},
		&bookingModel.Booking{}, &bookingModel.BookingTraveler{},
		&bookingModel.BookingOffer{}, &bookingModel.BookingStatusHistory{},
		&bookingModel.BookingDocument{}, &bookingModel.SLAEvent{},
		&bookingModel.TravelerDocument{}, &bookingModel.DocumentAccessLog{},
		&catalog.PackageView{}, &paymentModel.Payout{}, &paymentModel.PaymentTransaction{},
		&engage.Notification{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	global.GVA_DB = db
	global.GVA_CONFIG = config.Server{}
	global.GVA_LOG = zap.NewNop()

	// seed agent, user, package
	agent := dxapi.DxAgentProfile{CompanyName: "Acme Travel", BusinessEmail: "a@t.com", Status: "active", SLAHours: 2}
	if err := db.Create(&agent).Error; err != nil {
		t.Fatalf("seed agent: %v", err)
	}
	user := bxapi.BxsUserProfile{FullName: "Tester", Username: "tester", KYCStatus: "pending"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	pkg := catalog.Package{
		PackageCode: "PK-1", AgentID: agent.ID, Type: constants.PackageTypeOutbound,
		Name: "Dubai to Turkey", Status: constants.PackageStatusActive,
		EstimatedPrice: 100_000, DisplayPrice: 100_000, // 1000.00 AED
	}
	if err := db.Create(&pkg).Error; err != nil {
		t.Fatalf("seed package: %v", err)
	}
	return db, agent.ID, user.ID, pkg.ID
}

func TestBookingLifecycleSmoke(t *testing.T) {
	db, agentID, userID, pkgID := setupDB(t)
	ctx := context.Background()

	var svc booking.BookingServices
	var docs booking.DocServices
	var pay payment.PaymentServices

	// 1) BOOK: user creates a booking against the package
	pkgIDptr := pkgID
	bk, err := svc.CreateBooking(userID, bkReq.CreateBookingReq{
		Type: constants.BookingTypePackage, PackageID: &pkgIDptr,
		StartDate: "2026-01-10", EndDate: "2026-01-15", TravelerCount: 1,
		Travelers: []bkReq.TravelerReq{{FullName: "Tester Traveler", Email: "t@t.com"}},
	})
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}
	if bk.BookingStatus != constants.BookingStatusRequested {
		t.Fatalf("expected requested, got %s", bk.BookingStatus)
	}
	if bk.AmountGross != 100_000 {
		t.Fatalf("expected gross 100000, got %d", bk.AmountGross)
	}

	// 2) AGENT reviews
	if err := svc.ReviewBooking(agentID, bk.ID, "reviewing"); err != nil {
		t.Fatalf("review: %v", err)
	}

	// 3) OFFER -> price_confirmed, user pay window opens
	offer, err := svc.SubmitOffer(agentID, bkReq.SubmitOfferReq{BookingID: bk.ID, FinalPrice: 95_000})
	if err != nil {
		t.Fatalf("submit offer: %v", err)
	}
	if offer.FinalPrice != 95_000 {
		t.Fatalf("offer price mismatch: %d", offer.FinalPrice)
	}
	var bkNow bookingModel.Booking
	if err := db.First(&bkNow, bk.ID).Error; err != nil {
		t.Fatalf("reload booking: %v", err)
	}
	if bkNow.BookingStatus != constants.BookingStatusPriceConfirmed {
		t.Fatalf("expected price_confirmed, got %s", bkNow.BookingStatus)
	}

	// 4) PAY (simulated): awaiting_payment -> authorize -> capture -> paid_confirmed
	// (Real path: payment.CreatePayment -> Stripe webhook -> Capture. Stripe is unavailable in
	// this DB-only smoke test, so we drive the guarded transitions as the system would.)
	if err := svc.Advance(ctx, bk.ID, constants.BookingStatusAwaitingPayment, "payment opened"); err != nil {
		t.Fatalf("awaiting payment: %v", err)
	}
	if err := svc.Advance(ctx, bk.ID, constants.BookingStatusPaymentAuthorized, "auth simulated"); err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if err := svc.Advance(ctx, bk.ID, constants.BookingStatusPaidConfirmed, "capture simulated"); err != nil {
		t.Fatalf("capture: %v", err)
	}

	// 5) DOCS: user submits traveler documents -> docs_pending -> docs_submitted
	// first resolve the created traveler id
	var traveler bookingModel.BookingTraveler
	if err := db.Where("booking_id = ?", bk.ID).First(&traveler).Error; err != nil {
		t.Fatalf("traveler: %v", err)
	}
	if err := docs.SubmitDocuments(ctx, userID, bkReq.SubmitDocumentsReq{
		BookingID: bk.ID,
		Documents: []bkReq.TravelerDocReq{{TravelerID: traveler.ID, DocumentTypeID: 1, FileURL: "s3://passport.pdf", OriginalName: "passport.pdf", MIME: "application/pdf"}},
	}); err != nil {
		t.Fatalf("submit docs: %v", err)
	}

	// 6) KYC (admin approves) + agent approves docs
	if err := docs.AdminReviewKYC(ctx, 1, bkReq.KYCReviewReq{BookingID: bk.ID, Approve: true}); err != nil {
		t.Fatalf("kyc: %v", err)
	}
	var doc bookingModel.TravelerDocument
	if err := db.Where("booking_traveler_id = ?", traveler.ID).First(&doc).Error; err != nil {
		t.Fatalf("doc: %v", err)
	}
	if err := docs.ApproveDocument(ctx, agentID, bkReq.ApproveDocumentReq{DocumentID: doc.ID, Approve: true}); err != nil {
		t.Fatalf("approve doc: %v", err)
	}

	// 7) processing -> issued -> completed
	if err := docs.MarkProcessing(ctx, agentID, bk.ID); err != nil {
		t.Fatalf("mark processing: %v", err)
	}
	if err := docs.IssueBooking(ctx, agentID, bk.ID); err != nil {
		t.Fatalf("issue: %v", err)
	}
	if err := docs.CompleteBooking(ctx, agentID, bk.ID); err != nil {
		t.Fatalf("complete: %v", err)
	}

	// 8) verify terminal state + history written
	if err := db.First(&bkNow, bk.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if bkNow.BookingStatus != constants.BookingStatusCompleted {
		t.Fatalf("expected completed, got %s", bkNow.BookingStatus)
	}
	var history int64
	if err := db.Model(&bookingModel.BookingStatusHistory{}).Where("booking_id = ?", bk.ID).Count(&history).Error; err != nil {
		t.Fatalf("history: %v", err)
	}
	if history < 8 {
		t.Fatalf("expected >=8 status transitions, got %d", history)
	}

	// 9) SLA timer should be clear on terminal state
	var activeSLA int64
	_ = db.Model(&bookingModel.SLAEvent{}).Where("booking_id = ? AND active = ?", bk.ID, true).Count(&activeSLA).Error
	if activeSLA != 0 {
		t.Fatalf("expected no active SLA on completed booking, got %d", activeSLA)
	}

	// 10) payout: agent earnings aggregates exist (no crash); eligibility requires captured tx,
	// so we assert the read paths return without error.
	earn, err := pay.AgentEarnings(agentID)
	if err != nil {
		t.Fatalf("agent earnings: %v", err)
	}
	if earn.AgentID != agentID {
		t.Fatalf("earnings agent mismatch")
	}

	t.Logf("smoke ok: booking %s reached %s with %d history rows; earnings=%v",
		bk.BookingCode, bkNow.BookingStatus, history, earn)
}

func TestCancelIsTerminal(t *testing.T) {
	_, agentID, userID, pkgID := setupDB(t)
	var svc booking.BookingServices

	pkgIDptr := pkgID
	bk, err := svc.CreateBooking(userID, bkReq.CreateBookingReq{
		Type: constants.BookingTypePackage, PackageID: &pkgIDptr, TravelerCount: 1,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.CancelBooking(userID, bk.ID, "changed mind"); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	// a cancel from a terminal state must be rejected
	if err := svc.CancelBooking(userID, bk.ID, "again"); err == nil {
		t.Fatal("expected error cancelling a cancelled booking")
	}
	// agent cannot review a cancelled booking
	if err := svc.ReviewBooking(agentID, bk.ID, "review"); err == nil {
		t.Fatal("expected error reviewing a cancelled booking")
	}
}
