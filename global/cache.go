package global

import (
	"github.com/mo7amed4522/tayara_api/utils/cache"
	"github.com/redis/go-redis/v9"
)

// InitCache initializes the two-tier cache system (L1: Ristretto, L2: Redis).
// maxCostMB: maximum cache memory in megabytes (default 100MB if 0)
// metricsEnabled: whether to enable cache metrics collection
func InitCache(maxCostMB int, metricsEnabled bool) error {
	if maxCostMB <= 0 {
		maxCostMB = 100 // default 100MB
	}
	maxCost := int64(maxCostMB) << 20 // convert MB to bytes
	return cache.Init(maxCost, metricsEnabled)
}

// SetCacheRedisClient sets the Redis client for the cache manager.
// Should be called after Redis is initialized.
func SetCacheRedisClient(client redis.UniversalClient) {
	cache.SetRedisClient(client)
}
