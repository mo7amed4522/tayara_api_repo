package support

import (
	"errors"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	bookingModel "github.com/mo7amed4522/tayara_api/model/booking"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	supportResp "github.com/mo7amed4522/tayara_api/model/support/response"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

// TicketServices owns ticket creation, listing and detail reads.
type TicketServices struct{}

// Open creates a ticket, its conversation and the system welcome message in one transaction.
func (s TicketServices) Open(actor Actor, req supportReq.OpenTicketReq) (*supportResp.TicketView, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	if !actor.IsCustomer() {
		return nil, errors.New("only customers can open a support ticket")
	}
	subject := strings.TrimSpace(req.Subject)
	if subject == "" {
		return nil, response.ValidationErr("subject is required")
	}
	priority := req.Priority
	switch priority {
	case "", constants.SupportPriorityLow, constants.SupportPriorityNormal,
		constants.SupportPriorityHigh, constants.SupportPriorityUrgent:
	default:
		return nil, response.ValidationErr("invalid priority")
	}
	if priority == "" {
		priority = constants.SupportPriorityNormal
	}
	if err := s.checkLinks(actor, req); err != nil {
		return nil, err
	}

	var view *supportResp.TicketView
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		ticket := supportModel.Ticket{
			TicketCode: utils.NextCode(constants.CodePrefixSupportTicket),
			Subject:    subject,
			CustomerID: actor.ID,
			BookingID:  req.BookingID,
			DisputeID:  req.DisputeID,
			Status:     constants.SupportTicketStatusOpen,
			Priority:   priority,
		}
		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}
		conv := supportModel.Conversation{
			TicketID: ticket.ID,
			Status:   constants.SupportConversationStatusActive,
			OpenedAt: &now,
		}
		if err := tx.Create(&conv).Error; err != nil {
			return err
		}
		body := utils.SanitizeChatText((SettingsServices{}).WelcomeMessage(), (SettingsServices{}).MaxMessageBytes())
		msg := supportModel.Message{
			MessageUUID:    utils.NewUUIDv7(),
			ConversationID: conv.ID,
			TicketID:       ticket.ID,
			Seq:            1,
			SenderType:     constants.ActorSystem,
			Body:           body,
			PrevHash:       utils.HashChainZero,
			BodyHash:       utils.HashChainStep(utils.HashChainZero, body),
			ClientMsgID:    "welcome-" + uintString(conv.ID),
		}
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}
		if err := tx.Model(&supportModel.Conversation{}).Where("id = ?", conv.ID).Updates(map[string]interface{}{
			"first_message_id": msg.ID,
			"message_count":    1,
			"last_seq":         1,
			"last_body_hash":   msg.BodyHash,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&supportModel.Ticket{}).Where("id = ?", ticket.ID).
			Update("last_message_at", msg.CreatedAt).Error; err != nil {
			return err
		}
		conv.FirstMessageID = &msg.ID
		conv.MessageCount = 1
		conv.LastSeq = 1
		conv.LastBodyHash = msg.BodyHash
		ticket.LastMessageAt = &msg.CreatedAt
		view = &supportResp.TicketView{Ticket: ticket, Conversation: conversationView(conv)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	publishTicketEvent("ticket.opened", view.Conversation.ID, view.Ticket.ID, uintString(view.Ticket.ID), map[string]any{
		"ticket_id": view.Ticket.ID, "ticket_code": view.Ticket.TicketCode, "customer_id": actor.ID,
	})
	return view, nil
}

// checkLinks verifies that an optional booking/dispute actually belongs to the caller.
func (s TicketServices) checkLinks(actor Actor, req supportReq.OpenTicketReq) error {
	if req.BookingID != nil && *req.BookingID > 0 {
		var bk bookingModel.Booking
		if err := global.GVA_DB.First(&bk, *req.BookingID).Error; err != nil {
			return response.ValidationErr("booking not found")
		}
		if bk.UserID != actor.ID {
			return response.ForbiddenErr("booking does not belong to the caller")
		}
	}
	if req.DisputeID != nil && *req.DisputeID > 0 {
		var d bookingModel.Dispute
		if err := global.GVA_DB.First(&d, *req.DisputeID).Error; err != nil {
			return response.ValidationErr("dispute not found")
		}
		var bk bookingModel.Booking
		if err := global.GVA_DB.First(&bk, d.BookingID).Error; err != nil || bk.UserID != actor.ID {
			return response.ForbiddenErr("dispute does not belong to the caller")
		}
	}
	return nil
}

// ListOwn returns the customer's own tickets.
func (s TicketServices) ListOwn(actor Actor, req supportReq.TicketPageReq) ([]supportModel.Ticket, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	if !actor.IsCustomer() {
		return nil, 0, 0, response.ForbiddenErr("customers only")
	}
	db := global.GVA_DB.Model(&supportModel.Ticket{}).Where("customer_id = ?", actor.ID)
	return paginateTickets(db, req)
}

// ListAdmin returns the inbox scoped to the acting staff role.
func (s TicketServices) ListAdmin(actor Actor, req supportReq.TicketPageReq) ([]supportModel.Ticket, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	db := global.GVA_DB.Model(&supportModel.Ticket{})
	switch actor.Type {
	case constants.ActorSupport:
		db = db.Where("assigned_employee_id = ? OR assigned_employee_id IS NULL", actor.ID)
	case constants.ActorQuality:
		if !(SettingsServices{}).QualityCanReadAll() {
			return nil, 0, 0, response.ForbiddenErr("quality review is not enabled")
		}
	case constants.ActorAdmin:
		// all
	default:
		return nil, 0, 0, response.ForbiddenErr("staff only")
	}
	return paginateTickets(db, req)
}

func paginateTickets(db *gorm.DB, req supportReq.TicketPageReq) ([]supportModel.Ticket, int64, int64, error) {
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Priority != "" {
		db = db.Where("priority = ?", req.Priority)
	}
	if req.CustomerID > 0 {
		db = db.Where("customer_id = ?", req.CustomerID)
	}
	if q := strings.TrimSpace(req.Query); q != "" {
		like := "%" + q + "%"
		db = db.Where("subject LIKE ? OR ticket_code LIKE ?", like, like)
	}
	var list []supportModel.Ticket
	page := req.Page
	size := req.PageSize
	if size <= 0 || size > 100 {
		size = 20
	}
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("last_message_at desc, id desc"), page, size, &list)
	if err != nil {
		return nil, 0, 0, err
	}
	return list, pr.Total, pr.TotalPage, nil
}

