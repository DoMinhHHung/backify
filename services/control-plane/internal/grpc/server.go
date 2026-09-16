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

	controlv1 "backify/pkg/proto/control/v1"
	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/usecase"
)

// Server implements controlv1.ControlPlaneServiceServer. Failures use
// standard gRPC status codes (NotFound / InvalidArgument / Internal)
// instead of an in-band error field, so callers can use
// google.golang.org/grpc/status and their circuit breaker trips on real
// failures, not on successful responses that merely describe an error.
type Server struct {
	controlv1.UnimplementedControlPlaneServiceServer
	getProject       *usecase.GetProject
	getProjectConfig *usecase.GetProjectConfig
}

// NewServer trả về Server đã gắn hai usecase cần cho ControlPlaneService.
func NewServer(getProject *usecase.GetProject, getProjectConfig *usecase.GetProjectConfig) *Server {
	return &Server{getProject: getProject, getProjectConfig: getProjectConfig}
}

// GetProject trả gRPC InvalidArgument nếu thiếu project_id, NotFound nếu
// project không tồn tại.
func (s *Server) GetProject(ctx context.Context, req *controlv1.GetProjectRequest) (*controlv1.GetProjectResponse, error) {
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	p, err := s.getProject.Execute(ctx, req.GetProjectId())
	if err != nil {
		return nil, toStatusError(err, req.GetProjectId(), "GetProject")
	}

	return &controlv1.GetProjectResponse{Project: toProtoProject(p)}, nil
}

// GetProjectConfig trả gRPC InvalidArgument nếu thiếu project_id, NotFound
// nếu project không tồn tại; ngược lại trả project + entities + fields +
// modules + functions (kèm field ID đang bật) trong một lượt gọi.
func (s *Server) GetProjectConfig(ctx context.Context, req *controlv1.GetProjectConfigRequest) (*controlv1.GetProjectConfigResponse, error) {
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	cfg, err := s.getProjectConfig.Execute(ctx, req.GetProjectId())
	if err != nil {
		return nil, toStatusError(err, req.GetProjectId(), "GetProjectConfig")
	}

	return &controlv1.GetProjectConfigResponse{Config: toProtoProjectConfig(cfg)}, nil
}

// toStatusError map lỗi domain sang gRPC status: ErrProjectNotFound thành
// NotFound, còn lại thành Internal kèm log — không lộ chi tiết lỗi hạ tầng
// qua gRPC message.
func toStatusError(err error, projectID, rpc string) error {
	if err == domain.ErrProjectNotFound {
		return status.Error(codes.NotFound, "project not found")
	}
	log.Error().Err(err).Str("project_id", projectID).Str("rpc", rpc).Msg("grpc call failed")
	return status.Error(codes.Internal, "internal error")
}

func toProtoProject(p *domain.Project) *controlv1.Project {
	if p == nil {
		return nil
	}
	return &controlv1.Project{
		Id:        p.ID,
		Name:      p.Name,
		Subdomain: p.Subdomain,
		Plan:      string(p.Plan),
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt.Unix(),
		UpdatedAt: p.UpdatedAt.Unix(),
	}
}

func toProtoProjectConfig(cfg *usecase.ProjectConfig) *controlv1.ProjectConfig {
	entities := make([]*controlv1.Entity, 0, len(cfg.Entities))
	for _, ec := range cfg.Entities {
		fields := make([]*controlv1.Field, 0, len(ec.Fields))
		for _, f := range ec.Fields {
			fields = append(fields, &controlv1.Field{
				Id:     f.ID,
				Name:   f.Name,
				Type:   string(f.Type),
				System: f.IsSystem,
			})
		}
		entities = append(entities, &controlv1.Entity{
			Id:       ec.Entity.ID,
			Name:     ec.Entity.Name,
			Fields:   fields,
			IsSystem: ec.Entity.IsSystem,
		})
	}

	modules := make([]*controlv1.Module, 0, len(cfg.Modules))
	for _, mc := range cfg.Modules {
		functions := make([]*controlv1.Function, 0, len(mc.Functions))
		for _, fn := range mc.Functions {
			functions = append(functions, &controlv1.Function{
				Id:              fn.ID,
				Name:            fn.Name,
				EnabledFieldIds: fn.EnabledFieldIDs,
			})
		}
		modules = append(modules, &controlv1.Module{
			Id:        mc.Module.ID,
			Name:      string(mc.Module.Name),
			Functions: functions,
		})
	}

	return &controlv1.ProjectConfig{
		Project:  toProtoProject(cfg.Project),
		Entities: entities,
		Modules:  modules,
	}
}

// NewGRPCServer dựng grpc.Server đã đăng ký ControlPlaneService và chuẩn
// health checking protocol (grpc.health.v1.Health) — Auth Service dùng
// health service này để kiểm tra kết nối mà không phải gọi vào một RPC
// nghiệp vụ thật (tránh tốn round-trip DB chỉ để đo liveness).
func NewGRPCServer(srv *Server) *grpc.Server {
	gs := grpc.NewServer()
	controlv1.RegisterControlPlaneServiceServer(gs, srv)

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(gs, healthSrv)

	return gs
}

// Listen mở listener TCP tại addr; tách khỏi việc Serve để caller giữ được
// cả *grpc.Server (để GracefulStop khi shutdown) lẫn listener.
func Listen(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}
