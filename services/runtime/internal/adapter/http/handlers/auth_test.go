package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

func TestWriteDomainErrorMapsExpectedStatusAndPayload(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "validation", err: domain.ErrValidation("invalid"), status: http.StatusBadRequest, code: "validation_error"},
		{name: "unauthorized", err: domain.ErrUnauthorized(), status: http.StatusUnauthorized, code: "unauthorized"},
		{name: "invalid credentials", err: domain.ErrInvalidCredentials(), status: http.StatusUnauthorized, code: "invalid_credentials"},
		{name: "forbidden", err: domain.ErrForbidden(), status: http.StatusForbidden, code: "forbidden"},
		{name: "not found", err: domain.ErrNotFound("missing"), status: http.StatusNotFound, code: "not_found"},
		{name: "email taken", err: domain.ErrEmailTaken(), status: http.StatusConflict, code: "email_taken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			writeDomainError(recorder, tt.err)

			require.Equal(t, tt.status, recorder.Code)
			require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			require.JSONEq(t, `{"error":"`+tt.err.Error()+`","code":"`+tt.code+`"}`, recorder.Body.String())
		})
	}
}

func TestWriteDomainErrorSanitizesUnexpectedErrors(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeDomainError(recorder, errors.New("database password leaked"))

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.JSONEq(t, `{"error":"internal error"}`, recorder.Body.String())
	require.NotContains(t, recorder.Body.String(), "database password")
}

func TestPublicUserExcludesPasswordAndExtraFields(t *testing.T) {
	now := time.Now().UTC()
	got := publicUser(&domain.User{
		ID: "user-1", Email: "person@example.com", PasswordHash: "secret", FullName: "Person", Phone: "0900", Role: "user",
		CreatedAt: now, UpdatedAt: now, Extra: map[string]any{"private": true},
	})

	require.Equal(t, map[string]any{
		"id": "user-1", "email": "person@example.com", "fullName": "Person", "phone": "0900", "role": "user",
		"createdAt": now, "updatedAt": now,
	}, got)
	require.NotContains(t, got, "password")
	require.NotContains(t, got, "passwordHash")
	require.NotContains(t, got, "extra")
}
