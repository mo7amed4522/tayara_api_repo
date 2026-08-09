package payment

import (
	"context"
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	paymentModel "github.com/mo7amed4522/tayara_api/model/payment"
	paymentReq "github.com/mo7amed4522/tayara_api/model/payment/request"
	paymentResp "github.com/mo7amed4522/tayara_api/model/payment/response"
	"github.com/mo7amed4522/tayara_api/service/engage"
	"github.com/mo7amed4522/tayara_api/utils"
)

// eligibleForPayout evaluates the four payout gates from the PRD: payment captured AND agent
// uploaded a confirmation file AND admin verified AND booking not cancelled/refunded.
func (p *PaymentServices) eligibleForPayout(bk *bookingModel.Booking, tx *paymentModel.PaymentTransaction) bool {
	if bk == nil || tx == nil {
		return false
	}
	if tx.Status != constants.PaymentLifecycleCaptured {
		return false
	}
	if bk.BookingStatus == constants.BookingStatusCancelled || bk.BookingStatus == constants.BookingStatusRefunded {
		return false
	}
	// agent confirmation file gate (configurable)
	if global.GVA_CONFIG.Travel.Payout.RequireConfirmationFile {
		var n int64
		_ = global.GVA_DB.Model(&bookingModel.BookingDocument{}).
			Where("booking_id = ? AND doc_role = ? AND status = ?", bk.ID, constants.DocRoleAgentConfirmation, "uploaded").
			Count(&n).Error
		if n == 0 {
			return false
		}
	}
	// admin verification gate (configurable)
	if global.GVA_CONFIG.Travel.Payout.RequireAdminVerification {
		if bk.AdminNotes == "" {
			return false
		}
	}
	return true
}

// PreparePayouts scans captured bookings and creates/updates a payout record per eligible booking
// (pending) and flags ineligible ones (not_eligible). Called by admin or on-demand.
func (p *PaymentServices) PreparePayouts() (created int, err error) {
	if global.GVA_DB == nil {
		return 0, errors.New("db not init")
	}
	var bookings []bookingModel.Booking
	if err := global.GVA_DB.Where("booking_status NOT IN ?", []string{
		constants.BookingStatusCancelled, constants.BookingStatusRefunded,
	}).Find(&bookings).Error; err != nil {
		return 0, err
	}
	now := time.Now()
	for i := range bookings {
		bk := &bookings[i]
		tx, terr := p.paymentByBooking(bk.ID)
		if terr != nil {
			continue
		}
		var payout paymentModel.Payout
		errGet := global.GVA_DB.Where("booking_id = ?", bk.ID).First(&payout).Error
		if errGet == nil {
			continue // already prepared
		}
		status := constants.PayoutStatusNotEligible
		if p.eligibleForPayout(bk, tx) {
			status = constants.PayoutStatusPending
		}
		row := &paymentModel.Payout{
			BookingID: bk.ID, AgentID: bk.AgentID, AgentShare: tx.AgentShare,
			PayoutStatus: status, Method: constants.PayoutMethodManual,
			RequestedAt: &now,
		}
		if err := global.GVA_DB.Create(row).Error; err != nil {
			continue
		}
		created++
	}
	return created, nil
}

// PayAgent is the admin manual payout action: it pays an eligible pending payout. Runs under the
// payment lifecycle red lock so payout cannot race capture/refund. Reference + notes are recorded.
func (p *PaymentServices) PayAgent(ctx context.Context, adminID uint, req paymentReq.PayAgentReq) (*paymentResp.PayoutDetail, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var payout paymentModel.Payout
	if err := global.GVA_DB.Where("booking_id = ?", req.BookingID).First(&payout).Error; err != nil {
		return nil, errors.New("payout not found")
	}
	if payout.PayoutStatus != constants.PayoutStatusPending {
		return nil, errors.New("only pending payouts can be paid")
	}
	release, err := lockPayment(ctx, req.BookingID)
	if err != nil {
		return nil, err
	}
	defer release()

	// re-verify eligibility under the lock (booking may have moved)
	var bk bookingModel.Booking
	if err := global.GVA_DB.First(&bk, req.BookingID).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	tx, terr := p.paymentByBooking(req.BookingID)
	if terr != nil {
		return nil, errors.New("payment transaction not found")
	}
	if !p.eligibleForPayout(&bk, tx) {
		return nil, errors.New("booking is no longer eligible for payout")
	}

	method := req.Method
	if method == "" {
		method = constants.PayoutMethodManual
	}
	now := time.Now()
	if err := global.GVA_DB.Model(&paymentModel.Payout{}).Where("id = ?", payout.ID).Updates(map[string]interface{}{
		"payout_status": constants.PayoutStatusPaid, "method": method,
		"reference_number": req.ReferenceNumber, "admin_notes": req.Notes,
		"paid_at": &now, "paid_by": adminID,
	}).Error; err != nil {
		return nil, err
	}
	engage.NotificationServices{}.Emit(constants.NotifEventPayout, constants.NotifRecipientAgent,
		bk.AgentID, constants.NotifChannelPush, "Payout released", "Your payout was released")
	return p.payoutDetail(req.BookingID)
}

