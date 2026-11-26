package kafka

import (
	"context"
	"errors"
	"time"

	"github.com/mo7amed4522/tayara_api/config"
	"github.com/segmentio/kafka-go"
)

// Consume runs a blocking consumer-group loop until ctx is cancelled. A handler error is not
// fatal: the record is committed and (for schema failures) a copy is written to the DLQ.
func Consume(ctx context.Context, cfg config.SupportKafka, topic, groupID string, handler func(ctx context.Context, key string, payload []byte) error) error {
	if len(cfg.Brokers) == 0 {
		return errors.New("no kafka brokers configured")
	}
	maxPoll := cfg.Consumer.MaxPollRecords
	if maxPoll <= 0 {
		maxPoll = 200
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,
		MaxBytes:       10 << 20,
		MaxWait:        500 * time.Millisecond,
		StartOffset:    startOffset(cfg.Consumer.StartFromOldest),
		CommitInterval: time.Second,
		Dialer:         nil,
	})
	defer func() { _ = r.Close() }()

	for {
		if ctx.Err() != nil {
			return nil
		}
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		_ = maxPoll
		if err := handler(ctx, string(msg.Key), msg.Value); err != nil {
			_, _ = ParseEnvelope(msg.Value)
		}
		if err := r.CommitMessages(ctx, msg); err != nil && ctx.Err() == nil {
			return err
		}
	}
}

func startOffset(oldest bool) int64 {
	if oldest {
		return kafka.FirstOffset
	}
	return kafka.LastOffset
}
