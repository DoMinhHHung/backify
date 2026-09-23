package usecase

import (
	"context"
	"fmt"

	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

type BootstrapSchema struct {
	configClient   port.ConfigClient
	schemaMigrator port.SchemaMigrator
	configCache    port.ConfigCache
}

func NewBootstrapSchema(
	configClient port.ConfigClient,
	schemaMigrator port.SchemaMigrator,
	configCache port.ConfigCache,
) *BootstrapSchema {
	return &BootstrapSchema{
		configClient:   configClient,
		schemaMigrator: schemaMigrator,
		configCache:    configCache,
	}
}

type BootstrapSchemaInput struct {
	ProjectID string
}

type BootstrapSchemaOutput struct {
	SchemaName string
	Version    int32
	Applied    bool
}

func (uc *BootstrapSchema) Execute(ctx context.Context, in BootstrapSchemaInput) (*BootstrapSchemaOutput, error) {
	if uc.configCache != nil {
		uc.configCache.Invalidate(ctx, in.ProjectID)
	}

	cfg, err := uc.configClient.GetProjectConfig(ctx, in.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	applied, err := uc.schemaMigrator.AppliedVersion(ctx, cfg.SchemaName)
	if err != nil {
		return nil, err
	}

	if applied >= cfg.Version {
		return &BootstrapSchemaOutput{
			SchemaName: cfg.SchemaName,
			Version:    applied,
			Applied:    false,
		}, nil
	}

	if err := uc.schemaMigrator.EnsureSchema(ctx, cfg); err != nil {
		return nil, err
	}

	return &BootstrapSchemaOutput{
		SchemaName: cfg.SchemaName,
		Version:    cfg.Version,
		Applied:    true,
	}, nil
}
