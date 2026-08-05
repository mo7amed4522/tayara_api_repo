package booking

import (
	"context"
	"errors"
	"fmt"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/booking/request"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

// DocServices owns the Docs / KYC workflow: per-traveler document submission and agent review,
// processing/issuance to COMPLETED, and admin KYC review. Every lifecycle change routes through
// the guarded booking.transition (same red lock as the rest of the booking module).
type DocServices struct{}

// SubmitDocuments records the user's uploaded traveler documents and advances the booking into the
// docs flow. Safe to resubmit after a correction.
func (d DocServices) SubmitDocuments(ctx context.Context, userID uint, req request.SubmitDocumentsReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var bk booking.Booking
	if err := global.GVA_DB.Where("id = ? AND user_id = ?", req.BookingID, userID).First(&bk).Error; err != nil {
		return errors.New("booking not found")
	}
	if !(bk.BookingStatus == constants.BookingStatusPaidConfirmed ||
		bk.BookingStatus == constants.BookingStatusDocsPending ||
		bk.BookingStatus == constants.BookingStatusDocsSubmitted ||
		bk.BookingStatus == constants.BookingStatusDocsPendingCorrection) {
		return errors.New("documents can only be submitted during the docs phase")
	}

	for _, doc := range req.Documents {
		if err := d.upsertTravelerDocument(ctx, bk.ID, doc); err != nil {
			return err
		}
	}

	// move the lifecycle forward through the legal edges
	switch bk.BookingStatus {
	case constants.BookingStatusPaidConfirmed:
		_ = d.transition(ctx, bk.ID, constants.BookingStatusDocsPending, constants.TimelineActorUser, userID, "documents phase started")
		fallthrough
	case constants.BookingStatusDocsPending, constants.BookingStatusDocsPendingCorrection:
		if err := d.transition(ctx, bk.ID, constants.BookingStatusDocsSubmitted, constants.TimelineActorUser, userID, "traveler documents submitted"); err != nil {
			return err
		}
	}
	return nil
}

// ApproveDocument approves a traveler document, or requests a correction (which drops the booking
// back to docs_pending_correction so the user can fix it). Agent scoping enforced.
func (d DocServices) ApproveDocument(ctx context.Context, agentID uint, req request.ApproveDocumentReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var doc booking.TravelerDocument
	if err := global.GVA_DB.First(&doc, req.DocumentID).Error; err != nil {
		return errors.New("document not found")
	}
	// resolve the owning booking through the traveler and enforce agent scoping
	var tr booking.BookingTraveler
	if err := global.GVA_DB.First(&tr, doc.BookingTravelerID).Error; err != nil {
		return errors.New("traveler not found")
	}
	var bk booking.Booking
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", tr.BookingID, agentID).First(&bk).Error; err != nil {
		return errors.New("agent does not own this booking")
	}

	status := constants.DocumentStatusApproved
	if !req.Approve {
		status = constants.DocumentStatusPendingCorrection
	}
	if err := global.GVA_DB.Model(&booking.TravelerDocument{}).Where("id = ?", doc.ID).
		Updates(map[string]interface{}{"doc_status": status, "review_note": req.Note}).Error; err != nil {
		return err
	}
	if !req.Approve {
		return d.transition(ctx, bk.ID, constants.BookingStatusDocsPendingCorrection, constants.TimelineActorAgent, agentID, "correction requested")
	}
	return nil
}

// MarkProcessing advances the booking to processing once docs are OK and KYC is approved.
func (d DocServices) MarkProcessing(ctx context.Context, agentID uint, bookingID uint) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var bk booking.Booking
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", bookingID, agentID).First(&bk).Error; err != nil {
		return errors.New("booking not found")
	}
	if bk.KYCStatus != constants.KYCStatusApproved {
		return errors.New("KYC must be approved before processing")
	}
	if bk.BookingStatus != constants.BookingStatusDocsSubmitted && bk.BookingStatus != constants.BookingStatusKycApproved {
		return errors.New("booking is not ready to process")
	}
	return d.transition(ctx, bookingID, constants.BookingStatusProcessing, constants.TimelineActorAgent, agentID, "documents approved, processing")
}

// IssueBooking advances a processing booking to issued_ready (tickets/confirmation issued).
func (d DocServices) IssueBooking(ctx context.Context, agentID uint, bookingID uint) error {
	return d.transition(ctx, bookingID, constants.BookingStatusIssuedReady, constants.TimelineActorAgent, agentID, "tickets issued")
}

// CompleteBooking advances the booking to completed via BookingConfirmed.
func (d DocServices) CompleteBooking(ctx context.Context, agentID uint, bookingID uint) error {
	if err := d.transition(ctx, bookingID, constants.BookingStatusBookingConfirmed, constants.TimelineActorAgent, agentID, "booking confirmed"); err != nil {
		return err
	}
	return d.transition(ctx, bookingID, constants.BookingStatusCompleted, constants.TimelineActorAgent, agentID, "booking completed")
}

// AdminReviewKYC sets the KYC identity axis; approval is required before processing.
func (d DocServices) AdminReviewKYC(ctx context.Context, adminID uint, req request.KYCReviewReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	status := constants.KYCStatusApproved
	if !req.Approve {
		status = constants.KYCStatusRejected
	}
	if err := global.GVA_DB.Model(&booking.Booking{}).
		Where("id = ?", req.BookingID).
		Updates(map[string]interface{}{"kyc_status": status}).Error; err != nil {
		return err
	}
	utils.LogAccess(global.GVA_DB, req.BookingID, nil, constants.TimelineActorAdmin, adminID, "kyc_review",
		fmt.Sprintf("KYC %s: %s", status, req.Note))
	return nil
}

