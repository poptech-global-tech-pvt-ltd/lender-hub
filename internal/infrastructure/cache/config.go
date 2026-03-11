package cache

import "time"

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Address      string        `yaml:"address"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	PoolSize     int           `yaml:"poolSize"`
	MinIdleConns int           `yaml:"minIdleConns"`
	DialTimeout  time.Duration `yaml:"dialTimeout"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

// DefaultRedisConfig returns default Redis configuration
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Address:      "localhost:6379",
		DB:           0,
		PoolSize:     25,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}
