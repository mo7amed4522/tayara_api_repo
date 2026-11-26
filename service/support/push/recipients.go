package push

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	"gorm.io/gorm"
)

// Participant is one actor who may receive a push.
type Participant struct {
	ActorType string
	ActorID   uint
}

// Key is the canonical actor key shared with the SSE stream (`customer:51`).
func (p Participant) Key() string { return p.ActorType + ":" + uintString(p.ActorID) }

// ResolveRecipients decides who should be woken for a message. It is deliberately pure: the
// ticket is re-read from the database by the caller so a reassignment that happened between
// publish and consume is respected.
//
//   - sender is a customer: the assigned employee if any, else the whole staff queue.
//   - sender is staff (support/quality/admin/system): the ticket's customer.
//   - the sender is never a recipient.
//
// The staff queue is supplied by the caller because this schema stores no employee->authority
// mapping (the authority lives in the JWT claims only).
func ResolveRecipients(senderType string, senderID uint, ticket supportModel.Ticket, staffQueue []Participant) []Participant {
	candidates := make([]Participant, 0, len(staffQueue)+1)
	if senderType == constants.ActorCustomer {
		if ticket.AssignedEmployeeID != nil && *ticket.AssignedEmployeeID > 0 {
			typ := constants.ActorSupport
			if ticket.AssignedAuthorityID != nil && *ticket.AssignedAuthorityID == constants.AuthorityQuality {
				typ = constants.ActorQuality
			}
			candidates = append(candidates, Participant{ActorType: typ, ActorID: *ticket.AssignedEmployeeID})
		} else {
			candidates = append(candidates, staffQueue...)
		}
	} else if ticket.CustomerID > 0 {
		candidates = append(candidates, Participant{ActorType: constants.ActorCustomer, ActorID: ticket.CustomerID})
	}

	out := make([]Participant, 0, len(candidates))
	seen := map[string]bool{}
	for _, c := range candidates {
		if c.ActorID == 0 || c.ActorType == "" {
			continue
		}
		if c.ActorType == senderType && c.ActorID == senderID {
			continue // never push to the sender
		}
		key := c.Key()
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, c)
	}
	return out
}

// loadConversationContext re-reads the conversation and its ticket. Recipients are never taken
// from the event payload: a reassignment between publish and consume must win.
func loadConversationContext(conversationID uint) (*supportModel.Conversation, *supportModel.Ticket, error) {
	if global.GVA_DB == nil {
		return nil, nil, errors.New("db not init")
	}
	var conv supportModel.Conversation
	if err := global.GVA_DB.First(&conv, conversationID).Error; err != nil {
		return nil, nil, err
	}
	var ticket supportModel.Ticket
	if err := global.GVA_DB.First(&ticket, conv.TicketID).Error; err != nil {
		return nil, nil, err
	}
	return &conv, &ticket, nil
}

// StaffQueue returns the support/quality actors who have an active push token. The schema keeps
// no employee->authority table (authority is a JWT claim), so a registered device is the only
// durable "this staff member can receive support pushes" signal.
func StaffQueue() ([]Participant, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var tokens []supportModel.PushToken
	if err := activeTokensQuery().
		Where("actor_type IN ?", []string{constants.ActorSupport, constants.ActorQuality}).
		Find(&tokens).Error; err != nil {
		return nil, err
	}
	out := make([]Participant, 0, len(tokens))
	seen := map[string]bool{}
	for _, t := range tokens {
		key := t.ActorType + ":" + uintString(t.ActorID)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, Participant{ActorType: t.ActorType, ActorID: t.ActorID})
	}
	return out, nil
}

// activeTokensQuery selects the latest token row per (actor, platform) that is not revoked.
// Correlated MAX(id) keeps it portable across MySQL/PG/SQLite (no window functions).
func activeTokensQuery() *gorm.DB {
	sub := global.GVA_DB.Model(&supportModel.PushToken{}).
		Select("MAX(id)").
		Group("actor_type, actor_id, platform")
	return global.GVA_DB.Model(&supportModel.PushToken{}).
		Where("id IN (?)", sub).
		Where("revoked_at IS NULL")
}
