package push

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mo7amed4522/tayara_api/constants"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/mo7amed4522/tayara_api/utils"
)

// Payload is the client-facing push content. It carries a short preview only: the real message is
// fetched through the existing REST/SSE path once the user taps the notification.
type Payload struct {
	Title  string            `json:"title"`
	Body   string            `json:"body"`
	Data   map[string]string `json:"data"`
	Silent bool              `json:"silent"`
}

// BuildPayload renders the title/body/data for one `message.created` event. maxTitleBytes and
// maxPreviewBytes are byte caps applied on valid UTF-8 boundaries (Arabic and emoji survive).
func BuildPayload(conversationID, ticketID uint, subject string, msg supportKafka.MessageCreatedData, maxTitleBytes, maxPreviewBytes int) Payload {
	if maxTitleBytes <= 0 {
		maxTitleBytes = 60
	}
	if maxPreviewBytes <= 0 {
		maxPreviewBytes = 140
	}
	title := utils.SanitizeChatText(subject, 0)
	if title == "" {
		title = "Support chat"
	}
	title = truncateRunes(title, maxTitleBytes)

	preview := utils.SanitizeChatText(msg.Body, 0)
	body := ""
	switch {
	case preview != "":
		body = senderLabel(msg.SenderType) + ": " + preview
	case len(msg.Attachments) > 0:
		body = "📎 Attachment"
	default:
		body = senderLabel(msg.SenderType) + ": (new message)"
	}
	body = truncateRunes(body, maxPreviewBytes)

	return Payload{
		Title: title,
		Body:  body,
		Data: map[string]string{
			"event_type":      supportKafka.EventMessageCreated,
			"conversation_id": strconv.FormatUint(uint64(conversationID), 10),
			"ticket_id":       strconv.FormatUint(uint64(ticketID), 10),
			"message_uuid":    msg.MessageUUID,
			"seq":             strconv.FormatInt(msg.Seq, 10),
		},
	}
}

// senderLabel is the short role prefix shown in the notification body.
func senderLabel(senderType string) string {
	switch senderType {
	case constants.ActorCustomer:
		return "Customer"
	case constants.ActorSupport:
		return "Support"
	case constants.ActorQuality:
		return "Quality"
	case constants.ActorSystem:
		return "System"
	default:
		return "Admin"
	}
}

// truncateRunes cuts s to at most maxBytes without splitting a rune, then trims trailing space.
func truncateRunes(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return strings.TrimSpace(s[:cut])
}
