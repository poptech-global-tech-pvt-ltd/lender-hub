package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	profileResp "lending-hub-service/internal/domain/profile/dto/response"
	profilePort "lending-hub-service/internal/domain/profile/port"
)

const (
	profileCacheTTL  = 60 * time.Second
	profileKeyPrefix = "payin3:profile:"
)

// RedisProfileCache implements profilePort.ProfileCache using Redis
type RedisProfileCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisProfileCache creates a new Redis-backed profile cache
func NewRedisProfileCache(cfg RedisConfig) (*RedisProfileCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})
	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisProfileCache{client: client, ttl: profileCacheTTL}, nil
}

// Verify interface compliance
var _ profilePort.ProfileCache = (*RedisProfileCache)(nil)

// Get retrieves a cached profile response
func (c *RedisProfileCache) Get(ctx context.Context, userID, lender string) (*profileResp.CustomerStatusResponse, bool, error) {
	key := profileKeyPrefix + userID + ":" + lender
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var resp profileResp.CustomerStatusResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, false, err
	}
	return &resp, true, nil
}

// Set stores a profile response in cache
func (c *RedisProfileCache) Set(ctx context.Context, userID, lender string, value *profileResp.CustomerStatusResponse) error {
	key := profileKeyPrefix + userID + ":" + lender
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// Invalidate removes a cached profile response
func (c *RedisProfileCache) Invalidate(ctx context.Context, userID, lender string) error {
	key := profileKeyPrefix + userID + ":" + lender
	return c.client.Del(ctx, key).Err()
}

// Close gracefully shuts down the Redis client
func (c *RedisProfileCache) Close() error {
	return c.client.Close()
}
