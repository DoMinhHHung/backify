package port

import (
	"context"

	"backify/services/auth/internal/domain"
)

// UserRepository, RefreshTokenRepository, PasswordResetRepository đều nhận
// projectID làm tham số đầu tiên thay vì gắn cố định vào một project lúc
// khởi tạo — một instance repository duy nhất phục vụ TẤT CẢ project trong
// suốt vòng đời process (multi-tenant), triển khai Postgres tự định tuyến
// tới đúng database auth_proj_<projectID> cho từng lời gọi (xem
// adapter/postgres/conn_manager.go).

// UserRepository lưu trữ User trong database riêng của project.
type UserRepository interface {
	// Create trả ErrEmailTaken nếu email đã tồn tại trong project này (email
	// chỉ unique trong phạm vi project, xem domain.User).
	Create(ctx context.Context, projectID string, user *domain.User) error

	// GetByID trả ErrUserNotFound nếu không có user với id này trong project.
	GetByID(ctx context.Context, projectID, userID string) (*domain.User, error)

	// GetByEmail trả ErrUserNotFound nếu không có user với email này trong
	// project — email đã được domain.NewUser chuẩn hóa lowercase/trim,
	// caller phải tự chuẩn hóa trước khi gọi nếu input đến từ nơi khác.
	GetByEmail(ctx context.Context, projectID, email string) (*domain.User, error)

	// Update ghi lại toàn bộ field có thể đổi (hiện tại chỉ PasswordHash qua
	// SetPasswordHash) — không có UpdateEmail vì đổi email nằm ngoài scope MVP.
	Update(ctx context.Context, projectID string, user *domain.User) error
}

// RefreshTokenRepository lưu trữ RefreshToken trong database riêng của project.
type RefreshTokenRepository interface {
	Create(ctx context.Context, projectID string, token *domain.RefreshToken) error

	// GetByTokenHash trả ErrTokenInvalid nếu không có token với hash này —
	// tra theo hash vì token thật không bao giờ được lưu (domain.HashToken).
	GetByTokenHash(ctx context.Context, projectID, tokenHash string) (*domain.RefreshToken, error)

	// Update ghi lại UsedAt/RevokedAt sau khi usecase gọi MarkUsed/Revoke.
	Update(ctx context.Context, projectID string, token *domain.RefreshToken) error

	// RevokeAllByUser thu hồi mọi refresh token còn sống của user — dùng khi
	// phát hiện reuse (BACKIFY.md §7.9: "revoke toàn bộ refresh token của
	// user", không chỉ family của token bị dùng lại) hoặc khi reset password.
	RevokeAllByUser(ctx context.Context, projectID, userID string) error
}

// PasswordResetRepository lưu trữ PasswordResetToken trong database riêng
// của project.
type PasswordResetRepository interface {
	Create(ctx context.Context, projectID string, token *domain.PasswordResetToken) error

	// GetByTokenHash trả ErrTokenInvalid nếu không có token với hash này.
	GetByTokenHash(ctx context.Context, projectID, tokenHash string) (*domain.PasswordResetToken, error)

	// Update ghi lại UsedAt sau khi usecase gọi MarkUsed.
	Update(ctx context.Context, projectID string, token *domain.PasswordResetToken) error
}
