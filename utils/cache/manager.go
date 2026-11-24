package cache

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	mrand "math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const (
	lockTTL      = 30 * time.Second
	lockWait     = 3 * time.Second
	pollInterval = 50 * time.Millisecond
)

// CacheManager manages the local Ristretto cache and shared Redis cache.
type CacheManager struct {
	l1Cache     *ristretto.Cache
	sfGroup     *singleflight.Group
	redisClient redis.UniversalClient
}

// DefaultManager is the global cache manager instance.
var DefaultManager = &CacheManager{sfGroup: &singleflight.Group{}}

var redisLockReleaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

// Stats holds cache statistics for monitoring.
type Stats struct {
	HitCount      int64
	MissCount     int64
	EvictionCount int64
	Size          int
}

// Init initializes the local L1 cache with Ristretto.
func Init(maxCost int64, metricsEnabled bool) error {
	config := &ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     maxCost,
		BufferItems: 64,
		Metrics:     metricsEnabled,
	}
	l1, err := ristretto.NewCache(config)
	if err != nil {
		return fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	DefaultManager.l1Cache = l1
	DefaultManager.sfGroup = &singleflight.Group{}
	go startInvalidationPublisher(DefaultManager)
	return nil
}

// SetRedisClient sets the Redis client for L2 cache operations.
func SetRedisClient(client redis.UniversalClient) {
	if DefaultManager != nil {
		DefaultManager.redisClient = client
	}
}

// AddJitter adds random jitter (±10%) to TTL to prevent synchronized expiry.
func AddJitter(baseTTL time.Duration) time.Duration {
	if baseTTL <= 0 || baseTTL/10 <= 0 {
		return baseTTL
	}
	jitter := time.Duration(mrand.Int63n(int64(baseTTL / 10)))
	if mrand.Intn(2) == 0 {
		return baseTTL + jitter
	}
	return baseTTL - jitter
}

// Get retrieves a value from the local cache.
func Get(key string) (interface{}, bool, error) {
	if DefaultManager == nil || DefaultManager.l1Cache == nil {
		return nil, false, nil
	}
	value, found := DefaultManager.l1Cache.Get(key)
	if !found {
		return nil, false, nil
	}
	return value, true, nil
}

// GetWithLoader retrieves a value with a custom loader function.
func GetWithLoader[T any](key string, loader func() (T, int64, error), ttl time.Duration) (T, bool, error) {
	if loader == nil {
		var zero T
		return zero, false, errors.New("loader is nil")
	}
	return GetOrLoadJSON(context.Background(), key, ttl, func() (T, error) {
		value, _, err := loader()
		return value, err
	})
}

// GetOrLoadJSON retrieves a JSON-serializable value from L1 or Redis, loading it
// once per key on a cache miss. Values are JSON-round-tripped before being
// stored locally so L1 and L2 contain the same independently-owned value.
func GetOrLoadJSON[T any](ctx context.Context, key string, ttl time.Duration, loader func() (T, error)) (T, bool, error) {
	var zero T
	if loader == nil {
		return zero, false, errors.New("loader is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if DefaultManager == nil {
		return loaderValue(ctx, loader)
	}

	started := time.Now()
	if value, ok := getL1[T](key); ok {
		log.Printf("cache key=%q source=l1 duration=%s", key, time.Since(started))
		return value, true, nil
	}

	result, err, _ := DefaultManager.sfGroup.Do(key, func() (interface{}, error) {
		if value, ok := getL1[T](key); ok {
			return value, nil
		}
		if value, ok, remainingTTL := getL2[T](ctx, key); ok {
			log.Printf("cache key=%q source=l2 duration=%s", key, time.Since(started))
			storeL1WithTTL(key, value, remainingTTL)
			return value, nil
		}

		client := DefaultManager.getRedis()
		if client != nil {
			release, acquired := acquireLock(ctx, key, lockTTL)
			if acquired {
				defer release()
				if value, ok, remainingTTL := getL2[T](ctx, key); ok {
					log.Printf("cache key=%q source=l2-after-lock duration=%s", key, time.Since(started))
					storeL1WithTTL(key, value, remainingTTL)
					return value, nil
				}
				return loadAndStore(ctx, key, ttl, loader, started)
			}

			if value, ok, remainingTTL := waitForRedisValue[T](ctx, key); ok {
				log.Printf("cache key=%q source=l2-after-wait duration=%s", key, time.Since(started))
				storeL1WithTTL(key, value, remainingTTL)
				return value, nil
			}
		}

		return loadAndStore(ctx, key, ttl, loader, started)
	})
	if err != nil {
		return zero, false, err
	}
	value, ok := result.(T)
	if !ok {
		return zero, false, fmt.Errorf("cache value type mismatch for key %q: got %T", key, result)
	}
	return value, true, nil
}

func loaderValue[T any](ctx context.Context, loader func() (T, error)) (T, bool, error) {
	value, err := loader()
	if err != nil {
		var zero T
		return zero, false, err
	}
	return value, true, nil
}

func loadAndStore[T any](ctx context.Context, key string, ttl time.Duration, loader func() (T, error), started time.Time) (interface{}, error) {
	value, err := loader()
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode cache value for key %q: %w", key, err)
	}
	var owned T
	if err := json.Unmarshal(encoded, &owned); err != nil {
		return nil, fmt.Errorf("decode cache value for key %q: %w", key, err)
	}
	if !isNilValue(owned) {
		cacheTTL := AddJitter(ttl)
		storeL1WithTTL(key, owned, cacheTTL)
		if client := DefaultManager.getRedis(); client != nil && cacheTTL > 0 {
			if err := client.Set(ctx, key, encoded, cacheTTL).Err(); err != nil {
				log.Printf("cache key=%q redis_write_error=%v", key, err)
			}
		}
	}
	log.Printf("cache key=%q source=loaded duration=%s", key, time.Since(started))
	return owned, nil
}

