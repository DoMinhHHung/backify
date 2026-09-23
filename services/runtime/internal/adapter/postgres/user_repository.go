package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
			%s, %s, %s, %s, %s, %s, %s
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		r.table(schemaName),
		quoteIdent("id"),
		quoteIdent("email"),
		quoteIdent("password"),
		quoteIdent("full_name"),
		quoteIdent("phone"),
		quoteIdent("created_at"),
		quoteIdent("updated_at"),
	)

	_, err := r.pool.Exec(ctx, q,
		user.ID,
		user.Email,
		user.PasswordHash,
		nullIfEmpty(user.FullName),
		nullIfEmpty(user.Phone),
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
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
			%s, %s
		FROM %s
		WHERE %s = $1
	`,
		quoteIdent("id"),
		quoteIdent("email"),
		quoteIdent("password"),
		quoteIdent("full_name"),
		quoteIdent("phone"),
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
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find by email: %w", err)
	}
	u.Role = "user"
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, schemaName, id string) (*domain.User, error) {
	q := fmt.Sprintf(`
		SELECT
			%s, %s, %s,
			COALESCE(%s, ''),
			COALESCE(%s, ''),
			%s, %s
		FROM %s
		WHERE %s = $1
	`,
		quoteIdent("id"),
		quoteIdent("email"),
		quoteIdent("password"),
		quoteIdent("full_name"),
		quoteIdent("phone"),
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
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find by id: %w", err)
	}
	u.Role = "user"
	return &u, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
