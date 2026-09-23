package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/permission"
)

func sampleCRUDConfig() *domain.ProjectConfig {
	return &domain.ProjectConfig{
		ProjectID:  "p1",
		SchemaName: "proj_test",
		Entities: map[string]domain.Entity{
			"Order": {
				Name: "Order",
				Pool: []domain.Field{
					{Name: "id", Type: "uuid", System: true},
					{Name: "title", Type: "string"},
					{Name: "amount", Type: "number"},
					{Name: "status", Type: "string"},
					{Name: "userId", Type: "relation", RelationTo: "User", RelationCardinality: "n-1"},
				},
			},
		},
		Modules: map[string]domain.ModuleConfig{
			"crud": {
				Enabled: true,
				EntityFunctions: map[string]map[string]domain.FunctionConfig{
					"Order": {
						"create": {EnabledFields: []string{"title", "amount"}},
						"update": {EnabledFields: []string{"title"}},
					},
				},
			},
		},
	}
}

func TestCRUDEnabledFieldsCreate(t *testing.T) {
	cfg := sampleCRUDConfig()
	got := crudEnabledFields(cfg, "Order", "create")
	require.Equal(t, []string{"title", "amount"}, got)
}

func TestCRUDEnabledFieldsUpdate(t *testing.T) {
	cfg := sampleCRUDConfig()
	got := crudEnabledFields(cfg, "Order", "update")
	require.Equal(t, []string{"title"}, got)
}

func TestCRUDEnabledFieldsMissingModule(t *testing.T) {
	cfg := &domain.ProjectConfig{Entities: map[string]domain.Entity{"Order": {Name: "Order"}}}
	got := crudEnabledFields(cfg, "Order", "create")
	require.Nil(t, got)
}

func TestCRUDEnabledFieldsDisabledModule(t *testing.T) {
	cfg := sampleCRUDConfig()
	mod := cfg.Modules["crud"]
	mod.Enabled = false
	cfg.Modules["crud"] = mod
	got := crudEnabledFields(cfg, "Order", "create")
	require.Nil(t, got)
}

func TestFilterCRUDFieldsKeepsAllowedOnly(t *testing.T) {
	data := map[string]any{
		"title":       "ok",
		"amount":      10,
		"status":      "x",
		"hackerField": true,
		"id":          "should-drop",
		"createdAt":   "now",
	}
	got := filterCRUDFields(data, []string{"title", "amount"})
	require.Equal(t, map[string]any{"title": "ok", "amount": 10}, got)
}

func TestFilterCRUDFieldsEmptyAllowed(t *testing.T) {
	data := map[string]any{"title": "ok"}
	got := filterCRUDFields(data, nil)
	require.Empty(t, got)
}

func TestFilterCRUDFieldsStripsSystemEvenIfAllowed(t *testing.T) {
	data := map[string]any{"id": "x", "title": "t", "updated_at": "y"}
	got := filterCRUDFields(data, []string{"id", "title", "updated_at"})
	require.Equal(t, map[string]any{"title": "t"}, got)
}

func TestCRUDCreateFiltersInputAndForcesAuthenticatedOwner(t *testing.T) {
	records := &fakeRecordRepository{createFn: func(_ context.Context, schema, table string, columns map[string]any) (domain.Record, error) {
		require.Equal(t, "proj_test", schema)
		require.Equal(t, "Order", table)
		require.Equal(t, map[string]any{
			"title":  "kept",
			"amount": 42,
			"userId": "user-1",
		}, columns)
		return domain.Record{"id": "record-1"}, nil
	}}
	crud := NewCRUD(records, permission.NewEngine())

	record, err := crud.Create(context.Background(), CreateInput{
		ProjectConfig: sampleCRUDConfig(),
		Claims:        &domain.AuthClaims{UserID: "user-1", Role: "user"},
		EntityName:    "order",
		Data: map[string]any{
			"title":      "kept",
			"amount":     42,
			"status":     "not enabled",
			"id":         "caller-id",
			"userId":     "attacker",
			"user_id":    "attacker-snake",
			"created_at": "caller-time",
		},
	})

	require.NoError(t, err)
	require.Equal(t, "record-1", record["id"])
}

func TestCRUDCreateDoesNotMutateCallerInput(t *testing.T) {
	input := map[string]any{"title": "kept", "userId": "attacker"}
	records := &fakeRecordRepository{createFn: func(_ context.Context, _, _ string, columns map[string]any) (domain.Record, error) {
		columns["title"] = "repository mutation"
		return domain.Record{}, nil
	}}

	_, err := NewCRUD(records, permission.NewEngine()).Create(context.Background(), CreateInput{
		ProjectConfig: sampleCRUDConfig(),
		Claims:        &domain.AuthClaims{UserID: "user-1"},
		EntityName:    "Order",
		Data:          input,
	})

	require.NoError(t, err)
	require.Equal(t, map[string]any{"title": "kept", "userId": "attacker"}, input)
}

func TestCRUDCreateRejectsUnknownEntityAndAnonymousUser(t *testing.T) {
	crud := NewCRUD(&fakeRecordRepository{}, permission.NewEngine())

	_, err := crud.Create(context.Background(), CreateInput{
		ProjectConfig: sampleCRUDConfig(), Claims: &domain.AuthClaims{}, EntityName: "missing",
	})
	requireDomainError(t, err, "not_found", "entity not found")

	_, err = crud.Create(context.Background(), CreateInput{
		ProjectConfig: sampleCRUDConfig(), EntityName: "Order", Data: map[string]any{"title": "x"},
	})
	requireDomainError(t, err, "unauthorized", "unauthorized")
}

