package usecase

import (
	"context"
	"log"

	"github.com/DoMinhHHung/backify/services/runtime/internal/port"
)

type ConfigEvents struct {
	cache     port.ConfigCache
	bootstrap *BootstrapSchema
}

func NewConfigEvents(cache port.ConfigCache, bootstrap *BootstrapSchema) *ConfigEvents {
	return &ConfigEvents{cache: cache, bootstrap: bootstrap}
}

func (uc *ConfigEvents) HandleProjectCreated(ctx context.Context, projectID string) error {
	return uc.refresh(ctx, projectID, "project.created")
}

func (uc *ConfigEvents) HandleConfigUpdated(ctx context.Context, projectID string) error {
	return uc.refresh(ctx, projectID, "project.config.updated")
}

func (uc *ConfigEvents) refresh(ctx context.Context, projectID, reason string) error {
	if uc.cache != nil {
		uc.cache.Invalidate(ctx, projectID)
	}
	out, err := uc.bootstrap.Execute(ctx, BootstrapSchemaInput{ProjectID: projectID})
	if err != nil {
		log.Printf("config event %s project=%s bootstrap error: %v", reason, projectID, err)
		return err
	}
	log.Printf("config event %s project=%s schema=%s version=%d applied=%v",
		reason, projectID, out.SchemaName, out.Version, out.Applied)
	return nil
}
