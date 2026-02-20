package cache

import (
	"context"
	"sync"
	"time"

	"lending-hub-service/internal/domain/refund/entity"
	refundPort "lending-hub-service/internal/domain/refund/port"
)

// MemoryRefundCache implements refundPort.RefundCache for local dev
type MemoryRefundCache struct {
	store sync.Map
	stop  chan struct{}
}

type refundCacheEntry struct {
	status    entity.RefundStatus
	expiresAt time.Time
}

var _ refundPort.RefundCache = (*MemoryRefundCache)(nil)

// NewMemoryRefundCache creates a new in-memory refund cache
func NewMemoryRefundCache() *MemoryRefundCache {
	mc := &MemoryRefundCache{stop: make(chan struct{})}
	go mc.cleanup()
	return mc
}

// Get retrieves cached refund status
func (c *MemoryRefundCache) Get(ctx context.Context, refundID string) (entity.RefundStatus, bool, error) {
	key := "refund:" + refundID
	v, ok := c.store.Load(key)
	if !ok {
		return "", false, nil
	}
	entry := v.(*refundCacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.store.Delete(key)
		return "", false, nil
	}
	return entry.status, true, nil
}

// Set stores refund status with given TTL
func (c *MemoryRefundCache) Set(ctx context.Context, refundID string, status entity.RefundStatus, ttl time.Duration) error {
	key := "refund:" + refundID
	c.store.Store(key, &refundCacheEntry{
		status:    status,
		expiresAt: time.Now().Add(ttl),
	})
	return nil
}

// Invalidate removes cached refund status
func (c *MemoryRefundCache) Invalidate(ctx context.Context, refundID string) error {
	c.store.Delete("refund:" + refundID)
	return nil
}

func (c *MemoryRefundCache) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			now := time.Now()
			c.store.Range(func(key, value interface{}) bool {
				if entry, ok := value.(*refundCacheEntry); ok && now.After(entry.expiresAt) {
					c.store.Delete(key)
				}
				return true
			})
		}
	}
}

// Close stops the cleanup goroutine
func (c *MemoryRefundCache) Close() error {
	close(c.stop)
	return nil
}
