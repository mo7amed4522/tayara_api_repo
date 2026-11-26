package kafka

import (
	"context"

	"github.com/mo7amed4522/tayara_api/config"
)

// PresencePublisher is a deliberately separate writer contract for the presence topic: shorter
// retention, rare transitions, and its own batching budget. It must never inherit the chat
// writer's linger/retry behaviour.
type PresencePublisher interface {
	PublishPresence(ctx context.Context, key string, payload []byte) error
	Close() error
}

// NewPresenceClient builds a dedicated client whose only topic is the presence topic.
func NewPresenceClient(cfg config.SupportKafka) (*Client, error) {
	return New(cfg)
}

// PublishTransition sends one 0→1 or 1→0 presence transition to the presence topic.
func PublishTransition(ctx context.Context, key string, payload []byte) error {
	return PublishPresence(ctx, key, payload)
}
