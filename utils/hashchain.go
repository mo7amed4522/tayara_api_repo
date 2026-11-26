package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HashChainZero is the prev_hash of the first message in a conversation.
const HashChainZero = "0000000000000000000000000000000000000000000000000000000000000000"

// HashChainStep computes one tamper-evident link: SHA-256 over the previous hash and the
// supplied parts, joined with a NUL separator so field boundaries cannot be forged.
func HashChainStep(prev string, parts ...string) string {
	if prev == "" {
		prev = HashChainZero
	}
	h := sha256.New()
	h.Write([]byte(prev))
	for _, p := range parts {
		h.Write([]byte{0})
		h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyChain recomputes the chain over bodies (already in seq order) and reports whether the
// final link matches want. A tampered body breaks verification.
func VerifyChain(bodies []string, want string) bool {
	prev := HashChainZero
	for _, b := range bodies {
		prev = HashChainStep(prev, b)
	}
	return strings.EqualFold(prev, want)
}
