package stub

import (
	"context"
	"sync"

	res "lending-hub-service/internal/domain/profile/dto/response"
	"lending-hub-service/internal/domain/profile/port"
)

// StubProfileCache implements port.ProfileCache using in-memory sync.Map
type StubProfileCache struct {
	mu    sync.RWMutex
	cache map[string]*res.CustomerStatusResponse
}

// NewStubProfileCache creates a new stub cache
func NewStubProfileCache() port.ProfileCache {
	return &StubProfileCache{
		cache: make(map[string]*res.CustomerStatusResponse),
	}
}

// Get retrieves a cached value
func (c *StubProfileCache) Get(ctx context.Context, userID, lender string) (*res.CustomerStatusResponse, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := userID + ":" + lender
	value, found := c.cache[key]
	return value, found, nil
}

// Set stores a value in cache
func (c *StubProfileCache) Set(ctx context.Context, userID, lender string, value *res.CustomerStatusResponse) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := userID + ":" + lender
	c.cache[key] = value
	return nil
}

// Invalidate removes a value from cache
func (c *StubProfileCache) Invalidate(ctx context.Context, userID, lender string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := userID + ":" + lender
	delete(c.cache, key)
	return nil
}
