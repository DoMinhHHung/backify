package grpc

import (
	"context"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	controlv1 "backify/pkg/proto/control/v1"
	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/usecase"
)

type Server struct {
	controlv1.UnimplementedControlPlaneServiceServer
	getProject *usecase.GetProject
}

func NewServer(getProject *usecase.GetProject) *Server {
	return &Server{getProject: getProject}
}

func (s *Server) GetProject(ctx context.Context, req *controlv1.GetProjectRequest) (*controlv1.GetProjectResponse, error) {
	if req.GetProjectId() == "" {
		return &controlv1.GetProjectResponse{Error: "project_id is required"}, nil
	}

	p, err := s.getProject.Execute(ctx, req.GetProjectId())
	if err != nil {
		if err == domain.ErrProjectNotFound {
			return &controlv1.GetProjectResponse{Error: "project not found"}, nil
		}
		log.Error().Err(err).Str("project_id", req.GetProjectId()).Msg("GetProject failed")
		return &controlv1.GetProjectResponse{Error: "internal error"}, nil
	}

	return &controlv1.GetProjectResponse{
		Project: toProtoProject(p),
	}, nil
}

func (s *Server) GetProjectConfig(ctx context.Context, req *controlv1.GetProjectConfigRequest) (*controlv1.GetProjectConfigResponse, error) {
	if req.GetProjectId() == "" {
		return &controlv1.GetProjectConfigResponse{Error: "project_id is required"}, nil
	}

	p, err := s.getProject.Execute(ctx, req.GetProjectId())
	if err != nil {
		if err == domain.ErrProjectNotFound {
			return &controlv1.GetProjectConfigResponse{Error: "project not found"}, nil
		}
		log.Error().Err(err).Str("project_id", req.GetProjectId()).Msg("GetProjectConfig failed")
		return &controlv1.GetProjectConfigResponse{Error: "internal error"}, nil
	}

	return &controlv1.GetProjectConfigResponse{
		Config: &controlv1.ProjectConfig{
			Project: toProtoProject(p),
		},
	}, nil
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

func ListenAndServe(addr string, srv *Server) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	gs := grpc.NewServer()
	controlv1.RegisterControlPlaneServiceServer(gs, srv)

	log.Info().Str("addr", addr).Msg("control-plane gRPC starting")
	return gs.Serve(lis)
}
