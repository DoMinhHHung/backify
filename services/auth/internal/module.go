package app

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

// Config chứa toàn bộ tham số kết nối cần thiết để khởi tạo Auth Service.
//
// ControlDatabaseURL trỏ tới Postgres của Control Plane (đọc project config).
// AuthDatabaseURL trỏ tới database admin ("postgres") của Postgres instance
// riêng cho Auth — instance này sẽ chứa các database auth_proj_<project_id>
// được Database Manager tạo động ở Bước 3, tách biệt khỏi schema-per-project
// của Control Plane.
type Config struct {
	ControlDatabaseURL string
	AuthDatabaseURL    string
	RedisURL           string
	RabbitMQURL        string
}

// App giữ toàn bộ tài nguyên đã kết nối và router HTTP của Auth Service.
type App struct {
	Router      chi.Router
	ControlPool *pgxpool.Pool
	AuthPool    *pgxpool.Pool
	RedisClient *redis.Client
	RabbitConn  *amqp.Connection
}

// New kết nối Control DB, Auth DB, Redis và RabbitMQ rồi dựng router HTTP.
// Mọi tài nguyên đã mở được đóng trước khi trả về nếu một bước sau đó thất bại,
// để tránh rò rỉ connection khi wiring lỗi giữa chừng.
func New(ctx context.Context, cfg Config) (*App, error) {
	controlPool, err := pgxpool.New(ctx, cfg.ControlDatabaseURL)
	if err != nil {
		return nil, err
	}

	authPool, err := pgxpool.New(ctx, cfg.AuthDatabaseURL)
	if err != nil {
		controlPool.Close()
		return nil, err
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		authPool.Close()
		controlPool.Close()
		return nil, err
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient.Close()
		authPool.Close()
		controlPool.Close()
		return nil, err
	}

	rabbitConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		redisClient.Close()
		authPool.Close()
		controlPool.Close()
		return nil, err
	}

	router := chi.NewRouter()
	router.Get("/health", healthHandler(controlPool, authPool, redisClient, rabbitConn))

	return &App{
		Router:      router,
		ControlPool: controlPool,
		AuthPool:    authPool,
		RedisClient: redisClient,
		RabbitConn:  rabbitConn,
	}, nil
}

// Close đóng RabbitMQ, Redis, Auth DB pool và Control DB pool, theo thứ tự
// ngược lại lúc mở. Lỗi đóng RabbitMQ và Redis bị bỏ qua vì không còn hành
// động khắc phục nào khi service đang shutdown.
func (a *App) Close() {
	_ = a.RabbitConn.Close()
	_ = a.RedisClient.Close()
	a.AuthPool.Close()
	a.ControlPool.Close()
}

type healthResponse struct {
	Status   string `json:"status"`
	DB       string `json:"db"`
	Redis    string `json:"redis"`
	RabbitMQ string `json:"rabbitmq"`
}

// healthHandler báo "ok" khi Control DB, Auth DB, Redis và RabbitMQ đều phản
// hồi trong ba giây; ngược lại báo "degraded" kèm HTTP 503. Trường "db" gộp
// chung cả hai pool Postgres vì client bên ngoài chỉ cần biết tầng lưu trữ
// quan hệ có sẵn sàng hay không, không cần phân biệt Control DB với Auth DB.
func healthHandler(controlPool, authPool *pgxpool.Pool, redisClient *redis.Client, rabbitConn *amqp.Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := controlPool.Ping(ctx); err != nil {
			dbStatus = "down"
		}
		if err := authPool.Ping(ctx); err != nil {
			dbStatus = "down"
		}

		redisStatus := "ok"
		if err := redisClient.Ping(ctx).Err(); err != nil {
			redisStatus = "down"
		}

		rabbitStatus := "ok"
		if rabbitConn.IsClosed() {
			rabbitStatus = "down"
		}

		status := "ok"
		if dbStatus != "ok" || redisStatus != "ok" || rabbitStatus != "ok" {
			status = "degraded"
		}

		w.Header().Set("Content-Type", "application/json")
		if status != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(healthResponse{
			Status:   status,
			DB:       dbStatus,
			Redis:    redisStatus,
			RabbitMQ: rabbitStatus,
		})
	}
}
