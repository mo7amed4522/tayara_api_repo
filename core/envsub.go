package core

import (
	"os"
	"strings"
)

// expandEnv performs env substitution on a config payload, supporting the two forms used
// across config.yaml:
//
//	${VAR}          -> value of VAR, or "" if unset
//	${VAR:-default} -> value of VAR, or "default" if unset/empty
//
// Unlike os.ExpandEnv, this handles the `:-default` (and `-default`) fallback syntax, so the
// `${STRIPE_SECRET_KEY:-sk_test_...}` placeholders in config.yaml behave as intended and a
// deployed image can inject secrets through the process environment.
func expandEnv(raw string) string {
	var b strings.Builder
	b.Grow(len(raw))
	i := 0
	n := len(raw)
	for i < n {
		c := raw[i]
		if c != '$' || i+1 >= n || (raw[i+1] != '{') {
			b.WriteByte(c)
			i++
			continue
		}
		end := strings.IndexByte(raw[i+2:], '}')
		if end < 0 {
			// no closing brace — treat literally
			b.WriteString(raw[i:])
			break
		}
		end += i + 2
		expr := raw[i+2 : end]
		i = end + 1
		b.WriteString(evalEnvExpr(expr))
	}
	return b.String()
}

func evalEnvExpr(expr string) string {
	name := expr
	hasDefault := false
	preserveEmpty := false
	if j := strings.Index(expr, ":-"); j >= 0 {
		name = expr[:j]
		hasDefault = true
	} else if j := strings.Index(expr, "-"); j >= 0 {
		name = expr[:j]
		hasDefault = true
		preserveEmpty = true
	}
	val, ok := os.LookupEnv(name)
	if ok && val != "" {
		return val
	}
	if ok && preserveEmpty {
		// `${VAR-default}`: empty-but-set should yield the (empty) value, not the default
		return val
	}
	if hasDefault {
		d := expr[len(name):]
		d = strings.TrimPrefix(d, ":")
		d = strings.TrimPrefix(d, "-")
		return d
	}
	return ""
}