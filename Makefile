DATABASE_URL            ?= postgres://backify:backify@localhost:5432/backify?sslmode=disable
AUTH_DATABASE_URL       ?= postgres://backify:backify@localhost:5433/auth?sslmode=disable
REDIS_URL               ?= redis://localhost:6379/0
RABBITMQ_URL            ?= amqp://backify:backify@localhost:5672/
JWT_SECRET              ?= change-me-in-production-use-long-random-string
CONTROL_PLANE_GRPC_ADDR ?= localhost:9091

CONTROL_PLANE_DIR       := services/control-plane
AUTH_DIR                := services/auth
CONTROL_MIGRATIONS_DIR  := $(CONTROL_PLANE_DIR)/migrations
AUTH_MIGRATIONS_DIR     := $(AUTH_DIR)/migrations
COMPOSE_FILE            := deployments/docker-compose.yml

define source_cp_env
	set -a; \
	[ -f $(CONTROL_PLANE_DIR)/.env ] && . ./$(CONTROL_PLANE_DIR)/.env; \
	set +a;
endef

define source_auth_env
	set -a; \
	[ -f $(AUTH_DIR)/.env ] && . ./$(AUTH_DIR)/.env; \
	set +a;
endef

.PHONY: help \
	infra-up infra-down infra-logs \
	run run-auth run-all \
	build build-auth \
	test test-auth test-integration test-e2e \
	migrate-up migrate-down migrate-auth-up migrate-auth-down \
	proto

help:
	@echo "Backify — common targets"
	@echo ""
	@echo "  Infra"
	@echo "    make infra-up          Start local compose services"
	@echo "    make infra-down        Stop containers"
	@echo "    make infra-logs        Follow logs"
	@echo ""
	@echo "  Run (godotenv loads services/*/ .env — dùng KEY=value hoặc KEY=\"value\")"
	@echo "    make run               Control-plane  HTTP :8080  gRPC :9091"
	@echo "    make run-auth          Auth            HTTP :8081"
	@echo "    make run-all           Hướng dẫn chạy 2 terminal"
	@echo ""
	@echo "  Migrations (source .env bằng shell)"
	@echo "    make migrate-up | migrate-down"
	@echo "    make migrate-auth-up | migrate-auth-down"
	@echo ""
	@echo "  Proto / build / test"
	@echo "    make proto | build | build-auth | test | test-auth | ..."

# -----------------------------------------------------------------------------
# Infra
# -----------------------------------------------------------------------------
infra-up:
	docker compose -f $(COMPOSE_FILE) up -d
	docker compose -f $(COMPOSE_FILE) ps

infra-down:
	docker compose -f $(COMPOSE_FILE) down

infra-logs:
	docker compose -f $(COMPOSE_FILE) logs -f

# -----------------------------------------------------------------------------
# Run — không export env từ Make; godotenv tự load .env trong service dir
# -----------------------------------------------------------------------------
run:
	cd $(CONTROL_PLANE_DIR) && go run ./cmd/api

run-auth:
	cd $(AUTH_DIR) && go run ./cmd/api

run-all:
	@echo "Start control-plane (HTTP :8080, gRPC :9091) and auth (:8081) in separate terminals:"
	@echo "  make run"
	@echo "  make run-auth"

# -----------------------------------------------------------------------------
# Build
# -----------------------------------------------------------------------------
build:
	cd $(CONTROL_PLANE_DIR) && go build -o bin/control-plane ./cmd/api

build-auth:
	cd $(AUTH_DIR) && go build -o bin/auth ./cmd/api

# -----------------------------------------------------------------------------
# Test
# -----------------------------------------------------------------------------
test:
	cd $(CONTROL_PLANE_DIR) && go test ./... -v

test-auth:
	cd $(AUTH_DIR) && go test ./... -v

test-integration:
	cd $(CONTROL_PLANE_DIR) && go test ./... -tags=integration -v

test-e2e:
	cd $(CONTROL_PLANE_DIR) && go test ./internal/... -tags=e2e -v

# -----------------------------------------------------------------------------
# Migrations — source .env bằng bash (quote được bỏ đúng cách)
# -----------------------------------------------------------------------------
migrate-up:
	@$(source_cp_env) \
	db_url="$${DATABASE_URL:-$(DATABASE_URL)}"; \
	if [ -z "$$db_url" ]; then echo "DATABASE_URL is empty — check $(CONTROL_PLANE_DIR)/.env"; exit 1; fi; \
	echo "Migrating control-plane..."; \
	migrate -path $(CONTROL_MIGRATIONS_DIR) -database "$$db_url" up

migrate-down:
	@$(source_cp_env) \
	db_url="$${DATABASE_URL:-$(DATABASE_URL)}"; \
	if [ -z "$$db_url" ]; then echo "DATABASE_URL is empty — check $(CONTROL_PLANE_DIR)/.env"; exit 1; fi; \
	migrate -path $(CONTROL_MIGRATIONS_DIR) -database "$$db_url" down

migrate-auth-up:
	@$(source_auth_env) \
	db_url="$${AUTH_DATABASE_URL:-$(AUTH_DATABASE_URL)}"; \
	if [ -z "$$db_url" ]; then echo "AUTH_DATABASE_URL is empty — check $(AUTH_DIR)/.env"; exit 1; fi; \
	echo "Migrating auth..."; \
	migrate -path $(AUTH_MIGRATIONS_DIR) -database "$$db_url" up

migrate-auth-down:
	@$(source_auth_env) \
	db_url="$${AUTH_DATABASE_URL:-$(AUTH_DATABASE_URL)}"; \
	if [ -z "$$db_url" ]; then echo "AUTH_DATABASE_URL is empty — check $(AUTH_DIR)/.env"; exit 1; fi; \
	migrate -path $(AUTH_MIGRATIONS_DIR) -database "$$db_url" down

# -----------------------------------------------------------------------------
# Proto
# -----------------------------------------------------------------------------
proto:
	@if command -v buf >/dev/null 2>&1; then \
		buf generate; \
	elif command -v docker >/dev/null 2>&1; then \
		docker run --rm -v "$$(pwd):/workspace" -w /workspace bufbuild/buf:1.47.2 generate; \
	else \
		echo "Neither buf nor docker found. Install one of:"; \
		echo "  # buf: https://buf.build/docs/installation"; \
		echo "  go install github.com/bufbuild/buf/cmd/buf@latest"; \
		exit 1; \
	fi
