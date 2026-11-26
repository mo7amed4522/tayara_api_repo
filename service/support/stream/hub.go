// Package stream is the in-memory fan-out from Kafka (or the local publisher) to SSE clients.
package stream

import (
	"sync"
)

// Event is one SSE frame: an event name and its JSON payload.
type Event struct {
	Type string
	Data []byte
}

// Hub keeps a bounded subscriber set per conversation. Slow consumers are dropped rather than
// allowed to block the fan-out goroutine.
type Hub struct {
	mu   sync.RWMutex
	subs map[uint]map[*Subscription]struct{}
	done bool
}

// Subscription is one SSE client's channel plus its cancel function.
type Subscription struct {
	C  <-chan Event
	ch chan Event
	h  *Hub
	id uint
}

// NewHub builds an empty hub.
func NewHub() *Hub {
	return &Hub{subs: map[uint]map[*Subscription]struct{}{}}
}

// Default is the process-wide hub used by the HTTP handler and the Kafka relay.
var Default = NewHub()

// Subscribe registers a client for one conversation. Call cancel when the client disconnects.
func (h *Hub) Subscribe(conversationID uint) (*Subscription, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.done {
		ch := make(chan Event)
		close(ch)
		return &Subscription{C: ch, ch: ch, h: h, id: conversationID}, func() {}
	}
	ch := make(chan Event, 32)
	s := &Subscription{C: ch, ch: ch, h: h, id: conversationID}
	if h.subs[conversationID] == nil {
		h.subs[conversationID] = map[*Subscription]struct{}{}
	}
	h.subs[conversationID][s] = struct{}{}
	var once sync.Once
	cancel := func() {
		once.Do(func() { h.remove(conversationID, s) })
	}
	return s, cancel
}

func (h *Hub) remove(id uint, s *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.subs[id]; ok {
		if _, exists := set[s]; exists {
			delete(set, s)
			close(s.ch)
		}
		if len(set) == 0 {
			delete(h.subs, id)
		}
	}
}

// Broadcast delivers an event to every subscriber of a conversation, dropping a client whose
// buffer is full (it will reconnect and replay from the database).
func (h *Hub) Broadcast(conversationID uint, e Event) {
	h.mu.RLock()
	set := h.subs[conversationID]
	targets := make([]*Subscription, 0, len(set))
	for s := range set {
		targets = append(targets, s)
	}
	h.mu.RUnlock()
	for _, s := range targets {
		select {
		case s.ch <- e:
		default:
			h.remove(conversationID, s)
		}
	}
}

// Count reports the current subscriber count (tests and metrics).
func (h *Hub) Count(conversationID uint) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs[conversationID])
}

// CloseAll closes every subscription and marks the hub closed.
func (h *Hub) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.done = true
	for id, set := range h.subs {
		for s := range set {
			close(s.ch)
		}
		delete(h.subs, id)
	}
}
