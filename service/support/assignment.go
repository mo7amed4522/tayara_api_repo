package support

import (
	"context"
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	"gorm.io/gorm"
)

// contextBackground is a tiny indirection so tests can stub it if needed.
func contextBackground() context.Context { return context.Background() }

// ReviewServices appends quality verdicts without ever mutating message history.
type ReviewServices struct{}

// Review records a verdict (append-only) and flips the ticket status.
func (s ReviewServices) Review(actor Actor, ticketID uint, req supportReq.ReviewTicketReq) (*supportModel.Review, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if actor.Type != constants.ActorQuality && actor.Type != constants.ActorAdmin {
		return nil, response.ForbiddenErr("quality or admin only")
	}
	if req.Verdict != constants.SupportVerdictSolved && req.Verdict != constants.SupportVerdictNotSolved {
		return nil, response.ValidationErr("invalid verdict")
	}
	var tk supportModel.Ticket
	if err := global.GVA_DB.First(&tk, ticketID).Error; err != nil {
		return nil, response.NotFoundErr("ticket not found")
	}
	var conv supportModel.Conversation
	if err := global.GVA_DB.Where("ticket_id = ?", tk.ID).First(&conv).Error; err != nil {
		return nil, response.NotFoundErr("conversation not found")
	}
	var review supportModel.Review
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		review = supportModel.Review{
			TicketID:             tk.ID,
			ConversationID:       conv.ID,
			ReviewerID:           actor.ID,
			Verdict:              req.Verdict,
			Notes:                req.Notes,
			MessageCountAtReview: conv.MessageCount,
		}
		if err := tx.Create(&review).Error; err != nil {
			return err
		}
		return tx.Model(&supportModel.Ticket{}).Where("id = ?", tk.ID).Updates(map[string]interface{}{
			"status": req.Verdict,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	publishTicketEvent("ticket.reviewed", conv.ID, tk.ID, uintString(tk.ID), map[string]any{
		"ticket_id": tk.ID, "verdict": review.Verdict, "reviewer_id": actor.ID,
	})
	return &review, nil
}

// List returns a ticket's review trail (append-only, newest last).
func (s ReviewServices) List(ticketID uint) ([]supportModel.Review, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var rows []supportModel.Review
	err := global.GVA_DB.Where("ticket_id = ?", ticketID).Order("id asc").Find(&rows).Error
	return rows, err
}

// AssignmentServices appends assignment history and mirrors the assignee onto ticket + conversation.
type AssignmentServices struct{}

// Assign reassigns a ticket to an existing staff account.
func (s AssignmentServices) Assign(actor Actor, ticketID uint, req supportReq.AssignTicketReq) (*supportModel.AssignmentHistory, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	switch actor.Type {
	case constants.ActorAdmin, constants.ActorSupport, constants.ActorQuality:
	default:
		return nil, response.ForbiddenErr("staff only")
	}
	if global.GVA_CONFIG.Support.Chat.Assignment.RequireReason && req.Reason == "" {
		return nil, response.ValidationErr("reason is required")
	}
	var emp struct{ ID uint }
	if err := global.GVA_DB.Table("sys_admin_profile").Select("id").
		Where("id = ?", req.EmployeeID).Take(&emp).Error; err != nil {
		return nil, response.ValidationErr("target employee not found")
	}
	authority := req.AuthorityID
	if authority == 0 {
		authority = constants.AuthoritySupport
	}
	if authority != constants.AuthoritySupport && authority != constants.AuthorityQuality && authority != constants.AuthoritySystem {
		return nil, response.ValidationErr("invalid target authority")
	}
	var tk supportModel.Ticket
	if err := global.GVA_DB.First(&tk, ticketID).Error; err != nil {
		return nil, response.NotFoundErr("ticket not found")
	}
	newID := emp.ID
	history := supportModel.AssignmentHistory{
		TicketID:       tk.ID,
		FromEmployeeID: tk.AssignedEmployeeID,
		ToEmployeeID:   &newID,
		AssignedBy:     actor.ID,
		Reason:         req.Reason,
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		if err := tx.Model(&supportModel.Ticket{}).Where("id = ?", tk.ID).Updates(map[string]interface{}{
			"assigned_employee_id":  newID,
			"assigned_authority_id": authority,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&supportModel.Conversation{}).Where("ticket_id = ?", tk.ID).
			Update("assigned_employee_id", newID).Error
	})
	if err != nil {
		return nil, err
	}
	var conv supportModel.Conversation
	_ = global.GVA_DB.Where("ticket_id = ?", tk.ID).First(&conv).Error
	publishTicketEvent("ticket.assigned", conv.ID, tk.ID, uintString(tk.ID), map[string]any{
		"ticket_id": tk.ID, "employee_id": newID, "assigned_by": actor.ID,
	})
	return &history, nil
}

// History returns a ticket's assignment trail.
func (s AssignmentServices) History(ticketID uint) ([]supportModel.AssignmentHistory, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var rows []supportModel.AssignmentHistory
	err := global.GVA_DB.Where("ticket_id = ?", ticketID).Order("id asc").Find(&rows).Error
	return rows, err
}

// touch is reserved for future escalation logic.
func touchAssignment(now time.Time) *time.Time { return &now }
