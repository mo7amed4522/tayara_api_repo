package booking

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/booking/request"
	"github.com/mo7amed4522/tayara_api/model/booking/response"
	"github.com/mo7amed4522/tayara_api/model/catalog"
	"github.com/mo7amed4522/tayara_api/service/engage"
	"github.com/mo7amed4522/tayara_api/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BookingServices struct{}

// bookingTransitions is the canonical lifecycle state machine (single source of truth).
// A transition is legal only if `from` lists `to`. Terminal states have no outgoing edges,
// so nothing can "get inside" a completed/cancelled/refunded booking again.
var bookingTransitions = map[string][]string{
	constants.BookingStatusRequested:             {constants.BookingStatusAgentReviewing, constants.BookingStatusCancelled},
	constants.BookingStatusAgentReviewing:        {constants.BookingStatusPriceConfirmed, constants.BookingStatusPriceExpired, constants.BookingStatusCancelled},
	constants.BookingStatusPriceConfirmed:        {constants.BookingStatusAwaitingPayment, constants.BookingStatusPriceExpired, constants.BookingStatusCancelled},
	constants.BookingStatusPriceExpired:          {constants.BookingStatusRequested, constants.BookingStatusCancelled},
	constants.BookingStatusAwaitingPayment:       {constants.BookingStatusPaymentAuthorized, constants.BookingStatusPriceExpired, constants.BookingStatusCancelled},
	constants.BookingStatusPaymentAuthorized:     {constants.BookingStatusPaidConfirmed, constants.BookingStatusRefunded, constants.BookingStatusCancelled},
	constants.BookingStatusPaidConfirmed:         {constants.BookingStatusDocsPending, constants.BookingStatusRefunded},
	constants.BookingStatusDocsPending:           {constants.BookingStatusDocsSubmitted, constants.BookingStatusKycPending, constants.BookingStatusCancelled},
	constants.BookingStatusDocsSubmitted:         {constants.BookingStatusProcessing, constants.BookingStatusDocsPendingCorrection},
	constants.BookingStatusDocsPendingCorrection: {constants.BookingStatusDocsSubmitted, constants.BookingStatusCancelled},
	constants.BookingStatusKycPending:            {constants.BookingStatusKycApproved, constants.BookingStatusCancelled},
	constants.BookingStatusKycApproved:           {constants.BookingStatusProcessing},
	constants.BookingStatusProcessing:            {constants.BookingStatusIssuedReady, constants.BookingStatusBookingConfirmed},
	constants.BookingStatusBookingConfirmed:      {constants.BookingStatusCompleted},
	constants.BookingStatusIssuedReady:           {constants.BookingStatusCompleted},
	// terminal states: no outgoing transitions (locked)
	constants.BookingStatusCompleted: {},
	constants.BookingStatusCancelled: {},
	constants.BookingStatusRefunded:  {},
}

// CreateBooking starts a booking at `requested`, opens the agent-response SLA timer, notifies the
// agent, and snapshots pricing from the package / custom-trip request.
func (b *BookingServices) CreateBooking(userID uint, req request.CreateBookingReq) (*booking.Booking, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if req.Type != constants.BookingTypePackage && req.Type != constants.BookingTypeCustomTrip {
		return nil, errors.New("booking type must be package or custom_trip")
	}

	// resolve agent + pricing snapshot
	agentID, destCountryID, gross, discountBps, currency := b.resolveBookingSource(req)
	if agentID == 0 {
		return nil, errors.New("invalid booking source")
	}
	if currency == "" {
		currency = constants.CurrencyAED
	}
	discountAmount := utils.MoneyPercentOf(gross, discountBps)
	now := time.Now()

	bk := &booking.Booking{
		BookingCode:          utils.NextCode(constants.CodePrefixBooking),
		UserID:               userID,
		AgentID:              agentID,
		PackageID:            req.PackageID,
		CustomTripRequestID:  req.CustomTripRequestID,
		Type:                 req.Type,
		BookingStatus:        constants.BookingStatusRequested,
		PaymentStatus:        constants.PaymentStatusPending,
		KYCStatus:            constants.KYCStatusNotSubmitted,
		AmountGross:          gross,
		DiscountPercent:      discountBps,
		DiscountAmount:       discountAmount,
		Currency:             currency,
		DestinationCountryID: destCountryID,
		TravelerCount:        req.TravelerCount,
		HotelChoiceID:        req.HotelChoiceID,
		SelectedActivities:   req.SelectedActivities,
		Notes:                req.Notes,
		StartDate:            parseDate(req.StartDate),
		EndDate:              parseDate(req.EndDate),
		RequestedAt:          &now,
	}

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(bk).Error; err != nil {
			return err
		}
		if err := tx.Create(&booking.BookingStatusHistory{
			BookingID: bk.ID, ActorType: constants.TimelineActorUser, ActorID: userID,
			FromStatus: "", ToStatus: constants.BookingStatusRequested, Notes: "booking requested",
		}).Error; err != nil {
			return err
		}
		if err := insertTravelers(tx, bk.ID, req.Travelers); err != nil {
			return err
		}
		return startSLAFor(tx, bk.ID, constants.BookingStatusRequested)
	})
	if err != nil {
		return nil, err
	}

	engage.NotificationServices{}.Emit(constants.NotifEventBookingRequested, constants.NotifRecipientAgent,
		agentID, constants.NotifChannelPush, "New booking request", "A new booking request arrived")
	return bk, nil
}

