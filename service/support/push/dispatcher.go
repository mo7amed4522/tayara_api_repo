package push

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	"github.com/mo7amed4522/tayara_api/service/engage/dispatch"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"go.uber.org/zap"
)

// ErrInvalidToken marks a device token FCM permanently rejected; the caller appends a revocation
// row instead of retrying the token forever.
var ErrInvalidToken = errors.New("push token is no longer valid")

// DispatchMessage is one push to one device.
type DispatchMessage struct {
	Title  string
	Body   string
	Data   map[string]string
	Silent bool
}

// DeviceSender is the delivery seam so tests never reach FCM.
type DeviceSender interface {
	Send(ctx context.Context, token string, msg DispatchMessage) error
}

// FcmDeviceSender adapts the existing engage FCM integration. No new dependency.
type FcmDeviceSender struct{}

// Send delivers through the existing firebase messaging client.
func (FcmDeviceSender) Send(ctx context.Context, token string, msg DispatchMessage) error {
	err := dispatch.FcmSender{}.SendDataMessage(ctx, dispatch.DataMessage{
		Token: token, Title: msg.Title, Body: msg.Body, Data: msg.Data, Silent: msg.Silent,
	})
	if err != nil && dispatch.IsInvalidToken(err) {
		return ErrInvalidToken
	}
	return err
}

// PresenceGate is implemented by service/support (never queried directly by this package).
type PresenceGate interface {
	IsOnline(actorType string, actorID uint) (online bool, known bool)
}

// CollapseGate is the best-effort per-recipient debounce window.
type CollapseGate interface {
	ClaimPushWindow(recipient string, conversationID uint, window time.Duration) (collapsed bool, ok bool)
}

type unknownPresence struct{}

func (unknownPresence) IsOnline(string, uint) (bool, bool) { return false, false }

type noCollapse struct{}

func (noCollapse) ClaimPushWindow(string, uint, time.Duration) (bool, bool) { return false, false }

var (
	gateMu       sync.RWMutex
	presenceGate PresenceGate = unknownPresence{}
	collapseGate CollapseGate = noCollapse{}
)

// SetPresenceGate installs the presence implementation (called from service/support's init to
// avoid an import cycle: this package must not import its parent).
func SetPresenceGate(g PresenceGate) {
	gateMu.Lock()
	defer gateMu.Unlock()
	if g == nil {
		g = unknownPresence{}
	}
	presenceGate = g
}

// SetCollapseGate installs the collapse implementation.
func SetCollapseGate(g CollapseGate) {
	gateMu.Lock()
	defer gateMu.Unlock()
	if g == nil {
		g = noCollapse{}
	}
	collapseGate = g
}

func currentPresence() PresenceGate {
	gateMu.RLock()
	defer gateMu.RUnlock()
	return presenceGate
}

func currentCollapse() CollapseGate {
	gateMu.RLock()
	defer gateMu.RUnlock()
	return collapseGate
}

// Dispatcher orchestrates one event into zero or more push notifications.
type Dispatcher struct {
	Sender       DeviceSender
	Presence     PresenceGate
	Collapse     CollapseGate
	Notification NotificationSink
}

// NewDispatcher builds the production dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		Sender:       FcmDeviceSender{},
		Presence:     currentPresence(),
		Collapse:     currentCollapse(),
		Notification: EngageSink{},
	}
}

var (
	defaultOnce       sync.Once
	defaultDispatcher *Dispatcher
)

// DispatcherInstance returns the process-wide dispatcher.
func DispatcherInstance() *Dispatcher {
	defaultOnce.Do(func() { defaultDispatcher = NewDispatcher() })
	return defaultDispatcher
}

