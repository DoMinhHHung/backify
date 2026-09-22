package domain

import "context"

type contextKey int

const projectConfigKey contextKey = 1

func WithProjectConfig(ctx context.Context, cfg *ProjectConfig) context.Context {
	return context.WithValue(ctx, projectConfigKey, cfg)
}

func ProjectConfigFromContext(ctx context.Context) (*ProjectConfig, bool) {
	cfg, ok := ctx.Value(projectConfigKey).(*ProjectConfig)
	return cfg, ok
}
