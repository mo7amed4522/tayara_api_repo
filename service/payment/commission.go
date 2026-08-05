package payment

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	paymentModel "github.com/mo7amed4522/tayara_api/model/payment"
	paymentReq "github.com/mo7amed4522/tayara_api/model/payment/request"
	"github.com/mo7amed4522/tayara_api/utils"
)

// createCommissionRecord writes the per-booking commission snapshot at capture. Called from the
// payment Capture path under the same lifecycle red lock, so it is never missed.
func (p *PaymentServices) createCommissionRecord(tx *paymentModel.PaymentTransaction, planType string) error {
	if global.GVA_DB == nil {
		return errors.New("db not init")
	}
	rec := &paymentModel.CommissionRecord{
		BookingID:         tx.BookingID,
		AgentID:           tx.AgentID,
		PlanType:          planType,
		GrossAmount:       tx.Amount,
		StripeFee:         tx.FeeAmount,
		CommissionPercent: tx.CommissionPercent,
		CommissionAmount:  tx.CommissionAmount,
		AgentShare:        tx.AgentShare,
		Status:            constants.CommissionStatusCaptured,
	}
	return global.GVA_DB.Create(rec).Error
}

// OverrideCommission lets admin change the commission % on a booking; the reason + admin are
// recorded, totals recalculated, and the change is persisted as an override (audited).
func (p *PaymentServices) OverrideCommission(adminID uint, req paymentReq.CommissionOverrideReq) (*paymentModel.CommissionRecord, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if req.Reason == "" {
		return nil, errors.New("a reason is required to override commission")
	}
	var rec paymentModel.CommissionRecord
	if err := global.GVA_DB.Where("booking_id = ?", req.BookingID).First(&rec).Error; err != nil {
		return nil, errors.New("commission record not found")
	}
	// recompute on the captured gross
	commissionAmount := utils.MoneyPercentOf(rec.GrossAmount, req.CommissionPercent)
	agentShare := rec.GrossAmount - rec.StripeFee - commissionAmount
	updates := map[string]interface{}{
		"commission_percent": req.CommissionPercent,
		"commission_amount":  commissionAmount,
		"agent_share":        agentShare,
		"is_override":        true,
		"override_reason":    req.Reason,
		"override_by":        adminID,
	}
	if err := global.GVA_DB.Model(&paymentModel.CommissionRecord{}).Where("id = ?", rec.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	// keep the booking + payment transaction snapshots in sync
	_ = global.GVA_DB.Model(&bookingModel.Booking{}).Where("id = ?", rec.BookingID).Updates(map[string]interface{}{
		"commission_percent": req.CommissionPercent, "commission_amount": commissionAmount, "agent_share": agentShare,
	}).Error
	_ = global.GVA_DB.Model(&paymentModel.PaymentTransaction{}).Where("booking_id = ?", rec.BookingID).Updates(map[string]interface{}{
		"commission_percent": req.CommissionPercent, "commission_amount": commissionAmount, "agent_share": agentShare,
	}).Error
	return &rec, nil
}

// ListCommission returns commission records (admin: all; agent: scoped by agentID).
func (p *PaymentServices) ListCommission(agentID uint, req paymentReq.CommissionPageReq) ([]paymentModel.CommissionRecord, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&paymentModel.CommissionRecord{})
	if agentID > 0 {
		db = db.Where("agent_id = ?", agentID)
	}
	if req.AgentID > 0 {
		db = db.Where("agent_id = ?", req.AgentID)
	}
	if req.BookingID > 0 {
		db = db.Where("booking_id = ?", req.BookingID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []paymentModel.CommissionRecord
	if err := db.Scopes(req.Paginate()).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
