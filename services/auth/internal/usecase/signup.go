package usecase

import (
	"context"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/port"
)

// UserSignedUpEvent payload cho auth.user.signup.
type UserSignedUpEvent struct {
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id"`
	Email     string `json:"email"`
}

// SignUpInput tham số đăng ký.
type SignUpInput struct {
	ProjectID       string
	Email           string
	Password        string
	FullName        string
	Phone           string
	RefreshDuration string
}

// SignUpOutput kết quả đăng ký thành công.
type SignUpOutput struct {
	User   *domain.User
	Tokens AuthTokens
}

// SignUp usecase đăng ký user mới trong một project.
type SignUp struct {
	users         port.UserRepository
	refreshTokens port.RefreshTokenRepository
	hasher        port.PasswordHasher
	issuer        port.TokenIssuer
	publisher     port.EventPublisher
}

// NewSignUp inject dependencies qua constructor.
func NewSignUp(
	users port.UserRepository,
	refreshTokens port.RefreshTokenRepository,
	hasher port.PasswordHasher,
	issuer port.TokenIssuer,
	publisher port.EventPublisher,
) *SignUp {
	return &SignUp{
		users:         users,
		refreshTokens: refreshTokens,
		hasher:        hasher,
		issuer:        issuer,
		publisher:     publisher,
	}
}

// Execute tạo user, hash password, phát hành token và publish event.
func (uc *SignUp) Execute(ctx context.Context, in SignUpInput) (*SignUpOutput, error) {
	duration, err := domain.ParseRefreshDuration(in.RefreshDuration)
	if err != nil {
		return nil, err
	}
	if err := domain.ValidatePassword(in.Password); err != nil {
		return nil, err
	}

	user, err := domain.NewUser(in.ProjectID, in.Email, in.FullName, in.Phone)
	if err != nil {
		return nil, err
	}

	hash, err := uc.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}
	if err := user.SetPasswordHash(hash); err != nil {
		return nil, err
	}

	if err := uc.users.Create(ctx, in.ProjectID, user); err != nil {
		return nil, err // bao gồm domain.ErrEmailTaken
	}

	tokens, err := issueTokens(ctx, in.ProjectID, user.ID, user.Email, duration, uc.issuer, uc.refreshTokens)
	if err != nil {
		return nil, err
	}

	_ = uc.publisher.Publish(ctx, port.Event{
		Name: "auth.user.signup",
		Payload: UserSignedUpEvent{
			UserID:    user.ID,
			ProjectID: in.ProjectID,
			Email:     user.Email,
		},
	})

	return &SignUpOutput{User: user, Tokens: tokens}, nil
}
