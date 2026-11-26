package push_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/mo7amed4522/tayara_api/service/support/push"
	"github.com/segmentio/kafka-go"
)

var errDrained = errors.New("drained")

type fakeFetcher struct {
	msgs      []kafka.Message
	committed int
}

func (f *fakeFetcher) FetchMessage(context.Context) (kafka.Message, error) {
	if len(f.msgs) == 0 {
		return kafka.Message{}, errDrained
	}
	msg := f.msgs[0]
	f.msgs = f.msgs[1:]
	return msg, nil
}

func (f *fakeFetcher) CommitMessages(context.Context, ...kafka.Message) error {
	f.committed++
	return nil
}

func (f *fakeFetcher) Close() error { return nil }

func TestConsumerDispatchesMessageAndIgnoresPresence(t *testing.T) {
	db := setupDB(t)
	assigned := uint(9)
	ticket, conv := seedConversation(t, db, &assigned)
	registerToken(t, "support", 9, "device-token-support-9")

	created, err := json.Marshal(envelopeFor(t, conv.ID, ticket.ID, "customer", 51, "hello"))
	if err != nil {
		t.Fatalf("marshal created: %v", err)
	}
	presence, err := json.Marshal(supportKafka.PresenceChangedData{
		ActorType: "support", ActorID: 9, State: "online", PreviousState: "offline",
	})
	if err != nil {
		t.Fatalf("marshal presence: %v", err)
	}
	presenceEnv, err := json.Marshal(supportKafka.Envelope{
		SchemaVersion: 1, EventID: "evt-2", EventType: supportKafka.EventPresenceChange,
		OccurredAt: time.Now().UTC(), ConversationID: conv.ID, TicketID: ticket.ID, Data: presence,
	})
	if err != nil {
		t.Fatalf("marshal presence envelope: %v", err)
	}

	fetcher := &fakeFetcher{msgs: []kafka.Message{
		{Key: []byte("1"), Value: created},
		{Key: []byte("1"), Value: presenceEnv},
	}}
	sender := &fakeSender{}
	d := &push.Dispatcher{Sender: sender, Presence: fakePresence{online: false, known: true}, Collapse: fakeCollapse{}}

	err = push.ConsumeFetcher(context.Background(), fetcher, d)
	if !errors.Is(err, errDrained) {
		t.Fatalf("expected drained fetcher, got %v", err)
	}
	if sender.count() != 1 {
		t.Fatalf("only message.created must dispatch, got %d", sender.count())
	}
	if fetcher.committed != 2 {
		t.Fatalf("every record must be committed, got %d", fetcher.committed)
	}
	if global.GVA_DB == nil {
		t.Fatal("db should be restored by cleanup only")
	}
}
