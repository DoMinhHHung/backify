DATABASE_URL ?= postgres://backify:backify@localhost:5432/backify?sslmode=disable
CONTROL_PLANE_DIR := services/control-plane
MIGRATIONS_DIR := $(CONTROL_PLANE_DIR)/migrations

.PHONY: infra-up infra-down infra-logs run build test test-integration test-e2e migrate-up migrate-down

infra-up:
	docker compose -f deployments/docker-compose.yml up -d
	docker compose -f deployments/docker-compose.yml ps

infra-down:
	docker compose -f deployments/docker-compose.yml down

infra-logs:
	docker compose -f deployments/docker-compose.yml logs -f

run:
	cd $(CONTROL_PLANE_DIR) && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api

build:
	cd $(CONTROL_PLANE_DIR) && go build -o bin/control-plane ./cmd/api

test:
	cd $(CONTROL_PLANE_DIR) && go test ./... -v

test-integration:
	cd $(CONTROL_PLANE_DIR) && go test ./... -tags=integration -v

test-e2e:
	cd $(CONTROL_PLANE_DIR) && go test ./internal/... -tags=e2e -v

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down
