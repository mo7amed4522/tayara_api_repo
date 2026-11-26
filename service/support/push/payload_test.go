package push_test

import (
	"strings"
	"testing"

	"github.com/mo7amed4522/tayara_api/model/support/response"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/mo7amed4522/tayara_api/service/support/push"
)

func TestBuildPayloadTruncatesTitleAndPreview(t *testing.T) {
	subject := strings.Repeat("s", 200)
	body := strings.Repeat("b", 500)
	p := push.BuildPayload(42, 7, subject, supportKafka.MessageCreatedData{
		MessageUUID: "m-1", Seq: 13, SenderType: "customer", SenderID: 51, Body: body,
	}, 60, 140)
	if len(p.Title) > 60 {
		t.Fatalf("title exceeds cap: %d", len(p.Title))
	}
	if len(p.Body) > 140 {
		t.Fatalf("body exceeds preview cap: %d", len(p.Body))
	}
	if !strings.HasPrefix(p.Body, "Customer: ") {
		t.Fatalf("body must carry the sender prefix: %q", p.Body)
	}
	if len(p.Body) >= len(body) {
		t.Fatalf("body must only be a preview, got %d bytes", len(p.Body))
	}
}

func TestBuildPayloadAttachmentOnly(t *testing.T) {
	p := push.BuildPayload(42, 7, "Receipt", supportKafka.MessageCreatedData{
		MessageUUID: "m-2", Seq: 4, SenderType: "support", SenderID: 9,
		Attachments: []response.AttachmentView{{AttachmentUUID: "a-1", Kind: "image"}},
	}, 60, 140)
	if p.Body != "📎 Attachment" {
		t.Fatalf("attachment-only body mismatch: %q", p.Body)
	}
	if p.Data["seq"] != "4" || p.Data["conversation_id"] != "42" || p.Data["ticket_id"] != "7" {
		t.Fatalf("data keys incomplete: %+v", p.Data)
	}
	if p.Data["message_uuid"] != "m-2" || p.Data["event_type"] != supportKafka.EventMessageCreated {
		t.Fatalf("data keys incomplete: %+v", p.Data)
	}
}

func TestBuildPayloadArabicNotSplitMidRune(t *testing.T) {
	body := strings.Repeat("مرحبا بك في الدعم ", 20)
	p := push.BuildPayload(1, 2, "موضوع التذكرة", supportKafka.MessageCreatedData{
		MessageUUID: "m-3", Seq: 2, SenderType: "customer", SenderID: 51, Body: body,
	}, 12, 40)
	if !utf8Valid(p.Title) || !utf8Valid(p.Body) {
		t.Fatalf("truncation split a rune: title=%q body=%q", p.Title, p.Body)
	}
	if len(p.Body) > 40 {
		t.Fatalf("body exceeds preview cap: %d", len(p.Body))
	}
}

func TestBuildPayloadDropsMarkup(t *testing.T) {
	p := push.BuildPayload(1, 2, "<b>Subject</b>", supportKafka.MessageCreatedData{
		MessageUUID: "m-4", Seq: 1, SenderType: "customer", SenderID: 51, Body: "hello <script>x</script>there",
	}, 60, 140)
	if strings.Contains(p.Body, "<") || strings.Contains(p.Title, "<") {
		t.Fatalf("markup leaked into payload: %q / %q", p.Title, p.Body)
	}
}

func utf8Valid(s string) bool {
	for _, r := range s {
		if r == '\uFFFD' {
			return false
		}
	}
	return true
}
