DATABASE_URL ?= postgres://backify:backify@localhost:5432/backify?sslmode=disable
AUTH_DATABASE_URL ?= postgres://backify:backify@localhost:5433/auth?sslmode=disable
REDIS_URL ?= redis://localhost:6379/0
RABBITMQ_URL ?= amqp://backify:backify@localhost:5672/
JWT_SECRET ?= change-me-in-production-use-long-random-string
CONTROL_PLANE_GRPC_ADDR ?= localhost:9091

CONTROL_PLANE_DIR := services/control-plane
AUTH_DIR := services/auth
CONTROL_MIGRATIONS_DIR := $(CONTROL_PLANE_DIR)/migrations
AUTH_MIGRATIONS_DIR := $(AUTH_DIR)/migrations

.PHONY: infra-up infra-down infra-logs \
	run run-auth run-all \
	build build-auth \
	test test-auth test-integration test-e2e \
	migrate-up migrate-down \
	proto

infra-up:
	docker compose -f deployments/docker-compose.yml up -d
	docker compose -f deployments/docker-compose.yml ps

infra-down:
	docker compose -f deployments/docker-compose.yml down

infra-logs:
	docker compose -f deployments/docker-compose.yml logs -f

run:
	cd $(CONTROL_PLANE_DIR) && \
		DATABASE_URL="$(DATABASE_URL)" \
		RABBITMQ_URL="$(RABBITMQ_URL)" \
		GRPC_ADDR=:9091 \
		go run ./cmd/api

run-auth:
	cd $(AUTH_DIR) && \
		AUTH_DATABASE_URL="$(AUTH_DATABASE_URL)" \
		REDIS_URL="$(REDIS_URL)" \
		RABBITMQ_URL="$(RABBITMQ_URL)" \
		JWT_SECRET="$(JWT_SECRET)" \
		CONTROL_PLANE_GRPC_ADDR="$(CONTROL_PLANE_GRPC_ADDR)" \
		PORT=8081 \
		go run ./cmd/api

run-all:
	@echo "Start control-plane (HTTP :8080, gRPC :9091) and auth (:8081) in separate terminals:"
	@echo "  make run"
	@echo "  make run-auth"

build:
	cd $(CONTROL_PLANE_DIR) && go build -o bin/control-plane ./cmd/api

build-auth:
	cd $(AUTH_DIR) && go build -o bin/auth ./cmd/api

test:
	cd $(CONTROL_PLANE_DIR) && go test ./... -v

test-auth:
	cd $(AUTH_DIR) && go test ./... -v

test-integration:
	cd $(CONTROL_PLANE_DIR) && go test ./... -tags=integration -v

test-e2e:
	cd $(CONTROL_PLANE_DIR) && go test ./internal/... -tags=e2e -v

migrate-up:
	migrate -path $(CONTROL_MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(CONTROL_MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

proto:
	buf generate
