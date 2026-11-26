package push

import (
	"encoding/json"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	engageService "github.com/mo7amed4522/tayara_api/service/engage"
	"github.com/mo7amed4522/tayara_api/service/support/stream"
	"go.uber.org/zap"
)

// NotificationSink records a push in the notification log owned by the engage module
// (engage_notification) and applies its admin-managed template (engage_notification_template).
// It is a seam so dispatcher tests can run without the engage wiring.
type NotificationSink interface {
	Render(eventType, channel, title, body string) (string, string)
	Record(code, eventType, recipientType string, recipientID uint, channel, title, body, imageURL string) (uint, bool, error)
	SetStatusByCode(code, status string)
}

// EngageSink is the production sink. It delegates to service/engage.NotificationServices, so the
// support push side reuses the module that already owns notification templates, the notification
// log and the FCM provider instead of duplicating them.
type EngageSink struct{}

// Render applies an active notification template to the generated title/body.
func (EngageSink) Render(eventType, channel, title, body string) (string, string) {
	return (engageService.NotificationServices{}).Render(eventType, channel, title, body)
}

// Record writes the idempotent notification log row; created is false when it already existed.
func (EngageSink) Record(code, eventType, recipientType string, recipientID uint, channel, title, body, imageURL string) (uint, bool, error) {
	return (engageService.NotificationServices{}).Record(code, eventType, recipientType, recipientID, channel, title, body, imageURL)
}

// SetStatusByCode marks the delivery outcome of a recorded notification.
func (EngageSink) SetStatusByCode(code, status string) {
	(engageService.NotificationServices{}).SetStatusByCode(code, status)
}

// NotificationEvent is the real-time frame pushed to the recipient's own notification stream. It
// mirrors the FCM payload so the in-app list and the system notification show the same thing.
type NotificationEvent struct {
	NotificationID uint              `json:"notification_id"`
	Title          string            `json:"title"`
	Body           string            `json:"body"`
	Data           map[string]string `json:"data"`
	Silent         bool              `json:"silent"`
}

func (d *Dispatcher) notifier() NotificationSink {
	if d.Notification != nil {
		return d.Notification
	}
	return EngageSink{}
}

// recordNotification writes the notification log row. A failure is logged and never stops the
// push: the message itself is already committed and the database is the source of truth.
func (d *Dispatcher) recordNotification(messageUUID string, r Participant, p Payload) (uint, bool) {
	if global.GVA_DB == nil {
		return 0, false
	}
	id, created, err := d.notifier().Record(
		NotificationCode(messageUUID, r.Key()),
		constants.SupportPushNotificationEventType,
		r.ActorType, r.ActorID,
		constants.NotifChannelPush,
		p.Title, p.Body, "",
	)
	if err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Warn("support push notification log failed", zap.Error(err))
		}
		return 0, false
	}
	return id, created
}

// broadcastNotification delivers the recorded notification to any app instance of the recipient
// that is listening on the real-time notification stream.
func (d *Dispatcher) broadcastNotification(r Participant, id uint, p Payload) {
	raw, err := json.Marshal(NotificationEvent{
		NotificationID: id,
		Title:          p.Title,
		Body:           p.Body,
		Data:           p.Data,
		Silent:         p.Silent,
	})
	if err != nil {
		return
	}
	stream.Default.BroadcastActor(r.Key(), stream.Event{
		Type: constants.SupportPushNotificationSSEEvent,
		Data: raw,
	})
}
