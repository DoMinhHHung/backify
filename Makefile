DATABASE_URL ?= postgres://backify:backify@localhost:5432/backify?sslmode=disable
CONTROL_PLANE_DIR := services/control-plane
MIGRATIONS_DIR := $(CONTROL_PLANE_DIR)/migrations

AUTH_DIR := services/auth
AUTH_MIGRATIONS_DIR := $(AUTH_DIR)/migrations
CONTROL_DATABASE_URL ?= postgres://backify:backify@localhost:5432/backify?sslmode=disable
AUTH_DATABASE_URL ?= postgres://backify:backify@localhost:5433/postgres?sslmode=disable
REDIS_URL ?= redis://localhost:6379/0
RABBITMQ_URL ?= amqp://backify:backify@localhost:5672/

.PHONY: infra-up infra-down infra-logs run build test test-integration test-e2e migrate-up migrate-down \
	run-auth build-auth test-auth test-integration-auth test-e2e-auth migrate-up-auth migrate-down-auth

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

# --- Auth Service ---
# Target riêng suffix -auth thay vì đè lên run/build/test ở trên, để không
# phá các target sẵn có của Control Plane khi cả hai service cùng nằm trong
# một Makefile.

run-auth:
	cd $(AUTH_DIR) && \
		CONTROL_DATABASE_URL="$(CONTROL_DATABASE_URL)" \
		AUTH_DATABASE_URL="$(AUTH_DATABASE_URL)" \
		REDIS_URL="$(REDIS_URL)" \
		RABBITMQ_URL="$(RABBITMQ_URL)" \
		go run ./cmd/api

build-auth:
	cd $(AUTH_DIR) && go build -o bin/auth ./cmd/api

test-auth:
	cd $(AUTH_DIR) && go test ./... -v

test-integration-auth:
	cd $(AUTH_DIR) && go test ./... -tags=integration -v

test-e2e-auth:
	cd $(AUTH_DIR) && go test ./internal/... -tags=e2e -v

migrate-up-auth:
	migrate -path $(AUTH_MIGRATIONS_DIR) -database "$(AUTH_DATABASE_URL)" up

migrate-down-auth:
	migrate -path $(AUTH_MIGRATIONS_DIR) -database "$(AUTH_DATABASE_URL)" down
