package usecase

import (
	"context"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type fakeUserRepository struct {
	createFn      func(context.Context, string, *domain.User) error
	findByEmailFn func(context.Context, string, string) (*domain.User, error)
	findByIDFn    func(context.Context, string, string) (*domain.User, error)
	setRoleFn     func(context.Context, string, string, string) error
}

func (f *fakeUserRepository) Create(ctx context.Context, schema string, user *domain.User) error {
	if f.createFn != nil {
		return f.createFn(ctx, schema, user)
	}
	return nil
}

func (f *fakeUserRepository) FindByEmail(ctx context.Context, schema, email string) (*domain.User, error) {
	if f.findByEmailFn != nil {
		return f.findByEmailFn(ctx, schema, email)
	}
	return nil, nil
}

func (f *fakeUserRepository) FindByID(ctx context.Context, schema, id string) (*domain.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, schema, id)
	}
	return nil, nil
}

func (f *fakeUserRepository) SetRole(ctx context.Context, schema, userID, role string) error {
	if f.setRoleFn != nil {
		return f.setRoleFn(ctx, schema, userID, role)
	}
	return nil
}

type fakePasswordHasher struct {
	hashFn    func(string) (string, error)
	compareFn func(string, string) bool
}

func (f *fakePasswordHasher) Hash(password string) (string, error) {
	if f.hashFn != nil {
		return f.hashFn(password)
	}
	return "hash:" + password, nil
}

func (f *fakePasswordHasher) Compare(hash, password string) bool {
	return f.compareFn != nil && f.compareFn(hash, password)
}

type fakeTokenService struct {
	issueFn        func(string, string, string) (*domain.TokenPair, error)
	parseAccessFn  func(string) (*domain.AuthClaims, error)
	parseRefreshFn func(string) (*domain.AuthClaims, error)
}

func (f *fakeTokenService) Issue(userID, projectID, role string) (*domain.TokenPair, error) {
	if f.issueFn != nil {
		return f.issueFn(userID, projectID, role)
	}
	return &domain.TokenPair{}, nil
}

func (f *fakeTokenService) ParseAccess(token string) (*domain.AuthClaims, error) {
	if f.parseAccessFn != nil {
		return f.parseAccessFn(token)
	}
	return nil, domain.ErrUnauthorized()
}

func (f *fakeTokenService) ParseRefresh(token string) (*domain.AuthClaims, error) {
	if f.parseRefreshFn != nil {
		return f.parseRefreshFn(token)
	}
	return nil, domain.ErrUnauthorized()
}

type fakeConfigClient struct {
	getProjectFn       func(context.Context, string) (*domain.ProjectMeta, error)
	getProjectConfigFn func(context.Context, string) (*domain.ProjectConfig, error)
}

func (f *fakeConfigClient) GetProject(ctx context.Context, id string) (*domain.ProjectMeta, error) {
	if f.getProjectFn != nil {
		return f.getProjectFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeConfigClient) GetProjectConfig(ctx context.Context, id string) (*domain.ProjectConfig, error) {
	if f.getProjectConfigFn != nil {
		return f.getProjectConfigFn(ctx, id)
	}
	return nil, nil
}

type fakeSchemaMigrator struct {
	appliedVersionFn func(context.Context, string) (int32, error)
	ensureSchemaFn   func(context.Context, *domain.ProjectConfig) error
}

func (f *fakeSchemaMigrator) AppliedVersion(ctx context.Context, schema string) (int32, error) {
	if f.appliedVersionFn != nil {
		return f.appliedVersionFn(ctx, schema)
	}
	return 0, nil
}

func (f *fakeSchemaMigrator) EnsureSchema(ctx context.Context, cfg *domain.ProjectConfig) error {
	if f.ensureSchemaFn != nil {
		return f.ensureSchemaFn(ctx, cfg)
	}
	return nil
}

type fakeRecordRepository struct {
	createFn   func(context.Context, string, string, map[string]any) (domain.Record, error)
	findByIDFn func(context.Context, string, string, string) (domain.Record, error)
	listFn     func(context.Context, string, string, string, string, int, int) ([]domain.Record, int, error)
	updateFn   func(context.Context, string, string, string, map[string]any) (domain.Record, error)
	deleteFn   func(context.Context, string, string, string) error
}

func (f *fakeRecordRepository) Create(ctx context.Context, schema, table string, columns map[string]any) (domain.Record, error) {
	return f.createFn(ctx, schema, table, columns)
}

func (f *fakeRecordRepository) FindByID(ctx context.Context, schema, table, id string) (domain.Record, error) {
	return f.findByIDFn(ctx, schema, table, id)
}

func (f *fakeRecordRepository) List(ctx context.Context, schema, table, ownerColumn, ownerID string, limit, offset int) ([]domain.Record, int, error) {
	return f.listFn(ctx, schema, table, ownerColumn, ownerID, limit, offset)
}

func (f *fakeRecordRepository) Update(ctx context.Context, schema, table, id string, columns map[string]any) (domain.Record, error) {
	return f.updateFn(ctx, schema, table, id, columns)
}

func (f *fakeRecordRepository) Delete(ctx context.Context, schema, table, id string) error {
	return f.deleteFn(ctx, schema, table, id)
}
