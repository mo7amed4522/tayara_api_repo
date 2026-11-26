package push_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/global"
	engageModel "github.com/mo7amed4522/tayara_api/model/engage"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/mo7amed4522/tayara_api/service/support/push"
	"github.com/mo7amed4522/tayara_api/service/support/stream"
	"gorm.io/gorm"
)

type fakeSender struct {
	mu    sync.Mutex
	calls []string
	msgs  []push.DispatchMessage
	err   error
}

func (f *fakeSender) Send(_ context.Context, token string, msg push.DispatchMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, token)
	f.msgs = append(f.msgs, msg)
	return nil
}

func (f *fakeSender) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

type fakePresence struct {
	online bool
	known  bool
}

func (f fakePresence) IsOnline(string, uint) (bool, bool) { return f.online, f.known }

type fakeCollapse struct {
	collapsed bool
	ok        bool
}

func (f fakeCollapse) ClaimPushWindow(string, uint, time.Duration) (bool, bool) {
	return f.collapsed, f.ok
}

func testConfig() config.Server {
	return config.Server{Support: config.SupportConfig{
		Enabled: true,
		Push: config.SupportPush{
			Enabled: true, CollapseWindowSeconds: 30, MaxPreviewBytes: 140, MaxTitleBytes: 60,
			SilentForRoles: []string{"support", "quality"},
			Retry:          config.SupportPushRetry{MaxAttempts: 3},
		},
	}}
}

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&supportModel.Ticket{}, &supportModel.Conversation{},
		&supportModel.PushToken{}, &supportModel.PushDispatch{},
		&engageModel.DeviceToken{}, &engageModel.Notification{}, &engageModel.NotificationTemplate{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	prevDB, prevCfg := global.GVA_DB, global.GVA_CONFIG
	global.GVA_DB = db
	global.GVA_CONFIG = testConfig()
	t.Cleanup(func() { global.GVA_DB = prevDB; global.GVA_CONFIG = prevCfg })
	return db
}

func seedConversation(t *testing.T, db *gorm.DB, assigned *uint) (supportModel.Ticket, supportModel.Conversation) {
	t.Helper()
	ticket := supportModel.Ticket{
		TicketCode: "ST-TEST-" + strings.ReplaceAll(t.Name(), "/", "-"),
		Subject:    "Where is my refund?", CustomerID: 51, AssignedEmployeeID: assigned,
		Status: "open", Priority: "normal",
	}
	if err := db.Create(&ticket).Error; err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	conv := supportModel.Conversation{TicketID: ticket.ID, Status: "active"}
	if err := db.Create(&conv).Error; err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	return ticket, conv
}

func registerToken(t *testing.T, actorType string, actorID uint, token string) {
	t.Helper()
	if _, err := (push.PushServices{}).RegisterToken(actorType, actorID, supportReq.RegisterPushTokenReq{
		Platform: "android", Token: token, AppVersion: "1.0.0",
	}); err != nil {
		t.Fatalf("register token: %v", err)
	}
}

func envelopeFor(t *testing.T, conv, ticket uint, senderType string, senderID uint, body string) *supportKafka.Envelope {
	t.Helper()
	raw, err := json.Marshal(supportKafka.MessageCreatedData{
		MessageUUID: "msg-" + strings.ReplaceAll(t.Name(), "/", "-") + "-1",
		Seq:         7, SenderType: senderType, SenderID: senderID, Body: body,
	})
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	return &supportKafka.Envelope{
		SchemaVersion: 1, EventID: "evt-1", EventType: supportKafka.EventMessageCreated,
		OccurredAt: time.Now().UTC(), ConversationID: conv, TicketID: ticket, Data: raw,
	}
}

func TestPresenceOnlineSkipsPush(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	registerToken(t, "support", 9, "device-token-support-9")

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: true, known: true}, Collapse: fakeCollapse{}}
	if err := d.DispatchForEvent(context.Background(), envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "hello")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if sender.count() != 0 {
		t.Fatalf("online recipient must not get a push, got %d", sender.count())
	}
}

func TestOfflineSendsVisiblePush(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	registerToken(t, "support", 9, "device-token-support-9")

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: true}, Collapse: fakeCollapse{}}
	if err := d.DispatchForEvent(context.Background(), envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "hello there")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if sender.count() != 1 {
		t.Fatalf("offline recipient must get one push, got %d", sender.count())
	}
	if sender.msgs[0].Silent {
		t.Fatal("a definitively offline user must get a visible push")
	}
	if sender.msgs[0].Title != "Where is my refund?" || !strings.Contains(sender.msgs[0].Body, "Customer: hello there") {
		t.Fatalf("unexpected payload: %+v", sender.msgs[0])
	}
	var row supportModel.PushDispatch
	if err := db.Where("status = ?", "sent").First(&row).Error; err != nil {
		t.Fatalf("dispatch row missing: %v", err)
	}
	if row.RecipientType != "support" || row.RecipientID != 9 || row.Channel != "fcm" {
		t.Fatalf("unexpected dispatch row: %+v", row)
	}
}

