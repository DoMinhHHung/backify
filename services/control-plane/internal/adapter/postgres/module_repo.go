package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

type ModuleRepo struct {
	pool *pgxpool.Pool
}

func NewModuleRepo(pool *pgxpool.Pool) *ModuleRepo {
	return &ModuleRepo{pool: pool}
}

func newRowID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Create bảo đảm mỗi project chỉ có một module cùng tên. Nếu bản ghi đã tồn tại, module đầu vào
// được cập nhật bằng ID và thời điểm tạo đã lưu.
func (r *ModuleRepo) Create(ctx context.Context, module *domain.Module) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO modules (id, project_id, name, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (project_id, name) DO NOTHING
	`, module.ID, module.ProjectID, string(module.Name), module.CreatedAt)
	if err != nil {
		return err
	}

	persisted, err := r.GetByProjectAndName(ctx, module.ProjectID, module.Name)
	if err != nil {
		return err
	}
	module.ID = persisted.ID
	module.CreatedAt = persisted.CreatedAt
	return nil
}

func (r *ModuleRepo) GetByProjectAndName(ctx context.Context, projectID string, name domain.ModuleName) (*domain.Module, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, project_id, name, created_at
		FROM modules WHERE project_id = $1 AND name = $2
	`, projectID, string(name))

	var m domain.Module
	var moduleName string
	err := row.Scan(&m.ID, &m.ProjectID, &moduleName, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrModuleNotFound
		}
		return nil, err
	}
	m.Name = domain.ModuleName(moduleName)
	return &m, nil
}

func (r *ModuleRepo) ListByProject(ctx context.Context, projectID string) ([]*domain.Module, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, name, created_at
		FROM modules WHERE project_id = $1 ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []*domain.Module
	for rows.Next() {
		var m domain.Module
		var moduleName string
		if err := rows.Scan(&m.ID, &m.ProjectID, &moduleName, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Name = domain.ModuleName(moduleName)
		modules = append(modules, &m)
	}
	return modules, rows.Err()
}

// EnsureFunction tạo function nếu chưa tồn tại và luôn trả về ID của bản ghi đã lưu.
func (r *ModuleRepo) EnsureFunction(ctx context.Context, moduleID, functionName string) (string, error) {
	id, err := newRowID()
	if err != nil {
		return "", err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO functions (id, module_id, name, created_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (module_id, name) DO NOTHING
	`, id, moduleID, functionName)
	if err != nil {
		return "", err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id FROM functions WHERE module_id = $1 AND name = $2
	`, moduleID, functionName)

	var functionID string
	if err := row.Scan(&functionID); err != nil {
		return "", err
	}
	return functionID, nil
}

// ToggleFunctionField tạo hoặc cập nhật liên kết field của function với trạng thái enabled được yêu cầu.
func (r *ModuleRepo) ToggleFunctionField(ctx context.Context, functionID, fieldID string, enabled bool) error {
	id, err := newRowID()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO function_fields (id, function_id, field_id, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
		ON CONFLICT (function_id, field_id) DO UPDATE SET enabled = $4, updated_at = now()
	`, id, functionID, fieldID, enabled)
	return err
}

// ListFieldUsages trả về các function đang bật field; các liên kết đã tắt không được tính là đang sử dụng.
func (r *ModuleRepo) ListFieldUsages(ctx context.Context, fieldID string) ([]port.FieldUsage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.name, f.name
		FROM function_fields ff
		JOIN functions f ON f.id = ff.function_id
		JOIN modules m ON m.id = f.module_id
		WHERE ff.field_id = $1 AND ff.enabled = true
	`, fieldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []port.FieldUsage
	for rows.Next() {
		var usage port.FieldUsage
		if err := rows.Scan(&usage.ModuleName, &usage.FunctionName); err != nil {
			return nil, err
		}
		usages = append(usages, usage)
	}
	return usages, rows.Err()
}

// DisableFieldEverywhere tắt field trong mọi function nhưng không xóa field.
func (r *ModuleRepo) DisableFieldEverywhere(ctx context.Context, fieldID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE function_fields SET enabled = false, updated_at = now()
		WHERE field_id = $1
	`, fieldID)
	return err
}

// DisableAndDeleteField tắt mọi liên kết rồi xóa field trong cùng một transaction.
// Hàm trả về ErrFieldNotFound nếu field không tồn tại.
func (r *ModuleRepo) DisableAndDeleteField(ctx context.Context, fieldID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE function_fields SET enabled = false, updated_at = now()
		WHERE field_id = $1
	`, fieldID); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `DELETE FROM fields WHERE id = $1`, fieldID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrFieldNotFound
	}

	return tx.Commit(ctx)
}
