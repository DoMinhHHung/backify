package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backify/services/control-plane/internal/domain"
)

type rowScanner interface {
	Scan(dest ...interface{}) error
}

type ProjectRepo struct {
	pool *pgxpool.Pool
}

func NewProjectRepo(pool *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{pool: pool}
}

func (r *ProjectRepo) Create(ctx context.Context, project *domain.Project) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO projects (id, name, subdomain, schema_name, plan, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, project.ID, project.Name, project.Subdomain, project.SchemaName, string(project.Plan), string(project.Status), project.CreatedAt, project.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrSubdomainTaken
		}
		return err
	}
	return nil
}

func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, subdomain, schema_name, plan, status, created_at, updated_at
		FROM projects WHERE id = $1
	`, id)
	return scanProject(row)
}

func (r *ProjectRepo) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Project, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, subdomain, schema_name, plan, status, created_at, updated_at
		FROM projects WHERE subdomain = $1
	`, subdomain)
	return scanProject(row)
}

func (r *ProjectRepo) List(ctx context.Context) ([]*domain.Project, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, subdomain, schema_name, plan, status, created_at, updated_at
		FROM projects ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepo) Update(ctx context.Context, project *domain.Project) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE projects SET name = $2, plan = $3, status = $4, updated_at = $5
		WHERE id = $1
	`, project.ID, project.Name, string(project.Plan), string(project.Status), project.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func scanProject(row rowScanner) (*domain.Project, error) {
	var p domain.Project
	var plan, status string
	err := row.Scan(&p.ID, &p.Name, &p.Subdomain, &p.SchemaName, &plan, &status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}
	p.Plan = domain.Plan(plan)
	p.Status = domain.ProjectStatus(status)
	return &p, nil
}
