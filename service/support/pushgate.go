package support

import (
	"time"

	"github.com/mo7amed4522/tayara_api/service/support/push"
)

// The push package must not import this package (its ServiceGroup is embedded here), so the
// concrete presence/collapse implementations are injected at init. Presence is therefore still
// only ever read through service/support/presence.go.
func init() {
	push.SetPresenceGate(presenceGate{})
	push.SetCollapseGate(collapseGate{})
}

type presenceGate struct{}

func (presenceGate) IsOnline(actorType string, actorID uint) (bool, bool) {
	return (PresenceServices{}).IsOnline(actorType, actorID)
}

type collapseGate struct{}

func (collapseGate) ClaimPushWindow(recipient string, conversationID uint, window time.Duration) (bool, bool) {
	return (PresenceServices{}).ClaimPushWindow(recipient, conversationID, window)
}
