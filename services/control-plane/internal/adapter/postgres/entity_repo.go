package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backify/services/control-plane/internal/domain"
)

type EntityRepo struct {
	pool *pgxpool.Pool
}

// NewEntityRepo tạo kho lưu trữ entity dùng pool PostgreSQL đã cho.
func NewEntityRepo(pool *pgxpool.Pool) *EntityRepo {
	return &EntityRepo{pool: pool}
}

// Create lưu entity và trả về ErrEntityNameTaken khi tên đã tồn tại trong project.
func (r *EntityRepo) Create(ctx context.Context, entity *domain.Entity) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO entities (id, project_id, name, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, entity.ID, entity.ProjectID, entity.Name, entity.IsSystem, entity.CreatedAt, entity.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEntityNameTaken
		}
		return err
	}
	return nil
}

// GetByID lấy entity theo mã định danh và trả về ErrEntityNotFound nếu không tồn tại.
func (r *EntityRepo) GetByID(ctx context.Context, id string) (*domain.Entity, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, project_id, name, is_system, created_at, updated_at
		FROM entities WHERE id = $1
	`, id)
	return scanEntity(row)
}

// GetByProjectAndName lấy entity theo project và tên.
func (r *EntityRepo) GetByProjectAndName(ctx context.Context, projectID, name string) (*domain.Entity, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, project_id, name, is_system, created_at, updated_at
		FROM entities WHERE project_id = $1 AND name = $2
	`, projectID, name)
	return scanEntity(row)
}

// ListByProject liệt kê entity của project theo thứ tự tạo tăng dần.
func (r *EntityRepo) ListByProject(ctx context.Context, projectID string) ([]*domain.Entity, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, name, is_system, created_at, updated_at
		FROM entities WHERE project_id = $1 ORDER BY created_at ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*domain.Entity
	for rows.Next() {
		e, err := scanEntity(rows)
		if err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

func scanEntity(row rowScanner) (*domain.Entity, error) {
	var e domain.Entity
	err := row.Scan(&e.ID, &e.ProjectID, &e.Name, &e.IsSystem, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEntityNotFound
		}
		return nil, err
	}
	return &e, nil
}
