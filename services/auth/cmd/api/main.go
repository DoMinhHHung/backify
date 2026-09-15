package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	app "backify/services/auth/internal"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	controlDatabaseURL := os.Getenv("CONTROL_DATABASE_URL")
	if controlDatabaseURL == "" {
		controlDatabaseURL = "postgres://backify:backify@localhost:5432/backify?sslmode=disable"
	}

	// AuthDatabaseURL trỏ vào database admin "postgres" của Postgres instance
	// riêng cho Auth — Database Manager (Bước 3) dùng connection này để chạy
	// CREATE DATABASE / DROP DATABASE cho từng project.
	authDatabaseURL := os.Getenv("AUTH_DATABASE_URL")
	if authDatabaseURL == "" {
		authDatabaseURL = "postgres://backify:backify@localhost:5433/postgres?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://backify:backify@localhost:5672/"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	application, err := app.New(ctx, app.Config{
		ControlDatabaseURL: controlDatabaseURL,
		AuthDatabaseURL:    authDatabaseURL,
		RedisURL:           redisURL,
		RabbitMQURL:        rabbitmqURL,
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
