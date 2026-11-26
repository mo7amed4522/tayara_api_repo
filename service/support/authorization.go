package support

import (
	"errors"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	"gorm.io/gorm"
)

// Actor is the authenticated caller of a support endpoint. Type is one of the constants.Actor*
// values; ID is a bx_user_profile id for customers and a sys_admin_profile id otherwise.
type Actor struct {
	Type        string
	ID          uint
	AuthorityID uint
}

// Valid reports whether the actor is usable.
func (a Actor) Valid() bool { return a.ID > 0 && a.Type != "" }

// IsCustomer reports whether the actor is a customer.
func (a Actor) IsCustomer() bool { return a.Type == constants.ActorCustomer }

// AuthorityActorType maps a JWT authority id to a support actor type.
func AuthorityActorType(authorityID uint) string {
	switch authorityID {
	case constants.AuthoritySystem:
		return constants.ActorAdmin
	case constants.AuthoritySupport:
		return constants.ActorSupport
	case constants.AuthorityQuality:
		return constants.ActorQuality
	default:
		return ""
	}
}

// Access bundles the conversation and its ticket after an authorization lookup.
type Access struct {
	Conversation supportModel.Conversation
	Ticket       supportModel.Ticket
}

// AuthorizationServices is the single place that answers "may actor X read/write conversation Y".
// Ownership is checked in the service (this repo's pattern), never trusted from the request.
type AuthorizationServices struct{}

// Resolve loads a conversation and its ticket.
func (s AuthorizationServices) Resolve(conversationID uint) (*Access, error) {
	if global.GVA_DB == nil {
		return nil, errors.New("db not init")
	}
	var conv supportModel.Conversation
	if err := global.GVA_DB.First(&conv, conversationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("conversation not found")
		}
		return nil, err
	}
	var tk supportModel.Ticket
	if err := global.GVA_DB.First(&tk, conv.TicketID).Error; err != nil {
		return nil, errors.New("ticket not found")
	}
	return &Access{Conversation: conv, Ticket: tk}, nil
}

// CanRead reports whether the actor may read the conversation.
func (s AuthorizationServices) CanRead(actor Actor, conversationID uint) (*Access, bool, error) {
	access, err := s.Resolve(conversationID)
	if err != nil {
		return nil, false, err
	}
	if !actor.Valid() {
		return access, false, nil
	}
	switch actor.Type {
	case constants.ActorCustomer:
		return access, access.Ticket.CustomerID == actor.ID, nil
	case constants.ActorAdmin:
		return access, true, nil
	case constants.ActorSupport:
		return access, access.Ticket.AssignedEmployeeID != nil && *access.Ticket.AssignedEmployeeID == actor.ID, nil
	case constants.ActorQuality:
		return access, (SettingsServices{}).QualityCanReadAll(), nil
	default:
		return access, false, nil
	}
}

// CanWrite reports whether the actor may post into the conversation. Quality reviewers are
// read-only in v1; admins and the assigned support employee may post.
func (s AuthorizationServices) CanWrite(actor Actor, conversationID uint) (*Access, bool, error) {
	access, ok, err := s.CanRead(actor, conversationID)
	if err != nil || !ok {
		return access, false, err
	}
	switch actor.Type {
	case constants.ActorCustomer:
		return access, access.Ticket.CustomerID == actor.ID, nil
	case constants.ActorAdmin:
		return access, true, nil
	case constants.ActorSupport:
		return access, access.Ticket.AssignedEmployeeID != nil && *access.Ticket.AssignedEmployeeID == actor.ID, nil
	default:
		return access, false, nil
	}
}

// AllowedPresenceActors returns the "<type>:<id>" keys whose presence this actor may see: only
// actors sharing a ticket the caller can read. Reuses the same ownership rules as CanRead.
func (s AuthorizationServices) AllowedPresenceActors(actor Actor) (map[string]bool, error) {
	out := map[string]bool{}
	if global.GVA_DB == nil || !actor.Valid() {
		return out, nil
	}
	db := global.GVA_DB.Model(&supportModel.Ticket{})
	switch actor.Type {
	case constants.ActorCustomer:
		db = db.Where("customer_id = ?", actor.ID)
	case constants.ActorSupport:
		db = db.Where("assigned_employee_id = ?", actor.ID)
	case constants.ActorQuality:
		if !(SettingsServices{}).QualityCanReadAll() {
			return out, nil
		}
	case constants.ActorAdmin:
		// sees all
	default:
		return out, nil
	}
	var tickets []supportModel.Ticket
	if err := db.Find(&tickets).Error; err != nil {
		return out, err
	}
	for _, t := range tickets {
		if t.CustomerID > 0 {
			out[constants.ActorCustomer+":"+uintString(t.CustomerID)] = true
		}
		if t.AssignedEmployeeID != nil && *t.AssignedEmployeeID > 0 {
			typ := constants.ActorSupport
			if t.AssignedAuthorityID != nil && *t.AssignedAuthorityID == constants.AuthorityQuality {
				typ = constants.ActorQuality
			}
			out[typ+":"+uintString(*t.AssignedEmployeeID)] = true
		}
	}
	return out, nil
}
