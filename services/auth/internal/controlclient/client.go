// Package controlclient bọc gRPC client tới Control Plane cho Auth Service.
// Auth không tự kết nối Postgres của Control Plane (đã bỏ ở Bước 1 theo
// quyết định A2) — mọi truy vấn config đi qua ControlPlaneService, và gói
// này chịu trách nhiệm cho việc gọi đó chịu được Control Plane down/chậm.
package controlclient

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	controlv1 "backify/pkg/proto/control/v1"
)

// retryableCodes là các mã lỗi transient đáng thử lại (Control Plane chưa kịp
// phục vụ / mạng chập chờn). NotFound và InvalidArgument KHÔNG nằm trong danh
// sách — đó là câu trả lời hợp lệ của nghiệp vụ, thử lại không đổi kết quả và
// chỉ làm chậm request của end-user.
var retryableCodes = map[codes.Code]bool{
	codes.Unavailable:       true,
	codes.DeadlineExceeded:  true,
	codes.ResourceExhausted: true,
	codes.Aborted:           true,
}

// Client gọi ControlPlaneService qua circuit breaker + retry. Đây là điểm
// tích hợp duy nhất Auth phụ thuộc vào Control Plane còn sống; mọi usecase
// khác chỉ nên phụ thuộc vào interface này (chưa tách interface ở Bước 1 vì
// chưa có usecase nào gọi tới — sẽ tách khi Bước 6 cần mock trong test).
type Client struct {
	grpcClient controlv1.ControlPlaneServiceClient
	health     grpc_health_v1.HealthClient
	breaker    *gobreaker.CircuitBreaker
	maxRetries int
	baseDelay  time.Duration
}

// New tạo Client từ một ControlPlaneServiceClient/HealthClient đã Dial sẵn
// (Dial/Close do caller quản lý vòng đời, giống cách App quản lý
// pgxpool/redis) và cấu hình circuit breaker: mở mạch sau 5 lỗi liên tiếp,
// thử lại (half-open) sau 15 giây.
func New(grpcClient controlv1.ControlPlaneServiceClient, health grpc_health_v1.HealthClient) *Client {
	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "control-plane",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})

	return &Client{
		grpcClient: grpcClient,
		health:     health,
		breaker:    breaker,
		maxRetries: 3,
		baseDelay:  100 * time.Millisecond,
	}
}

// GetProject gọi ControlPlaneService.GetProject qua breaker + retry.
func (c *Client) GetProject(ctx context.Context, projectID string) (*controlv1.Project, error) {
	resp, err := callWithRetry(c, ctx, func(ctx context.Context) (*controlv1.GetProjectResponse, error) {
		return c.grpcClient.GetProject(ctx, &controlv1.GetProjectRequest{ProjectId: projectID})
	})
	if err != nil {
		return nil, err
	}
	return resp.GetProject(), nil
}

// GetProjectConfig gọi ControlPlaneService.GetProjectConfig qua breaker + retry.
func (c *Client) GetProjectConfig(ctx context.Context, projectID string) (*controlv1.ProjectConfig, error) {
	resp, err := callWithRetry(c, ctx, func(ctx context.Context) (*controlv1.GetProjectConfigResponse, error) {
		return c.grpcClient.GetProjectConfig(ctx, &controlv1.GetProjectConfigRequest{ProjectId: projectID})
	})
	if err != nil {
		return nil, err
	}
	return resp.GetConfig(), nil
}

// callWithRetry là phần dùng chung cho mọi RPC: circuit breaker bọc ngoài,
// retry với backoff tuyến tính bọc trong. Thứ tự này để breaker đếm một "yêu
// cầu logic" là một lần thất bại (sau khi đã thử lại hết), không đếm từng
// lần retry riêng lẻ thành nhiều lỗi liên tiếp — tránh mở mạch sớm chỉ vì
// một request chậm thoáng qua.
func callWithRetry[T any](c *Client, ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	result, err := c.breaker.Execute(func() (interface{}, error) {
		var lastErr error
		var zero T
		for attempt := 0; attempt <= c.maxRetries; attempt++ {
			if attempt > 0 {
				select {
				case <-ctx.Done():
					return zero, ctx.Err()
				case <-time.After(c.baseDelay * time.Duration(attempt)):
				}
			}

			resp, err := fn(ctx)
			if err == nil {
				return resp, nil
			}
			lastErr = err

			st, ok := status.FromError(err)
			if !ok || !retryableCodes[st.Code()] {
				return zero, err
			}
		}
		return zero, lastErr
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return result.(T), nil
}

// Healthy gọi chuẩn gRPC health checking protocol thay vì một RPC nghiệp vụ
// thật — không tốn round-trip DB của Control Plane chỉ để đo liveness, và
// không phụ thuộc project nào có tồn tại hay không.
func (c *Client) Healthy(ctx context.Context) error {
	resp, err := c.health.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return err
	}
	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		return status.Errorf(codes.Unavailable, "control-plane health status: %s", resp.GetStatus())
	}
	return nil
}
