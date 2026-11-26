package support_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/common"
	supportModel "github.com/mo7amed4522/tayara_api/model/support"
	supportReq "github.com/mo7amed4522/tayara_api/model/support/request"
	supportService "github.com/mo7amed4522/tayara_api/service/support"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/mo7amed4522/tayara_api/utils"
	"gorm.io/gorm"
)

// fakePublisher records publishes per topic so tests can assert exactly-once transitions.
type fakePublisher struct {
	mu     sync.Mutex
	topics map[string]int
}

func (f *fakePublisher) Publish(_ context.Context, topic, _ string, _ []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.topics == nil {
		f.topics = map[string]int{}
	}
	f.topics[topic]++
	return nil
}

func (f *fakePublisher) Close() error { return nil }

func (f *fakePublisher) count(topic string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.topics[topic]
}

func testConfig() config.Server {
	return config.Server{
		System: config.System{OssType: "local"},
		Local:  config.Local{StorePath: os.TempDir(), BaseURL: "http://localhost", Path: "/uploads"},
		Support: config.SupportConfig{
			Enabled: true,
			Kafka: config.SupportKafka{
				Enabled: true,
				Topics:  config.SupportKafkaTopics{Presence: "presence-test", ChatMessages: "messages-test", ChatEvents: "events-test"},
			},
			Chat: config.SupportChat{
				Enabled:            true,
				WelcomeMessage:     "Hello! Thanks for reaching out.",
				MaxMessageBytes:    4000,
				PageSize:           50,
				QualityCanReadAll:  true,
				HashChainEnabled:   true,
				Assignment:         config.SupportAssignment{RequireReason: true},
				AllowedSenderTypes: []string{"customer", "support", "quality", "system"},
				Attachments: config.SupportAttachments{
					Enabled:                 true,
					MaxFileBytes:            10 << 20,
					MaxFilesPerMessage:      5,
					MaxTotalBytesPerMessage: 20 << 20,
					AllowedMime:             []string{"image/png", "image/jpeg", "application/pdf"},
					OrphanCleanupHours:      24,
				},
			},
			Presence: config.SupportPresence{
				Enabled:                  true,
				TTLSeconds:               45,
				LastSeenDisplay:          true,
				BatchReadMaxIds:          50,
				HeartbeatIntervalSeconds: 15,
			},
		},
	}
}

