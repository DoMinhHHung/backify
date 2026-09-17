package port

import "context"

// DatabaseManager tạo và xóa database riêng cho từng project theo mô hình
// database-per-project (auth_proj_<project_id>) — quyết định #4 đã chốt,
// trên Postgres instance riêng tách khỏi schema-per-project của Control
// Plane. Tách interface khỏi implementation Postgres để subscriber/usecase
// không phụ thuộc trực tiếp vào pgx, dễ mock khi test.
type DatabaseManager interface {
	// CreateDatabase tạo database auth_proj_<projectID> và áp schema từ
	// migrations/001_template.sql vào đó ngay sau khi tạo. Idempotent: gọi
	// lại với project đã có database phải là no-op, không lỗi — RabbitMQ có
	// thể redeliver event project.created.
	CreateDatabase(ctx context.Context, projectID string) error

	// DropDatabase xóa database auth_proj_<projectID>. Idempotent: project
	// chưa từng có database (hoặc đã bị xóa trước đó) không được lỗi.
	DropDatabase(ctx context.Context, projectID string) error
}
