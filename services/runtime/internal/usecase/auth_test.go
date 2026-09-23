package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

func authConfig() *domain.ProjectConfig {
	return &domain.ProjectConfig{
		ProjectID:  "project-1",
		SchemaName: "project_schema",
		Modules: map[string]domain.ModuleConfig{
			"auth": {
				Enabled: true,
				Functions: map[string]domain.FunctionConfig{
					"signup": {EnabledFields: []string{"email", "password", "fullName", "phone"}},
				},
			},
		},
	}
}

func requireDomainError(t *testing.T, err error, code, message string) {
	t.Helper()
	var domainErr *domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	require.Equal(t, code, domainErr.Code)
	require.Equal(t, message, domainErr.Message)
}

func TestAuthSignupCreatesNormalizedUserAndIssuesTokens(t *testing.T) {
	var created domain.User
	users := &fakeUserRepository{
		findByEmailFn: func(_ context.Context, schema, email string) (*domain.User, error) {
			require.Equal(t, "project_schema", schema)
			require.Equal(t, "person@example.com", email)
			return nil, nil
		},
		createFn: func(_ context.Context, schema string, user *domain.User) error {
			require.Equal(t, "project_schema", schema)
			user.ID = "user-1"
			created = *user
			return nil
		},
	}
	hasher := &fakePasswordHasher{hashFn: func(password string) (string, error) {
		require.Equal(t, "password123", password)
		return "stored-hash", nil
	}}
	wantTokens := &domain.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 900}
	tokens := &fakeTokenService{issueFn: func(userID, projectID, role string) (*domain.TokenPair, error) {
		require.Equal(t, "user-1", userID)
		require.Equal(t, "project-1", projectID)
		require.Equal(t, "user", role)
		return wantTokens, nil
	}}

	out, err := NewAuth(users, hasher, tokens).Signup(context.Background(), SignupInput{
		ProjectConfig: authConfig(),
		Fields: map[string]string{
			"email":    "  PERSON@Example.COM ",
			"password": "password123",
			"fullName": "  Ada Lovelace  ",
			"phone":    "  0900  ",
		},
	})

	require.NoError(t, err)
	require.Equal(t, "person@example.com", created.Email)
	require.Equal(t, "stored-hash", created.PasswordHash)
	require.Equal(t, "Ada Lovelace", created.FullName)
	require.Equal(t, "0900", created.Phone)
	require.Equal(t, "user", created.Role)
	require.Empty(t, out.User.PasswordHash)
	require.Same(t, wantTokens, out.Tokens)
}

func TestAuthSignupValidationStopsBeforeRepository(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*domain.ProjectConfig)
		fields  map[string]string
		message string
	}{
		{
			name: "signup disabled",
			mutate: func(cfg *domain.ProjectConfig) {
				mod := cfg.Modules["auth"]
				mod.Enabled = false
				cfg.Modules["auth"] = mod
			},
			fields:  map[string]string{"email": "a@example.com", "password": "password"},
			message: "signup is not enabled",
		},
		{
			name:    "required value missing",
			mutate:  func(*domain.ProjectConfig) {},
			fields:  map[string]string{"email": "a@example.com"},
			message: "email and password are required",
		},
		{
			name: "required field disabled",
			mutate: func(cfg *domain.ProjectConfig) {
				mod := cfg.Modules["auth"]
				mod.Functions["signup"] = domain.FunctionConfig{EnabledFields: []string{"email"}}
				cfg.Modules["auth"] = mod
			},
			fields:  map[string]string{"email": "a@example.com", "password": "password"},
			message: "email and password must be enabled for signup",
		},
		{
			name:    "short password",
			mutate:  func(*domain.ProjectConfig) {},
			fields:  map[string]string{"email": "a@example.com", "password": "1234567"},
			message: "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := authConfig()
			tt.mutate(cfg)
			called := false
			users := &fakeUserRepository{findByEmailFn: func(context.Context, string, string) (*domain.User, error) {
				called = true
				return nil, nil
			}}

			_, err := NewAuth(users, &fakePasswordHasher{}, &fakeTokenService{}).Signup(context.Background(), SignupInput{
				ProjectConfig: cfg,
				Fields:        tt.fields,
			})

			requireDomainError(t, err, "validation_error", tt.message)
			require.False(t, called)
		})
	}
}