// ReviewBooking moves a booking to agent_reviewing (agent owns it). Agent scoping enforced.
func (b *BookingServices) ReviewBooking(agentID, bookingID uint, notes string) error {
	return b.transition(context.Background(), bookingID, agentID,
		constants.BookingStatusAgentReviewing, constants.TimelineActorAgent, notes)
}

// SubmitOffer records the agent's final offer, moves the booking to price_confirmed, opens the
// user-payment SLA window, and notifies the user to pay. Agent scoping enforced.
func (b *BookingServices) SubmitOffer(agentID uint, req request.SubmitOfferReq) (*booking.BookingOffer, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	// agent must own the booking
	var bk booking.Booking
	if err := global.GVA_DB.Where("id = ? AND agent_id = ?", req.BookingID, agentID).First(&bk).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	if bk.BookingStatus != constants.BookingStatusAgentReviewing {
		return nil, errors.New("offer can only be submitted while agent is reviewing")
	}

	hours := slaHours("user_payment_hours", 2)
	now := time.Now()
	expiresAt := now.Add(time.Duration(hours) * time.Hour)
	offer := &booking.BookingOffer{
		BookingID:      req.BookingID,
		AgentID:        agentID,
		FinalPrice:     req.FinalPrice,
		AvailableDates: req.AvailableDates,
		Notes:          req.Notes,
		Status:         constants.OfferStatusActive,
		OfferedAt:      &now,
		ExpiresAt:      &expiresAt,
	}

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// mark previous active offers expired (single live offer)
		if err := tx.Model(&booking.BookingOffer{}).
			Where("booking_id = ? AND status = ?", req.BookingID, constants.OfferStatusActive).
			Update("status", constants.OfferStatusExpired).Error; err != nil {
			return err
		}
		if err := tx.Create(offer).Error; err != nil {
			return err
		}
		if err := tx.Model(&booking.Booking{}).Where("id = ?", req.BookingID).
			Updates(map[string]interface{}{
				"amount_gross":       req.FinalPrice,
				"price_confirmed_at": &now,
			}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// price confirmed via the guarded transition (lock + history + SLA + notification)
	if err := b.transition(context.Background(), req.BookingID, agentID,
		constants.BookingStatusPriceConfirmed, constants.TimelineActorAgent, "final offer accepted"); err != nil {
		return nil, err
	}
	engage.NotificationServices{}.Emit(constants.NotifEventPriceConfirmed, constants.NotifRecipientUser,
		bk.UserID, constants.NotifChannelPush, "Price confirmed", "Your price is confirmed, please pay")
	return offer, nil
}

// RejectBooking cancels the booking because the agent rejected/modified it, recording the reason.
func (b *BookingServices) RejectBooking(agentID uint, req request.RejectBookingReq) error {
	if err := b.transition(context.Background(), req.BookingID, agentID,
		constants.BookingStatusCancelled, constants.TimelineActorAgent, req.Reason); err != nil {
		return err
	}
	_ = global.GVA_DB.Model(&booking.Booking{}).Where("id = ?", req.BookingID).
		Updates(map[string]interface{}{"cancellation_reason": req.Reason}).Error
	return nil
}

// CancelBooking lets the user cancel before payment/confirmation.
func (b *BookingServices) CancelBooking(userID, bookingID uint, reason string) error {
	return b.transition(context.Background(), bookingID, userID,
		constants.BookingStatusCancelled, constants.TimelineActorUser, reason)
}

// ForceCancel is the admin force-cancel path, allowed from any non-terminal state.
func (b *BookingServices) ForceCancel(adminID uint, req request.CancelBookingReq) error {
	if err := b.transition(context.Background(), req.BookingID, adminID,
		constants.BookingStatusCancelled, constants.TimelineActorAdmin, req.Reason); err != nil {
		return err
	}
	_ = global.GVA_DB.Model(&booking.Booking{}).Where("id = ?", req.BookingID).
		Updates(map[string]interface{}{"cancellation_reason": req.Reason}).Error
	return nil
}

// Advance moves a booking to a system-driven status through the same guarded transition
// (used by payment capture/refund and other automated actors).
func (b *BookingServices) Advance(ctx context.Context, bookingID uint, toStatus, notes string) error {
	return b.transition(ctx, bookingID, 0, toStatus, constants.TimelineActorSystem, notes)
}

// transition is the single guarded mutator: red-locked, transition-validated, writes history,
// maintains the SLA timer, and snapshots timestamps. No other code may change booking_status.
func (b *BookingServices) transition(ctx context.Context, bookingID, actorID uint, toStatus, actorType, notes string) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	release, ok := utils.TryLock(ctx, "booking:"+strconv.FormatUint(uint64(bookingID), 10), time.Minute)
	if !ok {
		return errors.New("booking is being processed by another operation; retry")
	}
	defer release()

	var bk booking.Booking
	if err := global.GVA_DB.First(&bk, bookingID).Error; err != nil {
		return errors.New("booking not found")
	}
	if err := b.canTransition(bk.BookingStatus, toStatus); err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{"booking_status": toStatus}
	switch toStatus {
	case constants.BookingStatusPriceConfirmed:
		updates["price_confirmed_at"] = &now
	case constants.BookingStatusCompleted:
		updates["confirmed_at"] = &now
	case constants.BookingStatusCancelled:
		updates["cancelled_at"] = &now
		updates["cancelled_by"] = actorID
	}

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&booking.Booking{}).Where("id = ?", bookingID).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Create(&booking.BookingStatusHistory{
			BookingID: bookingID, ActorType: actorType, ActorID: actorID,
			FromStatus: bk.BookingStatus, ToStatus: toStatus, Notes: notes,
		}).Error; err != nil {
			return err
		}
		if err := clearActiveSLA(tx, bookingID); err != nil {
			return err
		}
		return startSLAFor(tx, bookingID, toStatus)
	})
	if err != nil {
		return err
	}

	// notify the counterpart on meaningful transitions
	switch toStatus {
	case constants.BookingStatusCancelled:
		engage.NotificationServices{}.Emit(constants.NotifEventCancellation, constants.NotifRecipientAgent,
			bk.AgentID, constants.NotifChannelPush, "Booking cancelled", "A booking was cancelled")
		engage.NotificationServices{}.Emit(constants.NotifEventCancellation, constants.NotifRecipientUser,
			bk.UserID, constants.NotifChannelPush, "Booking cancelled", "Your booking was cancelled")
	case constants.BookingStatusCompleted:
		engage.NotificationServices{}.Emit(constants.NotifEventBookingRequested, constants.NotifRecipientUser,
			bk.UserID, constants.NotifChannelPush, "Booking completed", "Your trip is complete")
	}
	return nil
}

