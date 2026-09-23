package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/DoMinhHHung/backify/services/runtime/internal/adapter/grpc/pb"
	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type ConfigClient struct {
	client pb.ControlPlaneServiceClient
	apiKey string
}

// NewConfigClient tạo client control plane dùng TLS hệ thống khi useTLS là true,
// hoặc kết nối không mã hóa khi useTLS là false.
func NewConfigClient(addr string, apiKey string, useTLS bool) (*ConfigClient, error) {
	var opts grpc.DialOption
	if useTLS {
		opts = grpc.WithTransportCredentials(credentials.NewTLS(nil))
	} else {
		opts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	conn, err := grpc.NewClient(addr, opts)
	if err != nil {
		return nil, fmt.Errorf("dial control-plane: %w", err)
	}
	return &ConfigClient{
		client: pb.NewControlPlaneServiceClient(conn),
		apiKey: apiKey,
	}, nil
}

// withAuth gắn khóa API nội bộ vào metadata của request gửi đi.
func (c *ConfigClient) withAuth(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-internal-key", c.apiKey)
}

// GetProject lấy metadata dự án từ control plane và truyền nguyên lỗi gRPC cho caller.
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

// GetProjectConfig lấy cấu hình runtime của dự án và giải mã entities cùng modules từ config_json.
// Hàm truyền nguyên lỗi gRPC và bọc lỗi JSON với ngữ cảnh parse config_json.
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
