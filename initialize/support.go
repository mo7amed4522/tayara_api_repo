package initialize

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/global"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/mo7amed4522/tayara_api/service/support/stream"
	"go.uber.org/zap"
)

var (
	supportOnce     sync.Once
	supportCancel   context.CancelFunc
	supportClient   *supportKafka.Client
	supportClientMu sync.Mutex
)

// Support boots the Kafka transport and the SSE fan-out consumers. It is non-fatal: a disabled
// or unreachable broker leaves chat working through the database + local hub.
func Support() {
	supportOnce.Do(func() {
		supportKafka.Configure(global.GVA_CONFIG.Support)
		if !global.GVA_CONFIG.Support.Kafka.Enabled {
			if global.GVA_LOG != nil {
				global.GVA_LOG.Info("support kafka disabled; messages persist and fan out locally")
			}
			return
		}
		client, err := supportKafka.New(global.GVA_CONFIG.Support.Kafka)
		if err != nil {
			if global.GVA_LOG != nil {
				global.GVA_LOG.Warn("support kafka init failed", zap.Error(err))
			}
			return
		}
		supportClientMu.Lock()
		supportClient = client
		supportClientMu.Unlock()
		supportKafka.SetPublisher(client)

		ctx, cancel := context.WithCancel(context.Background())
		supportCancel = cancel
		cfg := global.GVA_CONFIG.Support.Kafka
		group := cfg.FanoutGroupPrefix
		if group == "" {
			group = "tayara-support-fanout"
		}
		group += "-" + instanceID()

		go consumeFanout(ctx, cfg, supportKafka.MessagesTopic(), group)
		go consumeFanout(ctx, cfg, supportKafka.PresenceTopic(), group+"-presence")
	})
}

// consumeFanout relays bus records onto this instance's SSE hub.
func consumeFanout(ctx context.Context, cfg config.SupportKafka, topic, group string) {
	for {
		err := supportKafka.Consume(ctx, cfg, topic, group, func(_ context.Context, _ string, payload []byte) error {
			env, perr := supportKafka.ParseEnvelope(payload)
			if perr != nil {
				if global.GVA_LOG != nil {
					global.GVA_LOG.Warn("support kafka envelope rejected", zap.Error(perr))
				}
				return perr
			}
			if env.ConversationID > 0 {
				stream.Default.Broadcast(env.ConversationID, stream.Event{Type: env.EventType, Data: payload})
			}
			return nil
		})
		if ctx.Err() != nil {
			return
		}
		if global.GVA_LOG != nil {
			global.GVA_LOG.Warn("support fanout consumer stopped; retrying", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

// ShutdownSupport stops consumers, flushes the producers and closes SSE subscribers.
func ShutdownSupport() {
	if supportCancel != nil {
		supportCancel()
	}
	time.Sleep(100 * time.Millisecond)
	supportClientMu.Lock()
	client := supportClient
	supportClient = nil
	supportClientMu.Unlock()
	if client != nil {
		_ = client.Close()
	}
	supportKafka.SetPublisher(nil)
	stream.Default.CloseAll()
}

// instanceID identifies this replica's fanout consumer group.
func instanceID() string {
	if v := os.Getenv("INSTANCE_ID"); v != "" {
		return v
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "node"
	}
	return host + "-" + strconv.Itoa(os.Getpid())
}