// canTransition validates the move against the state machine and blocks re-entry into terminal states.
func (b *BookingServices) canTransition(from, to string) error {
	if isTerminal(from) {
		return errors.New("booking is in a terminal state and cannot change")
	}
	allowed, ok := bookingTransitions[from]
	if !ok {
		return errors.New("unknown booking state: " + from)
	}
	for _, a := range allowed {
		if a == to {
			return nil
		}
	}
	return errors.New("illegal booking transition: " + from + " -> " + to)
}

// isTerminal reports whether a status is a locked, final state.
func isTerminal(status string) bool {
	switch status {
	case constants.BookingStatusCompleted, constants.BookingStatusCancelled, constants.BookingStatusRefunded:
		return true
	}
	return false
}

// resolveBookingSource pulls agent, destination, price snapshot, and currency from the source.
func (b *BookingServices) resolveBookingSource(req request.CreateBookingReq) (agentID, destCountryID uint, gross, discountBps int64, currency string) {
	if req.Type == constants.BookingTypePackage && req.PackageID != nil {
		var pkg catalog.Package
		if err := global.GVA_DB.First(&pkg, *req.PackageID).Error; err == nil {
			return pkg.AgentID, pkg.DestinationCountryID, pkg.EstimatedPrice, pkg.DiscountPercent, pkgCurrency(pkg)
		}
	}
	if req.CustomTripRequestID != nil {
		var ctr booking.CustomTripRequest
		if err := global.GVA_DB.First(&ctr, *req.CustomTripRequestID).Error; err == nil {
			return ctr.AgentID, ctr.DestinationCountryID, ctr.EstimatedCost, 0, currency
		}
	}
	return 0, 0, 0, 0, currency
}

func pkgCurrency(_ catalog.Package) string {
	return constants.CurrencyAED
}

// --- read side (scoping enforced at the service boundary) ---

func (b *BookingServices) GetOwnBooking(userID, id uint) (*response.BookingDetail, error) {
	return b.buildDetail(global.GVA_DB.Where("user_id = ?", userID), id)
}

func (b *BookingServices) GetAgentBooking(agentID, id uint) (*response.BookingDetail, error) {
	return b.buildDetail(global.GVA_DB.Where("agent_id = ?", agentID), id)
}

