package usecase

import (
	"context"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/port"
)

// Me trả thông tin user hiện tại (GET /auth/me) — không có trong Bước 6
// gốc nhưng cần để handler không gọi thẳng repository.
type Me struct {
	users port.UserRepository
}

// NewMe inject UserRepository.
func NewMe(users port.UserRepository) *Me {
	return &Me{users: users}
}

// Execute lấy user theo projectID + userID từ context auth.
func (uc *Me) Execute(ctx context.Context, projectID, userID string) (*domain.User, error) {
	return uc.users.GetByID(ctx, projectID, userID)
}
