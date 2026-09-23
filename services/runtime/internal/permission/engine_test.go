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

func TestOwnerColumnSupportsDocumentedCardinalityAndCaseInsensitiveUser(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	order := cfg.Entities["Order"]
	order.Pool[2].RelationTo = "uSeR"
	order.Pool[2].RelationCardinality = "many_to_one"
	cfg.Entities["Order"] = order

	col, ok := e.OwnerColumn(cfg, "Order")

	require.True(t, ok)
	require.Equal(t, "userId", col)
}

func TestOwnerColumnIgnoresNonOwningRelations(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	order := cfg.Entities["Order"]
	order.Pool = []domain.Field{
		{Name: "customerId", Type: "relation", RelationTo: "Customer", RelationCardinality: "n-1"},
		{Name: "userIds", Type: "relation", RelationTo: "User", RelationCardinality: "1-n"},
	}
	cfg.Entities["Order"] = order

	_, ok := e.OwnerColumn(cfg, "Order")

	require.False(t, ok)
}

func TestPermissionOperationsRejectMissingClaims(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	record := domain.Record{"userId": "user-a"}

	_, err := e.CanCreate(cfg, "Order", nil, map[string]any{})
	requireDomainCode(t, err, "unauthorized")
	requireDomainCode(t, e.CanReadOne(cfg, "Order", nil, record), "unauthorized")
	_, _, err = e.CanListFilter(cfg, "Order", nil)
	requireDomainCode(t, err, "unauthorized")
	requireDomainCode(t, e.CanUpdate(cfg, "Order", nil, record), "unauthorized")
	requireDomainCode(t, e.CanDelete(cfg, "Order", nil, record), "unauthorized")
}

func TestCreateOwnerAssignmentDoesNotMutateInput(t *testing.T) {
	e := permission.NewEngine()
	input := map[string]any{"title": "order", "userId": "spoofed"}

	out, err := e.CanCreate(sampleConfig(), "Order", &domain.AuthClaims{UserID: "user-a"}, input)

	require.NoError(t, err)
	require.Equal(t, "spoofed", input["userId"])
	require.Equal(t, "user-a", out["userId"])
	out["title"] = "changed"
	require.Equal(t, "order", input["title"])
}

func TestReadAndUpdateAcceptSnakeCaseOwnerColumn(t *testing.T) {
	e := permission.NewEngine()
	claims := &domain.AuthClaims{UserID: "user-a", Role: "user"}
	record := domain.Record{"user_id": "user-a"}

	require.NoError(t, e.CanReadOne(sampleConfig(), "Order", claims, record))
	require.NoError(t, e.CanUpdate(sampleConfig(), "Order", claims, record))
	require.NoError(t, e.CanDelete(sampleConfig(), "Order", claims, record))
}

func TestEntityWithoutOwnerRelationIsAccessibleToAuthenticatedUser(t *testing.T) {
	e := permission.NewEngine()
	cfg := sampleConfig()
	claims := &domain.AuthClaims{UserID: "user-a", Role: "user"}

	require.NoError(t, e.CanReadOne(cfg, "User", claims, domain.Record{"id": "user-b"}))
	require.NoError(t, e.CanUpdate(cfg, "User", claims, domain.Record{"id": "user-b"}))
	col, id, err := e.CanListFilter(cfg, "User", claims)
	require.NoError(t, err)
	require.Empty(t, col)
	require.Empty(t, id)
}

func requireDomainCode(t *testing.T, err error, code string) {
	t.Helper()
	var domainErr *domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	require.Equal(t, code, domainErr.Code)
}
