package cache

import (
	"context"
	"sync"
	"time"

	profileResp "lending-hub-service/internal/domain/profile/dto/response"
	profilePort "lending-hub-service/internal/domain/profile/port"
)

// MemoryProfileCache implements profilePort.ProfileCache for local dev
// Uses sync.Map with TTL-based expiry via background goroutine
type MemoryProfileCache struct {
	store sync.Map
	ttl   time.Duration
	stop  chan struct{}
}

type cacheEntry struct {
	value     *profileResp.CustomerStatusResponse
	expiresAt time.Time
}

// Verify interface compliance
var _ profilePort.ProfileCache = (*MemoryProfileCache)(nil)

// NewMemoryProfileCache creates a new in-memory profile cache
func NewMemoryProfileCache(ttl time.Duration) *MemoryProfileCache {
	mc := &MemoryProfileCache{
		ttl:  ttl,
		stop: make(chan struct{}),
	}
	// Start background cleanup goroutine every 30s
	go mc.cleanup()
	return mc
}

// Get retrieves a cached profile response
func (c *MemoryProfileCache) Get(ctx context.Context, userID, lender string) (*profileResp.CustomerStatusResponse, bool, error) {
	key := userID + ":" + lender
	v, ok := c.store.Load(key)
	if !ok {
		return nil, false, nil
	}
	entry := v.(*cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.store.Delete(key)
		return nil, false, nil
	}
	return entry.value, true, nil
}

// Set stores a profile response in cache
func (c *MemoryProfileCache) Set(ctx context.Context, userID, lender string, value *profileResp.CustomerStatusResponse) error {
	key := userID + ":" + lender
	c.store.Store(key, &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	})
	return nil
}

// Invalidate removes a cached profile response
func (c *MemoryProfileCache) Invalidate(ctx context.Context, userID, lender string) error {
	key := userID + ":" + lender
	c.store.Delete(key)
	return nil
}

// cleanup periodically removes expired entries
func (c *MemoryProfileCache) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			now := time.Now()
			c.store.Range(func(key, value interface{}) bool {
				if entry, ok := value.(*cacheEntry); ok && now.After(entry.expiresAt) {
					c.store.Delete(key)
				}
				return true
			})
		}
	}
}

// Close stops the cleanup goroutine
func (c *MemoryProfileCache) Close() error {
	close(c.stop)
	return nil
}
