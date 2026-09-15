package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backify/services/control-plane/internal/domain"
)

type FieldRepo struct {
	pool *pgxpool.Pool
}

func NewFieldRepo(pool *pgxpool.Pool) *FieldRepo {
	return &FieldRepo{pool: pool}
}

// Create lưu field và chuyển vi phạm ràng buộc duy nhất thành ErrFieldNameTaken.
func (r *FieldRepo) Create(ctx context.Context, field *domain.Field) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO fields (id, entity_id, name, type, is_system, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, field.ID, field.EntityID, field.Name, string(field.Type), field.IsSystem, field.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrFieldNameTaken
		}
		return err
	}
	return nil
}

func (r *FieldRepo) GetByID(ctx context.Context, id string) (*domain.Field, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, entity_id, name, type, is_system, created_at
		FROM fields WHERE id = $1
	`, id)
	return scanField(row)
}

func (r *FieldRepo) GetByEntityAndName(ctx context.Context, entityID, name string) (*domain.Field, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, entity_id, name, type, is_system, created_at
		FROM fields WHERE entity_id = $1 AND name = $2
	`, entityID, name)
	return scanField(row)
}

// ListByEntity trả về các field của entity theo thứ tự tạo tăng dần.
func (r *FieldRepo) ListByEntity(ctx context.Context, entityID string) ([]*domain.Field, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, entity_id, name, type, is_system, created_at
		FROM fields WHERE entity_id = $1 ORDER BY created_at ASC
	`, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []*domain.Field
	for rows.Next() {
		f, err := scanField(rows)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

// Delete xóa field theo ID và trả về ErrFieldNotFound nếu không có bản ghi tương ứng.
func (r *FieldRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM fields WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrFieldNotFound
	}
	return nil
}

func scanField(row rowScanner) (*domain.Field, error) {
	var f domain.Field
	var fieldType string
	err := row.Scan(&f.ID, &f.EntityID, &f.Name, &fieldType, &f.IsSystem, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFieldNotFound
		}
		return nil, err
	}
	f.Type = domain.FieldType(fieldType)
	return &f, nil
}
