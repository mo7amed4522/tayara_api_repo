package utils

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Code generator for the PRD's per-domain sequence prefixes (BK-, CT-, PK-, AG-, US-, NT-, AD-).
// Each prefix has its own monotonically increasing, concurrency-safe counter.
// Seed from the table's current max sequence at startup (SeedCode), then call NextCode to mint.

var codeCounters sync.Map // string -> *atomic.Uint64

// SeedCode sets the next sequence for a prefix (call with the max existing sequence number).
func SeedCode(prefix string, lastSeq uint64) {
	next := &atomic.Uint64{}
	next.Store(lastSeq + 1)
	codeCounters.Store(prefix, next)
}

// NextCode mints the next code for a prefix, e.g. NextCode("BK-") -> "BK-10021".
// If the prefix was never seeded it starts at 1.
func NextCode(prefix string) string {
	v, _ := codeCounters.LoadOrStore(prefix, &atomic.Uint64{})
	counter := v.(*atomic.Uint64)
	seq := counter.Add(1)
	return fmt.Sprintf("%s%d", prefix, seq)
}
