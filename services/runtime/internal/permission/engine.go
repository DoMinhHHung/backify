package permission

import (
	"strings"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) OwnerColumn(cfg *domain.ProjectConfig, entityName string) (string, bool) {
	entity, ok := cfg.Entities[entityName]
	if !ok {
		return "", false
	}
	for _, f := range entity.Pool {
		if f.Type != "relation" {
			continue
		}
		if !strings.EqualFold(f.RelationTo, "User") {
			continue
		}
		if f.RelationCardinality == "n-1" || f.RelationCardinality == "many_to_one" {
			return f.Name, true
		}
	}
	return "", false
}

func (e *Engine) CanCreate(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims, input map[string]any) (map[string]any, error) {
	if claims == nil {
		return nil, domain.ErrUnauthorized()
	}
	out := make(map[string]any, len(input)+1)
	for k, v := range input {
		out[k] = v
	}
	if col, ok := e.OwnerColumn(cfg, entityName); ok {
		out[col] = claims.UserID
	}
	return out, nil
}

func (e *Engine) CanReadOne(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims, record domain.Record) error {
	if claims == nil {
		return domain.ErrUnauthorized()
	}
	if claims.Role == "admin" {
		return nil
	}
	col, ok := e.OwnerColumn(cfg, entityName)
	if !ok {
		return nil
	}
	owner, _ := record[col].(string)
	if owner == "" {
		if v, ok := record[toSnake(col)].(string); ok {
			owner = v
		}
	}
	if owner != claims.UserID {
		return domain.ErrForbidden()
	}
	return nil
}

func (e *Engine) CanListFilter(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims) (string, string, error) {
	if claims == nil {
		return "", "", domain.ErrUnauthorized()
	}
	if claims.Role == "admin" {
		return "", "", nil
	}
	col, ok := e.OwnerColumn(cfg, entityName)
	if !ok {
		return "", "", nil
	}
	return col, claims.UserID, nil
}

func (e *Engine) CanUpdate(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims, record domain.Record) error {
	if claims == nil {
		return domain.ErrUnauthorized()
	}
	col, ok := e.OwnerColumn(cfg, entityName)
	if !ok {
		return nil
	}
	owner := recordString(record, col)
	if owner != claims.UserID {
		return domain.ErrForbidden()
	}
	return nil
}

func (e *Engine) CanDelete(cfg *domain.ProjectConfig, entityName string, claims *domain.AuthClaims, record domain.Record) error {
	return e.CanUpdate(cfg, entityName, claims, record)
}

func recordString(r domain.Record, field string) string {
	if v, ok := r[field].(string); ok && v != "" {
		return v
	}
	if v, ok := r[toSnake(field)].(string); ok {
		return v
	}
	return ""
}

func toSnake(s string) string {
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
