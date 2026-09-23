package usecase

import (
	"context"
	"strings"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

type CRUD struct {
	records    port.RecordRepository
	permission port.PermissionEngine
}

func NewCRUD(records port.RecordRepository, permission port.PermissionEngine) *CRUD {
	return &CRUD{records: records, permission: permission}
}

func (uc *CRUD) resolveEntity(cfg *domain.ProjectConfig, name string) (string, *domain.Entity, error) {
	for k, e := range cfg.Entities {
		if strings.EqualFold(k, name) {
			return k, &e, nil
		}
	}
	return "", nil, domain.ErrNotFound("entity not found")
}

type CreateInput struct {
	ProjectConfig *domain.ProjectConfig
	Claims        *domain.AuthClaims
	EntityName    string
	Data          map[string]any
}

func (uc *CRUD) Create(ctx context.Context, in CreateInput) (domain.Record, error) {
	name, entity, err := uc.resolveEntity(in.ProjectConfig, in.EntityName)
	if err != nil {
		return nil, err
	}
	_ = entity

	data, err := uc.permission.CanCreate(in.ProjectConfig, name, in.Claims, in.Data)
	if err != nil {
		return nil, err
	}

	return uc.records.Create(ctx, in.ProjectConfig.SchemaName, name, data)
}

type GetInput struct {
	ProjectConfig *domain.ProjectConfig
	Claims        *domain.AuthClaims
	EntityName    string
	ID            string
}

func (uc *CRUD) Get(ctx context.Context, in GetInput) (domain.Record, error) {
	name, _, err := uc.resolveEntity(in.ProjectConfig, in.EntityName)
	if err != nil {
		return nil, err
	}

	rec, err := uc.records.FindByID(ctx, in.ProjectConfig.SchemaName, name, in.ID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, domain.ErrNotFound("record not found")
	}

	if err := uc.permission.CanReadOne(in.ProjectConfig, name, in.Claims, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

type ListInput struct {
	ProjectConfig *domain.ProjectConfig
	Claims        *domain.AuthClaims
	EntityName    string
	Limit         int
	Offset        int
}

func (uc *CRUD) List(ctx context.Context, in ListInput) (*domain.ListResult, error) {
	name, _, err := uc.resolveEntity(in.ProjectConfig, in.EntityName)
	if err != nil {
		return nil, err
	}

	ownerCol, ownerID, err := uc.permission.CanListFilter(in.ProjectConfig, name, in.Claims)
	if err != nil {
		return nil, err
	}

	items, total, err := uc.records.List(ctx, in.ProjectConfig.SchemaName, name, ownerCol, ownerID, in.Limit, in.Offset)
	if err != nil {
		return nil, err
	}
	return &domain.ListResult{Items: items, Total: total}, nil
}

type UpdateInput struct {
	ProjectConfig *domain.ProjectConfig
	Claims        *domain.AuthClaims
	EntityName    string
	ID            string
	Data          map[string]any
}

func (uc *CRUD) Update(ctx context.Context, in UpdateInput) (domain.Record, error) {
	name, _, err := uc.resolveEntity(in.ProjectConfig, in.EntityName)
	if err != nil {
		return nil, err
	}

	existing, err := uc.records.FindByID(ctx, in.ProjectConfig.SchemaName, name, in.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound("record not found")
	}

	if err := uc.permission.CanUpdate(in.ProjectConfig, name, in.Claims, existing); err != nil {
		return nil, err
	}

	if col, ok := uc.permission.OwnerColumn(in.ProjectConfig, name); ok {
		delete(in.Data, col)
		delete(in.Data, toSnakeLocal(col))
	}

	return uc.records.Update(ctx, in.ProjectConfig.SchemaName, name, in.ID, in.Data)
}

type DeleteInput struct {
	ProjectConfig *domain.ProjectConfig
	Claims        *domain.AuthClaims
	EntityName    string
	ID            string
}

func (uc *CRUD) Delete(ctx context.Context, in DeleteInput) error {
	name, _, err := uc.resolveEntity(in.ProjectConfig, in.EntityName)
	if err != nil {
		return err
	}

	existing, err := uc.records.FindByID(ctx, in.ProjectConfig.SchemaName, name, in.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound("record not found")
	}

	if err := uc.permission.CanDelete(in.ProjectConfig, name, in.Claims, existing); err != nil {
		return err
	}

	return uc.records.Delete(ctx, in.ProjectConfig.SchemaName, name, in.ID)
}

func toSnakeLocal(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