// Detail returns a ticket the actor may read, with its conversation summary.
func (s TicketServices) Detail(actor Actor, ticketID uint) (*supportResp.TicketView, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var tk supportModel.Ticket
	if err := global.GVA_DB.First(&tk, ticketID).Error; err != nil {
		return nil, response.NotFoundErr("ticket not found")
	}
	var conv supportModel.Conversation
	if err := global.GVA_DB.Where("ticket_id = ?", tk.ID).First(&conv).Error; err != nil {
		return nil, response.NotFoundErr("conversation not found")
	}
	if _, ok, err := (AuthorizationServices{}).CanRead(actor, conv.ID); err != nil {
		return nil, err
	} else if !ok {
		return nil, response.ForbiddenErr("not allowed to read this ticket")
	}
	return &supportResp.TicketView{Ticket: tk, Conversation: conversationView(conv)}, nil
}

func conversationView(c supportModel.Conversation) *supportResp.ConversationView {
	return &supportResp.ConversationView{
		ID:                 c.ID,
		TicketID:           c.TicketID,
		Status:             c.Status,
		MessageCount:       c.MessageCount,
		LastSeq:            c.LastSeq,
		AssignedEmployeeID: c.AssignedEmployeeID,
		OpenedAt:           c.OpenedAt,
	}
}
