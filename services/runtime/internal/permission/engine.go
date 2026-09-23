package permission

import (
	"strings"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

// OwnerColumn trả field relation đầu tiên trỏ tới User theo cardinality n-1 hoặc
// many_to_one; entity hay relation không phù hợp trả "", false.
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

// CanCreate sao chép input và, nếu entity có owner, luôn ghi đè owner bằng user trong claims.
// Claims nil trả ErrUnauthorized.
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

// CanReadOne cho admin đọc mọi bản ghi và cho phép entity không có owner; người dùng
// thường chỉ được đọc bản ghi có owner khớp UserID.
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

// CanListFilter trả owner column cùng UserID cho người dùng thường; admin hoặc entity
// không có owner nhận bộ lọc rỗng. Claims nil trả ErrUnauthorized.
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

// CanUpdate chỉ cho phép cập nhật khi entity không có owner hoặc owner khớp UserID;
// hàm không áp dụng ngoại lệ admin.
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

// CanDelete áp dụng cùng quy tắc quyền sở hữu như CanUpdate.
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