func setup(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&supportModel.Ticket{}, &supportModel.Conversation{}, &supportModel.Message{},
		&supportModel.AssignmentHistory{}, &supportModel.Review{}, &supportModel.Faq{},
		&supportModel.FaqVote{}, &supportModel.Presence{}, &supportModel.Attachment{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	prevDB, prevCfg := global.GVA_DB, global.GVA_CONFIG
	global.GVA_DB = db
	global.GVA_CONFIG = testConfig()
	t.Cleanup(func() { global.GVA_DB = prevDB; global.GVA_CONFIG = prevCfg })
	return db
}

func seedTicket(t *testing.T, db *gorm.DB, customerID uint) (supportModel.Ticket, supportModel.Conversation) {
	t.Helper()
	ticket := supportModel.Ticket{
		TicketCode: utils.NextCode(constants.CodePrefixSupportTicket),
		Subject:    "help", CustomerID: customerID,
		Status: constants.SupportTicketStatusOpen, Priority: constants.SupportPriorityNormal,
	}
	if err := db.Create(&ticket).Error; err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	conv := supportModel.Conversation{TicketID: ticket.ID, Status: constants.SupportConversationStatusActive}
	if err := db.Create(&conv).Error; err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	return ticket, conv
}

func TestOpenTicketCreatesConversationAndWelcome(t *testing.T) {
	db := setup(t)
	view, err := supportService.TicketServices{}.Open(
		supportService.Actor{Type: constants.ActorCustomer, ID: 51},
		supportReq.OpenTicketReq{Subject: "Where is my refund?"},
	)
	if err != nil {
		t.Fatalf("open ticket: %v", err)
	}
	if view.Conversation == nil || view.Conversation.MessageCount != 1 || view.Conversation.LastSeq != 1 {
		t.Fatalf("unexpected conversation: %+v", view.Conversation)
	}
	var msgs []supportModel.Message
	db.Where("conversation_id = ?", view.Conversation.ID).Find(&msgs)
	if len(msgs) != 1 || msgs[0].SenderType != constants.ActorSystem || msgs[0].Body == "" {
		t.Fatalf("welcome message missing: %+v", msgs)
	}
	if msgs[0].PrevHash != utils.HashChainZero {
		t.Fatalf("first message must chain from zero hash")
	}
}

func TestSendMessageDenseSeqAndIdempotency(t *testing.T) {
	db := setup(t)
	_, conv := seedTicket(t, db, 51)
	actor := supportService.Actor{Type: constants.ActorCustomer, ID: 51}
	svc := supportService.MessageServices{}

	first, err := svc.Send(actor, conv.ID, supportReq.SendMessageReq{Body: "hello <b>there</b>", ClientMsgID: "client-msg-0001"})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if first.Seq != 1 || first.Body != "hello there" {
		t.Fatalf("unexpected message: %+v", first)
	}
	second, err := svc.Send(actor, conv.ID, supportReq.SendMessageReq{Body: "hello again", ClientMsgID: "client-msg-0001"})
	if err != nil {
		t.Fatalf("resend: %v", err)
	}
	if second.MessageUUID != first.MessageUUID {
		t.Fatalf("retry must return the original message")
	}
	var count int64
	db.Model(&supportModel.Message{}).Where("conversation_id = ?", conv.ID).Count(&count)
	if count != 1 {
		t.Fatalf("idempotency broken, %d rows", count)
	}
	third, err := svc.Send(actor, conv.ID, supportReq.SendMessageReq{Body: "second", ClientMsgID: "client-msg-0002"})
	if err != nil {
		t.Fatalf("send second: %v", err)
	}
	if third.Seq != 2 {
		t.Fatalf("seq must be dense: %d", third.Seq)
	}
	if third.PrevHash == "" || third.BodyHash == third.PrevHash {
		t.Fatalf("hash chain not linked")
	}
}

func TestMessageAppendOnlyGuards(t *testing.T) {
	db := setup(t)
	_, conv := seedTicket(t, db, 51)
	view, err := supportService.MessageServices{}.Send(
		supportService.Actor{Type: constants.ActorCustomer, ID: 51}, conv.ID,
		supportReq.SendMessageReq{Body: "immutable", ClientMsgID: "client-msg-0003"},
	)
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	var msg supportModel.Message
	db.Where("message_uuid = ?", view.MessageUUID).First(&msg)
	if err := db.Model(&supportModel.Message{}).Where("id = ?", msg.ID).Update("body", "tampered").Error; err == nil {
		t.Fatal("expected update to be refused")
	}
	if err := db.Delete(&supportModel.Message{}, msg.ID).Error; err == nil {
		t.Fatal("expected delete to be refused")
	}
}

func TestAuthorizationScoping(t *testing.T) {
	db := setup(t)
	ticket, conv := seedTicket(t, db, 51)
	auth := supportService.AuthorizationServices{}

	if _, ok, _ := auth.CanRead(supportService.Actor{Type: constants.ActorCustomer, ID: 999}, conv.ID); ok {
		t.Fatal("foreign customer must not read")
	}
	if _, ok, _ := auth.CanRead(supportService.Actor{Type: constants.ActorSupport, ID: 9}, conv.ID); ok {
		t.Fatal("unassigned support employee must not read")
	}
	emp := uint(9)
	db.Model(&supportModel.Ticket{}).Where("id = ?", ticket.ID).Update("assigned_employee_id", emp)
	if _, ok, _ := auth.CanRead(supportService.Actor{Type: constants.ActorSupport, ID: 9}, conv.ID); !ok {
		t.Fatal("assigned support employee must read")
	}
	if _, ok, _ := auth.CanWrite(supportService.Actor{Type: constants.ActorQuality, ID: 3}, conv.ID); ok {
		t.Fatal("quality is read-only in v1")
	}
	if _, ok, _ := auth.CanWrite(supportService.Actor{Type: constants.ActorAdmin, ID: 1}, conv.ID); !ok {
		t.Fatal("admin must be able to write")
	}
}

func TestPresenceRefcountAndSingleTransition(t *testing.T) {
	db := setup(t)
	ticket, conv := seedTicket(t, db, 51)
	fake := &fakePublisher{}
	supportKafka.Configure(global.GVA_CONFIG.Support)
	supportKafka.SetPublisher(fake)
	supportService.SetPresenceStore(supportService.NewMemoryPresenceStore())
	t.Cleanup(func() {
		supportService.SetPresenceStore(nil)
		supportKafka.SetPublisher(nil)
	})

	svc := supportService.PresenceServices{}
	actor := supportService.Actor{Type: constants.ActorCustomer, ID: 51}
	ctx := context.Background()
	svc.Connect(ctx, actor, conv.ID, ticket.ID, "conn-1")
	svc.Connect(ctx, actor, conv.ID, ticket.ID, "conn-2")
	if got := fake.count("presence-test"); got != 1 {
		t.Fatalf("two connections must publish exactly one online transition, got %d", got)
	}
	svc.Disconnect(ctx, constants.ActorCustomer, 51, "conn-1", conv.ID, ticket.ID)
	if got := fake.count("presence-test"); got != 1 {
		t.Fatalf("dropping one of two connections must not publish, got %d", got)
	}
	svc.Disconnect(ctx, constants.ActorCustomer, 51, "conn-2", conv.ID, ticket.ID)
	if got := fake.count("presence-test"); got != 2 {
		t.Fatalf("final disconnect must publish exactly one offline transition, got %d", got)
	}
	var rows int64
	db.Model(&supportModel.Presence{}).Where("actor_type = ? AND actor_id = ?", constants.ActorCustomer, 51).Count(&rows)
	if rows != 1 {
		t.Fatalf("graceful disconnect must append exactly one last-seen row, got %d", rows)
	}
}

func TestPresenceSweepPublishesOfflineOnce(t *testing.T) {
	_, conv := seedTicket(t, setup(t), 51)
	fake := &fakePublisher{}
	supportKafka.Configure(global.GVA_CONFIG.Support)
	supportKafka.SetPublisher(fake)
	store := supportService.NewMemoryPresenceStore()
	supportService.SetPresenceStore(store)
	t.Cleanup(func() {
		supportService.SetPresenceStore(nil)
		supportKafka.SetPublisher(nil)
	})
	svc := supportService.PresenceServices{}
	actor := supportService.Actor{Type: constants.ActorCustomer, ID: 51}
	svc.Connect(context.Background(), actor, conv.ID, 0, "conn-x")

	expired, err := store.Sweep(context.Background(), time.Now().Add(2*time.Hour), 10)
	if err != nil || len(expired) != 1 {
		t.Fatalf("sweep should find one expired connection: %v %d", err, len(expired))
	}
	if again, _ := store.Sweep(context.Background(), time.Now().Add(2*time.Hour), 10); len(again) != 0 {
		t.Fatal("second sweep must not repeat the transition")
	}
}

func TestLastSeenHiding(t *testing.T) {
	setup(t)
	cfg := global.GVA_CONFIG
	cfg.Support.Presence.HideLastSeenForRoles = []string{"support"}
	global.GVA_CONFIG = cfg
	lastSeen := supportService.LastSeenServices{}
	if lastSeen.Visible("support") {
		t.Fatal("support role must be hidden")
	}
	if !lastSeen.Visible("customer") {
		t.Fatal("customer last seen must be visible")
	}
	cfg.Support.Presence.LastSeenDisplay = false
	global.GVA_CONFIG = cfg
	if lastSeen.Visible("customer") {
		t.Fatal("display disabled must hide everyone")
	}
}

func pngHeader(t *testing.T) *multipart.FileHeader {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var raw bytes.Buffer
	if err := png.Encode(&raw, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("file", "../../evil image.png")
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	_, _ = fw.Write(raw.Bytes())
	_ = w.Close()
	reader := multipart.NewReader(&body, w.Boundary())
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	return form.File["file"][0]
}

func TestAttachmentUploadLinkAndOrphanSweep(t *testing.T) {
	db := setup(t)
	dir := t.TempDir()
	cfg := global.GVA_CONFIG
	cfg.Local.StorePath = dir
	global.GVA_CONFIG = cfg
	_, conv := seedTicket(t, db, 51)
	actor := supportService.Actor{Type: constants.ActorCustomer, ID: 51}

	att, err := supportService.AttachmentServices{}.Upload(actor, conv.ID, pngHeader(t))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if att.Status != constants.SupportAttachmentStatusUploaded || att.MimeType != "image/png" {
		t.Fatalf("unexpected attachment: %+v", att)
	}
	if att.Width == nil || *att.Width != 1 || att.Height == nil || *att.Height != 1 {
		t.Fatalf("image dimensions not decoded: %+v", att)
	}
	if att.OriginalName != "evil_image.png" {
		t.Fatalf("filename not sanitized: %q", att.OriginalName)
	}
	if att.StorageKey == "" || att.StorageName == "" {
		t.Fatalf("storage returns missing: %+v", att)
	}

	view, err := supportService.MessageServices{}.Send(actor, conv.ID, supportReq.SendMessageReq{
		AttachmentUuids: []string{att.AttachmentUUID}, ClientMsgID: "client-msg-0004",
	})
	if err != nil {
		t.Fatalf("send with attachment: %v", err)
	}
	if len(view.Attachments) != 1 || view.Attachments[0].AttachmentUUID != att.AttachmentUUID {
		t.Fatalf("attachment not on message: %+v", view.Attachments)
	}
	var linked supportModel.Attachment
	db.Where("attachment_uuid = ?", att.AttachmentUUID).First(&linked)
	if linked.Status != constants.SupportAttachmentStatusAttached || linked.MessageID == nil {
		t.Fatalf("promotion did not happen: %+v", linked)
	}
	if err := db.Save(&linked).Error; err == nil {
		t.Fatal("second update after linking must be refused")
	}
	if err := db.Delete(&linked).Error; err == nil {
		t.Fatal("delete must be refused")
	}

	orphan := supportModel.Attachment{
		AttachmentUUID: utils.NewUUIDv7(), ConversationID: conv.ID, TicketID: conv.TicketID,
		UploaderType: constants.ActorCustomer, UploaderID: 51, Kind: "file",
		MimeType: "application/pdf", SizeBytes: 10, OriginalName: "x.pdf",
		StorageKey: "/uploads/x.pdf", StorageName: "x.pdf", Sha256: "x",
		Status: constants.SupportAttachmentStatusUploaded,
	}
	if err := db.Create(&orphan).Error; err != nil {
		t.Fatalf("create orphan: %v", err)
	}
	db.Exec("UPDATE support_attachment SET created_at = ? WHERE id = ?", time.Now().Add(-48*time.Hour), orphan.ID)
	if err := (supportService.AttachmentServices{}).OrphanCleanup(); err != nil {
		t.Fatalf("orphan sweep: %v", err)
	}
	var swept supportModel.Attachment
	db.First(&swept, orphan.ID)
	if swept.Status != constants.SupportAttachmentStatusRejected {
		t.Fatalf("orphan not rejected: %q", swept.Status)
	}
}

func TestAttachmentRejectsDisallowedType(t *testing.T) {
	db := setup(t)
	_, conv := seedTicket(t, db, 51)
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, _ := w.CreateFormFile("file", "notes.exe")
	_, _ = fw.Write([]byte("MZ"))
	_ = w.Close()
	form, _ := multipart.NewReader(&body, w.Boundary()).ReadForm(1 << 20)
	_, err := supportService.AttachmentServices{}.Upload(
		supportService.Actor{Type: constants.ActorCustomer, ID: 51}, conv.ID, form.File["file"][0])
	if err == nil {
		t.Fatal("disallowed extension must be rejected before storage")
	}
}

func TestFaqVoteAppendsLatestWins(t *testing.T) {
	db := setup(t)
	faq := supportModel.Faq{
		Code: "poll", Question: "q", Answer: "a", IsActive: true,
		Options: common.JSONMap{"options": []map[string]string{
			{"key": constants.SupportVoteHelpful, "label": "Helpful"},
			{"key": constants.SupportVoteNotHelpful, "label": "Not helpful"},
		}},
	}
	if err := db.Create(&faq).Error; err != nil {
		t.Fatalf("create faq: %v", err)
	}
	svc := supportService.FaqServices{}
	if err := svc.Vote(51, faq.ID, supportReq.VoteReq{OptionKey: constants.SupportVoteHelpful}); err != nil {
		t.Fatalf("vote: %v", err)
	}
	if err := svc.Vote(51, faq.ID, supportReq.VoteReq{OptionKey: constants.SupportVoteHelpful}); err != nil {
		t.Fatalf("repeat vote: %v", err)
	}
	var count int64
	db.Model(&supportModel.FaqVote{}).Count(&count)
	if count != 1 {
		t.Fatalf("same option must be idempotent, got %d rows", count)
	}
	if err := svc.Vote(51, faq.ID, supportReq.VoteReq{OptionKey: constants.SupportVoteNotHelpful}); err != nil {
		t.Fatalf("change vote: %v", err)
	}
	db.Model(&supportModel.FaqVote{}).Count(&count)
	if count != 2 {
		t.Fatalf("changed vote must append, got %d rows", count)
	}
	views, err := svc.List(51)
	if err != nil || len(views) != 1 {
		t.Fatalf("list: %v %d", err, len(views))
	}
	if views[0].Tally[constants.SupportVoteNotHelpful] != 1 || views[0].MyVote != constants.SupportVoteNotHelpful {
		t.Fatalf("tally must count only the latest vote: %+v", views[0])
	}
}

func TestMessageRequiresBodyOrAttachment(t *testing.T) {
	db := setup(t)
	_, conv := seedTicket(t, db, 51)
	_, err := supportService.MessageServices{}.Send(
		supportService.Actor{Type: constants.ActorCustomer, ID: 51}, conv.ID,
		supportReq.SendMessageReq{Body: "   ", ClientMsgID: "client-msg-0005"})
	if err == nil {
		t.Fatal("empty message without attachment must be rejected")
	}
}
