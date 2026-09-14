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

func NewDeleteField(fields port.FieldRepository, modules port.ModuleRepository, publisher port.EventPublisher) *DeleteField {
	return &DeleteField{fields: fields, modules: modules, publisher: publisher}
}

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
		if err := uc.modules.DisableFieldEverywhere(ctx, field.ID); err != nil {
			return err
		}
	}

	if err := uc.fields.Delete(ctx, field.ID); err != nil {
		return err
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
