package usecase

import (
	"context"
	"time"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/port"
)

// ResetPasswordInput tham số đặt lại mật khẩu bằng reset token.
type ResetPasswordInput struct {
	ProjectID   string
	ResetToken  string
	NewPassword string
}

// ResetPassword usecase đổi password, mark reset token used, revoke mọi refresh token.
type ResetPassword struct {
	users          port.UserRepository
	passwordResets port.PasswordResetRepository
	refreshTokens  port.RefreshTokenRepository
	hasher         port.PasswordHasher
}

// NewResetPassword inject dependencies qua constructor.
func NewResetPassword(
	users port.UserRepository,
	passwordResets port.PasswordResetRepository,
	refreshTokens port.RefreshTokenRepository,
	hasher port.PasswordHasher,
) *ResetPassword {
	return &ResetPassword{
		users:          users,
		passwordResets: passwordResets,
		refreshTokens:  refreshTokens,
		hasher:         hasher,
	}
}

// Execute redeem reset token, đổi password, thu hồi mọi phiên cũ.
func (uc *ResetPassword) Execute(ctx context.Context, in ResetPasswordInput) error {
	now := time.Now().UTC()

	hash := domain.HashToken(in.ResetToken)
	token, err := uc.passwordResets.GetByTokenHash(ctx, in.ProjectID, hash)
	if err != nil {
		return err
	}
	if err := token.CanRedeem(now); err != nil {
		return err
	}

	if err := domain.ValidatePassword(in.NewPassword); err != nil {
		return err
	}

	user, err := uc.users.GetByID(ctx, in.ProjectID, token.UserID)
	if err != nil {
		return err
	}

	pwHash, err := uc.hasher.Hash(in.NewPassword)
	if err != nil {
		return err
	}
	if err := user.SetPasswordHash(pwHash); err != nil {
		return err
	}
	if err := uc.users.Update(ctx, in.ProjectID, user); err != nil {
		return err
	}

	token.MarkUsed(now)
	if err := uc.passwordResets.Update(ctx, in.ProjectID, token); err != nil {
		return err
	}

	// Đổi password xong → mọi phiên cũ (kể cả thiết bị khác) bị đăng xuất.
	if err := uc.refreshTokens.RevokeAllByUser(ctx, in.ProjectID, user.ID); err != nil {
		return err
	}

	return nil
}
