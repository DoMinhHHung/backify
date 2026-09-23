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
	CRUD            *usecase.CRUD
	PromoteUser     *usecase.PromoteUser
	ConfigEvents    *usecase.ConfigEvents
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
	records port.RecordRepository,
	permission port.PermissionEngine,
) *Container {

	c := &Container{
		Config:         cfg,
		ConfigClient:   configClient,
		ConfigCache:    configCache,
		SchemaMigrator: schemaMigrator,
		TokenService:   tokens,
	}

	c.BootstrapSchema = usecase.NewBootstrapSchema(configClient, schemaMigrator, configCache)
	c.Auth = usecase.NewAuth(users, hasher, tokens)
	c.CRUD = usecase.NewCRUD(records, permission)
	c.PromoteUser = usecase.NewPromoteUser(users, configClient)
	c.BootstrapSchema = usecase.NewBootstrapSchema(configClient, schemaMigrator, configCache)
	c.ConfigEvents = usecase.NewConfigEvents(configCache, c.BootstrapSchema)
	return c
}
