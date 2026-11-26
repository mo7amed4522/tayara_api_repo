package support

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	supportResp "github.com/mo7amed4522/tayara_api/model/support/response"
	supportKafka "github.com/mo7amed4522/tayara_api/service/support/kafka"
	"github.com/mo7amed4522/tayara_api/service/support/stream"
	"github.com/mo7amed4522/tayara_api/utils"
	"github.com/redis/go-redis/v9"
)

// PresenceMember identifies one live connection of one actor. Multi-tab collapses into a single
// presence state through the per-actor connection count.
type PresenceMember struct {
	ActorType string
	ActorID   uint
	ConnID    string
}

func (m PresenceMember) actorKey() string { return m.ActorType + ":" + uintString(m.ActorID) }
func (m PresenceMember) indexMember() string {
	return m.actorKey() + ":" + m.ConnID
}

func connKey(actorType string, actorID uint) string {
	return constants.SupportPresenceConnKeyPrefix + actorType + ":" + uintString(actorID)
}

// PresenceStore is the Redis seam. A nil store (Redis unavailable) means every read is
// "unknown" and every write is a silent no-op — chat is never blocked by presence.
type PresenceStore interface {
	Touch(ctx context.Context, m PresenceMember, ttl time.Duration) (int64, error)
	Drop(ctx context.Context, m PresenceMember) (int64, error)
	Sweep(ctx context.Context, now time.Time, limit int) ([]PresenceMember, error)
	OnlineActors(ctx context.Context) ([]PresenceMember, error)
}

// Store returns the active presence store, or nil when Redis is not configured.
func Store() PresenceStore {
	storeOverrideMu.RLock()
	override := storeOverride
	storeOverrideMu.RUnlock()
	if override != nil {
		return override
	}
	if global.GVA_REDIS == nil {
		return nil
	}
	return redisPresenceStore{client: global.GVA_REDIS}
}

type redisPresenceStore struct {
	client redis.UniversalClient
}

func (s redisPresenceStore) Touch(ctx context.Context, m PresenceMember, ttl time.Duration) (int64, error) {
	now := time.Now().UTC()
	expiry := float64(now.Add(ttl).Unix())
	ck := connKey(m.ActorType, m.ActorID)
	pipe := s.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, ck, "-inf", strconv.FormatInt(now.Unix(), 10))
	pipe.ZAdd(ctx, ck, redis.Z{Score: expiry, Member: m.ConnID})
	pipe.ZAdd(ctx, constants.SupportPresenceIndexKey, redis.Z{Score: expiry, Member: m.indexMember()})
	pipe.Expire(ctx, ck, ttl*3)
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	_ = cmds
	return s.client.ZCard(ctx, ck).Result()
}

func (s redisPresenceStore) Drop(ctx context.Context, m PresenceMember) (int64, error) {
	ck := connKey(m.ActorType, m.ActorID)
	pipe := s.client.Pipeline()
	pipe.ZRem(ctx, ck, m.ConnID)
	pipe.ZRem(ctx, constants.SupportPresenceIndexKey, m.indexMember())
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return s.client.ZCard(ctx, ck).Result()
}

func (s redisPresenceStore) Sweep(ctx context.Context, now time.Time, limit int) ([]PresenceMember, error) {
	if limit <= 0 {
		limit = 200
	}
	members, err := s.client.ZRangeByScore(ctx, constants.SupportPresenceIndexKey, &redis.ZRangeBy{
		Min: "-inf", Max: strconv.FormatInt(now.Unix(), 10), Offset: 0, Count: int64(limit),
	}).Result()
	if err != nil {
		return nil, err
	}
	var expired []PresenceMember
	for _, raw := range members {
		m, ok := parseIndexMember(raw)
		if !ok {
			if removed, _ := s.client.ZRem(ctx, constants.SupportPresenceIndexKey, raw).Result(); removed == 0 {
				continue
			}
			continue
		}
		// ZREM returning 1 is the exactly-once gate: two replicas cannot both report the drop.
		removed, err := s.client.ZRem(ctx, constants.SupportPresenceIndexKey, raw).Result()
		if err != nil || removed == 0 {
			continue
		}
		ck := connKey(m.ActorType, m.ActorID)
		_, _ = s.client.ZRem(ctx, ck, m.ConnID).Result()
		count, err := s.client.ZCard(ctx, ck).Result()
		if err != nil || count > 0 {
			continue
		}
		expired = append(expired, m)
	}
	return expired, nil
}

