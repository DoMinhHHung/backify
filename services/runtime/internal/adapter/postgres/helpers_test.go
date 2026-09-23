package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

func TestQuoteIdentEscapesEmbeddedQuotes(t *testing.T) {
	require.Equal(t, `"project""name"`, quoteIdent(`project"name`))
}

func TestNameCaseConversions(t *testing.T) {
	tests := []struct {
		camel string
		snake string
	}{
		{camel: "userId", snake: "user_id"},
		{camel: "createdAt", snake: "created_at"},
		{camel: "name", snake: "name"},
	}

	for _, tt := range tests {
		t.Run(tt.camel, func(t *testing.T) {
			require.Equal(t, tt.snake, toSnake(tt.camel))
			require.Equal(t, tt.camel, fromSnake(tt.snake))
		})
	}
}

func TestFromSnakeSkipsEmptySegments(t *testing.T) {
	require.Equal(t, "userId", fromSnake("user__id"))
}

func TestColumnDefMapsSupportedFieldTypesAndConstraints(t *testing.T) {
	tests := []struct {
		name  string
		field domain.Field
		want  string
	}{
		{name: "uuid primary key", field: domain.Field{Name: "id", Type: "uuid"}, want: `"id" UUID PRIMARY KEY DEFAULT gen_random_uuid()`},
		{name: "required unique string", field: domain.Field{Name: "email", Type: "email", Required: true, Unique: true}, want: `"email" TEXT NOT NULL UNIQUE`},
		{name: "integer", field: domain.Field{Name: "count", Type: "int"}, want: `"count" BIGINT`},
		{name: "boolean", field: domain.Field{Name: "active", Type: "bool"}, want: `"active" BOOLEAN`},
		{name: "date", field: domain.Field{Name: "dueDate", Type: "date"}, want: `"due_date" DATE`},
		{name: "enum", field: domain.Field{Name: "status", Type: "enum"}, want: `"status" TEXT`},
		{name: "relation", field: domain.Field{Name: "userId", Type: "relation", Required: true}, want: `"user_id" UUID NOT NULL`},
		{name: "password", field: domain.Field{Name: "password", Type: "password"}, want: `"password" TEXT`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := columnDef(tt.field)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestColumnDefRejectsUnsupportedType(t *testing.T) {
	got, err := columnDef(domain.Field{Name: "payload", Type: "json"})

	require.Empty(t, got)
	require.EqualError(t, err, `unsupported field type "json"`)
}

func TestNullIfEmpty(t *testing.T) {
	require.Nil(t, nullIfEmpty(""))
	require.Equal(t, "value", nullIfEmpty("value"))
}
