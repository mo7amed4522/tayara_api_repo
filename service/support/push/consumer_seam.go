package push

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// PushConsumerGroup is the single shared consumer group for offline delivery. It is intentionally
// NOT the per-instance fanout group: a push must be sent once per recipient, not once per replica.
const PushConsumerGroup = "tayara-support-push"

// Fetcher is the seam over kafka.Reader so tests never dial a broker.
type Fetcher interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}
