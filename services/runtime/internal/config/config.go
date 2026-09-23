package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr             string
	ControlPlaneGRPCAddr string
	InternalAPIKey       string
	DatabaseURL          string
	JWTSecret            string
	JWTAccessTTL         time.Duration
	JWTRefreshTTL        time.Duration
	ConfigCacheTTL       time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load(".env")
	_ = godotenv.Load("services/runtime/.env")

	cfg := &Config{
		HTTPAddr:             getEnv("HTTP_ADDR", ":8081"),
		ControlPlaneGRPCAddr: getEnv("CONTROL_PLANE_GRPC_ADDR", "localhost:9091"),
		InternalAPIKey:       getEnv("INTERNAL_API_KEY", ""),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		JWTSecret:            getEnv("JWT_SECRET", "change-me-runtime"),
		JWTAccessTTL:         getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:        getDurationEnv("JWT_REFRESH_TTL", 7*24*time.Hour),
		ConfigCacheTTL:       getDurationEnv("CONFIG_CACHE_TTL", 30*time.Second),
	}
	if cfg.InternalAPIKey == "" {
		return nil, errString("INTERNAL_API_KEY is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, errString("DATABASE_URL is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

type errString string

func (e errString) Error() string { return string(e) }
