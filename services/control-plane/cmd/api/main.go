package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"backify/services/control-plane/internal/adapter/postgres"
	"backify/services/control-plane/internal/handler"
	"backify/services/control-plane/internal/port"
	"backify/services/control-plane/internal/usecase"
)

type noopPublisher struct{}

func (noopPublisher) Publish(ctx context.Context, event port.Event) error {
	return nil
}

type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://backify:backify@localhost:5432/backify?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := pgxpool.New(ctx, databaseURL)
	cancel()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create postgres pool")
	}
	defer pool.Close()

	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			dbStatus = "down"
			log.Error().Err(err).Msg("db health check failed")
		}

		w.Header().Set("Content-Type", "application/json")
		if dbStatus != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(healthResponse{Status: "ok", DB: dbStatus})
	})

	projectRepo := postgres.NewProjectRepo(pool)
	entityRepo := postgres.NewEntityRepo(pool)
	fieldRepo := postgres.NewFieldRepo(pool)
	moduleRepo := postgres.NewModuleRepo(pool)
	publisher := noopPublisher{}

	h := handler.New(
		usecase.NewCreateProject(projectRepo, publisher),
		usecase.NewGetProject(projectRepo),
		usecase.NewAddEntity(projectRepo, entityRepo),
		usecase.NewAddField(entityRepo, fieldRepo),
		usecase.NewDeleteField(fieldRepo, moduleRepo, publisher),
		usecase.NewConfigModule(projectRepo, fieldRepo, moduleRepo, publisher),
	)
	h.RegisterRoutes(router)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("graceful shutdown failed")
	}

	log.Info().Msg("control-plane stopped")
}
