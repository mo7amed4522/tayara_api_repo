package booking

import (
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	bkReq "github.com/mo7amed4522/tayara_api/model/booking/request"
)

// DisputeServices owns the booking dispute lifecycle (open -> resolved/closed by admin).
type DisputeServices struct{}

// Create opens a dispute on a booking (raised by user or agent). Verifies the raiser owns it.
func (s DisputeServices) Create(raisedBy string, raiserID uint, req bkReq.CreateDisputeReq) (*bookingModel.Dispute, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var bk bookingModel.Booking
	if err := global.GVA_DB.First(&bk, req.BookingID).Error; err != nil {
		return nil, errors.New("booking not found")
	}
	if raisedBy == constants.TimelineActorUser && bk.UserID != raiserID {
		return nil, errors.New("user does not own this booking")
	}
	if raisedBy == constants.TimelineActorAgent && bk.AgentID != raiserID {
		return nil, errors.New("agent does not own this booking")
	}
	// a booking cannot have two open disputes
	var open int64
	_ = global.GVA_DB.Model(&bookingModel.Dispute{}).
		Where("booking_id = ? AND status = ?", req.BookingID, constants.DisputeStatusOpen).Count(&open).Error
	if open > 0 {
		return nil, errors.New("a dispute is already open for this booking")
	}
	dispute := &bookingModel.Dispute{
		BookingID: req.BookingID, RaisedBy: raisedBy, RaisedByID: raiserID,
		Status: constants.DisputeStatusOpen, Details: req.Details,
	}
	if err := global.GVA_DB.Create(dispute).Error; err != nil {
		return nil, err
	}
	return dispute, nil
}

// Resolve closes or resolves an open dispute (admin). Records resolution + acting admin.
func (s DisputeServices) Resolve(adminID uint, req bkReq.ResolveDisputeReq) (*bookingModel.Dispute, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var dispute bookingModel.Dispute
	if err := global.GVA_DB.First(&dispute, req.DisputeID).Error; err != nil {
		return nil, errors.New("dispute not found")
	}
	if dispute.Status != constants.DisputeStatusOpen {
		return nil, errors.New("only open disputes can be resolved")
	}
	status := req.Status
	if status == "" {
		status = constants.DisputeStatusResolved
	}
	if status != constants.DisputeStatusResolved && status != constants.DisputeStatusClosed {
		return nil, errors.New("invalid dispute status")
	}
	now := time.Now()
	if err := global.GVA_DB.Model(&bookingModel.Dispute{}).Where("id = ?", dispute.ID).Updates(map[string]interface{}{
		"status": status, "resolution": req.Resolution, "resolved_by": adminID, "resolved_at": &now,
	}).Error; err != nil {
		return nil, err
	}
	dispute.Status = status
	return &dispute, nil
}

// List returns disputes filtered by booking/status (admin).
func (s DisputeServices) List(req bkReq.DisputePageReq) ([]bookingModel.Dispute, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&bookingModel.Dispute{})
	if req.BookingID > 0 {
		db = db.Where("booking_id = ?", req.BookingID)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []bookingModel.Dispute
	if err := db.Scopes(req.Paginate()).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListForBooking returns disputes for a specific booking (user/agent scoped).
func (s DisputeServices) ListForBooking(bookingID uint) ([]bookingModel.Dispute, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var list []bookingModel.Dispute
	if err := global.GVA_DB.Where("booking_id = ?", bookingID).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
