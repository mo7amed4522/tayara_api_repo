package cache

import (
	"time"
)

// CacheEntry represents a cached item with metadata.
type CacheEntry struct {
	Key        string
	Value      interface{}
	Cost       int64
	ExpiresAt  time.Time
	AccessTime time.Time
	Version    int64
}

// CacheStats holds detailed cache statistics for monitoring.
type CacheStats struct {
	Name          string
	HitCount      int64
	MissCount     int64
	LoadCount     int64
	EvictionCount int64
	TotalLoadTime time.Duration
	Size          int
	Cost          int64 // Total memory cost of cached items
}

// GetCacheStats is a wrapper for getting detailed cache stats.
func GetCacheStats() *Stats {
	return GetStats()
}
