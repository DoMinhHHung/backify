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
	ControlPlaneGRPCTLS  bool
	RabbitMQURL          string
	RabbitMQEnabled      bool
	RabbitMQQueue        string
}

// Load nạp lần lượt các file .env được hỗ trợ mà không ghi đè biến môi trường đã có,
// áp dụng giá trị mặc định và kiểm tra khóa nội bộ, URL cơ sở dữ liệu cùng JWT secret.
// Duration không hợp lệ được thay bằng giá trị mặc định tương ứng.
func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load(".env")
	_ = godotenv.Load("services/runtime/.env")

	cfg := &Config{
		HTTPAddr:             getEnv("HTTP_ADDR", ":8081"),
		ControlPlaneGRPCAddr: getEnv("CONTROL_PLANE_GRPC_ADDR", "localhost:9091"),
		InternalAPIKey:       getEnv("INTERNAL_API_KEY", ""),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTAccessTTL:         getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:        getDurationEnv("JWT_REFRESH_TTL", 7*24*time.Hour),
		ConfigCacheTTL:       getDurationEnv("CONFIG_CACHE_TTL", 30*time.Second),
		ControlPlaneGRPCTLS:  getEnv("CONTROL_PLANE_GRPC_TLS", "false") == "true",
		RabbitMQURL:          getEnv("RABBITMQ_URL", "amqp://backify:backify@localhost:5672/"),
		RabbitMQEnabled:      getEnv("RABBITMQ_ENABLED", "false") == "true",
		RabbitMQQueue:        getEnv("RABBITMQ_QUEUE", "runtime.config"),
	}
	if cfg.InternalAPIKey == "" {
		return nil, errString("INTERNAL_API_KEY is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, errString("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, errString("JWT_SECRET is required and must be at least 32 bytes")
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