func getL1[T any](key string) (T, bool) {
	var zero T
	if DefaultManager == nil || DefaultManager.l1Cache == nil {
		return zero, false
	}
	value, found := DefaultManager.l1Cache.Get(key)
	if !found {
		return zero, false
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache key=%q l1_encode_error=%v", key, err)
		DefaultManager.l1Cache.Del(key)
		return zero, false
	}
	var owned T
	if err := json.Unmarshal(encoded, &owned); err != nil {
		log.Printf("cache key=%q l1_type_mismatch=%v", key, err)
		DefaultManager.l1Cache.Del(key)
		return zero, false
	}
	return owned, true
}

func getL2[T any](ctx context.Context, key string) (T, bool, time.Duration) {
	var zero T
	client := DefaultManager.getRedis()
	if client == nil {
		return zero, false, 0
	}
	payload, err := client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return zero, false, 0
	}
	if err != nil {
		log.Printf("cache key=%q redis_read_error=%v", key, err)
		return zero, false, 0
	}
	var value T
	if err := json.Unmarshal(payload, &value); err != nil {
		log.Printf("cache key=%q redis_decode_error=%v", key, err)
		if delErr := client.Del(ctx, key).Err(); delErr != nil {
			log.Printf("cache key=%q redis_delete_error=%v", key, delErr)
		}
		return zero, false, 0
	}
	if isNilValue(value) {
		return zero, false, 0
	}
	remainingTTL, err := client.PTTL(ctx, key).Result()
	if err != nil {
		log.Printf("cache key=%q redis_ttl_error=%v", key, err)
		return value, true, 0
	}
	return value, true, remainingTTL
}

func waitForRedisValue[T any](ctx context.Context, key string) (T, bool, time.Duration) {
	var zero T
	deadline := time.NewTimer(lockWait)
	ticker := time.NewTicker(pollInterval)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		if value, ok, remainingTTL := getL2[T](ctx, key); ok {
			return value, true, remainingTTL
		}
		select {
		case <-ctx.Done():
			return zero, false, 0
		case <-deadline.C:
			return zero, false, 0
		case <-ticker.C:
		}
	}
}

func storeL1WithTTL[T any](key string, value T, ttl time.Duration) {
	if DefaultManager == nil || DefaultManager.l1Cache == nil || ttl <= 0 {
		return
	}
	cost := (&DefaultCostEstimator{}).EstimateCost(value)
	if cost <= 0 {
		cost = 1
	}
	DefaultManager.l1Cache.SetWithTTL(key, value, cost, ttl)
}

