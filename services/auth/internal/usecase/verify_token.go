package usecase

import (
	"context"

	"backify/services/auth/internal/port"
)

// VerifyTokenInput access token + project kỳ vọng.
type VerifyTokenInput struct {
	Token             string
	ExpectedProjectID string
}

// VerifyTokenOutput Valid=false là kết quả bình thường (không phải lỗi RPC).
// Execute chỉ trả error khi hạ tầng hỏng (ví dụ Redis không kết nối được).
type VerifyTokenOutput struct {
	Valid  bool
	UserID string
	Email  string
	Error  string // mô tả lý do invalid khi Valid=false
}

// VerifyToken usecase kiểm tra chữ ký + hạn + project + blacklist.
type VerifyToken struct {
	verifier  port.TokenVerifier
	blacklist port.TokenBlacklist
}

// NewVerifyToken inject dependencies qua constructor.
func NewVerifyToken(verifier port.TokenVerifier, blacklist port.TokenBlacklist) *VerifyToken {
	return &VerifyToken{
		verifier:  verifier,
		blacklist: blacklist,
	}
}

// Execute thứ tự: verify chữ ký/hạn → khớp project → chưa blacklist.
func (uc *VerifyToken) Execute(ctx context.Context, in VerifyTokenInput) (*VerifyTokenOutput, error) {
	claims, err := uc.verifier.Verify(in.Token)
	if err != nil {
		return &VerifyTokenOutput{Valid: false, Error: "token invalid or expired"}, nil
	}

	if !claims.BelongsToProject(in.ExpectedProjectID) {
		return &VerifyTokenOutput{Valid: false, Error: "token does not belong to project"}, nil
	}

	blacklisted, err := uc.blacklist.IsBlacklisted(ctx, claims.JTI)
	if err != nil {
		// Hạ tầng Redis hỏng — đây mới là error thật.
		return nil, err
	}
	if blacklisted {
		return &VerifyTokenOutput{Valid: false, Error: "token has been revoked"}, nil
	}

	return &VerifyTokenOutput{
		Valid:  true,
		UserID: claims.Sub,
		Email:  claims.Email,
	}, nil
}
