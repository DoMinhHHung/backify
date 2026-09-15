package domain

import (
	"regexp"
	"time"
)

type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeText    FieldType = "text"
	FieldTypeEmail   FieldType = "email"
	FieldTypePhone   FieldType = "phone"
	FieldTypeDate    FieldType = "date"
	FieldTypeNumber  FieldType = "number"
	FieldTypeEnum    FieldType = "enum"
	FieldTypeImage   FieldType = "image"
	FieldTypeFile    FieldType = "file"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeUUID    FieldType = "uuid"
)

// Valid cho biết kiểu field có thuộc tập kiểu được hỗ trợ hay không.
func (t FieldType) Valid() bool {
	switch t {
	case FieldTypeString, FieldTypeText, FieldTypeEmail, FieldTypePhone,
		FieldTypeDate, FieldTypeNumber, FieldTypeEnum, FieldTypeImage,
		FieldTypeFile, FieldTypeBoolean, FieldTypeUUID:
		return true
	default:
		return false
	}
}

var fieldNamePattern = regexp.MustCompile(`^[a-z0-9_]{1,50}$`)

var systemFieldNames = map[string]bool{
	"id":       true,
	"email":    true,
	"password": true,
}

var reservedFieldTypes = map[string]FieldType{
	"email": FieldTypeEmail,
}

type Field struct {
	ID        string
	EntityID  string
	Name      string
	Type      FieldType
	IsSystem  bool
	CreatedAt time.Time
}

func validateFieldName(name string) error {
	if !fieldNamePattern.MatchString(name) {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "name",
			"reason": "must be lowercase alphanumeric with underscores, 1-50 chars",
		})
	}
	return nil
}

// NewField tạo field mới sau khi kiểm tra tên, kiểu và quy tắc dành riêng cho tên "email".
// Các tên id, email và password được đánh dấu là field hệ thống.
func NewField(entityID, name string, fieldType FieldType) (*Field, error) {
	if entityID == "" {
		return nil, ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "entity_id",
			"reason": "must not be empty",
		})
	}

	if err := validateFieldName(name); err != nil {
		return nil, err
	}

	if !fieldType.Valid() {
		return nil, ErrInvalidFieldType.WithDetails(map[string]interface{}{
			"field": "type",
			"value": string(fieldType),
		})
	}

	if requiredType, isReserved := reservedFieldTypes[name]; isReserved && fieldType != requiredType {
		return nil, ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "type",
			"reason": "field name \"" + name + "\" is reserved and must have type \"" + string(requiredType) + "\"",
		})
	}

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	return &Field{
		ID:        id,
		EntityID:  entityID,
		Name:      name,
		Type:      fieldType,
		IsSystem:  systemFieldNames[name],
		CreatedAt: time.Now().UTC(),
	}, nil
}

// CanDelete trả về lỗi nếu field là tài nguyên hệ thống không được phép xóa.
func (f *Field) CanDelete() error {
	if f.IsSystem {
		return ErrSystemFieldCannotDelete
	}
	return nil
}