// DispatchForEvent decides per recipient whether a push is needed, and sends it best-effort.
// A failure never rolls back the message: the row is already committed and the DB is the truth.
//
// For each recipient that is not known-online the flow is: apply the notification template,
// write the idempotent engage_notification row, emit the real-time notification frame, then
// collapse/FCM. The notification row and the real-time frame are written for every such
// recipient; only the device push is gated on having a registered token, so a recipient with no
// device still sees the notification in the app. A recipient already connected to the
// conversation stream is skipped: the SSE frame is their notification and it has already been
// delivered.
func (d *Dispatcher) DispatchForEvent(ctx context.Context, env *supportKafka.Envelope) error {
	if env == nil || env.EventType != supportKafka.EventMessageCreated {
		return nil
	}
	if !Enabled() {
		return nil
	}
	var msg supportKafka.MessageCreatedData
	if err := json.Unmarshal(env.Data, &msg); err != nil {
		return err
	}
	conv, ticket, err := loadConversationContext(env.ConversationID)
	if err != nil {
		return err
	}
	queue, err := StaffQueue()
	if err != nil {
		return err
	}
	recipients := ResolveRecipients(msg.SenderType, msg.SenderID, *ticket, queue)
	if len(recipients) == 0 {
		return nil
	}
	basePayload := BuildPayload(conv.ID, ticket.ID, ticket.Subject, msg, titleBytes(), previewBytes())
	// An admin-managed template may wrap or replace the generated subject/preview.
	basePayload.Title, basePayload.Body = d.notifier().Render(
		constants.SupportPushNotificationEventType, constants.NotifChannelPush,
		basePayload.Title, basePayload.Body)

	for _, r := range recipients {
		online, known := d.Presence.IsOnline(r.ActorType, r.ActorID)
		if known && online {
			continue // the conversation SSE stream already delivered it
		}
		payload := basePayload
		payload.Silent = !known && silentRole(r.ActorType)

		// The notification log row is also the frame the app receives in real time. Its
		// deterministic code keeps a redelivered Kafka record from writing it twice. This runs
		// regardless of device tokens: the in-app notification is delivery on its own.
		if id, created := d.recordNotification(msg.MessageUUID, r, payload); created {
			d.broadcastNotification(r, id, payload)
		}

		// Only the device push needs a token.
		tokens, err := activeDeviceTokens(r)
		if err != nil || len(tokens) == 0 {
			continue
		}
		if collapsed, ok := d.Collapse.ClaimPushWindow(r.Key(), conv.ID, collapseWindow()); ok && collapsed {
			d.recordCollapsed(*conv, *ticket, msg, r, tokens[0], payload)
			d.notifier().SetStatusByCode(NotificationCode(msg.MessageUUID, r.Key()), constants.NotificationStatusSent)
			continue
		}
		for _, tk := range tokens {
			key := IdempotencyKey(msg.MessageUUID, r.ActorType, r.ActorID, tk)
			attempts, settled := dispatchState(key)
			if settled || attempts >= maxAttempts() {
				continue // redelivered record: already delivered, or out of attempts
			}
			d.sendAttempt(ctx, attempt{
				MessageUUID: msg.MessageUUID, ConversationID: conv.ID, TicketID: ticket.ID,
				Recipient: r, Token: tk, Payload: payload, Attempt: attempts + 1,
			})
		}
	}
	return nil
}

// attempt is one delivery try.
type attempt struct {
	MessageUUID    string
	ConversationID uint
	TicketID       uint
	Recipient      Participant
	Token          string
	Payload        Payload
	Attempt        int
}