func (s redisPresenceStore) OnlineActors(ctx context.Context) ([]PresenceMember, error) {
	now := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	members, err := s.client.ZRangeByScore(ctx, constants.SupportPresenceIndexKey, &redis.ZRangeBy{
		Min: now, Max: "+inf",
	}).Result()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := make([]PresenceMember, 0, len(members))
	for _, raw := range members {
		m, ok := parseIndexMember(raw)
		if !ok || seen[m.actorKey()] {
			continue
		}
		seen[m.actorKey()] = true
		out = append(out, m)
	}
	return out, nil
}

func parseIndexMember(raw string) (PresenceMember, bool) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) != 3 {
		return PresenceMember{}, false
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return PresenceMember{}, false
	}
	return PresenceMember{ActorType: parts[0], ActorID: uint(id), ConnID: parts[2]}, true
}

// PresenceServices turns connection lifecycle events into at most one state transition each.
type PresenceServices struct{}

// Enabled reports whether presence is configured on.
func (s PresenceServices) Enabled() bool { return global.GVA_CONFIG.Support.Presence.Enabled }

func (s PresenceServices) ttl() time.Duration {
	secs := global.GVA_CONFIG.Support.Presence.TTLSeconds
	if secs <= 0 {
		secs = 45
	}
	return time.Duration(secs) * time.Second
}

// Connect registers a connection and publishes online only on the 0→1 transition.
func (s PresenceServices) Connect(ctx context.Context, actor Actor, conversationID, ticketID uint, connID string) {
	if !s.Enabled() {
		return
	}
	store := Store()
	if store == nil {
		return
	}
	m := PresenceMember{ActorType: senderTypeForActor(actor), ActorID: actor.ID, ConnID: connID}
	count, err := store.Touch(ctx, m, s.ttl())
	if err != nil {
		return
	}
	if count == 1 {
		s.transition(ctx, m, constants.SupportPresenceOnline, conversationID, ticketID)
	}
}

// Heartbeat refreshes a connection's expiry without touching Kafka.
func (s PresenceServices) Heartbeat(ctx context.Context, actorType string, actorID uint, connID string) {
	if !s.Enabled() {
		return
	}
	store := Store()
	if store == nil {
		return
	}
	_, _ = store.Touch(ctx, PresenceMember{ActorType: actorType, ActorID: actorID, ConnID: connID}, s.ttl())
}

// Disconnect removes a connection and publishes offline (plus last seen) only on 1→0.
func (s PresenceServices) Disconnect(ctx context.Context, actorType string, actorID uint, connID string, conversationID, ticketID uint) {
	if !s.Enabled() {
		return
	}
	store := Store()
	if store == nil {
		return
	}
	m := PresenceMember{ActorType: actorType, ActorID: actorID, ConnID: connID}
	count, err := store.Drop(ctx, m)
	if err != nil {
		return
	}
	if count == 0 {
		_ = (LastSeenServices{}).Record(actorType, actorID, time.Now().UTC())
		s.transition(ctx, m, constants.SupportPresenceOffline, conversationID, ticketID)
	}
}

