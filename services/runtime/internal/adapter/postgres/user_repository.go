package postgres

import (
	"context"
	"errors"
	"fmt"

	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) table(schemaName string) string {
	return quoteIdent(schemaName) + "." + quoteIdent("user")
}

func (r *UserRepository) Create(ctx context.Context, schemaName string, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.Role == "" {
		user.Role = "user"
	}

	q := fmt.Sprintf(`
	INSERT INTO %s (
		%s, %s, %s, %s, %s, %s, %s, %s
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`,
		r.table(schemaName),
		quoteIdent("id"),
		quoteIdent("email"),
		quoteIdent("password"),
		quoteIdent("full_name"),
		quoteIdent("phone"),
		quoteIdent("role"),
		quoteIdent("created_at"),
		quoteIdent("updated_at"),
	)

	_, err := r.pool.Exec(ctx, q,
		user.ID,
		user.Email,
		user.PasswordHash,
		nullIfEmpty(user.FullName),
		nullIfEmpty(user.Phone),
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailTaken()
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, schemaName, email string) (*domain.User, error) {
	q := fmt.Sprintf(`
	SELECT
		%s, %s, %s,
		COALESCE(%s, ''),
		COALESCE(%s, ''),
		COALESCE(%s, 'user'),
		%s, %s
	FROM %s
	WHERE %s = $1
	`,
		quoteIdent("id"),
		quoteIdent("email"),
		quoteIdent("password"),
		quoteIdent("full_name"),
		quoteIdent("phone"),
		quoteIdent("role"),
		quoteIdent("created_at"),
		quoteIdent("updated_at"),
		r.table(schemaName),
		quoteIdent("email"),
	)

	var u domain.User
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.Phone,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, schemaName, id string) (*domain.User, error) {
	q := fmt.Sprintf(`
		SELECT
			%s, %s, %s,
			COALESCE(%s, ''),
			COALESCE(%s, ''),
			COALESCE(%s, 'user'),
			%s, %s
		FROM %s
		WHERE %s = $1
	`,
		quoteIdent("id"),
		quoteIdent("email"),
		quoteIdent("password"),
		quoteIdent("full_name"),
		quoteIdent("phone"),
		quoteIdent("role"),
		quoteIdent("created_at"),
		quoteIdent("updated_at"),
		r.table(schemaName),
		quoteIdent("id"),
	)

	var u domain.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.Phone,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) SetRole(ctx context.Context, schemaName, userID, role string) error {
	q := fmt.Sprintf(
		`UPDATE %s SET %s = $1, %s = $2 WHERE %s = $3`,
		r.table(schemaName),
		quoteIdent("role"),
		quoteIdent("updated_at"),
		quoteIdent("id"),
	)
	ct, err := r.pool.Exec(ctx, q, role, time.Now().UTC(), userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound("user not found")
	}
	return nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
