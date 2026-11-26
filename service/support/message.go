package support

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common/response"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	supportResp "github.com/mo7amed4522/tayara_api/model/support/response"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var clientMsgIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

// MessageServices owns the append path and history reads. A message row doubles as the outbox
// (published_at NULL = needs re-publish), so persistence always wins over delivery.
type MessageServices struct{}

// Send appends a message to a conversation: authorization, attachment claim, sanitization,
// seq allocation under a lock, hash chain, one transaction, then a publish attempt.
func (s MessageServices) Send(actor Actor, conversationID uint, req supportReq.SendMessageReq) (*supportResp.MessageView, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	settings := SettingsServices{}
	if !settings.ChatEnabled() {
		return nil, response.ConflictErr("support chat is disabled")
	}
	if !clientMsgIDPattern.MatchString(req.ClientMsgID) {
		return nil, response.ValidationErr("clientMsgId must be 8-64 chars of [A-Za-z0-9_-]")
	}
	access, ok, err := (AuthorizationServices{}).CanWrite(actor, conversationID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, response.ForbiddenErr("not allowed to write to this conversation")
	}
	if access.Conversation.Status != constants.SupportConversationStatusActive {
		return nil, response.ConflictErr("conversation is closed")
	}
	body := utils.SanitizeChatText(req.Body, settings.MaxMessageBytes())
	if body == "" && len(req.AttachmentUuids) == 0 {
		return nil, response.ValidationErr("body or at least one attachment is required")
	}

	// Idempotency: a retried clientMsgId returns the original row and never double-inserts.
	if existing, err := s.findByClientMsgID(conversationID, req.ClientMsgID); err == nil && existing != nil {
		return s.viewFor(*existing)
	} else if err != nil {
		return nil, err
	}

	release, ok := utils.TryLock(context.Background(), constants.SupportConversationLockPrefix+uintString(conversationID), 5*time.Second)
	if !ok {
		return nil, response.ConflictErr("another message is being appended; retry")
	}
	defer release()

	var (
		view      *supportResp.MessageView
		created   supportModel.Message
		published bool
	)
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var conv supportModel.Conversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&conv, conversationID).Error; err != nil {
			return errors.New("conversation not found")
		}
		attachments, err := (AttachmentServices{}).claim(tx, actor, &conv, req.AttachmentUuids)
		if err != nil {
			return err
		}
		seq := conv.LastSeq + 1
		prev := conv.LastBodyHash
		if prev == "" {
			prev = utils.HashChainZero
		}
		hash := utils.HashChainStep(prev, body, actor.Type, uintString(actor.ID), strconv.FormatInt(seq, 10))
		msg := supportModel.Message{
			MessageUUID:    utils.NewUUIDv7(),
			ConversationID: conv.ID,
			TicketID:       conv.TicketID,
			Seq:            seq,
			SenderType:     senderTypeForActor(actor),
			SenderID:       actor.ID,
			Body:           body,
			PrevHash:       prev,
			BodyHash:       hash,
			ClientMsgID:    req.ClientMsgID,
		}
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}
		if len(req.AttachmentUuids) > 0 {
			ids := make([]uint, 0, len(attachments))
			for _, a := range attachments {
				ids = append(ids, a.ID)
			}
			if err := (AttachmentServices{}).linkToMessage(tx, ids, msg.ID); err != nil {
				return err
			}
		}
		if err := tx.Model(&supportModel.Conversation{}).Where("id = ?", conv.ID).Updates(map[string]interface{}{
			"message_count":  gorm.Expr("message_count + 1"),
			"last_seq":       seq,
			"last_body_hash": hash,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&supportModel.Ticket{}).Where("id = ?", conv.TicketID).Updates(map[string]interface{}{
			"last_message_at": msg.CreatedAt,
			"status":          ticketStatusAfterMessage(actor),
		}).Error; err != nil {
			return err
		}
		created = msg
		view = ptrMessageView(messageView(msg, attachments))
		return nil
	})
	if err != nil {
		return nil, err
	}
	published = publishMessage(*view, created.ConversationID, created.TicketID)
	if published {
		_ = global.GVA_DB.Model(&supportModel.Message{}).Where("id = ?", created.ID).
			UpdateColumn("published_at", time.Now().UTC()).Error
	}
	return view, nil
}

// ticketStatusAfterMessage moves the ticket between the two pending states.
func ticketStatusAfterMessage(actor Actor) string {
	if actor.IsCustomer() {
		return constants.SupportTicketStatusPendingSupport
	}
	return constants.SupportTicketStatusPendingCustomer
}

func (s MessageServices) findByClientMsgID(conversationID uint, clientMsgID string) (*supportModel.Message, error) {
	var msg supportModel.Message
	err := global.GVA_DB.Where("conversation_id = ? AND client_msg_id = ?", conversationID, clientMsgID).First(&msg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (s MessageServices) viewFor(msg supportModel.Message) (*supportResp.MessageView, error) {
	var attachments []supportModel.Attachment
	if err := global.GVA_DB.Where("message_id = ?", msg.ID).Find(&attachments).Error; err != nil {
		return nil, err
	}
	v := messageView(msg, attachments)
	return &v, nil
}

func ptrMessageView(v supportResp.MessageView) *supportResp.MessageView { return &v }

// History pages a conversation's messages in seq order.
func (s MessageServices) History(actor Actor, conversationID uint, req supportReq.MessagePageReq) ([]supportResp.MessageView, int64, int64, error) {
	if global.GVA_DB == nil {
		return nil, 0, 0, errors.New("db not init")
	}
	if _, ok, err := (AuthorizationServices{}).CanRead(actor, conversationID); err != nil {
		return nil, 0, 0, err
	} else if !ok {
		return nil, 0, 0, response.ForbiddenErr("not allowed to read this conversation")
	}
	size := req.PageSize
	if size <= 0 || size > 100 {
		size = (SettingsServices{}).PageSize()
	}
	db := global.GVA_DB.Model(&supportModel.Message{}).Where("conversation_id = ?", conversationID)
	if req.SinceSeq > 0 {
		db = db.Where("seq > ?", req.SinceSeq)
	}
	var rows []supportModel.Message
	pr, err := utils.PaginateWithCustomQuery(db, db.Order("seq asc"), req.Page, size, &rows)
	if err != nil {
		return nil, 0, 0, err
	}
	out := make([]supportResp.MessageView, 0, len(rows))
	for _, m := range rows {
		v, err := s.viewFor(m)
		if err != nil {
			return nil, 0, 0, err
		}
		out = append(out, *v)
	}
	return out, pr.Total, pr.TotalPage, nil
}
