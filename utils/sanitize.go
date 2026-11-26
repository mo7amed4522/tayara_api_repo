package utils

import (
	"html"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// SanitizeChatText makes user chat text safe to persist and re-render as plain text:
// tags are stripped, HTML entities are decoded, control characters (except newline and tab)
// are dropped, runs of blank lines are collapsed, and the result is capped to maxBytes on a
// valid UTF-8 boundary. Arabic and emoji are preserved. Returns "" for whitespace-only input.
func SanitizeChatText(raw string, maxBytes int) string {
	if raw == "" {
		return ""
	}
	// Decode entities first, then strip tags, then decode once more (handles &lt;b&gt;).
	s := html.UnescapeString(raw)
	s = stripTags(s)
	s = html.UnescapeString(s)

	var b strings.Builder
	b.Grow(len(s))
	lastNewline := false
	for _, r := range s {
		switch {
		case r == '\n':
			if lastNewline {
				continue
			}
			lastNewline = true
			b.WriteRune('\n')
			continue
		case r == '\t':
			b.WriteRune(' ')
			continue
		case r == '\r':
			continue
		case unicode.IsControl(r):
			continue
		case unicode.IsSpace(r):
			lastNewline = false
			b.WriteRune(' ')
			continue
		}
		lastNewline = false
		b.WriteRune(r)
	}

	out := strings.Join(strings.Fields(strings.ReplaceAll(b.String(), "\n", " \n ")), " ")
	// Fields collapses the padding we added; rebuild newlines from the collapsed lines.
	lines := make([]string, 0, 8)
	for _, line := range strings.Split(b.String(), "\n") {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			lines = append(lines, line)
		}
	}
	out = strings.Join(lines, "\n")
	out = strings.TrimSpace(out)
	if out == "" {
		return ""
	}
	if maxBytes > 0 {
		out = truncateUTF8(out, maxBytes)
	}
	return out
}

// stripTags removes anything that looks like an HTML/XML tag and any dangling '<' or '>'.
func stripTags(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag && r != '<' && r != '>':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// truncateUTF8 cuts s so that it is at most maxBytes long without splitting a rune.
func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return strings.TrimSpace(s[:cut])
}

// SanitizeFilename reduces an uploaded filename to a safe stored original name: the basename
// only (no directory components), no traversal sequences, no separators or control characters,
// the extension preserved, and capped at 255 bytes.
func SanitizeFilename(name string) string {
	if name == "" {
		return "file"
	}
	name = strings.ReplaceAll(name, "\\", "/")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	name = strings.ReplaceAll(name, "..", "")
	ext := strings.ToLower(filepath.Ext(name))
	base := strings.TrimSuffix(name, filepath.Ext(name))

	var b strings.Builder
	for _, r := range base {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' ||
			r == '<' || r == '>' || r == '|' || unicode.IsControl(r) {
			continue
		}
		if unicode.IsSpace(r) {
			b.WriteRune('_')
			continue
		}
		b.WriteRune(r)
	}
	clean := strings.Trim(b.String(), "._-")
	if clean == "" {
		clean = "file"
	}
	if len(ext) > 16 {
		ext = ""
	}
	return truncateUTF8(clean, 255-len(ext)) + ext
}
