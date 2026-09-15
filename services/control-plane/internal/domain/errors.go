package domain

import "fmt"

type ErrorCode string

const (
	CodeInvalidInput             ErrorCode = "INVALID_INPUT"
	CodeProjectNotFound          ErrorCode = "PROJECT_NOT_FOUND"
	CodeSubdomainTaken           ErrorCode = "SUBDOMAIN_TAKEN"
	CodeEntityNotFound           ErrorCode = "ENTITY_NOT_FOUND"
	CodeEntityNameTaken          ErrorCode = "ENTITY_NAME_TAKEN"
	CodeSystemEntityCannotDelete ErrorCode = "SYSTEM_ENTITY_CANNOT_DELETE"
	CodeFieldNotFound            ErrorCode = "FIELD_NOT_FOUND"
	CodeFieldNameTaken           ErrorCode = "FIELD_NAME_TAKEN"
	CodeSystemFieldCannotDelete  ErrorCode = "SYSTEM_FIELD_CANNOT_DELETE"
	CodeFieldInUse               ErrorCode = "FIELD_IN_USE"
	CodeInvalidFieldType         ErrorCode = "INVALID_FIELD_TYPE"
	CodeModuleNotFound           ErrorCode = "MODULE_NOT_FOUND"
	CodeModuleNotEnabled         ErrorCode = "MODULE_NOT_ENABLED"
	CodeInvalidModuleName        ErrorCode = "INVALID_MODULE_NAME"
	CodeFunctionNotFound         ErrorCode = "FUNCTION_NOT_FOUND"
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
	ErrInvalidInput             = NewError(CodeInvalidInput, "invalid input")
	ErrProjectNotFound          = NewError(CodeProjectNotFound, "project not found")
	ErrSubdomainTaken           = NewError(CodeSubdomainTaken, "subdomain already taken")
	ErrEntityNotFound           = NewError(CodeEntityNotFound, "entity not found")
	ErrEntityNameTaken          = NewError(CodeEntityNameTaken, "entity name already taken")
	ErrSystemEntityCannotDelete = NewError(CodeSystemEntityCannotDelete, "system entity cannot be deleted")
	ErrFieldNotFound            = NewError(CodeFieldNotFound, "field not found")
	ErrFieldNameTaken           = NewError(CodeFieldNameTaken, "field name already taken")
	ErrSystemFieldCannotDelete  = NewError(CodeSystemFieldCannotDelete, "system field cannot be deleted")
	ErrInvalidFieldType         = NewError(CodeInvalidFieldType, "invalid field type")
	ErrModuleNotFound           = NewError(CodeModuleNotFound, "module not found")
	ErrModuleNotEnabled         = NewError(CodeModuleNotEnabled, "module is not enabled in this plan; only auth is available in MVP")
	ErrInvalidModuleName        = NewError(CodeInvalidModuleName, "invalid module name")
	ErrFunctionNotFound         = NewError(CodeFunctionNotFound, "function not found")
)

func NewFieldInUseError(usages []string) *Error {
	return &Error{
		Code:    CodeFieldInUse,
		Message: "field is in use by one or more functions",
		Details: map[string]interface{}{"usages": usages},
	}
}
