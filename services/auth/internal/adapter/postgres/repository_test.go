//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"backify/services/auth/internal/adapter/postgres"
	"backify/services/auth/internal/domain"
)

// setupConnManager dựng một Postgres container đóng vai trò instance admin
// của Auth (giống backify-auth-postgres thật), rồi dùng chính DatabaseManager
// của Bước 3 để tạo database cho projectID — bài test này vì vậy xác nhận cả
// Bước 3 lẫn Bước 4 hoạt động đúng với nhau, không chỉ mock ConnManager để
// né việc tạo database thật. Trả về *ConnManager (không phải pool đã resolve)
// vì đó chính là kiểu repo constructor cần.
func setupConnManager(t *testing.T, projectID string) *postgres.ConnManager {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16",
		tcpostgres.WithDatabase("auth"),
		tcpostgres.WithUsername("backify"),
		tcpostgres.WithPassword("backify"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	adminPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect admin pool: %v", err)
	}
	t.Cleanup(adminPool.Close)

	if err := waitUntilReachable(ctx, adminPool, 10, 500*time.Millisecond); err != nil {
		t.Fatalf("admin db not reachable: %v", err)
	}

	dbManager := postgres.NewDatabaseManager(adminPool)
	if err := dbManager.CreateDatabase(ctx, projectID); err != nil {
		t.Fatalf("failed to create project database: %v", err)
	}

	conns := postgres.NewConnManager(adminPool)
	t.Cleanup(conns.Close)
	return conns
}

func waitUntilReachable(ctx context.Context, pool *pgxpool.Pool, attempts int, delay time.Duration) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = pool.Ping(ctx); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return err
}

func TestUserRepo_CreateAndGet(t *testing.T) {
	projectID := "int-test-user"
	conns := setupConnManager(t, projectID)
	repo := postgres.NewUserRepo(conns)
	ctx := context.Background()

	user, err := domain.NewUser(projectID, "alice@example.com", "Alice", "0901234567")
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	if err := user.SetPasswordHash("hashed"); err != nil {
		t.Fatalf("SetPasswordHash: %v", err)
	}

	if err := repo.Create(ctx, projectID, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Create(ctx, projectID, user); err != domain.ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken on duplicate email, got %v", err)
	}

	fetched, err := repo.GetByEmail(ctx, projectID, "alice@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if fetched.ID != user.ID {
		t.Fatalf("expected id %s, got %s", user.ID, fetched.ID)
	}

	if _, err := repo.GetByID(ctx, projectID, "missing"); err != domain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRefreshTokenRepo_CreateGetUpdateRevoke(t *testing.T) {
	projectID := "int-test-refresh"
	conns := setupConnManager(t, projectID)
	users := postgres.NewUserRepo(conns)
	repo := postgres.NewRefreshTokenRepo(conns)
	ctx := context.Background()

	user, _ := domain.NewUser(projectID, "bob@example.com", "Bob", "")
	_ = user.SetPasswordHash("hashed")
	if err := users.Create(ctx, projectID, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	token, raw, err := domain.NewRefreshToken(projectID, user.ID, domain.RefreshDuration(domain.RefreshDuration7d))
	if err != nil {
		t.Fatalf("NewRefreshToken: %v", err)
	}
	if raw == "" {
		t.Fatal("expected non-empty raw token")
	}
	if err := repo.Create(ctx, projectID, token); err != nil {
		t.Fatalf("Create: %v", err)
	}

	fetched, err := repo.GetByTokenHash(ctx, projectID, token.TokenHash)
	if err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}
	if fetched.ID != token.ID {
		t.Fatalf("expected id %s, got %s", token.ID, fetched.ID)
	}

	fetched.MarkUsed(time.Now().UTC())
	if err := repo.Update(ctx, projectID, fetched); err != nil {
		t.Fatalf("Update: %v", err)
	}
	reloaded, _ := repo.GetByTokenHash(ctx, projectID, token.TokenHash)
	if !reloaded.IsUsed() {
		t.Fatal("expected token to be marked used after Update")
	}

	if err := repo.RevokeAllByUser(ctx, projectID, user.ID); err != nil {
		t.Fatalf("RevokeAllByUser: %v", err)
	}
	reloaded, _ = repo.GetByTokenHash(ctx, projectID, token.TokenHash)
	if !reloaded.IsRevoked() {
		t.Fatal("expected token to be revoked after RevokeAllByUser")
	}

	if _, err := repo.GetByTokenHash(ctx, projectID, "nonexistent-hash"); err != domain.ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestPasswordResetRepo_CreateGetUpdate(t *testing.T) {
	projectID := "int-test-reset"
	conns := setupConnManager(t, projectID)
	users := postgres.NewUserRepo(conns)
	repo := postgres.NewPasswordResetRepo(conns)
	ctx := context.Background()

	user, _ := domain.NewUser(projectID, "carol@example.com", "Carol", "")
	_ = user.SetPasswordHash("hashed")
	if err := users.Create(ctx, projectID, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	token, raw, err := domain.NewPasswordResetToken(projectID, user.ID)
	if err != nil {
		t.Fatalf("NewPasswordResetToken: %v", err)
	}
	if raw == "" {
		t.Fatal("expected non-empty raw token")
	}
	if err := repo.Create(ctx, projectID, token); err != nil {
		t.Fatalf("Create: %v", err)
	}

	fetched, err := repo.GetByTokenHash(ctx, projectID, token.TokenHash)
	if err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}

	fetched.MarkUsed(time.Now().UTC())
	if err := repo.Update(ctx, projectID, fetched); err != nil {
		t.Fatalf("Update: %v", err)
	}
	reloaded, _ := repo.GetByTokenHash(ctx, projectID, token.TokenHash)
	if !reloaded.IsUsed() {
		t.Fatal("expected token to be marked used after Update")
	}
}
