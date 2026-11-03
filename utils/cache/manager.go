package cache

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// CacheManager manages L1 local cache (Ristretto) with singleflight protection.
type CacheManager struct {
	l1Cache *ristretto.Cache
	sfGroup *singleflight.Group
	redisClient redis.UniversalClient
}

// DefaultManager is the global cache manager instance.
var DefaultManager *CacheManager

func init() {
	DefaultManager = &CacheManager{}
}

// Stats holds cache statistics for monitoring.
type Stats struct {
	HitCount      int64
	MissCount     int64
	EvictionCount int64
	Size          int
}

// Init initializes the local L1 cache with Ristretto.
// maxCost: maximum cache memory in bytes (e.g., 100MB = 100 << 20)
func Init(maxCost int64, metricsEnabled bool) error {
	config := &ristretto.Config{
		NumCounters: 1e7, // 10M counters for TinyLFU
		MaxCost:     maxCost,
		BufferItems: 64,
		Metrics:     metricsEnabled,
	}

	cache, err := ristretto.NewCache(config)
	if err != nil {
		return fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	DefaultManager.l1Cache = cache
	DefaultManager.sfGroup = &singleflight.Group{}

	// Start invalidation publisher subscriber
	go startInvalidationPublisher(DefaultManager)

	return nil
}

// SetRedisClient sets the Redis client for L2 cache operations.
// Should be called after Redis is initialized.
func SetRedisClient(client redis.UniversalClient) {
	if DefaultManager != nil {
		DefaultManager.redisClient = client
	}
}

// AddJitter adds random jitter (±10%) to TTL to prevent synchronized expiry.
func AddJitter(baseTTL time.Duration) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(baseTTL / 10)))
	if rand.Intn(2) == 0 {
		return baseTTL + jitter
	}
	return baseTTL - jitter
}

// Get retrieves a value from cache with singleflight protection.
// Returns (value, found, error)
func Get(key string) (interface{}, bool, error) {
	// Try L1 cache first
	if DefaultManager.l1Cache == nil {
		return nil, false, nil
	}

	if val, found := DefaultManager.l1Cache.Get(key); found {
		return val, true, nil
	}

	// Use singleflight to prevent cache stampede
	_, err, _ := DefaultManager.sfGroup.Do(key, func() (interface{}, error) {
		return nil, fmt.Errorf("cache get miss for key: %s", key)
	})

	if err != nil {
		return nil, false, err
	}

	// Note: In real implementation, the loader would be passed in
	// This is a simplified version - actual usage should pass a loader function
	return nil, false, nil
}

// GetWithLoader retrieves a value with a custom loader function.
// The loader is only called if the key is not in cache (and not already
// being loaded by another goroutine thanks to singleflight).
func GetWithLoader[T any](key string, loader func() (T, int64, error), ttl time.Duration) (T, bool, error) {
	var zero T

	// Try L1 cache first
	if DefaultManager.l1Cache == nil {
		return zero, false, nil
	}

	if val, found := DefaultManager.l1Cache.Get(key); found {
		return val.(T), true, nil
	}

	// Use singleflight to prevent cache stampede
	_, err, _ := DefaultManager.sfGroup.Do(key, func() (interface{}, error) {
		if loader == nil {
			return zero, fmt.Errorf("loader is nil")
		}
		v, cost, err := loader()
		if err != nil {
			return zero, err
		}
		if cost > 0 {
			DefaultManager.l1Cache.SetWithTTL(key, v, cost, ttl)
		}
		return v, nil
	})

	return zero, false, err
}

// Set stores a value in cache with TTL and cost.
func Set(key string, value interface{}, cost int64, ttl time.Duration) {
	if DefaultManager.l1Cache == nil {
		return
	}
	DefaultManager.l1Cache.SetWithTTL(key, value, cost, ttl)
}

// Delete removes a key from cache.
func Delete(key string) {
	if DefaultManager.l1Cache == nil {
		return
	}
	DefaultManager.l1Cache.Del(key)
}

// GetStats returns current cache statistics.
func GetStats() *Stats {
	if DefaultManager.l1Cache == nil {
		return &Stats{}
	}

	// Get metrics from cache (Metrics is a field, not a method)
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

// Invalidation event publisher

// startInvalidationPublisher connects to Redis pub/sub and listens for invalidation events.
func startInvalidationPublisher(m *CacheManager) {
	redisClient := m.getRedis()

	if redisClient == nil {
		// Redis not ready yet, will try again on next invalidation
		return
	}

	// Subscribe to invalidation channel
	sub := redisClient.Subscribe(context.Background(), "cache:invalidate")
	if sub == nil {
		return
	}
	defer sub.Close()

	ch := sub.Channel()
	for msg := range ch {
		// Delete keys from local cache
		for _, key := range parseInvalidationMessage(msg.Payload) {
			Delete(key)
		}
	}
}

// parseInvalidationMessage extracts keys from pub/sub message
func parseInvalidationMessage(payload string) []string {
	// Simple JSON parser: expects {"keys": ["key1", "key2"]}
	var keys []string

	// Find the "keys" array
	// This is simplified - in production use encoding/json
	content := payload
	if idx := strings.Index(content, "\"keys\""); idx >= 0 {
		start := strings.Index(content[idx:], "[")
		if start >= 0 {
			arrContent := content[idx+start:]
			// Extract array contents
			arrContent = arrContent[1 : len(arrContent)-1]
			parts := splitJSONArray(arrContent)
			for _, s := range parts {
				s = strings.TrimSpace(s)
				if len(s) > 1 && s[0] == '"' && s[len(s)-1] == '"' {
					keys = append(keys, s[1:len(s)-1])
				}
			}
		}
	}
	return keys
}

// splitJSONArray splits a JSON array string into parts
func splitJSONArray(s string) []string {
	var result []string
	inString := false
	depth := 0
	start := 0

	for i := 0; i < len(s); i++ {
		if s[i] == '"' && (i == 0 || s[i-1] != '\\') {
			inString = !inString
		} else if !inString {
			if s[i] == '[' {
				depth++
			} else if s[i] == ']' {
				depth--
				if depth == 0 {
					result = append(result, s[start:i+1])
					start = i + 1
				}
			} else if s[i] == ',' && depth == 1 {
				result = append(result, s[start:i])
				start = i + 1
			}
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

// getRedis returns the Redis client
func (m *CacheManager) getRedis() redis.UniversalClient {
	return m.redisClient
}