// transition publishes exactly one presence.changed event per state change.
func (s PresenceServices) transition(ctx context.Context, m PresenceMember, state string, conversationID, ticketID uint) {
	release, ok := utils.TryLock(ctx, constants.SupportPresenceLockPrefix+m.actorKey()+":"+state, 5*time.Second)
	if !ok {
		return
	}
	defer release()

	visible := (LastSeenServices{}).Visible(m.ActorType)
	var lastSeen *time.Time
	if state == constants.SupportPresenceOffline && visible {
		if t := (LastSeenServices{}).Latest(m.ActorType, m.ActorID); t != nil {
			lastSeen = t
		}
	}
	previous := constants.SupportPresenceOffline
	if state == constants.SupportPresenceOffline {
		previous = constants.SupportPresenceOnline
	}
	data := supportKafka.PresenceChangedData{
		ActorType:       m.ActorType,
		ActorID:         m.ActorID,
		State:           state,
		PreviousState:   previous,
		LastSeenAt:      lastSeen,
		LastSeenVisible: visible,
	}
	payload, err := supportKafka.NewEnvelope(supportKafka.EventPresenceChange, conversationID, ticketID, data)
	if err != nil {
		return
	}
	if supportKafka.Active() {
		_ = supportKafka.PublishPresence(ctx, m.actorKey(), payload)
	}
	if conversationID > 0 {
		stream.Default.Broadcast(conversationID, stream.Event{Type: supportKafka.EventPresenceChange, Data: payload})
	}
}

// SweepExpired repairs state left behind by a killed process. The store's ZREM is the gate that
// guarantees a single offline transition across replicas.
func (s PresenceServices) SweepExpired() error {
	if !s.Enabled() {
		return nil
	}
	store := Store()
	if store == nil {
		return nil
	}
	expired, err := store.Sweep(context.Background(), time.Now().UTC(), 200)
	if err != nil {
		return err
	}
	for _, m := range expired {
		_ = (LastSeenServices{}).Record(m.ActorType, m.ActorID, time.Now().UTC())
		s.transition(context.Background(), m, constants.SupportPresenceOffline, 0, 0)
	}
	return nil
}

// BatchRead answers presence for the actors the caller may see. Actors the caller does not share
// a readable ticket with are omitted entirely (no state, no last-seen leak).
func (s PresenceServices) BatchRead(ctx context.Context, caller Actor, keys []string) ([]supportResp.PresenceView, error) {
	allowed, err := (AuthorizationServices{}).AllowedPresenceActors(caller)
	if err != nil {
		return nil, err
	}
	store := Store()
	lastSeen := LastSeenServices{}
	out := make([]supportResp.PresenceView, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || !allowed[key] {
			continue
		}
		actorType, actorID, ok := parseActorKey(key)
		if !ok {
			continue
		}
		view := supportResp.PresenceView{ActorType: actorType, ActorID: actorID, State: constants.SupportPresenceUnknown}
		if store != nil {
			online := false
			if members, e := store.OnlineActors(ctx); e == nil {
				for _, m := range members {
					if m.ActorType == actorType && m.ActorID == actorID {
						online = true
						break
					}
				}
			}
			state := constants.SupportPresenceOffline
			if online {
				state = constants.SupportPresenceOnline
			}
			view.State = state
			if lastSeen.Visible(actorType) {
				view.LastSeenAt = lastSeen.Latest(actorType, actorID)
			}
		}
		out = append(out, view)
	}
	return out, nil
}

// IsOnline reports whether the actor has at least one live connection. known is false when the
// presence system cannot answer (disabled or Redis down) so callers treat it as "unknown"
// instead of "offline". This is the only presence query the push dispatcher uses.
func (s PresenceServices) IsOnline(actorType string, actorID uint) (online bool, known bool) {
	if !s.Enabled() || actorID == 0 || actorType == "" {
		return false, false
	}
	store := Store()
	if store == nil {
		return false, false
	}
	members, err := store.OnlineActors(context.Background())
	if err != nil {
		return false, false
	}
	for _, m := range members {
		if m.ActorType == actorType && m.ActorID == actorID {
			return true, true
		}
	}
	return false, true
}

