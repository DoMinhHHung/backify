package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

type AddFieldInput struct {
	EntityID string
	Name     string
	Type     domain.FieldType
}

type AddField struct {
	entities port.EntityRepository
	fields   port.FieldRepository
}

// NewAddField tạo use case quản lý field từ các kho lưu trữ cần thiết.
func NewAddField(entities port.EntityRepository, fields port.FieldRepository) *AddField {
	return &AddField{entities: entities, fields: fields}
}

// Execute kiểm tra entity tồn tại, tạo field hợp lệ rồi lưu và trả về field đó.
func (uc *AddField) Execute(ctx context.Context, input AddFieldInput) (*domain.Field, error) {
	if _, err := uc.entities.GetByID(ctx, input.EntityID); err != nil {
		return nil, err
	}

	field, err := domain.NewField(input.EntityID, input.Name, input.Type)
	if err != nil {
		return nil, err
	}

	if err := uc.fields.Create(ctx, field); err != nil {
		return nil, err
	}

	return field, nil
}

// List trả về các field của entity sau khi xác nhận entity tồn tại.
func (uc *AddField) List(ctx context.Context, entityID string) ([]*domain.Field, error) {
	if _, err := uc.entities.GetByID(ctx, entityID); err != nil {
		return nil, err
	}
	return uc.fields.ListByEntity(ctx, entityID)
}
