package port

import "context"

type ConfigEventHandler interface {
	HandleProjectCreated(ctx context.Context, projectID string) error
	HandleConfigUpdated(ctx context.Context, projectID string) error
}
