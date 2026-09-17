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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	controlv1 "backify/pkg/proto/control/v1"
	"backify/services/auth/internal/adapter/postgres"
	"backify/services/auth/internal/adapter/rabbitmq"
	"backify/services/auth/internal/controlclient"
)

// Config chứa toàn bộ tham số kết nối cần thiết để khởi tạo Auth Service.
//
// Không có ControlDatabaseURL — Auth không kết nối trực tiếp Postgres của
// Control Plane (quyết định A2): mọi config đi qua gRPC ControlPlaneService
// để giữ boundary giữa hai service, tránh một service đọc thẳng DB nội bộ
// của service khác.
type Config struct {
	AuthDatabaseURL      string
	RedisURL             string
	RabbitMQURL          string
	ControlPlaneGRPCAddr string
}

// App giữ toàn bộ tài nguyên đã kết nối và router HTTP của Auth Service.
type App struct {
	Router          chi.Router
	AuthPool        *pgxpool.Pool
	RedisClient     *redis.Client
	RabbitConn      *amqp.Connection
	Subscriber      *rabbitmq.Subscriber
	ControlPlane    *controlclient.Client
	DatabaseManager *postgres.DatabaseManager
	Users           *postgres.UserRepo
	RefreshTokens   *postgres.RefreshTokenRepo
	PasswordResets  *postgres.PasswordResetRepo
	conns           *postgres.ConnManager
	controlConn     *grpc.ClientConn
}

// New kết nối Auth DB, Redis, RabbitMQ và gRPC Control Plane, khởi động
// subscriber sự kiện project.* rồi dựng router HTTP. Mọi tài nguyên đã mở
// được đóng trước khi trả về nếu một bước sau đó thất bại.
func New(ctx context.Context, cfg Config) (*App, error) {
	authPool, err := pgxpool.New(ctx, cfg.AuthDatabaseURL)
	if err != nil {
		return nil, err
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		authPool.Close()
		return nil, err
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient.Close()
		authPool.Close()
		return nil, err
	}

	rabbitConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		redisClient.Close()
		authPool.Close()
		return nil, err
	}

	// dbManager dùng authPool (kết nối tới database "auth") làm connection
	// admin để CREATE DATABASE/DROP DATABASE — không cần pool riêng, admin
	// operations không tốn connection lâu.
	dbManager := postgres.NewDatabaseManager(authPool)

	// conns định tuyến tới đúng database auth_proj_<projectID> cho từng
	// lời gọi repository (Bước 4) — cũng dùng authPool làm mẫu cấu hình
	// (host/user/password), không phải để chạy query cho project nào.
	conns := postgres.NewConnManager(authPool)
	users := postgres.NewUserRepo(conns)
	refreshTokens := postgres.NewRefreshTokenRepo(conns)
	passwordResets := postgres.NewPasswordResetRepo(conns)

	subscriber, err := rabbitmq.NewSubscriber(rabbitConn, dbManager)
	if err != nil {
		rabbitConn.Close()
		redisClient.Close()
		authPool.Close()
		return nil, err
	}
	if err := subscriber.Start(ctx); err != nil {
		subscriber.Close()
		rabbitConn.Close()
		redisClient.Close()
		authPool.Close()
		return nil, err
	}

	// grpc.NewClient không tự Dial ngay (lazy) nên không cần context timeout
	// riêng ở đây; kết nối thật diễn ra ở lần gọi RPC đầu tiên, lỗi kết nối
	// (nếu có) sẽ lộ ra qua Healthy() hoặc qua chính lần gọi đó, không phải ở New.
	controlConn, err := grpc.NewClient(cfg.ControlPlaneGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		subscriber.Close()
		rabbitConn.Close()
		redisClient.Close()
		authPool.Close()
		return nil, err
	}
	controlClient := controlclient.New(
		controlv1.NewControlPlaneServiceClient(controlConn),
		grpc_health_v1.NewHealthClient(controlConn),
	)

	router := chi.NewRouter()
	router.Get("/health", healthHandler(authPool, redisClient, rabbitConn, controlClient))
	router.Get("/auth/health", healthHandler(authPool, redisClient, rabbitConn, controlClient))

	return &App{
		Router:          router,
		AuthPool:        authPool,
		RedisClient:     redisClient,
		RabbitConn:      rabbitConn,
		Subscriber:      subscriber,
		ControlPlane:    controlClient,
		DatabaseManager: dbManager,
		Users:           users,
		RefreshTokens:   refreshTokens,
		PasswordResets:  passwordResets,
		conns:           conns,
		controlConn:     controlConn,
	}, nil
}

// Close đóng gRPC connection, subscriber, RabbitMQ, Redis, mọi pool per-project
// (conns) rồi Auth DB pool, theo thứ tự ngược lại lúc mở. Lỗi đóng RabbitMQ/
// Redis/subscriber/gRPC bị bỏ qua vì không còn hành động khắc phục nào khi
// service đang shutdown.
func (a *App) Close() {
	_ = a.controlConn.Close()
	_ = a.Subscriber.Close()
	_ = a.RabbitConn.Close()
	_ = a.RedisClient.Close()
	a.conns.Close()
	a.AuthPool.Close()
}

type healthResponse struct {
	Status       string `json:"status"`
	DB           string `json:"db"`
	Redis        string `json:"redis"`
	RabbitMQ     string `json:"rabbitmq"`
	ControlPlane string `json:"control_plane"`
}

// healthHandler báo "ok" khi Auth DB, Redis, RabbitMQ và Control Plane (qua
// gRPC health protocol) đều phản hồi trong ba giây; ngược lại báo "degraded"
// kèm HTTP 503. ControlPlane tách riêng khỏi "db" (không còn gộp như bản cũ
// dùng Control DB trực tiếp) vì đây là một dependency mạng, không phải một
// pool Postgres — gộp chung dễ khiến người đọc log tưởng nhầm là lỗi DB.
func healthHandler(authPool *pgxpool.Pool, redisClient *redis.Client, rabbitConn *amqp.Connection, controlPlane *controlclient.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		dbStatus := "ok"
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

		controlStatus := "ok"
		if err := controlPlane.Healthy(ctx); err != nil {
			controlStatus = "down"
		}

		status := "ok"
		if dbStatus != "ok" || redisStatus != "ok" || rabbitStatus != "ok" || controlStatus != "ok" {
			status = "degraded"
		}

		w.Header().Set("Content-Type", "application/json")
		if status != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(healthResponse{
			Status:       status,
			DB:           dbStatus,
			Redis:        redisStatus,
			RabbitMQ:     rabbitStatus,
			ControlPlane: controlStatus,
		})
	}
}
