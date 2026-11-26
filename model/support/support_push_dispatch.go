package support

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// PushDispatch is the append-only delivery log. Each attempt appends a row: a retry never mutates
// the previous one. IdempotencyKey is the sha256 of
// (message_uuid, recipient_actor_type, recipient_actor_id, device_token); AttemptKey is
// IdempotencyKey + ":" + attempt, and carries the unique index so concurrent/redelivered
// dispatches of the same attempt cannot both be recorded.
type PushDispatch struct {
	ID              uint      `json:"id" gorm:"primarykey;comment:ID"`
	IdempotencyKey  string    `json:"idempotencyKey" gorm:"size:64;index;comment:sha256(message+recipient+token)"`
	AttemptKey      string    `json:"attemptKey" gorm:"size:80;uniqueIndex;comment:IdempotencyKey:attempt"`
	MessageUUID     string    `json:"messageUuid" gorm:"size:36;index;comment:Message UUID"`
	ConversationID  uint      `json:"conversationId" gorm:"index;comment:Conversation ID"`
	TicketID        uint      `json:"ticketId" gorm:"index;comment:Ticket ID"`
	RecipientType   string    `json:"recipientType" gorm:"size:16;index:idx_push_recipient;comment:customer|support|quality|admin"`
	RecipientID     uint      `json:"recipientId" gorm:"index:idx_push_recipient;comment:Recipient actor ID"`
	DeviceTokenHash string    `json:"deviceTokenHash" gorm:"size:64;comment:sha256(device token)"`
	Channel         string    `json:"channel" gorm:"size:16;comment:fcm"`
	Status          string    `json:"status" gorm:"size:16;index;comment:pending|sent|failed|collapsed"`
	Attempts        int       `json:"attempts" gorm:"default:1;comment:Attempt number (1-based)"`
	Silent          bool      `json:"silent" gorm:"comment:Data-only silent push"`
	Error           string    `json:"error" gorm:"size:500;comment:Last delivery error"`
	Payload         string    `json:"payload" gorm:"type:text;comment:Frozen payload (retry input)"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime;<-:create;comment:Server created at"`
}

func (PushDispatch) TableName() string { return "support_push_dispatch" }

func (PushDispatch) BeforeUpdate(tx *gorm.DB) error {
	return errors.New("support_push_dispatch is append-only")
}
func (PushDispatch) BeforeDelete(tx *gorm.DB) error {
	return errors.New("support_push_dispatch is append-only")
}
