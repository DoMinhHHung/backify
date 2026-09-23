package domain

type DomainError struct {
	Code    string
	Message string
}

// Error trả nguyên Message của lỗi domain.
func (e *DomainError) Error() string {
	return e.Message
}

// ErrInvalidCredentials tạo lỗi invalid_credentials không tiết lộ email hay mật khẩu sai.
func ErrInvalidCredentials() *DomainError {
	return &DomainError{Code: "invalid_credentials", Message: "invalid email or password"}
}

// ErrEmailTaken tạo lỗi email_taken cho địa chỉ email đã được đăng ký.
func ErrEmailTaken() *DomainError {
	return &DomainError{Code: "email_taken", Message: "email already registered"}
}

// ErrUnauthorized tạo lỗi unauthorized.
func ErrUnauthorized() *DomainError {
	return &DomainError{Code: "unauthorized", Message: "unauthorized"}
}

// ErrForbidden tạo lỗi forbidden.
func ErrForbidden() *DomainError {
	return &DomainError{Code: "forbidden", Message: "forbidden"}
}

// ErrValidation tạo lỗi validation_error và giữ nguyên thông báo đầu vào.
func ErrValidation(msg string) *DomainError {
	return &DomainError{Code: "validation_error", Message: msg}
}

// ErrNotFound tạo lỗi not_found và giữ nguyên thông báo đầu vào.
func ErrNotFound(msg string) *DomainError {
	return &DomainError{Code: "not_found", Message: msg}
}
