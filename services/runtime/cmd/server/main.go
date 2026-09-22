package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	grpclient "github.com/DoMinhHHung/backify/services/runtime/internal/adapter/grpc"
	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/http/middleware"
	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/memory"
	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/postgres"
	"github.com/DoMinhHHung/backify/services/runtime/internal/config"
	"github.com/DoMinhHHung/backify/services/runtime/internal/container"
	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	rawClient, err := grpclient.NewConfigClient(cfg.ControlPlaneGRPCAddr, cfg.InternalAPIKey)
	if err != nil {
		log.Fatalf("config client: %v", err)
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	schemaMigrator := postgres.NewSchemaMigrator(pool)

	cache := memory.NewConfigCache(cfg.ConfigCacheTTL)
	configClient := grpclient.NewCachedConfigClient(rawClient, cache)

	c := container.New(cfg, configClient, cache, schemaMigrator)

	projectMW := middleware.NewProjectMiddleware(c.ConfigClient)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Post("/internal/bootstrap/{projectID}", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectID")
		out, err := c.BootstrapSchema.Execute(r.Context(), usecase.BootstrapSchemaInput{
			ProjectID: projectID,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	})

	r.Get("/internal/config/{projectID}", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectID")
		cfg, err := c.ConfigClient.GetProjectConfig(r.Context(), projectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
	})

	r.Route("/v1", func(r chi.Router) {
		r.Use(projectMW.Handler)

		r.Get("/me/config", func(w http.ResponseWriter, r *http.Request) {
			cfg, ok := domain.ProjectConfigFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"project context missing"}`, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cfg)
		})
	})

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("runtime listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
