// Package kafka wraps the only new dependency (github.com/segmentio/kafka-go) behind a tiny
// Publisher interface so the rest of the support module can be unit-tested without a broker.
package kafka

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/support/response"
	"github.com/mo7amed4522/tayara_api/utils"
)

// SchemaVersion is the only envelope version this build accepts.
const SchemaVersion = 1

// Event types carried on the bus.
const (
	EventMessageCreated = "message.created"
	EventTicketOpened   = "ticket.opened"
	EventTicketAssigned = "ticket.assigned"
	EventTicketReviewed = "ticket.reviewed"
	EventFaqVoted       = "faq.voted"
	EventPresenceChange = "presence.changed"
)

var knownEvents = map[string]bool{
	EventMessageCreated: true,
	EventTicketOpened:   true,
	EventTicketAssigned: true,
	EventTicketReviewed: true,
	EventFaqVoted:       true,
	EventPresenceChange: true,
}

// Envelope is the versioned wire format: one shape for every topic.
type Envelope struct {
	SchemaVersion  int             `json:"schema_version"`
	EventID        string          `json:"event_id"`
	EventType      string          `json:"event_type"`
	OccurredAt     time.Time       `json:"occurred_at"`
	ConversationID uint            `json:"conversation_id,omitempty"`
	TicketID       uint            `json:"ticket_id,omitempty"`
	Data           json.RawMessage `json:"data"`
}

// MessageCreatedData is the `message.created` payload. Attachments is always a list (never
// null) and contains metadata only — bytes never travel on the bus.
type MessageCreatedData struct {
	MessageUUID string                    `json:"message_uuid"`
	Seq         int64                     `json:"seq"`
	SenderType  string                    `json:"sender_type"`
	SenderID    uint                      `json:"sender_id"`
	Body        string                    `json:"body"`
	BodyHash    string                    `json:"body_hash"`
	PrevHash    string                    `json:"prev_hash"`
	CreatedAt   time.Time                 `json:"created_at"`
	Attachments []response.AttachmentView `json:"attachments"`
}

// PresenceChangedData is the `presence.changed` payload (transitions only, never heartbeats).
type PresenceChangedData struct {
	ActorType       string     `json:"actor_type"`
	ActorID         uint       `json:"actor_id"`
	State           string     `json:"state"`
	PreviousState   string     `json:"previous_state"`
	LastSeenAt      *time.Time `json:"last_seen_at"`
	LastSeenVisible bool       `json:"last_seen_visible"`
}

// NewEnvelope builds and marshals a versioned envelope with a fresh UUIDv7 event id.
func NewEnvelope(eventType string, conversationID, ticketID uint, data any) ([]byte, error) {
	if !knownEvents[eventType] {
		return nil, errors.New("unknown event type")
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{
		SchemaVersion:  SchemaVersion,
		EventID:        utils.NewUUIDv7(),
		EventType:      eventType,
		OccurredAt:     time.Now().UTC(),
		ConversationID: conversationID,
		TicketID:       ticketID,
		Data:           raw,
	})
}

// ParseEnvelope decodes and validates an envelope. A schema_version above 1 or an unknown
// event type is rejected (the consumer routes those to the DLQ rather than guessing).
func ParseEnvelope(payload []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return nil, err
	}
	if env.SchemaVersion > SchemaVersion || env.SchemaVersion < 1 {
		return nil, errors.New("unsupported schema_version")
	}
	if !knownEvents[env.EventType] {
		return nil, errors.New("unknown event_type")
	}
	return &env, nil
}

// Info is used by the stream handler to log a rejected envelope.
func Info() string {
	if global.GVA_LOG == nil {
		return ""
	}
	return "support kafka"
}
