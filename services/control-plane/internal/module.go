package app

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"

	"backify/services/control-plane/internal/adapter/postgres"
	"backify/services/control-plane/internal/adapter/rabbitmq"
	"backify/services/control-plane/internal/handler"
	"backify/services/control-plane/internal/usecase"
)

type Config struct {
	DatabaseURL string
	RabbitMQURL string
}

type App struct {
	Router     chi.Router
	Pool       *pgxpool.Pool
	RabbitConn *amqp.Connection
	publisher  *rabbitmq.Publisher
	GetProject *usecase.GetProject
}

func New(ctx context.Context, cfg Config) (*App, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	rabbitConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		pool.Close()
		return nil, err
	}

	publisher, err := rabbitmq.NewPublisher(rabbitConn)
	if err != nil {
		rabbitConn.Close()
		pool.Close()
		return nil, err
	}

	projectRepo := postgres.NewProjectRepo(pool)
	entityRepo := postgres.NewEntityRepo(pool)
	fieldRepo := postgres.NewFieldRepo(pool)
	moduleRepo := postgres.NewModuleRepo(pool)

	getProject := usecase.NewGetProject(projectRepo)

	h := handler.New(
		usecase.NewCreateProject(projectRepo, publisher),
		getProject,
		usecase.NewAddEntity(projectRepo, entityRepo),
		usecase.NewAddField(entityRepo, fieldRepo),
		usecase.NewDeleteField(fieldRepo, moduleRepo, publisher),
		usecase.NewConfigModule(projectRepo, entityRepo, fieldRepo, moduleRepo, publisher),
	)

	router := chi.NewRouter()
	router.Get("/health", healthHandler(pool))
	h.RegisterRoutes(router)

	return &App{
		Router:     router,
		Pool:       pool,
		RabbitConn: rabbitConn,
		publisher:  publisher,
		GetProject: getProject,
	}, nil
}

func (a *App) Close() {
	_ = a.publisher.Close()
	_ = a.RabbitConn.Close()
	a.Pool.Close()
}

type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			dbStatus = "down"
		}

		status := "ok"
		w.Header().Set("Content-Type", "application/json")
		if dbStatus != "ok" {
			status = "degraded"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(healthResponse{Status: status, DB: dbStatus})
	}
}
