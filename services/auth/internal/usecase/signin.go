package usecase

import (
	"context"
	"errors"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/port"
)

// UserSignedInEvent payload cho auth.user.signin.
type UserSignedInEvent struct {
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id"`
	Email     string `json:"email"`
}

// SignInInput tham số đăng nhập.
type SignInInput struct {
	ProjectID       string
	Email           string
	Password        string
	RefreshDuration string
}

// SignInOutput kết quả đăng nhập thành công.
type SignInOutput struct {
	User   *domain.User
	Tokens AuthTokens
}

// SignIn usecase đăng nhập bằng email + password.
type SignIn struct {
	users         port.UserRepository
	refreshTokens port.RefreshTokenRepository
	hasher        port.PasswordHasher
	issuer        port.TokenIssuer
	publisher     port.EventPublisher
}

// NewSignIn inject dependencies qua constructor.
func NewSignIn(
	users port.UserRepository,
	refreshTokens port.RefreshTokenRepository,
	hasher port.PasswordHasher,
	issuer port.TokenIssuer,
	publisher port.EventPublisher,
) *SignIn {
	return &SignIn{
		users:         users,
		refreshTokens: refreshTokens,
		hasher:        hasher,
		issuer:        issuer,
		publisher:     publisher,
	}
}

// Execute xác thực credentials rồi phát hành token. User enumeration:
// ErrUserNotFound được map thành ErrInvalidCredentials.
func (uc *SignIn) Execute(ctx context.Context, in SignInInput) (*SignInOutput, error) {
	duration, err := domain.ParseRefreshDuration(in.RefreshDuration)
	if err != nil {
		return nil, err
	}

	user, err := uc.users.GetByEmail(ctx, in.ProjectID, in.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || isDomainCode(err, domain.CodeUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	ok, err := uc.hasher.Verify(in.Password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrInvalidCredentials
	}

	tokens, err := issueTokens(ctx, in.ProjectID, user.ID, user.Email, duration, uc.issuer, uc.refreshTokens)
	if err != nil {
		return nil, err
	}

	_ = uc.publisher.Publish(ctx, port.Event{
		Name: "auth.user.signin",
		Payload: UserSignedInEvent{
			UserID:    user.ID,
			ProjectID: in.ProjectID,
			Email:     user.Email,
		},
	})

	return &SignInOutput{User: user, Tokens: tokens}, nil
}

func isDomainCode(err error, code domain.ErrorCode) bool {
	var de *domain.Error
	if errors.As(err, &de) {
		return de.Code == code
	}
	return false
}
