package push_test

import (
	"context"
	"testing"

	"github.com/mo7amed4522/tayara_api/constants"
	engageModel "github.com/mo7amed4522/tayara_api/model/engage"
	"github.com/mo7amed4522/tayara_api/service/support/push"
)

// A customer's token registered through the existing engage module
// (POST /api/v1/engage/devices) must be a valid push target for support chat.
func TestEngageDeviceTokenIsUsedForCustomer(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	if err := db.Create(&engageModel.DeviceToken{
		UserID: 51, Token: "engage-fcm-token-customer-51", Platform: "android", IsActive: true,
	}).Error; err != nil {
		t.Fatalf("seed engage device token: %v", err)
	}

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: true}, Collapse: fakeCollapse{}}
	if err := d.DispatchForEvent(context.Background(), envelopeFor(t, conv.ID, ticket.ID, "support", 9, "your refund is on the way")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if sender.count() != 1 || sender.calls[0] != "engage-fcm-token-customer-51" {
		t.Fatalf("engage device token must be pushed to: %+v", sender.calls)
	}

	var notes []engageModel.Notification
	if err := db.Where("event_type = ?", constants.SupportPushNotificationEventType).Find(&notes).Error; err != nil {
		t.Fatalf("read notifications: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected one notification log row, got %d", len(notes))
	}
	if notes[0].RecipientType != "customer" || notes[0].RecipientID != 51 || notes[0].Channel != "push" {
		t.Fatalf("unexpected notification row: %+v", notes[0])
	}
	if notes[0].Status != constants.NotificationStatusSent {
		t.Fatalf("delivered notification must be marked sent, got %q", notes[0].Status)
	}
}

// An admin template may wrap the generated subject/preview with the placeholders.
func TestNotificationTemplateWrapsPayload(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	registerToken(t, "support", 9, "device-token-support-9")
	if err := db.Create(&engageModel.NotificationTemplate{
		EventType: constants.SupportPushNotificationEventType, Channel: "push",
		Subject: "Tayara · {{title}}", Body: "New reply — {{body}}", IsActive: true,
	}).Error; err != nil {
		t.Fatalf("seed template: %v", err)
	}

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: true}, Collapse: fakeCollapse{}}
	if err := d.DispatchForEvent(context.Background(), envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "hello")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if sender.count() != 1 {
		t.Fatalf("expected one push, got %d", sender.count())
	}
	if sender.msgs[0].Title != "Tayara · Where is my refund?" {
		t.Fatalf("template subject not applied: %q", sender.msgs[0].Title)
	}
	if sender.msgs[0].Body != "New reply — Customer: hello" {
		t.Fatalf("template body not applied: %q", sender.msgs[0].Body)
	}
}

// A redelivered Kafka record must not create a second notification log row.
func TestRedeliveryKeepsNotificationLogSingle(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	registerToken(t, "support", 9, "device-token-support-9")

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: true}, Collapse: fakeCollapse{}}
	env := envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "hello")
	for i := 0; i < 2; i++ {
		if err := d.DispatchForEvent(context.Background(), env); err != nil {
			t.Fatalf("dispatch %d: %v", i, err)
		}
	}
	var notes int64
	db.Model(&engageModel.Notification{}).Count(&notes)
	if notes != 1 {
		t.Fatalf("expected one notification row after redelivery, got %d", notes)
	}
}
