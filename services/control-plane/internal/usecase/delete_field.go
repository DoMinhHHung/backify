package usecase

import (
	"context"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/port"
)

type FieldDeletedEvent struct {
	FieldID  string `json:"field_id"`
	EntityID string `json:"entity_id"`
	Name     string `json:"name"`
}

type DeleteFieldInput struct {
	FieldID string
	Force   bool
}

type DeleteField struct {
	fields    port.FieldRepository
	modules   port.ModuleRepository
	publisher port.EventPublisher
}

// NewDeleteField tạo use case xóa field từ các dependency cần thiết.
func NewDeleteField(fields port.FieldRepository, modules port.ModuleRepository, publisher port.EventPublisher) *DeleteField {
	return &DeleteField{fields: fields, modules: modules, publisher: publisher}
}

// Execute xóa field không thuộc hệ thống và phát sự kiện field.deleted khi thành công.
// Field đang được sử dụng cần Force; khi đó việc tắt liên kết và xóa diễn ra trong một transaction.
func (uc *DeleteField) Execute(ctx context.Context, input DeleteFieldInput) error {
	field, err := uc.fields.GetByID(ctx, input.FieldID)
	if err != nil {
		return err
	}

	if err := field.CanDelete(); err != nil {
		return err
	}

	usages, err := uc.modules.ListFieldUsages(ctx, field.ID)
	if err != nil {
		return err
	}

	if len(usages) > 0 && !input.Force {
		usageNames := make([]string, len(usages))
		for i, usage := range usages {
			usageNames[i] = usage.ModuleName + "." + usage.FunctionName
		}
		return domain.NewFieldInUseError(usageNames)
	}

	if len(usages) > 0 {
		if err := uc.modules.DisableAndDeleteField(ctx, field.ID); err != nil {
			return err
		}
	} else {
		if err := uc.fields.Delete(ctx, field.ID); err != nil {
			return err
		}
	}

	return uc.publisher.Publish(ctx, port.Event{
		Name: "field.deleted",
		Payload: FieldDeletedEvent{
			FieldID:  field.ID,
			EntityID: field.EntityID,
			Name:     field.Name,
		},
	})
}
