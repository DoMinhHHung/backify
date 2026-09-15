package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	controlv1 "backify/pkg/proto/control/v1"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	port := envOr("PORT", "8081")
	authDBURL := envOr("AUTH_DATABASE_URL", "postgres://backify:backify@localhost:5433/auth?sslmode=disable")
	redisURL := envOr("REDIS_URL", "redis://localhost:6379/0")
	rabbitURL := envOr("RABBITMQ_URL", "amqp://backify:backify@localhost:5672/")
	controlGRPC := envOr("CONTROL_PLANE_GRPC_ADDR", "localhost:9091")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, authDBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect auth postgres")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("auth postgres ping failed")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid REDIS_URL")
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("redis ping failed")
	}

	rabbitConn, err := amqp091.Dial(rabbitURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect rabbitmq")
	}
	defer rabbitConn.Close()

	conn, err := grpc.NewClient(controlGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to dial control-plane gRPC")
	}
	defer conn.Close()
	controlClient := controlv1.NewControlPlaneServiceClient(conn)

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "control-plane",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(pool, rdb, rabbitConn, controlClient, cb))
	mux.HandleFunc("/auth/health", healthHandler(pool, rdb, rabbitConn, controlClient, cb))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
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

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func healthHandler(
	pool *pgxpool.Pool,
	rdb *redis.Client,
	rabbit *amqp091.Connection,
	ctrl controlv1.ControlPlaneServiceClient,
	cb *gobreaker.CircuitBreaker,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		status := map[string]any{
			"status": "ok",
			"db":     "ok",
			"db_detail": map[string]string{
				"auth":    "ok",
				"control": "ok",
			},
			"redis":    "ok",
			"rabbitmq": "ok",
		}
		code := http.StatusOK

		if err := pool.Ping(ctx); err != nil {
			status["db"] = "error"
			status["db_detail"].(map[string]string)["auth"] = "error"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		}

		_, err := cb.Execute(func() (any, error) {
			_, e := ctrl.GetProject(ctx, &controlv1.GetProjectRequest{ProjectId: "health-check"})
			return nil, e
		})
		if err != nil {
			status["db_detail"].(map[string]string)["control"] = "error"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		}

		if err := rdb.Ping(ctx).Err(); err != nil {
			status["redis"] = "error"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		}

		if rabbit.IsClosed() {
			status["rabbitmq"] = "error"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(status)
	}
}
