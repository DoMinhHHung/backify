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

	app "backify/services/control-plane/internal"
	cpgrpc "backify/services/control-plane/internal/grpc"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Load .env nếu có (CWD lúc go run là services/control-plane/, khớp với
	// `cd $(CONTROL_PLANE_DIR) && go run` trong Makefile). Lỗi bị bỏ qua có
	// chủ đích: không có .env là bình thường ở production (biến môi trường
	// đến từ hệ thống triển khai), godotenv không được ghi đè biến đã có sẵn
	// trong môi trường nên không xung đột với cách deploy hiện tại.
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://backify:backify@localhost:5432/backify?sslmode=disable"
	}

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://backify:backify@localhost:5672/"
	}

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9091"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	application, err := app.New(ctx, app.Config{DatabaseURL: databaseURL, RabbitMQURL: rabbitmqURL})
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
		log.Info().Str("port", port).Msg("control-plane starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	grpcListener, err := cpgrpc.Listen(grpcAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open gRPC listener")
	}
	grpcServer := cpgrpc.NewGRPCServer(cpgrpc.NewServer(application.GetProject, application.GetProjectConfig))
	go func() {
		log.Info().Str("addr", grpcAddr).Msg("control-plane gRPC starting")
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal().Err(err).Msg("gRPC server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// GracefulStop trước Shutdown HTTP: gRPC không nhận context timeout nên
	// nếu treo (client giữ stream mở) sẽ chặn tiến trình thoát; chạy trong
	// goroutine kèm timeout riêng để không phụ thuộc vào shutdownCtx của HTTP.
	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()
	select {
	case <-grpcStopped:
	case <-time.After(10 * time.Second):
		log.Warn().Msg("gRPC graceful stop timed out, forcing stop")
		grpcServer.Stop()
	}

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("graceful shutdown failed")
	}

	log.Info().Msg("control-plane stopped")
}
