package port

import (
	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type PermissionEngine interface {
	OwnerColumn(cfg *domain.ProjectConfig, entityName string) (string, bool)
	CanCreate(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims, input map[string]any) (map[string]any, error)
	CanReadOne(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims, record domain.Record) error
	CanListFilter(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims) (ownerColumn, ownerID string, err error)
	CanUpdate(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims, record domain.Record) error
	CanDelete(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims, record domain.Record) error
}
