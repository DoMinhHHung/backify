package usecase

import (
	"context"
	"sync"
	"time"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/port"
)

// --- UserRepository mock ---

type mockUserRepo struct {
	mu    sync.Mutex
	users map[string]*domain.User // key = projectID + ":" + id
	byEmail map[string]string     // key = projectID + ":" + email → userID
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:   make(map[string]*domain.User),
		byEmail: make(map[string]string),
	}
}

func (m *mockUserRepo) key(projectID, id string) string {
	return projectID + ":" + id
}

func (m *mockUserRepo) emailKey(projectID, email string) string {
	return projectID + ":" + email
}

func (m *mockUserRepo) Create(ctx context.Context, projectID string, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ek := m.emailKey(projectID, user.Email)
	if _, ok := m.byEmail[ek]; ok {
		return domain.ErrEmailTaken
	}
	cp := *user
	m.users[m.key(projectID, user.ID)] = &cp
	m.byEmail[ek] = user.ID
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, projectID, userID string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[m.key(projectID, userID)]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, projectID, email string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byEmail[m.emailKey(projectID, email)]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	u := m.users[m.key(projectID, id)]
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) Update(ctx context.Context, projectID string, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.key(projectID, user.ID)
	if _, ok := m.users[k]; !ok {
		return domain.ErrUserNotFound
	}
	cp := *user
	m.users[k] = &cp
	return nil
}

// --- RefreshTokenRepository mock ---

type mockRefreshTokenRepo struct {
	mu     sync.Mutex
	tokens map[string]*domain.RefreshToken // key = projectID + ":" + tokenHash
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepo {
	return &mockRefreshTokenRepo{tokens: make(map[string]*domain.RefreshToken)}
}

func (m *mockRefreshTokenRepo) hashKey(projectID, hash string) string {
	return projectID + ":" + hash
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, projectID string, token *domain.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *token
	m.tokens[m.hashKey(projectID, token.TokenHash)] = &cp
	return nil
}

func (m *mockRefreshTokenRepo) GetByTokenHash(ctx context.Context, projectID, tokenHash string) (*domain.RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[m.hashKey(projectID, tokenHash)]
	if !ok {
		return nil, domain.ErrTokenInvalid
	}
	cp := *t
	return &cp, nil
}

func (m *mockRefreshTokenRepo) Update(ctx context.Context, projectID string, token *domain.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.hashKey(projectID, token.TokenHash)
	if _, ok := m.tokens[k]; !ok {
		return domain.ErrTokenInvalid
	}
	cp := *token
	m.tokens[k] = &cp
	return nil
}

func (m *mockRefreshTokenRepo) RevokeAllByUser(ctx context.Context, projectID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	for _, t := range m.tokens {
		if t.ProjectID == projectID && t.UserID == userID && t.RevokedAt == nil {
			t.Revoke(now)
		}
	}
	return nil
}

// --- PasswordResetRepository mock ---

type mockPasswordResetRepo struct {
	mu     sync.Mutex
	tokens map[string]*domain.PasswordResetToken
}

func newMockPasswordResetRepo() *mockPasswordResetRepo {
	return &mockPasswordResetRepo{tokens: make(map[string]*domain.PasswordResetToken)}
}

func (m *mockPasswordResetRepo) hashKey(projectID, hash string) string {
	return projectID + ":" + hash
}

func (m *mockPasswordResetRepo) Create(ctx context.Context, projectID string, token *domain.PasswordResetToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *token
	m.tokens[m.hashKey(projectID, token.TokenHash)] = &cp
	return nil
}

func (m *mockPasswordResetRepo) GetByTokenHash(ctx context.Context, projectID, tokenHash string) (*domain.PasswordResetToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[m.hashKey(projectID, tokenHash)]
	if !ok {
		return nil, domain.ErrTokenInvalid
	}
	cp := *t
	return &cp, nil
}

func (m *mockPasswordResetRepo) Update(ctx context.Context, projectID string, token *domain.PasswordResetToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := m.hashKey(projectID, token.TokenHash)
	if _, ok := m.tokens[k]; !ok {
		return domain.ErrTokenInvalid
	}
	cp := *token
	m.tokens[k] = &cp
	return nil
}

// --- PasswordHasher mock (không chạy argon2id thật) ---

type mockHasher struct {
	hashFn   func(string) (string, error)
	verifyFn func(password, hash string) (bool, error)
}

func newMockHasher() *mockHasher {
	return &mockHasher{
		hashFn: func(p string) (string, error) {
			return "hashed:" + p, nil
		},
		verifyFn: func(password, hash string) (bool, error) {
			return hash == "hashed:"+password, nil
		},
	}
}

func (m *mockHasher) Hash(password string) (string, error) {
	return m.hashFn(password)
}

func (m *mockHasher) Verify(password, hash string) (bool, error) {
	return m.verifyFn(password, hash)
}

// --- TokenIssuer mock ---

type mockIssuer struct {
	issueFn func(*domain.AccessClaims) (string, error)
}

func newMockIssuer() *mockIssuer {
	return &mockIssuer{
		issueFn: func(c *domain.AccessClaims) (string, error) {
			return "access:" + c.Sub + ":" + c.JTI, nil
		},
	}
}

func (m *mockIssuer) Issue(claims *domain.AccessClaims) (string, error) {
	return m.issueFn(claims)
}

// --- TokenVerifier mock ---

type mockVerifier struct {
	verifyFn func(string) (*domain.AccessClaims, error)
}

func newMockVerifier() *mockVerifier {
	return &mockVerifier{
		verifyFn: func(token string) (*domain.AccessClaims, error) {
			return nil, domain.ErrTokenInvalid
		},
	}
}

func (m *mockVerifier) Verify(token string) (*domain.AccessClaims, error) {
	return m.verifyFn(token)
}

// --- TokenBlacklist mock ---

type mockBlacklist struct {
	mu    sync.Mutex
	items map[string]time.Time // jti → expiresAt
}

func newMockBlacklist() *mockBlacklist {
	return &mockBlacklist{items: make(map[string]time.Time)}
}

func (m *mockBlacklist) Add(ctx context.Context, jti string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ttl <= 0 {
		return nil
	}
	m.items[jti] = time.Now().UTC().Add(ttl)
	return nil
}

func (m *mockBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	exp, ok := m.items[jti]
	if !ok {
		return false, nil
	}
	if time.Now().UTC().After(exp) {
		delete(m.items, jti)
		return false, nil
	}
	return true, nil
}

// --- EventPublisher mock ---

type mockEventPublisher struct {
	mu     sync.Mutex
	events []port.Event
}

func newMockEventPublisher() *mockEventPublisher {
	return &mockEventPublisher{}
}

func (m *mockEventPublisher) Publish(ctx context.Context, event port.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventPublisher) lastEventName() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.events) == 0 {
		return ""
	}
	return m.events[len(m.events)-1].Name
}

func (m *mockEventPublisher) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}
