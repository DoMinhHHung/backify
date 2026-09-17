-- 001_template.sql — schema áp dụng cho MỌI database auth_proj_<project_id>.
-- Không dùng golang-migrate CLI vì không có một database "auth" cố định để
-- migrate: mỗi project có DB riêng, được tạo động lúc runtime (event
-- project.created), nên DatabaseManager exec thẳng file này vào database
-- mới ngay sau khi CREATE DATABASE (xem adapter/postgres/database_manager.go).
-- project_id trong mỗi bảng chỉ để audit/debug — email đã unique theo DB
-- (một DB = một project) nên không cần project_id trong UNIQUE constraint.

CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    project_id    TEXT NOT NULL,
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    full_name     TEXT NOT NULL DEFAULT '',
    phone         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (email)
);

CREATE TABLE refresh_tokens (
    id         TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    family_id  TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (token_hash)
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);

CREATE TABLE password_reset_tokens (
    id         TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (token_hash)
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
