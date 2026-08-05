package utils

import (
	"context"
	"time"

	"github.com/mo7amed4522/tayara_api/global"
)

// Red lock for guarded operations (booking transitions, payouts, etc.).
//
// It prevents two operations from entering the same protected section concurrently —
// in booking terms, it stops a second request "getting inside" the state machine while
// a transition is already in flight. When Redis is configured the lock is distributed
// (safe across instances); otherwise it degrades to the in-process KeyLocker.
//
// TryLock attempts to acquire the key once. If Redis is present and the key is already
// held it returns ok=false immediately so the caller can reject the concurrent operation.
// An expiry is set so a crashed holder never deadlocks other instances.
var lockFallbackLocker = &KeyLocker{}

// LockResult carries the release function; call release() when done. ok reports whether
// the lock was acquired. When ok is false release() is nil and must not be called.
func TryLock(ctx context.Context, key string, ttl time.Duration) (release func(), ok bool) {
	// No Redis: single-process mutex (blocks until the holder unlocks).
	if global.GVA_REDIS == nil {
		lockFallbackLocker.Lock(key)
		return func() { lockFallbackLocker.Unlock(key) }, true
	}
	lockKey := "tayara:lock:" + key
	ok, err := global.GVA_REDIS.SetNX(ctx, lockKey, 1, ttl).Result()
	if err != nil || !ok {
		return nil, false
	}
	return func() { _ = global.GVA_REDIS.Del(context.Background(), lockKey).Err() }, true
}