// UploadBookingDocument attaches a booking-level document (agent confirmation / issued ticket / KYC file).
func (d DocServices) UploadBookingDocument(ctx context.Context, actorType string, actorID uint, req request.BookingDocumentReq) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	doc := &booking.BookingDocument{
		BookingID: req.BookingID, DocRole: req.DocRole, FileURL: req.FileURL,
		OriginalName: req.OriginalName, UploaderRole: actorType, UploadedBy: actorID, Status: "uploaded",
	}
	return global.GVA_DB.Create(doc).Error
}

// GetAgentBookingDocuments lists booking-level documents the agent owns; access is audited.
func (d DocServices) GetAgentBookingDocuments(ctx context.Context, agentID, bookingID uint) ([]booking.BookingDocument, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var bk booking.Booking
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", bookingID, agentID).First(&bk).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	var list []booking.BookingDocument
	if err := global.GVA_DB.Where("booking_id = ?", bookingID).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	for _, row := range list {
		id := row.ID
		utils.LogAccess(global.GVA_DB, bookingID, &id, constants.TimelineActorAgent, agentID, "view", "agent viewed booking document")
	}
	return list, nil
}

// GetAgentTravelerDocuments lists traveler documents for the agent's booking; access is audited.
func (d DocServices) GetAgentTravelerDocuments(ctx context.Context, agentID, bookingID uint) ([]booking.TravelerDocument, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var bk booking.Booking
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", bookingID, agentID).First(&bk).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	var travelers []booking.BookingTraveler
	if err := global.GVA_DB.Where("booking_id = ?", bookingID).Find(&travelers).Error; err != nil {
		return nil, err
	}
	var ids []uint
	for _, t := range travelers {
		ids = append(ids, t.ID)
	}
	var list []booking.TravelerDocument
	if len(ids) > 0 {
		if err := global.GVA_DB.Where("booking_traveler_id IN ?", ids).Order("created_at asc").Find(&list).Error; err != nil {
			return nil, err
		}
	}
	for _, row := range list {
		id := row.ID
		utils.LogAccess(global.GVA_DB, bookingID, &id, constants.TimelineActorAgent, agentID, "view", "agent viewed traveler document")
	}
	return list, nil
}

// GetUserDocuments lists a user's own traveler + booking documents; access is audited.
func (d DocServices) GetUserDocuments(ctx context.Context, userID, bookingID uint) ([]booking.TravelerDocument, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var bk booking.Booking
	if err := global.GVA_DB.Where("id = ? AND user_id = ?", bookingID, userID).First(&bk).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	var travelers []booking.BookingTraveler
	if err := global.GVA_DB.Where("booking_id = ?", bookingID).Find(&travelers).Error; err != nil {
		return nil, err
	}
	var ids []uint
	for _, t := range travelers {
		ids = append(ids, t.ID)
	}
	var list []booking.TravelerDocument
	if len(ids) > 0 {
		if err := global.GVA_DB.Where("booking_traveler_id IN ?", ids).Order("created_at asc").Find(&list).Error; err != nil {
			return nil, err
		}
	}
	for _, row := range list {
		id := row.ID
		utils.LogAccess(global.GVA_DB, bookingID, &id, constants.TimelineActorUser, userID, "view", "user viewed traveler document")
	}
	return list, nil
}

// upsertTravelerDocument inserts or updates a traveler's document (one per type).
func (d DocServices) upsertTravelerDocument(ctx context.Context, bookingID uint, doc request.TravelerDocReq) error {
	var tr booking.BookingTraveler
	if err := global.GVA_DB.First(&tr, doc.TravelerID).Error; err != nil || tr.BookingID != bookingID {
		return errors.New("traveler does not belong to this booking")
	}
	var existing booking.TravelerDocument
	err := global.GVA_DB.Where("booking_traveler_id = ? AND document_type_id = ?", doc.TravelerID, doc.DocumentTypeID).First(&existing).Error
	if err == nil {
		return global.GVA_DB.Model(&booking.TravelerDocument{}).Where("id = ?", existing.ID).
			Updates(map[string]interface{}{"file_url": doc.FileURL, "original_name": doc.OriginalName,
				"mime": doc.MIME, "doc_status": constants.DocumentStatusPending}).Error
	}
	return global.GVA_DB.Create(&booking.TravelerDocument{
		BookingTravelerID: doc.TravelerID, DocumentTypeID: doc.DocumentTypeID,
		FileURL: doc.FileURL, OriginalName: doc.OriginalName, MIME: doc.MIME,
		DocStatus: constants.DocumentStatusPending,
	}).Error
}

// transition is a thin wrapper so DocServices can drive the guarded booking state machine.
func (d DocServices) transition(ctx context.Context, bookingID uint, toStatus, actorType string, actorID uint, notes string) error {
	return (&BookingServices{}).transition(ctx, bookingID, actorID, toStatus, actorType, notes)
}

var _ = gorm.ErrRecordNotFound // keep gorm import available for future scoping helpers
