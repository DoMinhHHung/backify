package usecase

import (
	"context"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

type PromoteUser struct {
	users        port.UserRepository
	configClient port.ConfigClient
}

func NewPromoteUser(users port.UserRepository, configClient port.ConfigClient) *PromoteUser {
	return &PromoteUser{users: users, configClient: configClient}
}

type PromoteUserInput struct {
	ProjectID string
	UserID    string
	Role      string
}

func (uc *PromoteUser) Execute(ctx context.Context, in PromoteUserInput) error {
	if in.Role != "admin" && in.Role != "user" {
		return domain.ErrValidation("role must be admin or user")
	}
	cfg, err := uc.configClient.GetProjectConfig(ctx, in.ProjectID)
	if err != nil {
		return err
	}
	return uc.users.SetRole(ctx, cfg.SchemaName, in.UserID, in.Role)
}
