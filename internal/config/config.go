package config

import (
	"fmt"
	"os"
	"time"
)

const (
	DefaultJWTTTL          = 15 * time.Minute
	DefaultRateLimitLogin  = 5
	DefaultRateLimitReset  = 3
	DefaultRateLimitWindow = 5 * time.Minute

	DBMaxOpenConns    = 25
	DBMaxIdleConns    = 5
	DBConnMaxLifetime = 5 * time.Minute
)

type Config struct {
	GRPCPort    string
	DatabaseURL string
	RedisURL    string
	JWTTTL      time.Duration

	RateLimitLogin  int
	RateLimitReset  int
	RateLimitWindow time.Duration
}

func Load() *Config {
	return &Config{
		GRPCPort:        envOrDefault("GRPC_PORT", "50051"),
		DatabaseURL:     envOrDefault("DATABASE_URL", "postgres://auth:auth@localhost:5432/auth?sslmode=disable"),
		RedisURL:        envOrDefault("REDIS_URL", "redis://localhost:6379/0"),
		JWTTTL:          DefaultJWTTTL,
		RateLimitLogin:  DefaultRateLimitLogin,
		RateLimitReset:  DefaultRateLimitReset,
		RateLimitWindow: DefaultRateLimitWindow,
	}
}

func (c *Config) GRPCAddr() string {
	return fmt.Sprintf(":%s", c.GRPCPort)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
