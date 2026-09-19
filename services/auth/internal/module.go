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
	authjwt "backify/services/auth/internal/jwt"
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
	JWTSecret            string
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
	JWTIssuer       *authjwt.Issuer
	JWTVerifier     *authjwt.Verifier
	conns           *postgres.ConnManager
	controlConn     *grpc.ClientConn
}