func TestUnknownPresenceSupportGetsSilentPush(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	registerToken(t, "support", 9, "device-token-support-9")

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: false}, Collapse: fakeCollapse{}}
	if err := d.DispatchForEvent(context.Background(), envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "hi")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if sender.count() != 1 || !sender.msgs[0].Silent {
		t.Fatalf("unknown presence must give staff a single silent push: %+v", sender.msgs)
	}
}

func TestUnknownPresenceCustomerGetsVisiblePush(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	registerToken(t, "customer", 51, "device-token-customer-51")

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: false}, Collapse: fakeCollapse{}}
	if err := d.DispatchForEvent(context.Background(), envelopeFor(t, conv.ID, ticket.ID, "support", 9, "on it")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if sender.count() != 1 || sender.msgs[0].Silent {
		t.Fatalf("customers always get a visible push: %+v", sender.msgs)
	}
}

func TestNoDeviceTokenStillRecordsAndStreamsNotification(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	// deliberately no device token for support:9

	sub, cancel := stream.Default.SubscribeActor("support:9")
	defer cancel()

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: true}, Collapse: fakeCollapse{}}
	if err := d.DispatchForEvent(context.Background(), envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "anyone there?")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if sender.count() != 0 {
		t.Fatalf("no token means no FCM call, got %d", sender.count())
	}

	var note engageModel.Notification
	if err := db.Where("recipient_type = ? AND recipient_id = ?", "support", 9).First(&note).Error; err != nil {
		t.Fatalf("notification row missing without a token: %v", err)
	}
	if note.Title != "Where is my refund?" || note.Status != "pending" {
		t.Fatalf("unexpected notification row: %+v", note)
	}

	select {
	case ev := <-sub.C:
		var frame push.NotificationEvent
		if err := json.Unmarshal(ev.Data, &frame); err != nil {
			t.Fatalf("unmarshal frame: %v", err)
		}
		if frame.NotificationID != note.ID || frame.Title != note.Title {
			t.Fatalf("unexpected real-time frame: %+v", frame)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("recipient without a token got no real-time notification frame")
	}

	var rows int64
	db.Model(&supportModel.PushDispatch{}).Count(&rows)
	if rows != 0 {
		t.Fatalf("no dispatch row expected without a token, got %d", rows)
	}
}

func TestRedeliveredEventDoesNotDoublePush(t *testing.T) {
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
	if sender.count() != 1 {
		t.Fatalf("redelivery must not double-push, got %d", sender.count())
	}
	var rows int64
	db.Model(&supportModel.PushDispatch{}).Count(&rows)
	if rows != 1 {
		t.Fatalf("expected exactly one dispatch row, got %d", rows)
	}
}

func TestCollapseWindowSuppressesSecondPush(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	registerToken(t, "support", 9, "device-token-support-9")

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: true}, Collapse: fakeCollapse{collapsed: true, ok: true}}
	if err := d.DispatchForEvent(context.Background(), envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "second")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if sender.count() != 0 {
		t.Fatalf("collapsed dispatch must not send, got %d", sender.count())
	}
	var row supportModel.PushDispatch
	if err := db.Where("status = ?", "collapsed").First(&row).Error; err != nil {
		t.Fatalf("collapsed marker missing: %v", err)
	}
}

func TestInvalidTokenRevokesAndRecordsFailure(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	registerToken(t, "support", 9, "device-token-support-9")

	sender := &fakeSender{err: push.ErrInvalidToken}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: true}, Collapse: fakeCollapse{}}
	if err := d.DispatchForEvent(context.Background(), envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "hello")); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	var failed supportModel.PushDispatch
	if err := db.Where("status = ?", "failed").First(&failed).Error; err != nil {
		t.Fatalf("failed dispatch row missing: %v", err)
	}
	var revoked int64
	db.Model(&supportModel.PushToken{}).Where("revoked_at IS NOT NULL").Count(&revoked)
	if revoked != 1 {
		t.Fatalf("expected one appended revocation row, got %d", revoked)
	}
}

func TestReassignmentBetweenPublishAndConsumeWins(t *testing.T) {
	db := setupDB(t)
	first := uint(9)
	ticket, conv := seedConversation(t, db, &first)
	registerToken(t, "support", 9, "device-token-support-9")
	registerToken(t, "support", 10, "device-token-support-10")
	env := envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "hello")

	// reassigned after the event was published
	second := uint(10)
	db.Model(&supportModel.Ticket{}).Where("id = ?", ticket.ID).Update("assigned_employee_id", second)

	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: true}, Collapse: fakeCollapse{}}
	if err := d.DispatchForEvent(context.Background(), env); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if sender.count() != 1 || sender.calls[0] != "device-token-support-10" {
		t.Fatalf("recipient must be re-read from storage: %+v", sender.calls)
	}
}
