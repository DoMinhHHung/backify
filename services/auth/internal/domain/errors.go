package domain

import "fmt"

// ErrorCode là mã lỗi ổn định để handler/gRPC map sang HTTP status hoặc
// gRPC status code mà không phải so sánh chuỗi message (message có thể đổi).
type ErrorCode string

const (
	CodeInvalidInput           ErrorCode = "INVALID_INPUT"
	CodeUserNotFound           ErrorCode = "USER_NOT_FOUND"
	CodeEmailTaken             ErrorCode = "EMAIL_TAKEN"
	CodeInvalidCredentials     ErrorCode = "INVALID_CREDENTIALS"
	CodeTokenInvalid           ErrorCode = "TOKEN_INVALID"
	CodeTokenExpired           ErrorCode = "TOKEN_EXPIRED"
	CodeTokenRevoked           ErrorCode = "TOKEN_REVOKED"
	CodeTokenReuseDetected     ErrorCode = "TOKEN_REUSE_DETECTED"
	CodeProjectNotFound        ErrorCode = "PROJECT_NOT_FOUND"
	CodeProjectDatabaseMissing ErrorCode = "PROJECT_DATABASE_MISSING"
)

// Error là lỗi domain có mã ổn định (Code) và Details tùy chọn cho ngữ cảnh
// gỡ lỗi (ví dụ field nào sai) mà không làm đổi Code dùng để so sánh bằng ==.
type Error struct {
	Code    ErrorCode
	Message string
	Details map[string]interface{}
}

// Error hiện thực interface error tiêu chuẩn của Go.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewError tạo lỗi domain mới không kèm Details.
func NewError(code ErrorCode, message string) *Error {
	return &Error{Code: code, Message: message}
}

// WithDetails trả về một lỗi mới cùng Code và Message, kèm Details; lỗi gốc
// không bị thay đổi (immutable) để các biến Err* dùng chung ở nhiều nơi an toàn.
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	return &Error{Code: e.Code, Message: e.Message, Details: details}
}

// Các lỗi domain dùng chung. Lỗi trả về nguyên trạng (không qua WithDetails)
// so được bằng == với biến Err* tương ứng. Lỗi đã qua WithDetails là một con
// trỏ mới nên phải so bằng type assertion + so Code, không dùng == (xem cách
// control-plane làm: err.(*Error).Code == CodeInvalidInput).
var (
	ErrInvalidInput           = NewError(CodeInvalidInput, "invalid input")
	ErrUserNotFound           = NewError(CodeUserNotFound, "user not found")
	ErrEmailTaken             = NewError(CodeEmailTaken, "email already taken")
	ErrInvalidCredentials     = NewError(CodeInvalidCredentials, "invalid email or password")
	ErrTokenInvalid           = NewError(CodeTokenInvalid, "token is invalid")
	ErrTokenExpired           = NewError(CodeTokenExpired, "token has expired")
	ErrTokenRevoked           = NewError(CodeTokenRevoked, "token has been revoked")
	ErrTokenReuseDetected     = NewError(CodeTokenReuseDetected, "refresh token reuse detected; all sessions revoked")
	ErrProjectNotFound        = NewError(CodeProjectNotFound, "project not found")
	ErrProjectDatabaseMissing = NewError(CodeProjectDatabaseMissing, "auth database for project does not exist")
)
