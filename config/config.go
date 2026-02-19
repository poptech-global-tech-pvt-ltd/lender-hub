package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Env                string                   `mapstructure:"env"`
	HTTP               HTTPConfig               `mapstructure:"http"`
	DB                 DBConfig                 `mapstructure:"db"`
	Lazypay            LazypayConfig            `mapstructure:"lazypay"`
	Redis              RedisConfig              `mapstructure:"redis"`
	Kafka              KafkaConfig              `mapstructure:"kafka"`
	UserProfileService UserProfileServiceConfig `mapstructure:"userProfileService"`
}

// HTTPConfig defines HTTP server settings
type HTTPConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// DBConfig defines PostgreSQL connection settings
type DBConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// LazypayConfig defines Lazypay provider settings (stub for now)
type LazypayConfig struct {
	BaseURL        string        `mapstructure:"base_url"`
	AccessKey      string        `mapstructure:"access_key"`
	SecretKey      string        `mapstructure:"secret_key"`
	MerchantID     string        `mapstructure:"merchant_id"`     // Optional - use subMerchantId if not provided
	SubMerchantID  string        `mapstructure:"sub_merchant_id"` // Used in onboarding and as fallback
	ReturnURL      string        `mapstructure:"return_url"`      // Callback URL for redirects
	ProfileTimeout time.Duration `mapstructure:"profile_timeout"`
	PaymentTimeout time.Duration `mapstructure:"payment_timeout"`
	Enabled        bool          `mapstructure:"enabled"`
}

// RedisConfig defines Redis cache settings
type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// UserProfileServiceConfig defines external User Profile Service settings
type UserProfileServiceConfig struct {
	BaseURL string        `mapstructure:"baseURL"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// KafkaConfig defines Kafka event streaming settings
type KafkaConfig struct {
	Brokers      []string `mapstructure:"brokers"`
	Enabled      bool     `mapstructure:"enabled"`
	ProfileTopic string   `mapstructure:"profile_topic"`
	OrderTopic   string   `mapstructure:"order_topic"`
	RefundTopic  string   `mapstructure:"refund_topic"`
}

// Load reads configuration from file and environment variables
// Priority: env vars (LENDING_HUB_*) > YAML file > defaults
func Load(path string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read from file if provided
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("../config")
	}

	// Enable environment variable override
	v.SetEnvPrefix("LENDING_HUB")
	v.AutomaticEnv()

	// Read config file (non-fatal if missing when using defaults)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("env", "local")
	v.SetDefault("http.port", "8080")
	v.SetDefault("http.read_timeout", "30s")
	v.SetDefault("http.write_timeout", "30s")
	v.SetDefault("http.shutdown_timeout", "10s")

	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", 5432)
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("db.max_open_conns", 50)
	v.SetDefault("db.max_idle_conns", 10)
	v.SetDefault("db.conn_max_lifetime", "5m")
	v.SetDefault("db.conn_max_idle_time", "2m")

	v.SetDefault("lazypay.profile_timeout", "10s")
	v.SetDefault("lazypay.payment_timeout", "5s")
	v.SetDefault("lazypay.enabled", false)

	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")

	v.SetDefault("kafka.enabled", false)
	v.SetDefault("kafka.profile_topic", "lending-hub.profile.events")
	v.SetDefault("kafka.order_topic", "lending-hub.order.events")
	v.SetDefault("kafka.refund_topic", "lending-hub.refund.events")

	// User Profile Service defaults
	v.SetDefault("userProfileService.baseURL", "https://userprofile-sit.popclub.co.in")
	v.SetDefault("userProfileService.timeout", "5s")
}

func validate(cfg *Config) error {
	if cfg.DB.Host == "" {
		return fmt.Errorf("db.host is required")
	}
	if cfg.DB.User == "" {
		return fmt.Errorf("db.user is required")
	}
	if cfg.DB.Name == "" {
		return fmt.Errorf("db.name is required")
	}
	if cfg.HTTP.Port == "" {
		return fmt.Errorf("http.port is required")
	}
	return nil
}
