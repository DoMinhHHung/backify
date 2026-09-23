package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

func TestBootstrapSchemaAppliesOnlyNewerConfig(t *testing.T) {
	tests := []struct {
		name           string
		appliedVersion int32
		wantApplied    bool
		wantVersion    int32
		wantMigration  bool
	}{
		{name: "new schema", appliedVersion: 0, wantApplied: true, wantVersion: 3, wantMigration: true},
		{name: "same version", appliedVersion: 3, wantApplied: false, wantVersion: 3},
		{name: "database ahead", appliedVersion: 4, wantApplied: false, wantVersion: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.ProjectConfig{ProjectID: "project-1", SchemaName: "project_schema", Version: 3}
			configClient := &fakeConfigClient{getProjectConfigFn: func(_ context.Context, id string) (*domain.ProjectConfig, error) {
				require.Equal(t, "project-1", id)
				return cfg, nil
			}}
			migrationCalled := false
			migrator := &fakeSchemaMigrator{
				appliedVersionFn: func(_ context.Context, schema string) (int32, error) {
					require.Equal(t, "project_schema", schema)
					return tt.appliedVersion, nil
				},
				ensureSchemaFn: func(_ context.Context, got *domain.ProjectConfig) error {
					migrationCalled = true
					require.Same(t, cfg, got)
					return nil
				},
			}

			out, err := NewBootstrapSchema(configClient, migrator).Execute(context.Background(), BootstrapSchemaInput{ProjectID: "project-1"})

			require.NoError(t, err)
			require.Equal(t, "project_schema", out.SchemaName)
			require.Equal(t, tt.wantVersion, out.Version)
			require.Equal(t, tt.wantApplied, out.Applied)
			require.Equal(t, tt.wantMigration, migrationCalled)
		})
	}
}

func TestBootstrapSchemaPropagatesFailures(t *testing.T) {
	wantErr := errors.New("failed")

	_, err := NewBootstrapSchema(
		&fakeConfigClient{getProjectConfigFn: func(context.Context, string) (*domain.ProjectConfig, error) { return nil, wantErr }},
		&fakeSchemaMigrator{},
	).Execute(context.Background(), BootstrapSchemaInput{ProjectID: "project-1"})
	require.ErrorIs(t, err, wantErr)
	require.ErrorContains(t, err, "get config")

	cfgClient := &fakeConfigClient{getProjectConfigFn: func(context.Context, string) (*domain.ProjectConfig, error) {
		return &domain.ProjectConfig{SchemaName: "schema", Version: 2}, nil
	}}
	_, err = NewBootstrapSchema(cfgClient, &fakeSchemaMigrator{
		appliedVersionFn: func(context.Context, string) (int32, error) { return 0, wantErr },
	}).Execute(context.Background(), BootstrapSchemaInput{ProjectID: "project-1"})
	require.ErrorIs(t, err, wantErr)

	_, err = NewBootstrapSchema(cfgClient, &fakeSchemaMigrator{
		ensureSchemaFn: func(context.Context, *domain.ProjectConfig) error { return wantErr },
	}).Execute(context.Background(), BootstrapSchemaInput{ProjectID: "project-1"})
	require.ErrorIs(t, err, wantErr)
}
