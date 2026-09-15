//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"backify/services/control-plane/internal/adapter/postgres"
	"backify/services/control-plane/internal/domain"
)

func setupTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	migrationSQL, err := os.ReadFile("../../../migrations/001_init.up.sql")
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	container, err := tcpostgres.Run(ctx, "postgres:16",
		tcpostgres.WithDatabase("backify_test"),
		tcpostgres.WithUsername("backify"),
		tcpostgres.WithPassword("backify"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := waitUntilReachable(ctx, pool, 10, 500*time.Millisecond); err != nil {
		t.Fatalf("database not reachable after container ready: %v", err)
	}

	if err := execWithRetry(ctx, pool, string(migrationSQL), 5, 500*time.Millisecond); err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	return pool
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

func execWithRetry(ctx context.Context, pool *pgxpool.Pool, sql string, attempts int, delay time.Duration) error {
	var err error
	for i := 0; i < attempts; i++ {
		if _, err = pool.Exec(ctx, sql); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return err
}

func TestProjectRepo_CreateAndGetByID(t *testing.T) {
	pool := setupTestPool(t)
	repo := postgres.NewProjectRepo(pool)
	ctx := context.Background()

	project, err := domain.NewProject("Shop App", "shop-app")
	if err != nil {
		t.Fatalf("failed to build project: %v", err)
	}

	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	fetched, err := repo.GetByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("failed to get project: %v", err)
	}
	if fetched.Subdomain != project.Subdomain {
		t.Fatalf("expected subdomain %s, got %s", project.Subdomain, fetched.Subdomain)
	}
	if fetched.SchemaName != project.SchemaName {
		t.Fatalf("expected schema name %s, got %s", project.SchemaName, fetched.SchemaName)
	}
}

func TestProjectRepo_CreateDuplicateSubdomain(t *testing.T) {
	pool := setupTestPool(t)
	repo := postgres.NewProjectRepo(pool)
	ctx := context.Background()

	first, _ := domain.NewProject("Shop App", "shop-app")
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("failed to create first project: %v", err)
	}

	second, _ := domain.NewProject("Another Shop", "shop-app")
	err := repo.Create(ctx, second)
	if err != domain.ErrSubdomainTaken {
		t.Fatalf("expected ErrSubdomainTaken, got %v", err)
	}
}

func TestProjectRepo_GetByID_NotFound(t *testing.T) {
	pool := setupTestPool(t)
	repo := postgres.NewProjectRepo(pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "does-not-exist")
	if err != domain.ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectRepo_Update(t *testing.T) {
	pool := setupTestPool(t)
	repo := postgres.NewProjectRepo(pool)
	ctx := context.Background()

	project, _ := domain.NewProject("Shop App", "shop-app")
	if err := repo.Create(ctx, project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	project.Suspend()
	if err := repo.Update(ctx, project); err != nil {
		t.Fatalf("failed to update project: %v", err)
	}

	fetched, err := repo.GetByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("failed to get project: %v", err)
	}
	if fetched.Status != domain.ProjectStatusSuspended {
		t.Fatalf("expected status suspended, got %s", fetched.Status)
	}
}
