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
	Auth            *usecase.Auth
	TokenService    port.TokenService
}

func New(
	cfg *config.Config,
	configClient port.ConfigClient,
	configCache port.ConfigCache,
	schemaMigrator port.SchemaMigrator,
	users port.UserRepository,
	hasher port.PasswordHasher,
	tokens port.TokenService,
) *Container {
	c := &Container{
		Config:         cfg,
		ConfigClient:   configClient,
		ConfigCache:    configCache,
		SchemaMigrator: schemaMigrator,
		TokenService:   tokens,
	}
	c.BootstrapSchema = usecase.NewBootstrapSchema(configClient, schemaMigrator)
	c.Auth = usecase.NewAuth(users, hasher, tokens)
	return c
}
