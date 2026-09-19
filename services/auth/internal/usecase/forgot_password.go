package usecase

import (
	"context"
	"errors"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/port"
)

// PasswordResetRequestedEvent payload cho auth.password_reset_requested.
// Token thô đi qua RabbitMQ để service khác (chưa build) gửi email.
type PasswordResetRequestedEvent struct {
	UserID     string `json:"user_id"`
	ProjectID  string `json:"project_id"`
	ResetToken string `json:"reset_token"`
}

// ForgotPasswordInput tham số quên mật khẩu.
type ForgotPasswordInput struct {
	ProjectID string
	Email     string
}

// ForgotPassword usecase tạo password-reset token nếu email tồn tại.
// User enumeration: ErrUserNotFound → trả nil (thành công giả).
type ForgotPassword struct {
	users          port.UserRepository
	passwordResets port.PasswordResetRepository
	publisher      port.EventPublisher
}

// NewForgotPassword inject dependencies qua constructor.
func NewForgotPassword(
	users port.UserRepository,
	passwordResets port.PasswordResetRepository,
	publisher port.EventPublisher,
) *ForgotPassword {
	return &ForgotPassword{
		users:          users,
		passwordResets: passwordResets,
		publisher:      publisher,
	}
}

// Execute tạo reset token và publish event. Email không tồn tại → nil, nil.
func (uc *ForgotPassword) Execute(ctx context.Context, in ForgotPasswordInput) error {
	user, err := uc.users.GetByEmail(ctx, in.ProjectID, in.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || isDomainCode(err, domain.CodeUserNotFound) {
			return nil // thành công giả — chống user enumeration
		}
		return err
	}

	token, raw, err := domain.NewPasswordResetToken(in.ProjectID, user.ID)
	if err != nil {
		return err
	}
	if err := uc.passwordResets.Create(ctx, in.ProjectID, token); err != nil {
		return err
	}

	_ = uc.publisher.Publish(ctx, port.Event{
		Name: "auth.password_reset_requested",
		Payload: PasswordResetRequestedEvent{
			UserID:     user.ID,
			ProjectID:  in.ProjectID,
			ResetToken: raw,
		},
	})

	return nil
}
