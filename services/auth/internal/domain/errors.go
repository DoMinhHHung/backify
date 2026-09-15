package domain

import "fmt"

type ErrorCode string

const (
	CodeInvalidInput          ErrorCode = "INVALID_INPUT"
	CodeUserNotFound          ErrorCode = "USER_NOT_FOUND"
	CodeEmailTaken            ErrorCode = "EMAIL_TAKEN"
	CodeInvalidCredentials    ErrorCode = "INVALID_CREDENTIALS"
	CodeTokenInvalid          ErrorCode = "TOKEN_INVALID"
	CodeTokenExpired          ErrorCode = "TOKEN_EXPIRED"
	CodeTokenRevoked          ErrorCode = "TOKEN_REVOKED"
	CodeTokenReuseDetected    ErrorCode = "TOKEN_REUSE_DETECTED"
	CodeProjectNotFound       ErrorCode = "PROJECT_NOT_FOUND"
	CodeProjectDatabaseMissing ErrorCode = "PROJECT_DATABASE_MISSING"
)

type Error struct {
	Code    ErrorCode
	Message string
	Details map[string]interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code ErrorCode, message string) *Error {
	return &Error{Code: code, Message: message}
}

func (e *Error) WithDetails(details map[string]interface{}) *Error {
	return &Error{Code: e.Code, Message: e.Message, Details: details}
}

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
