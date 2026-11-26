package kafka

import (
	"encoding/json"
	"testing"
)

func TestEnvelopeRoundTrip(t *testing.T) {
	payload, err := NewEnvelope(EventMessageCreated, 42, 7, MessageCreatedData{
		MessageUUID: "m-1", Seq: 3, SenderType: "customer", Body: "hi",
	})
	if err != nil {
		t.Fatalf("build envelope: %v", err)
	}
	env, err := ParseEnvelope(payload)
	if err != nil {
		t.Fatalf("parse envelope: %v", err)
	}
	if env.SchemaVersion != 1 || env.ConversationID != 42 || env.TicketID != 7 || env.EventType != EventMessageCreated {
		t.Fatalf("unexpected envelope: %+v", env)
	}
	var data MessageCreatedData
	if err := json.Unmarshal(env.Data, &data); err != nil || data.Seq != 3 {
		t.Fatalf("unexpected data: %v %+v", err, data)
	}
}

func TestEnvelopeRejectsFutureVersion(t *testing.T) {
	if _, err := ParseEnvelope([]byte(`{"schema_version":2,"event_type":"message.created","data":{}}`)); err == nil {
		t.Fatal("expected schema_version 2 to be rejected")
	}
}

func TestEnvelopeRejectsUnknownEvent(t *testing.T) {
	if _, err := ParseEnvelope([]byte(`{"schema_version":1,"event_type":"nope","data":{}}`)); err == nil {
		t.Fatal("expected unknown event to be rejected")
	}
}

func TestPresenceEnvelope(t *testing.T) {
	payload, err := NewEnvelope(EventPresenceChange, 42, 7, PresenceChangedData{
		ActorType: "support", ActorID: 9, State: "offline", PreviousState: "online",
	})
	if err != nil {
		t.Fatalf("build presence envelope: %v", err)
	}
	env, err := ParseEnvelope(payload)
	if err != nil || env.EventType != EventPresenceChange {
		t.Fatalf("parse presence envelope: %v %+v", err, env)
	}
}

func TestUnknownEventCannotBeBuilt(t *testing.T) {
	if _, err := NewEnvelope("bogus", 0, 0, nil); err == nil {
		t.Fatal("expected unknown event type to fail")
	}
}