// ClaimPushWindow is the best-effort collapse gate for push notifications: the first push for a
// (recipient, conversation) pair inside the window wins. ok=false means Redis is unavailable, in
// which case the caller must NOT collapse (better two pushes than a silent chat).
func (s PresenceServices) ClaimPushWindow(recipient string, conversationID uint, window time.Duration) (collapsed bool, ok bool) {
	if global.GVA_REDIS == nil || window <= 0 || recipient == "" {
		return false, false
	}
	key := constants.SupportPushCollapseKeyPrefix + recipient + ":" + uintString(conversationID)
	set, err := global.GVA_REDIS.SetNX(context.Background(), key,
		strconv.FormatInt(time.Now().UTC().UnixNano(), 10), window).Result()
	if err != nil {
		return false, false
	}
	return !set, true
}

func parseActorKey(key string) (string, uint, bool) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		return "", 0, false
	}
	switch parts[0] {
	case constants.ActorCustomer, constants.ActorSupport, constants.ActorQuality, constants.ActorAdmin:
	default:
		return "", 0, false
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 {
		return "", 0, false
	}
	return parts[0], uint(id), true
}

// memoryPresenceStore is the in-memory store used when Redis is absent in tests. It is exported
// through UseMemoryStore so tests in other packages can install it.
type memoryPresenceStore struct {
	mu     sync.Mutex
	conns  map[string]map[string]time.Time
	actors map[string]PresenceMember
}

// NewMemoryPresenceStore builds an in-memory store (tests / single-node Redis-less runs).
func NewMemoryPresenceStore() PresenceStore {
	return &memoryPresenceStore{conns: map[string]map[string]time.Time{}, actors: map[string]PresenceMember{}}
}

// UseMemoryPresenceStore swaps the store returned by Store() for tests.
var (
	storeOverrideMu sync.RWMutex
	storeOverride   PresenceStore
)

// SetPresenceStore installs an override store (nil restores the Redis-backed default).
func SetPresenceStore(s PresenceStore) {
	storeOverrideMu.Lock()
	defer storeOverrideMu.Unlock()
	storeOverride = s
}

func (m *memoryPresenceStore) Touch(_ context.Context, member PresenceMember, ttl time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := member.actorKey()
	if m.conns[key] == nil {
		m.conns[key] = map[string]time.Time{}
	}
	m.conns[key][member.ConnID] = time.Now().UTC().Add(ttl)
	m.actors[member.indexMember()] = member
	m.pruneLocked(key)
	return int64(len(m.conns[key])), nil
}

func (m *memoryPresenceStore) Drop(_ context.Context, member PresenceMember) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := member.actorKey()
	delete(m.conns[key], member.ConnID)
	delete(m.actors, member.indexMember())
	return int64(len(m.conns[key])), nil
}

func (m *memoryPresenceStore) Sweep(_ context.Context, now time.Time, limit int) ([]PresenceMember, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []PresenceMember
	for key, conns := range m.conns {
		for connID, expiry := range conns {
			if expiry.After(now) {
				continue
			}
			delete(conns, connID)
			member := PresenceMember{ConnID: connID}
			if parts := strings.SplitN(key, ":", 2); len(parts) == 2 {
				member.ActorType = parts[0]
				if id, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					member.ActorID = uint(id)
				}
			}
			delete(m.actors, member.indexMember())
			if len(conns) == 0 {
				out = append(out, member)
			}
			if limit > 0 && len(out) >= limit {
				return out, nil
			}
		}
	}
	return out, nil
}

func (m *memoryPresenceStore) OnlineActors(context.Context) ([]PresenceMember, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	seen := map[string]bool{}
	out := make([]PresenceMember, 0)
	for key, conns := range m.conns {
		if len(conns) == 0 || seen[key] {
			continue
		}
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			continue
		}
		id, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		seen[key] = true
		out = append(out, PresenceMember{ActorType: parts[0], ActorID: uint(id)})
	}
	return out, nil
}

func (m *memoryPresenceStore) pruneLocked(key string) {
	now := time.Now().UTC()
	for connID, expiry := range m.conns[key] {
		if expiry.Before(now) {
			delete(m.conns[key], connID)
		}
	}
}

var errNoPresenceStore = errors.New("presence store unavailable")