// --- read side ---

func (p *PaymentServices) payoutDetail(bookingID uint) (*paymentResp.PayoutDetail, error) {
	var payout paymentModel.Payout
	if err := global.GVA_DB.Where("booking_id = ?", bookingID).First(&payout).Error; err != nil {
		return nil, err
	}
	tx, err := p.paymentByBooking(bookingID)
	if err != nil {
		return &paymentResp.PayoutDetail{Payout: payout}, nil
	}
	return &paymentResp.PayoutDetail{
		Payout: payout, TransactionAmount: tx.Amount, FeeAmount: tx.FeeAmount,
		CommissionAmount: tx.CommissionAmount, Currency: tx.Currency, PaidAt: payout.PaidAt,
	}, nil
}

// ListPayouts lists payouts (admin: all; agent: scoped by agentID).
func (p *PaymentServices) ListPayouts(agentID uint, req paymentReq.PayoutPageReq) ([]paymentModel.Payout, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&paymentModel.Payout{})
	if agentID > 0 {
		db = db.Where("agent_id = ?", agentID)
	}
	if req.AgentID > 0 {
		db = db.Where("agent_id = ?", req.AgentID)
	}
	if req.Status != "" {
		db = db.Where("payout_status = ?", req.Status)
	}
	var list []paymentModel.Payout
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("updated_at desc"), req.Page, req.PageSize, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}

// AgentEarnings returns the agent's dashboard earnings aggregation.
func (p *PaymentServices) AgentEarnings(agentID uint) (*paymentResp.AgentEarnings, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	out := &paymentResp.AgentEarnings{AgentID: agentID}

	// total earned = sum of agent shares on captured payments (the booking's agent)
	_ = global.GVA_DB.Model(&paymentModel.PaymentTransaction{}).
		Where("agent_id = ? AND status = ?", agentID, constants.PaymentLifecycleCaptured).
		Select("COALESCE(SUM(agent_share),0)").
		Scan(&out.TotalEarned).Error
	// pending payouts
	_ = global.GVA_DB.Model(&paymentModel.Payout{}).
		Where("agent_id = ? AND payout_status = ?", agentID, constants.PayoutStatusPending).
		Select("COALESCE(SUM(agent_share),0)").
		Scan(&out.TotalPendingPayouts).Error
	// paid out
	_ = global.GVA_DB.Model(&paymentModel.Payout{}).
		Where("agent_id = ? AND payout_status = ?", agentID, constants.PayoutStatusPaid).
		Select("COALESCE(SUM(agent_share),0)").
		Scan(&out.TotalPaidOut).Error

	var pending []paymentModel.Payout
	_ = global.GVA_DB.Where("agent_id = ? AND payout_status = ?", agentID, constants.PayoutStatusPending).
		Order("created_at desc").Limit(20).Find(&pending).Error
	for _, po := range pending {
		if d, err := p.payoutDetail(po.BookingID); err == nil {
			out.PendingPayouts = append(out.PendingPayouts, *d)
		}
	}
	var recent []paymentModel.Payout
	_ = global.GVA_DB.Where("agent_id = ? AND payout_status = ?", agentID, constants.PayoutStatusPaid).
		Order("paid_at desc").Limit(20).Find(&recent).Error
	for _, po := range recent {
		if d, err := p.payoutDetail(po.BookingID); err == nil {
			out.RecentPayouts = append(out.RecentPayouts, *d)
		}
	}
	return out, nil
}
