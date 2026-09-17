package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"backify/services/auth/internal/domain"
)

// pgUniqueViolation là SQLSTATE Postgres trả khi vi phạm UNIQUE constraint —
// bảng users chỉ có đúng một UNIQUE (email), nên gặp mã này khi Create nghĩa
// là email đã tồn tại, không cần so tên constraint.
const pgUniqueViolation = "23505"

// UserRepo implement port.UserRepository, định tuyến qua ConnManager tới
// đúng database auth_proj_<projectID> cho mỗi lời gọi.
type UserRepo struct {
	conns *ConnManager
}

func NewUserRepo(conns *ConnManager) *UserRepo {
	return &UserRepo{conns: conns}
}

func (r *UserRepo) Create(ctx context.Context, projectID string, user *domain.User) error {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, project_id, email, password_hash, full_name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.ProjectID, user.Email, user.PasswordHash, user.FullName, user.Phone, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, projectID, userID string) (*domain.User, error) {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return scanUser(pool.QueryRow(ctx, `
		SELECT id, project_id, email, password_hash, full_name, phone, created_at, updated_at
		FROM users WHERE id = $1
	`, userID))
}

func (r *UserRepo) GetByEmail(ctx context.Context, projectID, email string) (*domain.User, error) {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return scanUser(pool.QueryRow(ctx, `
		SELECT id, project_id, email, password_hash, full_name, phone, created_at, updated_at
		FROM users WHERE email = $1
	`, email))
}

func (r *UserRepo) Update(ctx context.Context, projectID string, user *domain.User) error {
	pool, err := r.conns.Pool(ctx, projectID)
	if err != nil {
		return err
	}

	tag, err := pool.Exec(ctx, `
		UPDATE users SET password_hash = $1, full_name = $2, phone = $3, updated_at = $4
		WHERE id = $5
	`, user.PasswordHash, user.FullName, user.Phone, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.ProjectID, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}
