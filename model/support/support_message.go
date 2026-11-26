package support

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Message an append-only chat message. It deliberately does not embed global.GVA_MODEL: no
// updated_at and no soft-delete columns, so GORM has no update or delete path for it.
// The row also acts as the outbox: published_at NULL means "needs (re)publishing".
type Message struct {
	ID             uint       `json:"id" gorm:"primarykey;comment:ID"`
	MessageUUID    string     `json:"messageUuid" form:"messageUuid" gorm:"size:36;uniqueIndex;comment:Server-generated UUIDv7"`
	ConversationID uint       `json:"conversationId" form:"conversationId" gorm:"index:idx_conv_seq;comment:Conversation ID"`
	TicketID       uint       `json:"ticketId" form:"ticketId" gorm:"index;comment:Ticket ID (denormalized)"`
	Seq            int64      `json:"seq" form:"seq" gorm:"index:idx_conv_seq;comment:Dense per-conversation sequence"`
	SenderType     string     `json:"senderType" form:"senderType" gorm:"size:16;index;comment:customer|support|quality|system"`
	SenderID       uint       `json:"senderId" form:"senderId" gorm:"index;comment:Sender ID (0 for system)"`
	Body           string     `json:"body" form:"body" gorm:"type:text;comment:Sanitized plain text"`
	BodyHash       string     `json:"bodyHash" form:"bodyHash" gorm:"size:64;comment:SHA-256 chain hash"`
	PrevHash       string     `json:"prevHash" form:"prevHash" gorm:"size:64;comment:Previous body hash"`
	ClientMsgID    string     `json:"clientMsgId" form:"clientMsgId" gorm:"size:64;index:uk_client_msg,unique;comment:Client idempotency key"`
	KafkaOffset    *int64     `json:"kafkaOffset" form:"kafkaOffset" gorm:"comment:Kafka offset (observability)"`
	PublishedAt    *time.Time `json:"publishedAt" form:"publishedAt" gorm:"index;comment:NULL = needs re-publish"`
	CreatedAt      time.Time  `json:"createdAt" gorm:"autoCreateTime;<-:create;comment:Server created at"`
}

func (Message) TableName() string { return "support_message" }

// ErrAppendOnly is returned by every mutation hook on an append-only table.
var ErrAppendOnly = errors.New("support_message is append-only")

// BeforeUpdate blocks every update: messages can never be edited.
func (Message) BeforeUpdate(tx *gorm.DB) error { return ErrAppendOnly }

// BeforeDelete blocks every delete: messages can never be removed.
func (Message) BeforeDelete(tx *gorm.DB) error { return ErrAppendOnly }

// AssignmentHistory an append-only record of ticket reassignment.
type AssignmentHistory struct {
	ID             uint      `json:"id" gorm:"primarykey;comment:ID"`
	TicketID       uint      `json:"ticketId" form:"ticketId" gorm:"index:idx_ticket_created;comment:Ticket ID"`
	FromEmployeeID *uint     `json:"fromEmployeeId" form:"fromEmployeeId" gorm:"comment:Previous assignee"`
	ToEmployeeID   *uint     `json:"toEmployeeId" form:"toEmployeeId" gorm:"comment:New assignee"`
	AssignedBy     uint      `json:"assignedBy" form:"assignedBy" gorm:"comment:Acting admin ID"`
	Reason         string    `json:"reason" form:"reason" gorm:"size:500;comment:Assignment reason"`
	CreatedAt      time.Time `json:"createdAt" gorm:"autoCreateTime;<-:create;comment:Server created at"`
}

func (AssignmentHistory) TableName() string { return "support_assignment_history" }

func (AssignmentHistory) BeforeUpdate(tx *gorm.DB) error {
	return errors.New("support_assignment_history is append-only")
}
func (AssignmentHistory) BeforeDelete(tx *gorm.DB) error {
	return errors.New("support_assignment_history is append-only")
}

// Review an append-only quality review of a ticket. message_count_at_review snapshots the
// conversation length so a later message cannot silently invalidate the verdict.
type Review struct {
	ID                   uint      `json:"id" gorm:"primarykey;comment:ID"`
	TicketID             uint      `json:"ticketId" form:"ticketId" gorm:"index:idx_review_ticket_created;comment:Ticket ID"`
	ConversationID       uint      `json:"conversationId" form:"conversationId" gorm:"index;comment:Conversation ID"`
	ReviewerID           uint      `json:"reviewerId" form:"reviewerId" gorm:"index;comment:Reviewer (sys_admin_profile) ID"`
	Verdict              string    `json:"verdict" form:"verdict" gorm:"size:16;comment:solved|not_solved"`
	Notes                string    `json:"notes" form:"notes" gorm:"size:2000;comment:Reviewer notes"`
	MessageCountAtReview int       `json:"messageCountAtReview" form:"messageCountAtReview" gorm:"comment:Snapshot of conversation length"`
	CreatedAt            time.Time `json:"createdAt" gorm:"autoCreateTime;<-:create;comment:Server created at"`
}

func (Review) TableName() string { return "support_review" }

func (Review) BeforeUpdate(tx *gorm.DB) error { return errors.New("support_review is append-only") }
func (Review) BeforeDelete(tx *gorm.DB) error { return errors.New("support_review is append-only") }
