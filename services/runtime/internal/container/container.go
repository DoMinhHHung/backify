package container

import (
	"github.com/DoMinhHHung/backify/services/runtime/internal/config"
	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

type Container struct {
	Config       *config.Config
	ConfigClient port.ConfigClient
	ConfigCache  port.ConfigCache
}

func New(
	cfg *config.Config,
	configClient port.ConfigClient,
	configCache port.ConfigCache,
) *Container {
	return &Container{
		Config:       cfg,
		ConfigClient: configClient,
		ConfigCache:  configCache,
	}
}
