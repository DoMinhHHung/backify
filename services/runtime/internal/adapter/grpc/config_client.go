package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/DoMinhHHung/backify/services/runtime/internal/adapter/grpc/pb"
	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type ConfigClient struct {
	client pb.ControlPlaneServiceClient
	apiKey string
}

func NewConfigClient(addr string, apiKey string) (*ConfigClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial control-plane: %w", err)
	}
	return &ConfigClient{
		client: pb.NewControlPlaneServiceClient(conn),
		apiKey: apiKey,
	}, nil
}

func (c *ConfigClient) withAuth(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-internal-key", c.apiKey)
}

func (c *ConfigClient) GetProject(ctx context.Context, projectID string) (*domain.ProjectMeta, error) {
	ctx = c.withAuth(ctx)
	resp, err := c.client.GetProject(ctx, &pb.GetProjectRequest{ProjectId: projectID})
	if err != nil {
		return nil, err
	}
	return &domain.ProjectMeta{
		ID:         resp.Id,
		Name:       resp.Name,
		Slug:       resp.Slug,
		SchemaName: resp.SchemaName,
		OwnerID:    resp.OwnerId,
		Version:    resp.Version,
	}, nil
}

func (c *ConfigClient) GetProjectConfig(ctx context.Context, projectID string) (*domain.ProjectConfig, error) {
	ctx = c.withAuth(ctx)
	resp, err := c.client.GetProjectConfig(ctx, &pb.GetProjectConfigRequest{ProjectId: projectID})
	if err != nil {
		return nil, err
	}

	var raw struct {
		Entities map[string]domain.Entity       `json:"entities"`
		Modules  map[string]domain.ModuleConfig `json:"modules"`
	}
	if err := json.Unmarshal([]byte(resp.ConfigJson), &raw); err != nil {
		return nil, fmt.Errorf("parse config_json: %w", err)
	}

	return &domain.ProjectConfig{
		ProjectID:  resp.ProjectId,
		Slug:       resp.Slug,
		SchemaName: resp.SchemaName,
		Version:    resp.Version,
		Entities:   raw.Entities,
		Modules:    raw.Modules,
	}, nil
}
