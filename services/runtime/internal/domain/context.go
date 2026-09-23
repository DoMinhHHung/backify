package domain

import "context"

type contextKey int

const projectConfigKey contextKey = 1

// WithProjectConfig trả context con chứa cấu hình dự án cho request hiện tại.
func WithProjectConfig(ctx context.Context, cfg *ProjectConfig) context.Context {
	return context.WithValue(ctx, projectConfigKey, cfg)
}

// ProjectConfigFromContext lấy cấu hình dự án và báo false nếu context không chứa đúng kiểu dữ liệu.
func ProjectConfigFromContext(ctx context.Context) (*ProjectConfig, bool) {
	cfg, ok := ctx.Value(projectConfigKey).(*ProjectConfig)
	return cfg, ok
}