func TestAuthSignupRejectsExistingEmail(t *testing.T) {
	users := &fakeUserRepository{findByEmailFn: func(context.Context, string, string) (*domain.User, error) {
		return &domain.User{ID: "existing"}, nil
	}}

	_, err := NewAuth(users, &fakePasswordHasher{}, &fakeTokenService{}).Signup(context.Background(), SignupInput{
		ProjectConfig: authConfig(),
		Fields:        map[string]string{"email": "a@example.com", "password": "password"},
	})

	requireDomainError(t, err, "email_taken", "email already registered")
}

func TestAuthSignupPropagatesDependencyErrors(t *testing.T) {
	wantErr := errors.New("dependency failed")
	tests := []struct {
		name   string
		users  *fakeUserRepository
		hasher *fakePasswordHasher
		tokens *fakeTokenService
	}{
		{
			name: "lookup",
			users: &fakeUserRepository{findByEmailFn: func(context.Context, string, string) (*domain.User, error) {
				return nil, wantErr
			}},
			hasher: &fakePasswordHasher{}, tokens: &fakeTokenService{},
		},
		{
			name:  "hash",
			users: &fakeUserRepository{},
			hasher: &fakePasswordHasher{hashFn: func(string) (string, error) {
				return "", wantErr
			}},
			tokens: &fakeTokenService{},
		},
		{
			name: "create",
			users: &fakeUserRepository{createFn: func(context.Context, string, *domain.User) error {
				return wantErr
			}},
			hasher: &fakePasswordHasher{}, tokens: &fakeTokenService{},
		},
		{
			name: "issue tokens",
			users: &fakeUserRepository{createFn: func(_ context.Context, _ string, user *domain.User) error {
				user.ID = "user-1"
				return nil
			}},
			hasher: &fakePasswordHasher{},
			tokens: &fakeTokenService{issueFn: func(string, string, string) (*domain.TokenPair, error) {
				return nil, wantErr
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAuth(tt.users, tt.hasher, tt.tokens).Signup(context.Background(), SignupInput{
				ProjectConfig: authConfig(),
				Fields:        map[string]string{"email": "a@example.com", "password": "password"},
			})
			require.ErrorIs(t, err, wantErr)
		})
	}
}

func TestAuthSigninNormalizesEmailAndClearsPassword(t *testing.T) {
	users := &fakeUserRepository{findByEmailFn: func(_ context.Context, schema, email string) (*domain.User, error) {
		require.Equal(t, "project_schema", schema)
		require.Equal(t, "person@example.com", email)
		return &domain.User{ID: "user-1", Email: email, PasswordHash: "stored", Role: "admin"}, nil
	}}
	hasher := &fakePasswordHasher{compareFn: func(hash, password string) bool {
		require.Equal(t, "stored", hash)
		require.Equal(t, "password", password)
		return true
	}}
	tokens := &fakeTokenService{issueFn: func(userID, projectID, role string) (*domain.TokenPair, error) {
		require.Equal(t, "user-1", userID)
		require.Equal(t, "project-1", projectID)
		require.Equal(t, "admin", role)
		return &domain.TokenPair{AccessToken: "access"}, nil
	}}

	out, err := NewAuth(users, hasher, tokens).Signin(context.Background(), SigninInput{
		ProjectConfig: authConfig(), Email: " PERSON@EXAMPLE.COM ", Password: "password",
	})

	require.NoError(t, err)
	require.Empty(t, out.User.PasswordHash)
	require.Equal(t, "access", out.Tokens.AccessToken)
}

func TestAuthSigninRejectsMissingOrMismatchedCredentials(t *testing.T) {
	uc := NewAuth(&fakeUserRepository{}, &fakePasswordHasher{}, &fakeTokenService{})

	_, err := uc.Signin(context.Background(), SigninInput{ProjectConfig: authConfig(), Email: "", Password: "password"})
	requireDomainError(t, err, "validation_error", "email and password are required")

	_, err = uc.Signin(context.Background(), SigninInput{ProjectConfig: authConfig(), Email: "missing@example.com", Password: "password"})
	requireDomainError(t, err, "invalid_credentials", "invalid email or password")

	users := &fakeUserRepository{findByEmailFn: func(context.Context, string, string) (*domain.User, error) {
		return &domain.User{PasswordHash: "hash"}, nil
	}}
	_, err = NewAuth(users, &fakePasswordHasher{compareFn: func(string, string) bool { return false }}, &fakeTokenService{}).Signin(
		context.Background(), SigninInput{ProjectConfig: authConfig(), Email: "a@example.com", Password: "wrong"},
	)
	requireDomainError(t, err, "invalid_credentials", "invalid email or password")
}

func TestAuthRefreshUsesCurrentUserRole(t *testing.T) {
	users := &fakeUserRepository{findByIDFn: func(_ context.Context, schema, id string) (*domain.User, error) {
		require.Equal(t, "project_schema", schema)
		require.Equal(t, "user-1", id)
		return &domain.User{ID: id, Role: "admin"}, nil
	}}
	tokens := &fakeTokenService{
		parseRefreshFn: func(token string) (*domain.AuthClaims, error) {
			require.Equal(t, "refresh-token", token)
			return &domain.AuthClaims{UserID: "user-1", ProjectID: "project-1", Role: "user"}, nil
		},
		issueFn: func(userID, projectID, role string) (*domain.TokenPair, error) {
			require.Equal(t, "admin", role)
			return &domain.TokenPair{AccessToken: "new-access"}, nil
		},
	}

	out, err := NewAuth(users, &fakePasswordHasher{}, tokens).Refresh(context.Background(), RefreshInput{
		ProjectConfig: authConfig(), RefreshToken: "refresh-token",
	})

	require.NoError(t, err)
	require.Equal(t, "new-access", out.AccessToken)
}

func TestAuthRefreshRejectsCrossProjectAndMissingUser(t *testing.T) {
	lookupCalled := false
	users := &fakeUserRepository{findByIDFn: func(context.Context, string, string) (*domain.User, error) {
		lookupCalled = true
		return nil, nil
	}}
	tokens := &fakeTokenService{parseRefreshFn: func(string) (*domain.AuthClaims, error) {
		return &domain.AuthClaims{UserID: "user-1", ProjectID: "other-project"}, nil
	}}

	_, err := NewAuth(users, &fakePasswordHasher{}, tokens).Refresh(context.Background(), RefreshInput{
		ProjectConfig: authConfig(), RefreshToken: "refresh",
	})
	requireDomainError(t, err, "unauthorized", "unauthorized")
	require.False(t, lookupCalled)

	tokens.parseRefreshFn = func(string) (*domain.AuthClaims, error) {
		return &domain.AuthClaims{UserID: "missing", ProjectID: "project-1"}, nil
	}
	_, err = NewAuth(users, &fakePasswordHasher{}, tokens).Refresh(context.Background(), RefreshInput{
		ProjectConfig: authConfig(), RefreshToken: "refresh",
	})
	requireDomainError(t, err, "unauthorized", "unauthorized")
}

func TestAuthMeReturnsPublicUserOrNotFound(t *testing.T) {
	users := &fakeUserRepository{findByIDFn: func(_ context.Context, schema, id string) (*domain.User, error) {
		require.Equal(t, "project_schema", schema)
		if id == "missing" {
			return nil, nil
		}
		return &domain.User{ID: id, PasswordHash: "secret"}, nil
	}}
	uc := NewAuth(users, &fakePasswordHasher{}, &fakeTokenService{})

	user, err := uc.Me(context.Background(), MeInput{ProjectConfig: authConfig(), UserID: "user-1"})
	require.NoError(t, err)
	require.Empty(t, user.PasswordHash)

	_, err = uc.Me(context.Background(), MeInput{ProjectConfig: authConfig(), UserID: "missing"})
	requireDomainError(t, err, "not_found", "user not found")
}
