package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mo7amed4522/tayara_api/config"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// Publisher is the transport seam: everything in the support module publishes through it, and
// tests swap in a fake. Close flushes and releases writers.
type Publisher interface {
	Publish(ctx context.Context, topic, key string, payload []byte) error
	Close() error
}

type noopPublisher struct{}

func (noopPublisher) Publish(context.Context, string, string, []byte) error { return nil }
func (noopPublisher) Close() error                                          { return nil }

var (
	mu        sync.RWMutex
	publisher Publisher = noopPublisher{}
	topicsCfg config.SupportKafkaTopics
	enabled   bool
)

// Configure records the deployment's topic names and whether the bus is enabled. Called once at
// boot before any service publishes.
func Configure(cfg config.SupportConfig) {
	mu.Lock()
	defer mu.Unlock()
	topicsCfg = cfg.Kafka.Topics
	enabled = cfg.Kafka.Enabled
}

// SetPublisher installs the active publisher (a real client at boot, a fake in tests).
func SetPublisher(p Publisher) {
	mu.Lock()
	defer mu.Unlock()
	if p == nil {
		publisher = noopPublisher{}
		return
	}
	publisher = p
}

// Active reports whether the Kafka transport is enabled in configuration.
func Active() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// Topic names, defaulted so an empty config still yields stable names.
func MessagesTopic() string {
	return topicOrDefault(topicsCfg.ChatMessages, "tayara.support.chat.messages.v1")
}
func EventsTopic() string {
	return topicOrDefault(topicsCfg.ChatEvents, "tayara.support.chat.events.v1")
}
func DLQTopic() string {
	return topicOrDefault(topicsCfg.ChatMessagesDLQ, "tayara.support.chat.messages.dlq.v1")
}
func PresenceTopic() string { return topicOrDefault(topicsCfg.Presence, "tayara.support.presence.v1") }

func topicOrDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func current() Publisher {
	mu.RLock()
	defer mu.RUnlock()
	return publisher
}

// Publish sends one payload to one topic, keyed by key (conversation id or actor id).
func Publish(ctx context.Context, topic, key string, payload []byte) error {
	return current().Publish(ctx, topic, key, payload)
}

// PublishChat publishes a chat message envelope.
func PublishChat(ctx context.Context, key string, payload []byte) error {
	return Publish(ctx, MessagesTopic(), key, payload)
}

// PublishEvent publishes a ticket/faq event envelope.
func PublishEvent(ctx context.Context, key string, payload []byte) error {
	return Publish(ctx, EventsTopic(), key, payload)
}

// PublishPresence publishes a presence transition envelope.
func PublishPresence(ctx context.Context, key string, payload []byte) error {
	return Publish(ctx, PresenceTopic(), key, payload)
}

// Client is the kafka-go implementation: one Writer per topic, all sharing a transport built from
// the TLS/SASL config.
type Client struct {
	mu      sync.Mutex
	writers map[string]*kafka.Writer
	cfg     config.SupportKafka
}

// New builds a Client. It never dials, so an unreachable broker cannot block boot; the first
// publish reports the error and the outbox sweeper retries.
func New(cfg config.SupportKafka) (*Client, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("no kafka brokers configured")
	}
	switch cfg.SASL.Mechanism {
	case "", "plain", "scram-sha256", "scram-sha512":
	default:
		return nil, errors.New("unsupported sasl mechanism")
	}
	return &Client{writers: map[string]*kafka.Writer{}, cfg: cfg}, nil
}

// Publish writes one record, creating (and caching) the topic's writer lazily.
func (c *Client) Publish(ctx context.Context, topic, key string, payload []byte) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.publishTimeout())
		defer cancel()
	}
	return c.writerFor(topic).WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: payload,
	})
}

func (c *Client) publishTimeout() time.Duration {
	if c.cfg.Producer.TimeoutMS > 0 {
		return time.Duration(c.cfg.Producer.TimeoutMS) * time.Millisecond
	}
	return 5 * time.Second
}

func (c *Client) writerFor(topic string) *kafka.Writer {
	c.mu.Lock()
	defer c.mu.Unlock()
	if w, ok := c.writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:                   kafka.TCP(c.cfg.Brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		RequiredAcks:           acksFor(c.cfg.RequiredAcks),
		AllowAutoTopicCreation: c.cfg.AutoCreateTopics,
		Async:                  false,
		BatchTimeout:           time.Duration(maxInt(c.cfg.Producer.LingerMS, 1)) * time.Millisecond,
		BatchBytes:             int64(maxInt(c.cfg.Producer.BatchBytes, 1048576)),
		Transport:              transportFor(c.cfg),
	}
	c.writers[topic] = w
	return w
}

// Close flushes and closes every writer.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error
	for name, w := range c.writers {
		if e := w.Close(); e != nil {
			err = e
		}
		delete(c.writers, name)
	}
	return err
}

func acksFor(v string) kafka.RequiredAcks {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "none":
		return kafka.RequireNone
	case "leader", "one":
		return kafka.RequireOne
	default:
		return kafka.RequireAll
	}
}

func transportFor(cfg config.SupportKafka) *kafka.Transport {
	t := &kafka.Transport{ClientID: cfg.ClientID}
	if cfg.TLS.Enabled {
		t.TLS = buildTLS(cfg.TLS)
	}
	if cfg.SASL.Enabled {
		t.SASL = buildSASL(cfg.SASL)
	}
	return t
}

func buildTLS(c config.SupportKafkaTLS) *tls.Config {
	out := &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify} // #nosec G402 - opt-in via config
	if c.CAFile != "" {
		if pem, err := os.ReadFile(c.CAFile); err == nil {
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(pem)
			out.RootCAs = pool
		}
	}
	if c.CertFile != "" && c.KeyFile != "" {
		if cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile); err == nil {
			out.Certificates = []tls.Certificate{cert}
		}
	}
	return out
}

func buildSASL(c config.SupportKafkaSASL) sasl.Mechanism {
	switch strings.ToLower(strings.TrimSpace(c.Mechanism)) {
	case "plain":
		return plain.Mechanism{Username: c.Username, Password: c.Password}
	case "scram-sha256":
		m, err := scram.Mechanism(scram.SHA256, c.Username, c.Password)
		if err == nil {
			return m
		}
	case "scram-sha512":
		m, err := scram.Mechanism(scram.SHA512, c.Username, c.Password)
		if err == nil {
			return m
		}
	}
	return plain.Mechanism{Username: c.Username, Password: c.Password}
}

func maxInt(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
