package container

import (
	"github.com/DoMinhHHung/backify/services/runtime/internal/config"
	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

type Container struct {
	Config       *config.Config
	ConfigClient port.ConfigClient
}

func New(cfg *config.Config, configClient port.ConfigClient) *Container {
	return &Container{
		Config:       cfg,
		ConfigClient: configClient,
	}
}
