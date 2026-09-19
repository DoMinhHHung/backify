package grpc

import (
	"context"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	authv1 "backify/pkg/proto/auth/v1"
	"backify/services/auth/internal/usecase"
)

// Server implements authv1.AuthServiceServer.
type Server struct {
	authv1.UnimplementedAuthServiceServer
	verifyToken *usecase.VerifyToken
}

// NewServer gắn usecase VerifyToken.
func NewServer(verifyToken *usecase.VerifyToken) *Server {
	return &Server{verifyToken: verifyToken}
}

// VerifyToken — token không hợp lệ là kết quả nghiệp vụ (valid=false),
// không phải lỗi gRPC. Chỉ lỗi hạ tầng (Redis down) mới trả status Internal.
func (s *Server) VerifyToken(ctx context.Context, req *authv1.VerifyTokenRequest) (*authv1.VerifyTokenResponse, error) {
	if req.GetToken() == "" || req.GetExpectedProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "token and expected_project_id are required")
	}

	out, err := s.verifyToken.Execute(ctx, usecase.VerifyTokenInput{
		Token:             req.GetToken(),
		ExpectedProjectID: req.GetExpectedProjectId(),
	})
	if err != nil {
		log.Error().Err(err).Msg("VerifyToken infrastructure error")
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.VerifyTokenResponse{
		Valid:     out.Valid,
		UserId:    out.UserID,
		ProjectId: out.ProjectID,
		Email:     out.Email,
		Error:     out.Error,
	}, nil
}

// NewGRPCServer đăng ký AuthService + health check.
func NewGRPCServer(srv *Server) *grpc.Server {
	gs := grpc.NewServer()
	authv1.RegisterAuthServiceServer(gs, srv)

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(gs, healthSrv)

	return gs
}

// Listen mở TCP listener tại addr.
func Listen(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}
