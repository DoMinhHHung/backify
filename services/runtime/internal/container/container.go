package container

import (
	"github.com/DoMinhHHung/backify/services/runtime/internal/config"
	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
	"github.com/DoMinhHHung/backify/services/runtime/internal/usecase"
)

type Container struct {
	Config          *config.Config
	ConfigClient    port.ConfigClient
	ConfigCache     port.ConfigCache
	SchemaMigrator  port.SchemaMigrator
	BootstrapSchema *usecase.BootstrapSchema
}

func New(
	cfg *config.Config,
	configClient port.ConfigClient,
	configCache port.ConfigCache,
	schemaMigrator port.SchemaMigrator,
) *Container {
	c := &Container{
		Config:         cfg,
		ConfigClient:   configClient,
		ConfigCache:    configCache,
		SchemaMigrator: schemaMigrator,
	}
	c.BootstrapSchema = usecase.NewBootstrapSchema(configClient, schemaMigrator)
	return c
}
