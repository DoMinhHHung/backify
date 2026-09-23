package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

func TestPromoteUserAcceptsSupportedRoles(t *testing.T) {
	for _, role := range []string{"admin", "user"} {
		t.Run(role, func(t *testing.T) {
			client := &fakeConfigClient{getProjectConfigFn: func(_ context.Context, projectID string) (*domain.ProjectConfig, error) {
				require.Equal(t, "project-1", projectID)
				return &domain.ProjectConfig{SchemaName: "project_schema"}, nil
			}}
			users := &fakeUserRepository{setRoleFn: func(_ context.Context, schema, userID, gotRole string) error {
				require.Equal(t, "project_schema", schema)
				require.Equal(t, "user-1", userID)
				require.Equal(t, role, gotRole)
				return nil
			}}

			err := NewPromoteUser(users, client).Execute(context.Background(), PromoteUserInput{
				ProjectID: "project-1", UserID: "user-1", Role: role,
			})

			require.NoError(t, err)
		})
	}
}

func TestPromoteUserRejectsInvalidRoleWithoutFetchingConfig(t *testing.T) {
	called := false
	client := &fakeConfigClient{getProjectConfigFn: func(context.Context, string) (*domain.ProjectConfig, error) {
		called = true
		return nil, nil
	}}

	err := NewPromoteUser(&fakeUserRepository{}, client).Execute(context.Background(), PromoteUserInput{Role: "owner"})

	requireDomainError(t, err, "validation_error", "role must be admin or user")
	require.False(t, called)
}

func TestPromoteUserPropagatesConfigAndRepositoryErrors(t *testing.T) {
	wantErr := errors.New("dependency failed")
	err := NewPromoteUser(&fakeUserRepository{}, &fakeConfigClient{
		getProjectConfigFn: func(context.Context, string) (*domain.ProjectConfig, error) { return nil, wantErr },
	}).Execute(context.Background(), PromoteUserInput{Role: "admin"})
	require.ErrorIs(t, err, wantErr)

	err = NewPromoteUser(&fakeUserRepository{
		setRoleFn: func(context.Context, string, string, string) error { return wantErr },
	}, &fakeConfigClient{
		getProjectConfigFn: func(context.Context, string) (*domain.ProjectConfig, error) {
			return &domain.ProjectConfig{SchemaName: "schema"}, nil
		},
	}).Execute(context.Background(), PromoteUserInput{Role: "user"})
	require.ErrorIs(t, err, wantErr)
}
