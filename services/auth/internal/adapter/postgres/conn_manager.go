package postgres

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnManager định tuyến động tới đúng database auth_proj_<projectID> cho
// từng lời gọi repository — đây là phần triển khai "dynamic connection"
// database-per-project (quyết định #4): một Auth Service process phục vụ
// nhiều project cùng lúc, mỗi project có database Postgres riêng, nên
// UserRepository/RefreshTokenRepository/PasswordResetRepository không thể
// cầm một *pgxpool.Pool cố định như control-plane (schema-per-project, một
// database duy nhất) — chúng hỏi ConnManager mỗi lần gọi.
//
// Pool cho mỗi project được cache và tái dùng thay vì mở connection mới mỗi
// request. Chưa có eviction cho project không hoạt động lâu (LRU hay TTL) —
// ở quy mô MVP số project nhỏ, việc này để lại cho khi thật sự cần.
type ConnManager struct {
	adminPool *pgxpool.Pool

	mu    sync.RWMutex
	pools map[string]*pgxpool.Pool
}

// NewConnManager nhận adminPool (kết nối tới database "auth") chỉ để làm mẫu
// cấu hình (host/user/password) — không dùng adminPool để chạy query cho
// project, mỗi project có pool riêng trỏ đúng database của nó.
func NewConnManager(adminPool *pgxpool.Pool) *ConnManager {
	return &ConnManager{adminPool: adminPool, pools: make(map[string]*pgxpool.Pool)}
}

// Pool trả về pool đã kết nối tới auth_proj_<projectID>, tạo mới và cache
// nếu đây là lần đầu project này được truy cập trong process này. Không tự
// tạo database nếu chưa có — CreateDatabase (DatabaseManager, Bước 3) phải
// chạy trước qua event project.created; Pool ở đây chỉ kết nối, một
// projectID mà DatabaseManager chưa xử lý sẽ lỗi kết nối rõ ràng ("database
// ... does not exist") thay vì âm thầm tạo database thiếu schema.
func (m *ConnManager) Pool(ctx context.Context, projectID string) (*pgxpool.Pool, error) {
	m.mu.RLock()
	pool, ok := m.pools[projectID]
	m.mu.RUnlock()
	if ok {
		return pool, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	// Kiểm tra lại sau khi giành write lock — request khác có thể đã tạo
	// xong pool trong lúc ta chờ (double-checked locking, tránh mở 2 pool
	// cho cùng một project khi nhiều goroutine cùng truy cập lần đầu).
	if pool, ok := m.pools[projectID]; ok {
		return pool, nil
	}

	dbName, err := databaseName(projectID)
	if err != nil {
		return nil, err
	}

	cfg := m.adminPool.Config()
	cfg.ConnConfig.Database = dbName
	// MaxConns nhỏ vì đây là pool riêng cho một project, không phải pool
	// dùng chung — một Auth Service có thể phục vụ hàng chục project cùng
	// lúc, mỗi pool 4 connection tối đa đã đủ cho MVP và tránh làm cạn
	// max_connections của Postgres khi số project tăng.
	cfg.MaxConns = 4

	newPool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	m.pools[projectID] = newPool
	return newPool, nil
}

// Close đóng toàn bộ pool đã mở cho mọi project — gọi khi Auth Service shutdown.
func (m *ConnManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, pool := range m.pools {
		pool.Close()
	}
	m.pools = make(map[string]*pgxpool.Pool)
}
