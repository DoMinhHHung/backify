package domain

type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

func ErrInvalidCredentials() *DomainError {
	return &DomainError{Code: "invalid_credentials", Message: "invalid email or password"}
}

func ErrEmailTaken() *DomainError {
	return &DomainError{Code: "email_taken", Message: "email already registered"}
}

func ErrUnauthorized() *DomainError {
	return &DomainError{Code: "unauthorized", Message: "unauthorized"}
}

func ErrForbidden() *DomainError {
	return &DomainError{Code: "forbidden", Message: "forbidden"}
}

func ErrValidation(msg string) *DomainError {
	return &DomainError{Code: "validation_error", Message: msg}
}

func ErrNotFound(msg string) *DomainError {
	return &DomainError{Code: "not_found", Message: msg}
}
