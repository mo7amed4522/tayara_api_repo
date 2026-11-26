package response

import (
	"time"

	supportModel "github.com/mo7amed4522/tayara_api/model/support"
)

// TicketView is a ticket plus its conversation summary and the viewer's last-seen value.
type TicketView struct {
	Ticket       supportModel.Ticket `json:"ticket"`
	Conversation *ConversationView   `json:"conversation,omitempty"`
}

// ConversationView is the conversation read model handed to chat clients.
type ConversationView struct {
	ID                 uint       `json:"id"`
	TicketID           uint       `json:"ticketId"`
	Status             string     `json:"status"`
	MessageCount       int        `json:"messageCount"`
	LastSeq            int64      `json:"lastSeq"`
	AssignedEmployeeID *uint      `json:"assignedEmployeeId,omitempty"`
	OpenedAt           *time.Time `json:"openedAt,omitempty"`
}

// AttachmentView is attachment metadata; bytes are fetched from the download endpoint.
type AttachmentView struct {
	AttachmentUUID string  `json:"attachmentUuid"`
	Kind           string  `json:"kind"`
	MimeType       string  `json:"mimeType"`
	SizeBytes      int64   `json:"sizeBytes"`
	OriginalName   string  `json:"originalName"`
	Width          *int    `json:"width,omitempty"`
	Height         *int    `json:"height,omitempty"`
	ThumbnailUUID  *string `json:"thumbnailUuid"`
}

// MessageView is one message plus its attachments.
type MessageView struct {
	MessageUUID    string           `json:"messageUuid"`
	ConversationID uint             `json:"conversationId"`
	TicketID       uint             `json:"ticketId"`
	Seq            int64            `json:"seq"`
	SenderType     string           `json:"senderType"`
	SenderID       uint             `json:"senderId"`
	Body           string           `json:"body"`
	BodyHash       string           `json:"bodyHash"`
	PrevHash       string           `json:"prevHash"`
	ClientMsgID    string           `json:"clientMsgId,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
	Attachments    []AttachmentView `json:"attachments"`
}

// PresenceView is the batch presence read result for a single actor. State is "unknown" when
// Redis is not available; LastSeenAt is null when hidden by policy.
type PresenceView struct {
	ActorType  string     `json:"actorType"`
	ActorID    uint       `json:"actorId"`
	State      string     `json:"state"`
	LastSeenAt *time.Time `json:"lastSeenAt"`
}

// FaqView is a FAQ entry with tallies and the caller's current vote.
type FaqView struct {
	supportModel.Faq
	Tally  map[string]int `json:"tally"`
	MyVote string         `json:"myVote,omitempty"`
}

// ReviewView aliases the append-only review row.
type ReviewView = supportModel.Review

// AssignmentView aliases the append-only assignment row.
type AssignmentView = supportModel.AssignmentHistory
