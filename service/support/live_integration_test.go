package support_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	supportService "github.com/mo7amed4522/tayara_api/service/support"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// These two tests need real infrastructure and skip unless the matching env var is set:
//
//	SUPPORT_REDIS_ADDR=localhost:6379 go test ./service/support/... -run LiveRedis -v
//	KAFKA_BROKERS=localhost:9092       go test ./service/support/... -run LiveKafka -v
//
// docker compose up -d redis redpanda provides both locally.

// TestLiveRedisPresence exercises the Redis presence registry (refcount + index cleanup).
func TestLiveRedisPresence(t *testing.T) {
	addr := os.Getenv("SUPPORT_REDIS_ADDR")
	if addr == "" {
		t.Skip("SUPPORT_REDIS_ADDR not set; skipping live Redis test")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not reachable at %s: %v", addr, err)
	}
	prev := global.GVA_REDIS
	global.GVA_REDIS = client
	supportService.SetPresenceStore(nil)

	actorID := uint(time.Now().UnixNano()%900000) + 1000
	connKey := constants.SupportPresenceConnKeyPrefix + constants.ActorSupport + ":" + fmt.Sprint(actorID)
	indexA := constants.ActorSupport + ":" + fmt.Sprint(actorID) + ":conn-a"
	indexB := constants.ActorSupport + ":" + fmt.Sprint(actorID) + ":conn-b"
	defer func() {
		_ = client.Del(context.Background(), connKey).Err()
		_ = client.ZRem(context.Background(), constants.SupportPresenceIndexKey, indexA, indexB).Err()
		global.GVA_REDIS = prev
		supportService.SetPresenceStore(nil)
	}()

	store := supportService.Store()
	if store == nil {
		t.Fatal("Store() must return a Redis-backed store when GVA_REDIS is set")
	}
	m := supportService.PresenceMember{ActorType: constants.ActorSupport, ActorID: actorID, ConnID: "conn-a"}
	if n, err := store.Touch(ctx, m, 30*time.Second); err != nil || n != 1 {
		t.Fatalf("first connection: n=%d err=%v", n, err)
	}
	m2 := supportService.PresenceMember{ActorType: constants.ActorSupport, ActorID: actorID, ConnID: "conn-b"}
	if n, err := store.Touch(ctx, m2, 30*time.Second); err != nil || n != 2 {
		t.Fatalf("second connection must refcount to 2: n=%d err=%v", n, err)
	}
	if n, err := store.Drop(ctx, m); err != nil || n != 1 {
		t.Fatalf("dropping one connection must leave 1: n=%d err=%v", n, err)
	}
	if n, err := store.Drop(ctx, m2); err != nil || n != 0 {
		t.Fatalf("dropping the last connection must leave 0: n=%d err=%v", n, err)
	}
	if _, err := client.ZScore(ctx, constants.SupportPresenceIndexKey, indexA).Result(); err != redis.Nil {
		t.Fatalf("online index entry must be removed, got err=%v", err)
	}
	if actors, err := store.OnlineActors(ctx); err != nil {
		t.Fatalf("online actors: %v", err)
	} else {
		for _, a := range actors {
			if a.ActorType == constants.ActorSupport && a.ActorID == actorID {
				t.Fatal("actor must no longer be online")
			}
		}
	}
}

// TestLiveKafkaRoundTrip publishes through the production Client and reads the record back.
func TestLiveKafkaRoundTrip(t *testing.T) {
	raw := os.Getenv("KAFKA_BROKERS")
	if raw == "" {
		t.Skip("KAFKA_BROKERS not set; skipping live Kafka test")
	}
	brokers := []string{}
	for _, b := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(b); v != "" {
			brokers = append(brokers, v)
		}
	}
	topic := fmt.Sprintf("tayara.support.selftest.%d", time.Now().UnixNano())
	client, err := supportKafka.New(config.SupportKafka{
		Enabled:          true,
		Brokers:          brokers,
		ClientID:         "tayara-selftest",
		RequiredAcks:     "all",
		AutoCreateTopics: true,
		Producer:         config.SupportKafkaProducer{TimeoutMS: 5000, LingerMS: 1, BatchBytes: 1048576},
	})
	if err != nil {
		t.Fatalf("build client: %v", err)
	}
	defer func() { _ = client.Close() }()

	payload, err := supportKafka.NewEnvelope(supportKafka.EventPresenceChange, 42, 7, supportKafka.PresenceChangedData{
		ActorType: constants.ActorSupport, ActorID: 9, State: constants.SupportPresenceOnline,
	})
	if err != nil {
		t.Fatalf("build envelope: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if err := client.Publish(ctx, topic, "42", payload); err != nil {
		t.Fatalf("publish: %v", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		Partition:   0,
		StartOffset: kafka.FirstOffset,
		MaxWait:     time.Second,
		Dialer:      nil,
	})
	defer func() { _ = reader.Close() }()
	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(msg.Key) != "42" {
		t.Fatalf("record key must be the conversation id, got %q", msg.Key)
	}
	env, err := supportKafka.ParseEnvelope(msg.Value)
	if err != nil {
		t.Fatalf("parse envelope: %v", err)
	}
	if env.ConversationID != 42 || env.EventType != supportKafka.EventPresenceChange {
		t.Fatalf("unexpected envelope: %+v", env)
	}
}
