package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type SchemaMigrator struct {
	pool *pgxpool.Pool
}

func NewSchemaMigrator(pool *pgxpool.Pool) *SchemaMigrator {
	return &SchemaMigrator{pool: pool}
}

func (m *SchemaMigrator) AppliedVersion(ctx context.Context, schemaName string) (int32, error) {
	var version int32
	query := fmt.Sprintf(`
		SELECT COALESCE(
			(SELECT version FROM %s._schema_meta WHERE id = 1),
			0
		)
	`, quoteIdent(schemaName))

	err := m.pool.QueryRow(ctx, query).Scan(&version)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return 0, nil
		}
		return 0, err
	}
	return version, nil
}

func (m *SchemaMigrator) EnsureSchema(ctx context.Context, cfg *domain.ProjectConfig) error {
	applied, err := m.AppliedVersion(ctx, cfg.SchemaName)
	if err != nil {
		return err
	}
	if applied >= cfg.Version {
		return nil
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quoteIdent(cfg.SchemaName))); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	metaDDL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s._schema_meta (
			id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
			version INT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`, quoteIdent(cfg.SchemaName))
	if _, err := tx.Exec(ctx, metaDDL); err != nil {
		return fmt.Errorf("create _schema_meta: %w", err)
	}

	for _, entity := range cfg.Entities {
		if err := m.createTable(ctx, tx, cfg.SchemaName, entity); err != nil {
			return err
		}
	}
	for _, entity := range cfg.Entities {
		if err := m.ensureConstraints(ctx, tx, cfg.SchemaName, entity); err != nil {
			return err
		}
	}

	upsert := fmt.Sprintf(`
		INSERT INTO %s._schema_meta (id, version, updated_at)
		VALUES (1, $1, now())
		ON CONFLICT (id) DO UPDATE SET version = EXCLUDED.version, updated_at = now()
	`, quoteIdent(cfg.SchemaName))
	if _, err := tx.Exec(ctx, upsert, cfg.Version); err != nil {
		return fmt.Errorf("update schema version: %w", err)
	}

	return tx.Commit(ctx)
}

func (m *SchemaMigrator) createTable(ctx context.Context, tx pgx.Tx, schemaName string, entity domain.Entity) error {
	table := quoteIdent(schemaName) + "." + quoteIdent(toSnake(entity.Name))

	cols := make([]string, 0, len(entity.Pool)+3)
	hasID := false

	for _, f := range entity.Pool {
		colName := toSnake(f.Name)
		if colName == "id" {
			hasID = true
		}
		colDef, err := columnDef(f)
		if err != nil {
			return fmt.Errorf("entity %s field %s: %w", entity.Name, f.Name, err)
		}
		cols = append(cols, colDef)
	}

	if !hasID {
		cols = append([]string{`id UUID PRIMARY KEY DEFAULT gen_random_uuid()`}, cols...)
	}

	cols = append(cols,
		`created_at TIMESTAMPTZ NOT NULL DEFAULT now()`,
		`updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`,
	)

	ddl := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (\n  %s\n)",
		table,
		strings.Join(cols, ",\n  "),
	)

	if _, err := tx.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("create table %s: %w", entity.Name, err)
	}

	if strings.EqualFold(entity.Name, "User") {
		alterRole := fmt.Sprintf(
			`ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s TEXT NOT NULL DEFAULT 'user'`,
			table,
			quoteIdent("role"),
		)
		if _, err := tx.Exec(ctx, alterRole); err != nil {
			return fmt.Errorf("ensure role column: %w", err)
		}
	}

	return nil
}

func (m *SchemaMigrator) ensureConstraints(ctx context.Context, tx pgx.Tx, schemaName string, entity domain.Entity) error {
	table := quoteIdent(schemaName) + "." + quoteIdent(toSnake(entity.Name))
	tableName := toSnake(entity.Name)

	for _, f := range entity.Pool {
		if f.Type != "relation" || f.RelationCardinality != "n-1" || f.RelationTo == "" {
			continue
		}
		fkTable := quoteIdent(schemaName) + "." + quoteIdent(toSnake(f.RelationTo))
		col := toSnake(f.Name)
		constraint := fmt.Sprintf("fk_%s_%s", tableName, col)

		var exists bool
		check := `
			SELECT EXISTS (
				SELECT 1
				FROM pg_constraint c
				JOIN pg_class t ON t.oid = c.conrelid
				JOIN pg_namespace n ON n.oid = t.relnamespace
				WHERE c.conname = $1 AND n.nspname = $2 AND t.relname = $3
			)`
		if err := tx.QueryRow(ctx, check, constraint, schemaName, tableName).Scan(&exists); err != nil {
			return fmt.Errorf("check fk %s: %w", constraint, err)
		}
		if exists {
			continue
		}

		alter := fmt.Sprintf(
			`ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (id) ON DELETE RESTRICT`,
			table, quoteIdent(constraint), quoteIdent(col), fkTable,
		)
		if _, err := tx.Exec(ctx, alter); err != nil {
			return fmt.Errorf("add fk %s: %w", constraint, err)
		}
	}
	return nil
}

func columnDef(f domain.Field) (string, error) {
	name := quoteIdent(toSnake(f.Name))
	var typ string

	switch f.Type {
	case "uuid":
		typ = "UUID"
	case "string", "email", "phone", "password":
		typ = "TEXT"
	case "int":
		typ = "BIGINT"
	case "bool":
		typ = "BOOLEAN"
	case "date":
		typ = "DATE"
	case "enum":
		typ = "TEXT"
	case "relation":
		typ = "UUID"
	default:
		return "", fmt.Errorf("unsupported field type %q", f.Type)
	}

	parts := []string{name, typ}

	if f.Name == "id" || toSnake(f.Name) == "id" {
		parts = append(parts, "PRIMARY KEY")
		if f.Type == "uuid" {
			parts = append(parts, "DEFAULT gen_random_uuid()")
		}
	} else {
		if f.Required {
			parts = append(parts, "NOT NULL")
		}
		if f.Unique {
			parts = append(parts, "UNIQUE")
		}
	}

	return strings.Join(parts, " "), nil
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
