//go:build e2e

package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcrabbitmq "github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	app "backify/services/control-plane/internal"
)

func setupApp(t *testing.T) *app.App {
	t.Helper()
	ctx := context.Background()

	migrationSQL, err := os.ReadFile("../migrations/001_init.up.sql")
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	pgContainer, err := tcpostgres.Run(ctx, "postgres:16",
		tcpostgres.WithDatabase("backify_e2e"),
		tcpostgres.WithUsername("backify"),
		tcpostgres.WithPassword("backify"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	dbURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	rmqContainer, err := tcrabbitmq.Run(ctx, "rabbitmq:3.13-management")
	if err != nil {
		t.Fatalf("failed to start rabbitmq container: %v", err)
	}
	t.Cleanup(func() { _ = rmqContainer.Terminate(ctx) })

	rmqURL, err := rmqContainer.AmqpURL(ctx)
	if err != nil {
		t.Fatalf("failed to get rabbitmq url: %v", err)
	}

	application, err := app.New(ctx, app.Config{DatabaseURL: dbURL, RabbitMQURL: rmqURL})
	if err != nil {
		t.Fatalf("failed to wire app: %v", err)
	}
	t.Cleanup(application.Close)

	if _, err := application.Pool.Exec(ctx, string(migrationSQL)); err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	return application
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, dst interface{}) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(dst); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

func TestE2E_FullFieldLifecycle(t *testing.T) {
	application := setupApp(t)

	createBody, _ := json.Marshal(map[string]string{"name": "Shop App", "subdomain": "shop-app-e2e"})
	req := httptest.NewRequest("POST", "/api/v1/projects", bytes.NewReader(createBody))
	rec := httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("expected 201 creating project, got %d: %s", rec.Code, rec.Body.String())
	}

	var createResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &createResp)
	projectID := createResp.Data.ID

	entityBody, _ := json.Marshal(map[string]string{"name": "product"})
	req = httptest.NewRequest("POST", "/api/v1/projects/"+projectID+"/entities", bytes.NewReader(entityBody))
	rec = httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("expected 201 creating entity, got %d: %s", rec.Code, rec.Body.String())
	}

	var entityResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &entityResp)
	entityID := entityResp.Data.ID

	fieldBody, _ := json.Marshal(map[string]string{"name": "title", "type": "string"})
	req = httptest.NewRequest("POST", "/api/v1/entities/"+entityID+"/fields", bytes.NewReader(fieldBody))
	rec = httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("expected 201 creating field, got %d: %s", rec.Code, rec.Body.String())
	}

	var fieldResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &fieldResp)
	fieldID := fieldResp.Data.ID

	toggleBody, _ := json.Marshal(map[string]interface{}{"field_id": fieldID, "enabled": true})
	req = httptest.NewRequest("POST", "/api/v1/projects/"+projectID+"/modules/auth/functions/signup/toggle-field", bytes.NewReader(toggleBody))
	rec = httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatalf("expected 204 toggling field, got %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest("DELETE", "/api/v1/fields/"+fieldID, nil)
	rec = httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	if rec.Code != 409 {
		t.Fatalf("expected 409 deleting in-use field without force, got %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest("DELETE", "/api/v1/fields/"+fieldID+"?force=true", nil)
	rec = httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatalf("expected 204 force deleting field, got %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest("DELETE", "/api/v1/projects/"+projectID, nil)
	rec = httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatalf("expected 204 deleting project, got %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest("GET", "/api/v1/projects/"+projectID, nil)
	rec = httptest.NewRecorder()
	application.Router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 getting soft-deleted project, got %d: %s", rec.Code, rec.Body.String())
	}

	var getResp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	decodeBody(t, rec, &getResp)
	if getResp.Data.Status != "deleted" {
		t.Fatalf("expected status deleted, got %s", getResp.Data.Status)
	}
}
