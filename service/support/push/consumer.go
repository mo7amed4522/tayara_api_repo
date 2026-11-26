package push

import (
	"context"

	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/global"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"go.uber.org/zap"
)

// Consume runs the shared push consumer group until ctx is cancelled.
func Consume(ctx context.Context, cfg config.SupportKafka, d *Dispatcher) error {
	reader, err := supportKafka.NewPushReader(cfg, supportKafka.MessagesTopic(), PushConsumerGroup)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()
	return ConsumeFetcher(ctx, reader, d)
}

// ConsumeFetcher is the broker-free loop: parse, dispatch message.created, commit. Anything else
// on the topic (presence.changed, ticket events) is ignored here.
func ConsumeFetcher(ctx context.Context, f Fetcher, d *Dispatcher) error {
	for {
		if ctx.Err() != nil {
			return nil
		}
		msg, err := f.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		env, perr := supportKafka.ParseEnvelope(msg.Value)
		if perr == nil && env.EventType == supportKafka.EventMessageCreated {
			if derr := d.DispatchForEvent(ctx, env); derr != nil && global.GVA_LOG != nil {
				global.GVA_LOG.Warn("support push dispatch failed", zap.Error(derr))
			}
		}
		if err := f.CommitMessages(ctx, msg); err != nil && ctx.Err() == nil {
			return err
		}
	}
}