func TestCRUDGetHandlesMissingAndOwnership(t *testing.T) {
	var record domain.Record
	records := &fakeRecordRepository{findByIDFn: func(_ context.Context, schema, table, id string) (domain.Record, error) {
		require.Equal(t, "proj_test", schema)
		require.Equal(t, "Order", table)
		require.Equal(t, "record-1", id)
		return record, nil
	}}
	crud := NewCRUD(records, permission.NewEngine())
	input := GetInput{
		ProjectConfig: sampleCRUDConfig(), Claims: &domain.AuthClaims{UserID: "user-1"}, EntityName: "Order", ID: "record-1",
	}

	_, err := crud.Get(context.Background(), input)
	requireDomainError(t, err, "not_found", "record not found")

	record = domain.Record{"id": "record-1", "userId": "user-2"}
	_, err = crud.Get(context.Background(), input)
	requireDomainError(t, err, "forbidden", "forbidden")

	record["userId"] = "user-1"
	got, err := crud.Get(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, record, got)
}

func TestCRUDListForwardsOwnerFilterAndPagination(t *testing.T) {
	records := &fakeRecordRepository{listFn: func(_ context.Context, schema, table, ownerColumn, ownerID string, limit, offset int) ([]domain.Record, int, error) {
		require.Equal(t, "proj_test", schema)
		require.Equal(t, "Order", table)
		require.Equal(t, "userId", ownerColumn)
		require.Equal(t, "user-1", ownerID)
		require.Equal(t, 25, limit)
		require.Equal(t, 50, offset)
		return []domain.Record{{"id": "record-1"}}, 1, nil
	}}

	out, err := NewCRUD(records, permission.NewEngine()).List(context.Background(), ListInput{
		ProjectConfig: sampleCRUDConfig(),
		Claims:        &domain.AuthClaims{UserID: "user-1", Role: "user"},
		EntityName:    "ORDER",
		Limit:         25,
		Offset:        50,
	})

	require.NoError(t, err)
	require.Equal(t, 1, out.Total)
	require.Equal(t, "record-1", out.Items[0]["id"])
}

func TestCRUDUpdateChecksOwnershipAndFiltersMutableFields(t *testing.T) {
	existing := domain.Record{"id": "record-1", "user_id": "user-1"}
	records := &fakeRecordRepository{
		findByIDFn: func(context.Context, string, string, string) (domain.Record, error) { return existing, nil },
		updateFn: func(_ context.Context, schema, table, id string, columns map[string]any) (domain.Record, error) {
			require.Equal(t, "proj_test", schema)
			require.Equal(t, "Order", table)
			require.Equal(t, "record-1", id)
			require.Equal(t, map[string]any{"title": "updated"}, columns)
			return domain.Record{"id": id, "title": "updated"}, nil
		},
	}
	crud := NewCRUD(records, permission.NewEngine())
	input := UpdateInput{
		ProjectConfig: sampleCRUDConfig(),
		Claims:        &domain.AuthClaims{UserID: "user-1"},
		EntityName:    "Order",
		ID:            "record-1",
		Data: map[string]any{
			"title":   "updated",
			"amount":  99,
			"userId":  "other",
			"user_id": "other",
		},
	}

	got, err := crud.Update(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, "updated", got["title"])

	existing["user_id"] = "user-2"
	_, err = crud.Update(context.Background(), input)
	requireDomainError(t, err, "forbidden", "forbidden")
}

func TestCRUDDeleteChecksExistenceAndPermissionBeforeDeleting(t *testing.T) {
	var existing domain.Record
	deleteCalls := 0
	records := &fakeRecordRepository{
		findByIDFn: func(context.Context, string, string, string) (domain.Record, error) { return existing, nil },
		deleteFn: func(_ context.Context, schema, table, id string) error {
			deleteCalls++
			require.Equal(t, "proj_test", schema)
			require.Equal(t, "Order", table)
			require.Equal(t, "record-1", id)
			return nil
		},
	}
	crud := NewCRUD(records, permission.NewEngine())
	input := DeleteInput{
		ProjectConfig: sampleCRUDConfig(), Claims: &domain.AuthClaims{UserID: "user-1"}, EntityName: "Order", ID: "record-1",
	}

	err := crud.Delete(context.Background(), input)
	requireDomainError(t, err, "not_found", "record not found")
	require.Zero(t, deleteCalls)

	existing = domain.Record{"userId": "user-2"}
	err = crud.Delete(context.Background(), input)
	requireDomainError(t, err, "forbidden", "forbidden")
	require.Zero(t, deleteCalls)

	existing["userId"] = "user-1"
	require.NoError(t, crud.Delete(context.Background(), input))
	require.Equal(t, 1, deleteCalls)
}

func TestCRUDPropagatesRepositoryErrors(t *testing.T) {
	wantErr := errors.New("repository failed")
	records := &fakeRecordRepository{
		findByIDFn: func(context.Context, string, string, string) (domain.Record, error) { return nil, wantErr },
	}
	crud := NewCRUD(records, permission.NewEngine())

	_, err := crud.Get(context.Background(), GetInput{
		ProjectConfig: sampleCRUDConfig(), Claims: &domain.AuthClaims{}, EntityName: "Order", ID: "record-1",
	})
	require.ErrorIs(t, err, wantErr)
}
