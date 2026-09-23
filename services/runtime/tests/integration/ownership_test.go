//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/postgres"
	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/permission"
	"github.com/DoMinhHHung/backify/services/runtime/internal/usecase"
)

func testConfig() *domain.ProjectConfig {
	return &domain.ProjectConfig{
		ProjectID:  "test-project",
		SchemaName: "proj_ownership_test",
		Version:    1,
		Entities: map[string]domain.Entity{
			"User": {
				Name: "User",
				Pool: []domain.Field{
					{Name: "id", Type: "uuid", Required: true, Unique: true, System: true},
					{Name: "email", Type: "email", Required: true, Unique: true, System: true},
					{Name: "password", Type: "password", Required: true, System: true},
				},
			},
			"Order": {
				Name: "Order",
				Pool: []domain.Field{
					{Name: "id", Type: "uuid", Required: true, Unique: true, System: true},
					{Name: "title", Type: "string", Required: true},
					{Name: "userId", Type: "relation", RelationTo: "User", RelationCardinality: "n-1"},
				},
			},
		},
	}
}

func setup(t *testing.T) (*usecase.CRUD, *domain.ProjectConfig, func()) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dsn)
	require.NoError(t, err)

	cfg := testConfig()
	migrator := postgres.NewSchemaMigrator(pool)
	require.NoError(t, migrator.EnsureSchema(ctx, cfg))

	records := postgres.NewRecordRepository(pool)
	engine := permission.NewEngine()
	crud := usecase.NewCRUD(records, engine)

	cleanup := func() {
		_, _ = pool.Exec(ctx, `DROP SCHEMA IF EXISTS proj_ownership_test CASCADE`)
		pool.Close()
	}
	return crud, cfg, cleanup
}

func TestOwnershipFourMetrics(t *testing.T) {
	crud, cfg, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	userA := &domain.AuthClaims{UserID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", ProjectID: cfg.ProjectID, Role: "user"}
	userB := &domain.AuthClaims{UserID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", ProjectID: cfg.ProjectID, Role: "user"}
	admin := &domain.AuthClaims{UserID: "cccccccc-cccc-cccc-cccc-cccccccccccc", ProjectID: cfg.ProjectID, Role: "admin"}

	// seed users rows so FK passes (nếu có FK)
	// nếu FK strict fail, tắt FK trong test schema hoặc insert user trước qua raw SQL

	recA, err := crud.Create(ctx, usecase.CreateInput{
		ProjectConfig: cfg,
		Claims:        userA,
		EntityName:    "Order",
		Data:          map[string]any{"title": "A's order"},
	})
	require.NoError(t, err)
	idA, _ := recA["id"].(string)
	require.NotEmpty(t, idA)

	listB, err := crud.List(ctx, usecase.ListInput{
		ProjectConfig: cfg,
		Claims:        userB,
		EntityName:    "Order",
	})
	require.NoError(t, err)
	require.Equal(t, 0, listB.Total)

	_, err = crud.Get(ctx, usecase.GetInput{
		ProjectConfig: cfg,
		Claims:        userB,
		EntityName:    "Order",
		ID:            idA,
	})
	require.Error(t, err)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok)
	require.Equal(t, "forbidden", de.Code)

	listAdmin, err := crud.List(ctx, usecase.ListInput{
		ProjectConfig: cfg,
		Claims:        admin,
		EntityName:    "Order",
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, listAdmin.Total, 1)

	// Metric 4: implicit — không có policy file; ownership chỉ từ relation trong config
}
