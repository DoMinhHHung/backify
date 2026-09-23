package usecase

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
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
