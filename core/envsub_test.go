package core

import (
	"testing"
)

func TestExpandEnv(t *testing.T) {
	t.Setenv("TAYARA_SET", "real-value")
	t.Setenv("TAYARA_EMPTY", "")

	cases := []struct {
		in, want string
	}{
		// plain
		{in: "no placeholders here", want: "no placeholders here"},
		{in: "", want: ""},
		// ${VAR}
		{in: "x${KAYARA_UNSET}x", want: "xx"},
		{in: "x${TAYARA_SET}x", want: "xreal-valuex"},
		// ${VAR:-default} fallback
		{in: "${TAYARA_UNSET:-fallback}", want: "fallback"},
		{in: "${KAYARA_UNSET:-fallback}", want: "fallback"},
		{in: "${TAYARA_SET:-fallback}", want: "real-value"},
		// empty-but-set with :- -> default; with - -> empty
		{in: "${TAYARA_EMPTY:-dk}", want: "dk"},
		{in: "${TAYARA_EMPTY-dk}", want: ""},
		// literal around braces
		{in: "${KAYARA_UNSET}${KAYARA_UNSET}", want: ""},
		// unmatched brace stays literal
		{in: "${NOT_CLOSED", want: "${NOT_CLOSED"},
		// nested-in-text
		{in: "a ${KAYARA_UNSET:-D} b", want: "a D b"},
	}
	for _, c := range cases {
		if got := expandEnv(c.in); got != c.want {
			t.Errorf("expandEnv(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}