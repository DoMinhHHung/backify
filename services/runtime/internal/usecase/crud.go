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

// resolveEntity tìm entity không phân biệt hoa thường nhưng luôn ẩn entity User khỏi CRUD công khai.
func (uc *CRUD) resolveEntity(cfg *domain.ProjectConfig, name string) (string, *domain.Entity, error) {
	for k, e := range cfg.Entities {
		if strings.EqualFold(k, name) {
			if strings.EqualFold(k, "User") {
				return "", nil, domain.ErrNotFound("entity not found")
			}
			return k, &e, nil
		}
	}
	return "", nil, domain.ErrNotFound("entity not found")
}

// crudEnabledFields trả các field được bật cho thao tác của entity, hoặc nil khi
// module, entity hay thao tác chưa được bật.
func crudEnabledFields(cfg *domain.ProjectConfig, entityName, fn string) []string {
	if cfg == nil {
		return nil
	}
	mod, ok := cfg.Modules["crud"]
	if !ok || !mod.Enabled {
		return nil
	}
	byEnt, ok := mod.EntityFunctions[entityName]
	if !ok {
		return nil
	}
	fc, ok := byEnt[fn]
	if !ok {
		return nil
	}
	return fc.EnabledFields
}

func isSystemField(name string) bool {
	switch name {
	case "id", "createdAt", "created_at", "updatedAt", "updated_at":
		return true
	default:
		return false
	}
}

// filterCRUDFields chỉ sao chép field nằm trong allowlist và luôn loại các field hệ thống.
func filterCRUDFields(data map[string]any, allowed []string) map[string]any {
	out := make(map[string]any)
	if len(data) == 0 || len(allowed) == 0 {
		return out
	}
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		set[a] = struct{}{}
	}
	for k, v := range data {
		if isSystemField(k) {
			continue
		}
		if _, ok := set[k]; ok {
			out[k] = v
		}
	}
	return out
}

type CreateInput struct {
	ProjectConfig *domain.ProjectConfig
	Claims        *domain.AuthClaims
	EntityName    string
	Data          map[string]any
}

// Create lọc dữ liệu theo cấu hình, loại owner do client cung cấp, áp dụng owner từ
// claims rồi tạo bản ghi trong schema dự án.
func (uc *CRUD) Create(ctx context.Context, in CreateInput) (domain.Record, error) {
	name, entity, err := uc.resolveEntity(in.ProjectConfig, in.EntityName)
	if err != nil {
		return nil, err
	}
	_ = entity

	allowed := crudEnabledFields(in.ProjectConfig, name, "create")
	data := filterCRUDFields(in.Data, allowed)

	if col, ok := uc.permission.OwnerColumn(in.ProjectConfig, name); ok {
		delete(data, col)
		delete(data, toSnakeLocal(col))
	}

	data, err = uc.permission.CanCreate(in.ProjectConfig, name, in.Claims, data)
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

// Get lấy bản ghi theo ID và chỉ trả về sau khi permission engine cho phép đọc.
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

// List áp dụng bộ lọc owner từ permission engine rồi trả danh sách cùng tổng số bản ghi.
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

// Update kiểm tra bản ghi và quyền sở hữu trước khi lọc field cho phép; owner và
// field hệ thống do client gửi không được cập nhật.
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

	allowed := crudEnabledFields(in.ProjectConfig, name, "update")
	data := filterCRUDFields(in.Data, allowed)

	if col, ok := uc.permission.OwnerColumn(in.ProjectConfig, name); ok {
		delete(data, col)
		delete(data, toSnakeLocal(col))
	}

	return uc.records.Update(ctx, in.ProjectConfig.SchemaName, name, in.ID, data)
}

type DeleteInput struct {
	ProjectConfig *domain.ProjectConfig
	Claims        *domain.AuthClaims
	EntityName    string
	ID            string
}

// Delete kiểm tra bản ghi tồn tại và quyền xóa trước khi gọi repository.
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
