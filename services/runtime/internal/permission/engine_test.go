package permission_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/permission"
)

func sampleConfig() *domain.ProjectConfig {
	return &domain.ProjectConfig{
		ProjectID:  "p1",
		SchemaName: "proj_test",
		Entities: map[string]domain.Entity{
			"User": {Name: "User", Pool: []domain.Field{
				{Name: "id", Type: "uuid"},
			}},
			"Order": {Name: "Order", Pool: []domain.Field{
				{Name: "id", Type: "uuid"},
				{Name: "title", Type: "string"},
				{Name: "userId", Type: "relation", RelationTo: "User", RelationCardinality: "n-1"},
			}},
		},
	}
}

func TestOwnerColumn(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	col, ok := e.OwnerColumn(cfg, "Order")
	require.True(t, ok)
	require.Equal(t, "userId", col)
}

func TestUserCannotReadOthersOrder(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	claims := &domain.AuthClaims{UserID: "user-a", Role: "user"}
	record := domain.Record{"id": "o1", "userId": "user-b"}

	err := e.CanReadOne(cfg, "Order", claims, record)
	require.Error(t, err)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok)
	require.Equal(t, "forbidden", de.Code)
}

func TestAdminCanReadAll(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	claims := &domain.AuthClaims{UserID: "admin-1", Role: "admin"}
	record := domain.Record{"id": "o1", "userId": "user-b"}

	err := e.CanReadOne(cfg, "Order", claims, record)
	require.NoError(t, err)

	col, ownerID, err := e.CanListFilter(cfg, "Order", claims)
	require.NoError(t, err)
	require.Equal(t, "", col)
	require.Equal(t, "", ownerID)
}

func TestListFilterForNormalUser(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	claims := &domain.AuthClaims{UserID: "user-a", Role: "user"}

	col, ownerID, err := e.CanListFilter(cfg, "Order", claims)
	require.NoError(t, err)
	require.Equal(t, "userId", col)
	require.Equal(t, "user-a", ownerID)
}

func TestCreateAutoAssignsOwner(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	claims := &domain.AuthClaims{UserID: "user-a", Role: "user"}

	out, err := e.CanCreate(cfg, "Order", claims, map[string]any{"title": "x"})
	require.NoError(t, err)
	require.Equal(t, "user-a", out["userId"])
}

func TestUserCannotUpdateOthers(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	claims := &domain.AuthClaims{UserID: "user-a", Role: "user"}
	record := domain.Record{"id": "o1", "userId": "user-b"}

	err := e.CanUpdate(cfg, "Order", claims, record)
	require.Error(t, err)
}
