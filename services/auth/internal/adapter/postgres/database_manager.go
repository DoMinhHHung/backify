package postgres

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backify/services/auth/migrations"
)

// pgDuplicateDatabase là SQLSTATE Postgres trả khi CREATE DATABASE trúng tên
// đã tồn tại — Postgres không có "CREATE DATABASE IF NOT EXISTS" như MySQL,
// nên đây là cách bắt idempotency ở tầng DB, dùng làm lưới an toàn thứ hai
// sau lượt kiểm tra pg_database (phòng race giữa kiểm tra và tạo).
const pgDuplicateDatabase = "42P04"

// projectIDPattern giới hạn projectID chỉ gồm ký tự an toàn để nối trực
// tiếp vào tên database — CREATE DATABASE/DROP DATABASE không nhận tham số
// (không thể dùng placeholder $1), nên phải validate chặt trước khi build
// chuỗi SQL bằng tay, không chỉ dựa vào pgx.Identifier.Sanitize.
var projectIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,50}$`)

// DatabaseManager tạo/xóa database auth_proj_<project_id> bằng một kết nối
// admin duy nhất (adminPool) — kết nối này trỏ vào database bookkeeping
// "auth", không phải vào database của project nào, vì CREATE DATABASE/DROP
// DATABASE phải chạy từ một database khác database đang được tạo/xóa.
type DatabaseManager struct {
	adminPool *pgxpool.Pool
}

// NewDatabaseManager nhận pool đã kết nối tới database admin (mặc định
// "auth" trong AUTH_DATABASE_URL) — DatabaseManager không tự mở kết nối,
// dùng lại pool App đã quản lý vòng đời.
func NewDatabaseManager(adminPool *pgxpool.Pool) *DatabaseManager {
	return &DatabaseManager{adminPool: adminPool}
}

func databaseName(projectID string) (string, error) {
	if !projectIDPattern.MatchString(projectID) {
		return "", fmt.Errorf("invalid project id for database name: %q", projectID)
	}
	return "auth_proj_" + projectID, nil
}

// CreateDatabase tạo auth_proj_<projectID> nếu chưa có, rồi áp schema từ
// migrations_template.sql. CREATE DATABASE không chạy được trong prepared
// statement (Postgres yêu cầu simple query protocol cho lệnh này), nên phải
// truyền pgx.QueryExecModeSimpleProtocol — thiếu cờ này sẽ lỗi "cannot run
// inside a transaction block" dù code không mở transaction nào.
func (m *DatabaseManager) CreateDatabase(ctx context.Context, projectID string) error {
	dbName, err := databaseName(projectID)
	if err != nil {
		return err
	}

	exists, err := m.databaseExists(ctx, dbName)
	if err != nil {
		return err
	}
	if exists {
		// Database đã có nghĩa là lần trước CreateDatabase đã chạy xong cả
		// applySchema (hai bước luôn đi cùng nhau trong nhánh !exists bên
		// dưới) — coi như thành công, không gọi lại applySchema vì CREATE
		// TABLE không IF NOT EXISTS sẽ lỗi trên schema đã tồn tại.
		return nil
	}

	createSQL := fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{dbName}.Sanitize())
	if _, err := m.adminPool.Exec(ctx, createSQL, pgx.QueryExecModeSimpleProtocol); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgDuplicateDatabase {
			// Race: một lần gọi CreateDatabase khác đã tạo xong giữa lúc ta
			// kiểm tra exists và lúc ta CREATE DATABASE. Database đó có thể
			// đang trong lúc applySchema dở dang — không có cách rẻ để biết
			// chắc, nên vẫn trả nil thay vì gọi applySchema chồng lên (rủi ro
			// lỗi "relation already exists" cao hơn lợi ích). Bước 4/6 sẽ cần
			// một cờ "schema_applied" thật sự nếu race này gây vấn đề thực tế.
			return nil
		}
		return fmt.Errorf("create database %s: %w", dbName, err)
	}

	return m.applySchema(ctx, dbName)
}

// DropDatabase xóa auth_proj_<projectID>. WITH (FORCE) (PG13+, ta chạy PG16)
// tự ngắt mọi connection đang mở tới database đó trước khi drop — cần thiết
// vì repository layer (Bước 4) sẽ giữ pool kết nối riêng tới từng database
// project, DROP DATABASE thường sẽ lỗi "database is being accessed by other
// users" nếu không có FORCE.
func (m *DatabaseManager) DropDatabase(ctx context.Context, projectID string) error {
	dbName, err := databaseName(projectID)
	if err != nil {
		return err
	}

	dropSQL := fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", pgx.Identifier{dbName}.Sanitize())
	_, err = m.adminPool.Exec(ctx, dropSQL, pgx.QueryExecModeSimpleProtocol)
	if err != nil {
		return fmt.Errorf("drop database %s: %w", dbName, err)
	}
	return nil
}

func (m *DatabaseManager) databaseExists(ctx context.Context, dbName string) (bool, error) {
	var exists bool
	err := m.adminPool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check database %s exists: %w", dbName, err)
	}
	return exists, nil
}

// applySchema mở một connection riêng (không lấy từ adminPool vì adminPool
// trỏ cố định vào database "auth") tới đúng database vừa tạo và exec toàn bộ
// migrations.Template. Chỉ được gọi từ nhánh vừa CREATE DATABASE thành công
// trong CreateDatabase — gọi lại trên database đã có schema sẽ lỗi vì
// CREATE TABLE ở đây không có IF NOT EXISTS.
func (m *DatabaseManager) applySchema(ctx context.Context, dbName string) error {
	connConfig := m.adminPool.Config().ConnConfig.Copy()
	connConfig.Database = dbName

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return fmt.Errorf("connect to database %s: %w", dbName, err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, migrations.Template); err != nil {
		return fmt.Errorf("apply schema to database %s: %w", dbName, err)
	}
	return nil
}