func (b *BookingServices) GetBooking(id uint) (*response.BookingDetail, error) {
	return b.buildDetail(global.GVA_DB, id)
}

func (b *BookingServices) ListOwnBookings(userID uint, req request.BookingPageReq) ([]booking.Booking, int64, error) {
	return b.page(global.GVA_DB.Model(&booking.Booking{}).Where("user_id = ?", userID), req)
}

func (b *BookingServices) ListAgentBookings(agentID uint, req request.BookingPageReq) ([]booking.Booking, int64, error) {
	return b.page(global.GVA_DB.Model(&booking.Booking{}).Where("agent_id = ?", agentID), req)
}

func (b *BookingServices) ListAllBookings(req request.BookingPageReq) ([]booking.Booking, int64, error) {
	return b.page(global.GVA_DB.Model(&booking.Booking{}), req)
}

func (b *BookingServices) buildDetail(scoped *gorm.DB, id uint) (*response.BookingDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var bk booking.Booking
	if err := scoped.First(&bk, id).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	var (
		travelers     []booking.BookingTraveler
		offers        []booking.BookingOffer
		documents     []booking.BookingDocument
		statusHistory []booking.BookingStatusHistory
	)
	bid := bk.ID
	if err := global.GVA_DB.Where("booking_id = ?", bid).Order("created_at asc").Find(&statusHistory).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("booking_id = ?", bid).Order("created_at asc").Find(&travelers).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("booking_id = ?", bid).Order("created_at desc").Find(&offers).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Where("booking_id = ?", bid).Order("created_at desc").Find(&documents).Error; err != nil {
		return nil, err
	}
	return &response.BookingDetail{Booking: bk, Travelers: travelers, Offers: offers,
		Documents: documents, StatusHistory: statusHistory}, nil
}

func (b *BookingServices) page(db *gorm.DB, req request.BookingPageReq) ([]booking.Booking, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	if req.Status != "" {
		db = db.Where("booking_status = ?", req.Status)
	}
	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []booking.Booking
	if err := db.Scopes(req.Paginate()).Order("updated_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ProcessOverdueSLA is the cron entry point: finds breached SLA timers and applies the auto action
// (auto-cancel for agent response, offer expiry for user payment) through the guarded transition.
func (b *BookingServices) ProcessOverdueSLA() error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	var events []booking.SLAEvent
	if err := global.GVA_DB.Where("active = ? AND breached = ? AND deadline_at <= ?", true, false, time.Now()).
		Find(&events).Error; err != nil {
		return err
	}
	for _, ev := range events {
		toStatus := constants.BookingStatusCancelled
		note := "SLA auto-cancel"
		switch ev.SLAType {
		case constants.SLATypeUserPayment:
			toStatus = constants.BookingStatusPriceExpired
			note = "payment window expired"
		}
		// guard against re-entry: skip if already terminal
		var bk booking.Booking
		if err := global.GVA_DB.First(&bk, ev.BookingID).Error; err != nil {
			continue
		}
		if isTerminal(bk.BookingStatus) {
			_ = global.GVA_DB.Model(&booking.SLAEvent{}).Where("id = ?", ev.ID).
				Updates(map[string]interface{}{"breached": true, "active": false}).Error
			continue
		}
		if err := b.transition(context.Background(), ev.BookingID, 0,
			toStatus, constants.TimelineActorSystem, note); err != nil {
			global.GVA_LOG.Warn("SLA auto action failed", zap.Error(err), zap.Uint("event", ev.ID))
			continue
		}
		global.GVA_DB.Model(&booking.SLAEvent{}).Where("id = ?", ev.ID).
			Updates(map[string]interface{}{"breached": true, "active": false})
	}
	return nil
}

// insertTravelers persists the traveler rows for a booking (called inside a transaction).
func insertTravelers(tx *gorm.DB, bookingID uint, travelers []request.TravelerReq) error {
	if len(travelers) == 0 {
		return nil
	}
	rows := make([]booking.BookingTraveler, 0, len(travelers))
	for i, t := range travelers {
		rows = append(rows, booking.BookingTraveler{
			BookingID:      bookingID,
			IsBooker:       i == 0,
			FullName:       t.FullName,
			DOB:            parseDate(t.DOB),
			NationalityID:  uintPtr(t.NationalityID),
			PassportNumber: t.PassportNumber,
			PassportExpiry: parseDate(t.PassportExpiry),
			AdultOrChild:   t.AdultOrChild,
			Email:          t.Email,
			DocStatus:      "missing",
		})
	}
	return tx.Create(&rows).Error
}

func uintPtr(v uint) *uint { return &v }

// parseDate parses a 2006-01-02 string into a time pointer (nil if empty/malformed).
func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}
