package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type RecordRepository struct {
	pool *pgxpool.Pool
}

func NewRecordRepository(pool *pgxpool.Pool) *RecordRepository {
	return &RecordRepository{pool: pool}
}

func (r *RecordRepository) tableRef(schemaName, table string) string {
	return quoteIdent(schemaName) + "." + quoteIdent(toSnake(table))
}

func (r *RecordRepository) Create(ctx context.Context, schemaName, table string, columns map[string]any) (domain.Record, error) {
	if _, ok := columns["id"]; !ok {
		columns["id"] = uuid.NewString()
	}
	now := time.Now().UTC()
	columns["created_at"] = now
	columns["updated_at"] = now

	cols := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))
	placeholders := make([]string, 0, len(columns))
	i := 1
	for k, v := range columns {
		cols = append(cols, quoteIdent(toSnake(k)))
		args = append(args, v)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}

	q := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) RETURNING *`,
		r.tableRef(schemaName, table),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}
	defer rows.Close()

	return scanRecord(rows)
}

func (r *RecordRepository) FindByID(ctx context.Context, schemaName, table, id string) (domain.Record, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, domain.ErrNotFound("record not found")
	}
	q := fmt.Sprintf(`SELECT * FROM %s WHERE %s = $1`, r.tableRef(schemaName, table), quoteIdent("id"))
	rows, err := r.pool.Query(ctx, q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rec, err := scanRecord(rows)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, nil
	}
	return rec, nil
}

func (r *RecordRepository) List(ctx context.Context, schemaName, table string, ownerColumn, ownerID string, limit, offset int) ([]domain.Record, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	where := ""
	args := []any{}
	if ownerColumn != "" && ownerID != "" {
		where = fmt.Sprintf(" WHERE %s = $1", quoteIdent(toSnake(ownerColumn)))
		args = append(args, ownerID)
	}

	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM %s%s`, r.tableRef(schemaName, table), where)
	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	limIdx := len(args) + 1
	offIdx := len(args) + 2

	listQ := fmt.Sprintf(
		`SELECT * FROM %s%s ORDER BY %s DESC LIMIT $%d OFFSET $%d`,
		r.tableRef(schemaName, table),
		where,
		quoteIdent("created_at"),
		limIdx,
		offIdx,
	)

	rows, err := r.pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Record, 0)
	for {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		if rec == nil {
			break
		}
		items = append(items, rec)
	}
	return items, total, nil
}

func (r *RecordRepository) Update(ctx context.Context, schemaName, table, id string, columns map[string]any) (domain.Record, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, domain.ErrNotFound("record not found")
	}
	delete(columns, "id")
	delete(columns, "created_at")
	columns["updated_at"] = time.Now().UTC()

	if len(columns) == 0 {
		return r.FindByID(ctx, schemaName, table, id)
	}

	sets := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns)+1)
	i := 1
	for k, v := range columns {
		sets = append(sets, fmt.Sprintf("%s = $%d", quoteIdent(toSnake(k)), i))
		args = append(args, v)
		i++
	}
	args = append(args, id)

	q := fmt.Sprintf(
		`UPDATE %s SET %s WHERE %s = $%d RETURNING *`,
		r.tableRef(schemaName, table),
		strings.Join(sets, ", "),
		quoteIdent("id"),
		i,
	)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRecord(rows)
}

func (r *RecordRepository) Delete(ctx context.Context, schemaName, table, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return domain.ErrNotFound("record not found")
	}
	q := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1`, r.tableRef(schemaName, table), quoteIdent("id"))
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound("record not found")
	}
	return nil
}

func scanRecord(rows pgx.Rows) (domain.Record, error) {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	descs := rows.FieldDescriptions()
	values, err := rows.Values()
	if err != nil {
		return nil, err
	}
	rec := make(domain.Record, len(descs))
	for i, d := range descs {
		v := values[i]
		switch t := v.(type) {
		case [16]byte:
			v = uuid.UUID(t).String()
		case []byte:
			if len(t) == 16 {
				var arr [16]byte
				copy(arr[:], t)
				v = uuid.UUID(arr).String()
			}
		}
		rec[fromSnake(string(d.Name))] = v
	}
	return rec, nil
}

func fromSnake(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s
	}
	var b strings.Builder
	b.WriteString(parts[0])
	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}
