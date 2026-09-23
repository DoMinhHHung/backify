package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/DoMinhHHung/backify/services/runtime/internal/adapter/grpc/pb"
)

type fakeControlPlaneClient struct {
	getProjectFn       func(context.Context, *pb.GetProjectRequest) (*pb.GetProjectResponse, error)
	getProjectConfigFn func(context.Context, *pb.GetProjectConfigRequest) (*pb.GetProjectConfigResponse, error)
}

func (f *fakeControlPlaneClient) GetProject(ctx context.Context, in *pb.GetProjectRequest, _ ...grpcpkg.CallOption) (*pb.GetProjectResponse, error) {
	return f.getProjectFn(ctx, in)
}

func (f *fakeControlPlaneClient) GetProjectConfig(ctx context.Context, in *pb.GetProjectConfigRequest, _ ...grpcpkg.CallOption) (*pb.GetProjectConfigResponse, error) {
	return f.getProjectConfigFn(ctx, in)
}

func requireInternalKey(t *testing.T, ctx context.Context) {
	t.Helper()
	md, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)
	require.Equal(t, []string{"internal-secret"}, md.Get("x-internal-key"))
}

func TestConfigClientMapsProjectMetadataAndAddsAuthentication(t *testing.T) {
	client := &ConfigClient{apiKey: "internal-secret"}
	client.client = &fakeControlPlaneClient{getProjectFn: func(ctx context.Context, in *pb.GetProjectRequest) (*pb.GetProjectResponse, error) {
		requireInternalKey(t, ctx)
		require.Equal(t, "project-1", in.ProjectId)
		return &pb.GetProjectResponse{
			Id: "project-1", Name: "Example", Slug: "example", SchemaName: "project_schema", OwnerId: "owner-1", Version: 7,
		}, nil
	}}

	got, err := client.GetProject(context.Background(), "project-1")

	require.NoError(t, err)
	require.Equal(t, "project-1", got.ID)
	require.Equal(t, "Example", got.Name)
	require.Equal(t, "example", got.Slug)
	require.Equal(t, "project_schema", got.SchemaName)
	require.Equal(t, "owner-1", got.OwnerID)
	require.Equal(t, int32(7), got.Version)
}

func TestConfigClientMapsRuntimeConfiguration(t *testing.T) {
	client := &ConfigClient{apiKey: "internal-secret"}
	client.client = &fakeControlPlaneClient{getProjectConfigFn: func(ctx context.Context, in *pb.GetProjectConfigRequest) (*pb.GetProjectConfigResponse, error) {
		requireInternalKey(t, ctx)
		require.Equal(t, "project-1", in.ProjectId)
		return &pb.GetProjectConfigResponse{
			ProjectId:  "project-1",
			Slug:       "example",
			SchemaName: "project_schema",
			Version:    3,
			ConfigJson: `{
				"entities":{"Order":{"name":"Order","pool":[{"name":"title","type":"string","required":true}]}},
				"modules":{"crud":{"enabled":true,"entityFunctions":{"Order":{"create":{"enabledFields":["title"]}}}}}
			}`,
		}, nil
	}}

	got, err := client.GetProjectConfig(context.Background(), "project-1")

	require.NoError(t, err)
	require.Equal(t, "project-1", got.ProjectID)
	require.Equal(t, "example", got.Slug)
	require.Equal(t, "project_schema", got.SchemaName)
	require.Equal(t, int32(3), got.Version)
	require.Equal(t, "Order", got.Entities["Order"].Name)
	require.True(t, got.Entities["Order"].Pool[0].Required)
	require.True(t, got.Modules["crud"].Enabled)
	require.Equal(t, []string{"title"}, got.Modules["crud"].EntityFunctions["Order"]["create"].EnabledFields)
}

func TestConfigClientReturnsTransportAndJSONErrors(t *testing.T) {
	wantErr := errors.New("rpc failed")
	client := &ConfigClient{apiKey: "internal-secret"}
	client.client = &fakeControlPlaneClient{
		getProjectFn: func(context.Context, *pb.GetProjectRequest) (*pb.GetProjectResponse, error) {
			return nil, wantErr
		},
		getProjectConfigFn: func(context.Context, *pb.GetProjectConfigRequest) (*pb.GetProjectConfigResponse, error) {
			return &pb.GetProjectConfigResponse{ConfigJson: "{"}, nil
		},
	}

	_, err := client.GetProject(context.Background(), "project-1")
	require.ErrorIs(t, err, wantErr)

	_, err = client.GetProjectConfig(context.Background(), "project-1")
	require.Error(t, err)
	require.ErrorContains(t, err, "parse config_json")
}