func isNilValue[T any](value T) bool {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return true
	}
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// acquireLock attempts to acquire a cross-instance Redis lock for a cache key.
func acquireLock(ctx context.Context, key string, ttl time.Duration) (release func(), ok bool) {
	client := DefaultManager.getRedis()
	if client == nil {
		return func() {}, false
	}
	tokenBytes := make([]byte, 16)
	if _, err := crand.Read(tokenBytes); err != nil {
		log.Printf("cache key=%q lock_token_error=%v", key, err)
		return func() {}, false
	}
	token := hex.EncodeToString(tokenBytes)
	lockKey := "lock:" + key
	acquired, err := client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		log.Printf("cache key=%q lock_error=%v", key, err)
		return func() {}, false
	}
	if !acquired {
		return func() {}, false
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			if err := redisLockReleaseScript.Run(context.Background(), client, []string{lockKey}, token).Err(); err != nil {
				log.Printf("cache key=%q lock_release_error=%v", key, err)
			}
		})
	}, true
}

// Set stores a value in L1 and, when ttl is positive, Redis.
func Set(key string, value interface{}, cost int64, ttl time.Duration) {
	if DefaultManager == nil {
		return
	}
	storedValue := value
	var encoded []byte
	if ttl > 0 {
		var err error
		encoded, err = json.Marshal(value)
		if err != nil {
			log.Printf("cache key=%q redis_encode_error=%v", key, err)
			return
		}
		if value != nil {
			valueType := reflect.TypeOf(value)
			owned := reflect.New(valueType)
			if err := json.Unmarshal(encoded, owned.Interface()); err != nil {
				log.Printf("cache key=%q decode_for_l1_error=%v", key, err)
				return
			}
			storedValue = owned.Elem().Interface()
		}
	}
	if DefaultManager.l1Cache != nil {
		if cost <= 0 {
			cost = (&DefaultCostEstimator{}).EstimateCost(storedValue)
		}
		if cost <= 0 {
			cost = 1
		}
		DefaultManager.l1Cache.SetWithTTL(key, storedValue, cost, ttl)
	}
	if ttl > 0 {
		if client := DefaultManager.getRedis(); client != nil {
			if err := client.Set(context.Background(), key, encoded, ttl).Err(); err != nil {
				log.Printf("cache key=%q redis_write_error=%v", key, err)
			}
		}
	}
}

// Delete removes a key from the local cache.
func Delete(key string) {
	if DefaultManager != nil && DefaultManager.l1Cache != nil {
		DefaultManager.l1Cache.Del(key)
	}
}

// DeleteRedis removes a key from the shared Redis cache.
func DeleteRedis(ctx context.Context, key string) error {
	if DefaultManager == nil || DefaultManager.getRedis() == nil {
		return errors.New("redis cache client not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	client := DefaultManager.getRedis()
	if err := client.Del(ctx, key).Err(); err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		Keys []string `json:"keys"`
	}{Keys: []string{key}})
	if err != nil {
		return fmt.Errorf("encode cache invalidation for key %q: %w", key, err)
	}
	if err := client.Publish(ctx, "cache:invalidate", payload).Err(); err != nil {
		return fmt.Errorf("publish cache invalidation for key %q: %w", key, err)
	}
	return nil
}

// GetStats returns current local cache statistics.
func GetStats() *Stats {
	if DefaultManager == nil || DefaultManager.l1Cache == nil {
		return &Stats{}
	}
	metrics := DefaultManager.l1Cache.Metrics
	if metrics == nil {
		metrics = &ristretto.Metrics{}
	}
	return &Stats{
		HitCount:      int64(metrics.Hits()),
		MissCount:     int64(metrics.Misses()),
		EvictionCount: int64(metrics.KeysEvicted()),
	}
}

func startInvalidationPublisher(m *CacheManager) {
	client := m.getRedis()
	if client == nil {
		return
	}
	sub := client.Subscribe(context.Background(), "cache:invalidate")
	if sub == nil {
		return
	}
	defer sub.Close()
	for msg := range sub.Channel() {
		var event struct {
			Keys []string `json:"keys"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			log.Printf("cache invalidation message decode error: %v", err)
			continue
		}
		for _, key := range event.Keys {
			Delete(key)
		}
	}
}

func (m *CacheManager) getRedis() redis.UniversalClient {
	if m == nil {
		return nil
	}
	return m.redisClient
}
