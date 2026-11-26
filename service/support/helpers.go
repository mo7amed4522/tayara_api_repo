package support

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	supportResp "github.com/mo7amed4522/tayara_api/model/support/response"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/mo7amed4522/tayara_api/service/support/stream"
)

func uintString(v uint) string { return strconv.FormatUint(uint64(v), 10) }

// senderTypeForActor maps a support actor to the message sender_type value.
func senderTypeForActor(actor Actor) string {
	switch actor.Type {
	case constants.ActorCustomer:
		return constants.ActorCustomer
	case constants.ActorSupport:
		return constants.ActorSupport
	case constants.ActorQuality:
		return constants.ActorQuality
	default:
		return constants.ActorAdmin
	}
}

// attachmentViews projects stored attachment rows for the wire (no bytes, no storage keys).
func attachmentViews(rows []supportModel.Attachment) []supportResp.AttachmentView {
	out := make([]supportResp.AttachmentView, 0, len(rows))
	for _, a := range rows {
		out = append(out, supportResp.AttachmentView{
			AttachmentUUID: a.AttachmentUUID,
			Kind:           a.Kind,
			MimeType:       a.MimeType,
			SizeBytes:      a.SizeBytes,
			OriginalName:   a.OriginalName,
			Width:          a.Width,
			Height:         a.Height,
			ThumbnailUUID:  nil,
		})
	}
	return out
}

// messageView projects a message row plus its attachments.
func messageView(m supportModel.Message, attachments []supportModel.Attachment) supportResp.MessageView {
	return supportResp.MessageView{
		MessageUUID:    m.MessageUUID,
		ConversationID: m.ConversationID,
		TicketID:       m.TicketID,
		Seq:            m.Seq,
		SenderType:     m.SenderType,
		SenderID:       m.SenderID,
		Body:           m.Body,
		BodyHash:       m.BodyHash,
		PrevHash:       m.PrevHash,
		ClientMsgID:    m.ClientMsgID,
		CreatedAt:      m.CreatedAt,
		Attachments:    attachmentViews(attachments),
	}
}

// publishMessage marshals a message.created envelope, publishes it when the bus is enabled and
// always feeds the local hub (so a single-node dev box works with Kafka disabled).
func publishMessage(view supportResp.MessageView, conversationID, ticketID uint) bool {
	data := supportKafka.MessageCreatedData{
		MessageUUID: view.MessageUUID,
		Seq:         view.Seq,
		SenderType:  view.SenderType,
		SenderID:    view.SenderID,
		Body:        view.Body,
		BodyHash:    view.BodyHash,
		PrevHash:    view.PrevHash,
		CreatedAt:   view.CreatedAt,
		Attachments: view.Attachments,
	}
	payload, err := supportKafka.NewEnvelope(supportKafka.EventMessageCreated, conversationID, ticketID, data)
	if err != nil {
		return false
	}
	ok := supportKafka.PublishChat(context.Background(), uintString(conversationID), payload) == nil
	stream.Default.Broadcast(conversationID, stream.Event{Type: supportKafka.EventMessageCreated, Data: payload})
	return ok
}

// publishTicketEvent publishes a ticket/faq event; failures are non-fatal (the DB is the truth).
func publishTicketEvent(eventType string, conversationID, ticketID uint, key string, data any) bool {
	payload, err := supportKafka.NewEnvelope(eventType, conversationID, ticketID, data)
	if err != nil {
		return false
	}
	return supportKafka.PublishEvent(context.Background(), key, payload) == nil
}

func nowPtr() *time.Time {
	t := time.Now().UTC()
	return &t
}

// trimJoin is a small helper for comma lists.
func trimJoin(v []string) string { return strings.Join(v, ",") }
