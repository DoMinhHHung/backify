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
	httpadapter "github.com/DoMinhHHung/backify/services/runtime/internal/adapter/http"
	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/http/handlers"
	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/http/middleware"
	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/memory"
	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/messaging"
	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/postgres"
	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/security"
	"github.com/DoMinhHHung/backify/services/runtime/internal/config"
	"github.com/DoMinhHHung/backify/services/runtime/internal/container"
	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/permission"
	"github.com/DoMinhHHung/backify/services/runtime/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	rawClient, err := grpclient.NewConfigClient(cfg.ControlPlaneGRPCAddr, cfg.InternalAPIKey, cfg.ControlPlaneGRPCTLS)
	if err != nil {
		log.Fatalf("config client: %v", err)
	}

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	pool, err := postgres.NewPool(rootCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	schemaMigrator := postgres.NewSchemaMigrator(pool)
	userRepo := postgres.NewUserRepository(pool)
	recordRepo := postgres.NewRecordRepository(pool)
	permEngine := permission.NewEngine()
	hasher := security.NewBcryptHasher()
	tokenSvc := security.NewJWTService(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	cache := memory.NewConfigCache(cfg.ConfigCacheTTL)
	configClient := grpclient.NewCachedConfigClient(rawClient, cache)

	c := container.New(
		cfg,
		configClient,
		cache,
		schemaMigrator,
		userRepo,
		hasher,
		tokenSvc,
		recordRepo,
		permEngine,
	)

	var rabbit *messaging.RabbitConsumer
	if cfg.RabbitMQEnabled {
		rabbit = messaging.NewRabbitConsumer(cfg.RabbitMQURL, cfg.RabbitMQQueue, c.ConfigEvents)
		if err := rabbit.Start(rootCtx); err != nil {
			log.Printf("rabbit consumer failed to start (continuing without events): %v", err)
			rabbit = nil
		} else {
			defer rabbit.Close()
		}
	}

	projectMW := middleware.NewProjectMiddleware(c.ConfigClient)
	authMW := middleware.NewAuthMiddleware(c.TokenService)
	authHandler := handlers.NewAuthHandler(c.Auth)
	crudHandler := handlers.NewCRUDHandler(c.CRUD)
	rateLimit := middleware.NewRateLimiter(120, time.Minute)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(middleware.StructuredLog)
	r.Use(chimw.Recoverer)

	httpadapter.MountSwagger(r)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/internal", func(r chi.Router) {
		r.Use(middleware.InternalKey(cfg.InternalAPIKey))

		r.Post("/bootstrap/{projectID}", func(w http.ResponseWriter, r *http.Request) {
			projectID := chi.URLParam(r, "projectID")
			out, err := c.BootstrapSchema.Execute(r.Context(), usecase.BootstrapSchemaInput{
				ProjectID: projectID,
			})
			if err != nil {
				log.Printf("bootstrap error: %v", err)
				handlers.WriteDomainError(w, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(out)
		})

		r.Post("/projects/{projectID}/users/{userID}/role", func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Role string `json:"role"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"invalid json"}`))
				return
			}
			err := c.PromoteUser.Execute(r.Context(), usecase.PromoteUserInput{
				ProjectID: chi.URLParam(r, "projectID"),
				UserID:    chi.URLParam(r, "userID"),
				Role:      body.Role,
			})
			if err != nil {
				log.Printf("promote error: %v", err)
				handlers.WriteDomainError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		r.Get("/config/{projectID}", func(w http.ResponseWriter, r *http.Request) {
			projectID := chi.URLParam(r, "projectID")
			projectCfg, err := c.ConfigClient.GetProjectConfig(r.Context(), projectID)
			if err != nil {
				log.Printf("get project config error: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte(`{"error":"project config unavailable"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(projectCfg)
		})
	})

	r.Route("/v1", func(r chi.Router) {
		r.Use(projectMW.Handler)
		r.Use(rateLimit.Middleware)

		r.Post("/auth/signup", authHandler.Signup)
		r.Post("/auth/signin", authHandler.Signin)
		r.Post("/auth/refresh", authHandler.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(authMW.Handler)
			r.Get("/auth/me", authHandler.Me)

			r.Post("/{entity}", crudHandler.Create)
			r.Get("/{entity}", crudHandler.List)
			r.Get("/{entity}/{id}", crudHandler.Get)
			r.Patch("/{entity}/{id}", crudHandler.Update)
			r.Delete("/{entity}/{id}", crudHandler.Delete)
		})

		r.Get("/me/config", func(w http.ResponseWriter, r *http.Request) {
			projectCfg, ok := domain.ProjectConfigFromContext(r.Context())
			if !ok {
				http.Error(w, `{"error":"project context missing"}`, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(projectCfg)
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

	rootCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
