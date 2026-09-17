// Package migrations nhúng schema template áp dụng cho mọi database
// auth_proj_<project_id> vào binary — DatabaseManager (adapter/postgres)
// dùng Template thay vì đọc file từ đĩa lúc chạy, để service chạy được từ
// một binary duy nhất trong Docker mà không cần copy kèm thư mục migrations/.
package migrations

import _ "embed"

//go:embed 001_template.sql
var Template string
