package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	app "backify/services/auth/internal"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Load .env nếu có (CWD lúc go run là services/auth/, khớp với
	// `cd $(AUTH_DIR) && go run` trong Makefile). Lỗi bị bỏ qua có chủ đích:
	// không có .env là bình thường ở production, và godotenv không ghi đè
	// biến đã có sẵn trong môi trường nên không xung đột với cách deploy hiện tại.
	_ = godotenv.Load()

	port := envOr("PORT", "8081")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	application, err := app.New(ctx, app.Config{
		AuthDatabaseURL:      envOr("AUTH_DATABASE_URL", "postgres://backify:backify@localhost:5433/auth?sslmode=disable"),
		RedisURL:             envOr("REDIS_URL", "redis://localhost:6379/0"),
		RabbitMQURL:          envOr("RABBITMQ_URL", "amqp://backify:backify@localhost:5672/"),
		ControlPlaneGRPCAddr: envOr("CONTROL_PLANE_GRPC_ADDR", "localhost:9091"),
		// Không có default cho JWT_SECRET như các biến khác — một secret mặc
		// định tiện cho local dev cũng là một secret đoán được, tức là
		// service "chạy được" ở production với chữ ký JWT ai cũng giả mạo
		// được. app.New (qua jwt.NewIssuer/NewVerifier) sẽ trả lỗi rõ ràng
		// nếu JWT_SECRET rỗng, chặn ngay ở startup thay vì âm thầm ký sai.
		JWTSecret: os.Getenv("JWT_SECRET"),
	})
	cancel()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to wire application")
	}
	defer application.Close()

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      application.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", port).Msg("auth service starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("graceful shutdown failed")
	}

	log.Info().Msg("auth service stopped")
}

// envOr đọc biến môi trường k; trả def nếu chưa set — dùng cho mọi tham số
// kết nối để service chạy được ngay ở local dev kể cả không có .env lẫn
// không set biến môi trường nào (mặc định trỏ vào docker-compose local).
func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
