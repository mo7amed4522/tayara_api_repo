package utils

import (
	"strings"
	"testing"
)

func TestSanitizeChatTextStripsTagsAndKeepsUnicode(t *testing.T) {
	got := SanitizeChatText("  hello <script>alert(1)</script>\u0000  مرحبا 🎉  ", 500)
	if strings.Contains(got, "<") || strings.Contains(got, "script") {
		t.Fatalf("tags not stripped: %q", got)
	}
	if !strings.Contains(got, "مرحبا") || !strings.Contains(got, "🎉") {
		t.Fatalf("unicode lost: %q", got)
	}
	if got != strings.TrimSpace(got) {
		t.Fatalf("not trimmed: %q", got)
	}
}

func TestSanitizeChatTextWhitespaceOnly(t *testing.T) {
	if got := SanitizeChatText("   \n\t  ", 100); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestSanitizeChatTextCapsBytes(t *testing.T) {
	got := SanitizeChatText(strings.Repeat("a", 200), 50)
	if len(got) > 50 {
		t.Fatalf("byte cap not enforced: %d", len(got))
	}
}

func TestSanitizeFilename(t *testing.T) {
	cases := map[string]string{
		"../../etc/passwd":    "passwd",
		`C:\temp\..\evil.png`: "evil.png",
		"my receipt (1).png":  "my_receipt_(1).png",
		"a\x00b<>.pdf":        "ab.pdf",
		"":                    "file",
	}
	for in, want := range cases {
		if got := SanitizeFilename(in); got != want {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", in, got, want)
		}
	}
	long := strings.Repeat("x", 400) + ".png"
	if got := SanitizeFilename(long); len(got) > 255 {
		t.Fatalf("filename not capped: %d", len(got))
	}
}