// sendAttempt performs the FCM call and appends the immutable outcome row.
func (d *Dispatcher) sendAttempt(ctx context.Context, a attempt) {
	if global.GVA_DB == nil {
		return
	}
	msg := DispatchMessage{Title: a.Payload.Title, Body: a.Payload.Body, Data: a.Payload.Data, Silent: a.Payload.Silent}
	err := d.deliver(ctx, a.Token, msg)

	status := constants.PushStatusSent
	errText := ""
	if err != nil {
		status = constants.PushStatusFailed
		errText = truncateRunes(err.Error(), 500)
		if errors.Is(err, ErrInvalidToken) {
			_ = revokeTokenRow(a.Token)
		}
	}
	frozen, _ := json.Marshal(a.Payload)
	row := supportModel.PushDispatch{
		IdempotencyKey:  IdempotencyKey(a.MessageUUID, a.Recipient.ActorType, a.Recipient.ActorID, a.Token),
		MessageUUID:     a.MessageUUID,
		ConversationID:  a.ConversationID,
		TicketID:        a.TicketID,
		RecipientType:   a.Recipient.ActorType,
		RecipientID:     a.Recipient.ActorID,
		DeviceTokenHash: sha256Hex(a.Token),
		Channel:         constants.PushChannelFCM,
		Status:          status,
		Attempts:        a.Attempt,
		Silent:          a.Payload.Silent,
		Error:           errText,
		Payload:         string(frozen),
	}
	row.AttemptKey = row.IdempotencyKey + ":" + uintString(uint(a.Attempt))
	_ = global.GVA_DB.Create(&row).Error

	noteStatus := constants.NotificationStatusSent
	if err != nil {
		noteStatus = constants.NotificationStatusFailed
	}
	d.notifier().SetStatusByCode(NotificationCode(a.MessageUUID, a.Recipient.Key()), noteStatus)

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("support push dispatched",
			zap.String("uuid", a.MessageUUID),
			zap.String("recipient", a.Recipient.Key()),
			zap.String("channel", constants.PushChannelFCM),
			zap.Bool("ok", err == nil),
		)
	}
}

// deliver performs the bounded in-process retry: capped exponential backoff values are reused
// from support.kafka.retry. A permanently invalid token is never retried. The DB row is the source
// of truth, so giving up here is safe — the message is still reachable via sinceSeq.
func (d *Dispatcher) deliver(ctx context.Context, token string, msg DispatchMessage) error {
	limit := maxAttempts()
	var err error
	for i := 1; i <= limit; i++ {
		if err = d.Sender.Send(ctx, token, msg); err == nil {
			return nil
		}
		if errors.Is(err, ErrInvalidToken) || i == limit {
			return err
		}
		select {
		case <-ctx.Done():
			return err
		case <-time.After(backoffFor(i)):
		}
	}
	return err
}

// recordCollapsed appends the "suppressed by the collapse window" marker without touching the
// first dispatch row.
func (d *Dispatcher) recordCollapsed(conv supportModel.Conversation, ticket supportModel.Ticket, msg supportKafka.MessageCreatedData, r Participant, token string, payload Payload) {
	if global.GVA_DB == nil {
		return
	}
	frozen, _ := json.Marshal(payload)
	key := IdempotencyKey(msg.MessageUUID, r.ActorType, r.ActorID, token)
	row := supportModel.PushDispatch{
		IdempotencyKey:  key,
		AttemptKey:      key + ":collapsed",
		MessageUUID:     msg.MessageUUID,
		ConversationID:  conv.ID,
		TicketID:        ticket.ID,
		RecipientType:   r.ActorType,
		RecipientID:     r.ActorID,
		DeviceTokenHash: sha256Hex(token),
		Channel:         constants.PushChannelFCM,
		Status:          constants.PushStatusCollapsed,
		Attempts:        1,
		Silent:          payload.Silent,
		Payload:         string(frozen),
	}
	_ = global.GVA_DB.Create(&row).Error
}

// dispatchState reports how many attempts exist for a key and whether it is settled (sent or
// collapsed). A redelivered Kafka record is skipped once settled.
func dispatchState(key string) (attempts int, settled bool) {
	if global.GVA_DB == nil {
		return 0, false
	}
	var rows []supportModel.PushDispatch
	_ = global.GVA_DB.Select("attempts, status").Where("idempotency_key = ?", key).Find(&rows).Error
	for _, r := range rows {
		if r.Status == constants.PushStatusSent || r.Status == constants.PushStatusCollapsed {
			settled = true
		}
		if r.Attempts > attempts {
			attempts = r.Attempts
		}
	}
	return attempts, settled
}
