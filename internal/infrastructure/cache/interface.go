package cache

import (
	"context"
	"time"
)

// GenericCache interface for any key-value caching needs
type GenericCache interface {
	GetBytes(ctx context.Context, key string) ([]byte, bool, error)
	SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
