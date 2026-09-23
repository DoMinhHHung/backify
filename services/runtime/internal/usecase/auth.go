package usecase

import (
	"context"
	"strings"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

type Auth struct {
	users  port.UserRepository
	hasher port.PasswordHasher
	tokens port.TokenService
}

func NewAuth(users port.UserRepository, hasher port.PasswordHasher, tokens port.TokenService) *Auth {
	return &Auth{users: users, hasher: hasher, tokens: tokens}
}

type SignupInput struct {
	ProjectConfig *domain.ProjectConfig
	Fields        map[string]string
}

type AuthOutput struct {
	User   *domain.User
	Tokens *domain.TokenPair
}

func (uc *Auth) Signup(ctx context.Context, in SignupInput) (*AuthOutput, error) {
	cfg := in.ProjectConfig
	enabled := enabledFields(cfg, "signup")
	if len(enabled) == 0 {
		return nil, domain.ErrValidation("signup is not enabled")
	}

	email := strings.TrimSpace(strings.ToLower(in.Fields["email"]))
	password := in.Fields["password"]
	if email == "" || password == "" {
		return nil, domain.ErrValidation("email and password are required")
	}
	if !contains(enabled, "email") || !contains(enabled, "password") {
		return nil, domain.ErrValidation("email and password must be enabled for signup")
	}
	if len(password) < 8 {
		return nil, domain.ErrValidation("password must be at least 8 characters")
	}

	existing, err := uc.users.FindByEmail(ctx, cfg.SchemaName, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrEmailTaken()
	}

	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: hash,
		Role:         "user",
	}
	if contains(enabled, "fullName") {
		user.FullName = strings.TrimSpace(in.Fields["fullName"])
	}
	if contains(enabled, "phone") {
		user.Phone = strings.TrimSpace(in.Fields["phone"])
	}

	if err := uc.users.Create(ctx, cfg.SchemaName, user); err != nil {
		return nil, err
	}

	tokens, err := uc.tokens.Issue(user.ID, cfg.ProjectID, user.Role)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return &AuthOutput{User: user, Tokens: tokens}, nil
}

type SigninInput struct {
	ProjectConfig *domain.ProjectConfig
	Email         string
	Password      string
}

func (uc *Auth) Signin(ctx context.Context, in SigninInput) (*AuthOutput, error) {
	cfg := in.ProjectConfig
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if email == "" || in.Password == "" {
		return nil, domain.ErrValidation("email and password are required")
	}

	user, err := uc.users.FindByEmail(ctx, cfg.SchemaName, email)
	if err != nil {
		return nil, err
	}
	if user == nil || !uc.hasher.Compare(user.PasswordHash, in.Password) {
		return nil, domain.ErrInvalidCredentials()
	}

	tokens, err := uc.tokens.Issue(user.ID, cfg.ProjectID, user.Role)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return &AuthOutput{User: user, Tokens: tokens}, nil
}

type RefreshInput struct {
	ProjectConfig *domain.ProjectConfig
	RefreshToken  string
}

func (uc *Auth) Refresh(ctx context.Context, in RefreshInput) (*domain.TokenPair, error) {
	claims, err := uc.tokens.ParseRefresh(in.RefreshToken)
	if err != nil {
		return nil, err
	}
	if claims.ProjectID != in.ProjectConfig.ProjectID {
		return nil, domain.ErrUnauthorized()
	}

	user, err := uc.users.FindByID(ctx, in.ProjectConfig.SchemaName, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUnauthorized()
	}

	return uc.tokens.Issue(user.ID, in.ProjectConfig.ProjectID, user.Role)
}

type MeInput struct {
	ProjectConfig *domain.ProjectConfig
	UserID        string
}

func (uc *Auth) Me(ctx context.Context, in MeInput) (*domain.User, error) {
	user, err := uc.users.FindByID(ctx, in.ProjectConfig.SchemaName, in.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrNotFound("user not found")
	}
	user.PasswordHash = ""
	return user, nil
}

func enabledFields(cfg *domain.ProjectConfig, fn string) []string {
	mod, ok := cfg.Modules["auth"]
	if !ok || !mod.Enabled {
		return nil
	}
	fc, ok := mod.Functions[fn]
	if !ok {
		return nil
	}
	return fc.EnabledFields
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
