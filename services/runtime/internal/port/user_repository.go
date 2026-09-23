package port

import (
	"context"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, schemaName string, user *domain.User) error
	FindByEmail(ctx context.Context, schemaName, email string) (*domain.User, error)
	FindByID(ctx context.Context, schemaName, id string) (*domain.User, error)
	SetRole(ctx context.Context, schemaName, userID, role string) error
}
