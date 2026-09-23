CP := services/control-plane
RT := services/runtime
UV := uv --directory $(CP)

.PHONY: infra-up infra-down infra-logs \
	install install-runtime \
	migrate-up migrate-down grpc-gen \
	run run-control-plane run-runtime run-all \
	test test-runtime test-cov lint fmt clean \
	swagger

infra-up:
	docker compose up -d --wait

infra-down:
	docker compose down -v

infra-logs:
	docker compose logs -f

install:
	$(UV) sync --extra dev

install-runtime:
	cd $(RT) && go mod download

migrate-up:
	cd $(CP) && uv run alembic upgrade head

migrate-down:
	cd $(CP) && uv run alembic downgrade -1

grpc-gen:
	uv --directory $(CP) run python -m grpc_tools.protoc \
		-I ../../proto \
		--python_out=app/adapter/grpc/pb \
		--grpc_python_out=app/adapter/grpc/pb \
		--pyi_out=app/adapter/grpc/pb \
		../../proto/control_plane.proto
	sed -i 's/^import control_plane_pb2/from . import control_plane_pb2/' \
		$(CP)/app/adapter/grpc/pb/control_plane_pb2_grpc.py

run: run-control-plane

run-control-plane:
	cd $(CP) && uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

run-runtime:
	cd $(RT) && go run ./cmd/server

run-all:
	@echo "Start control-plane on :8000 and runtime on :8081 in two terminals:"
	@echo "  make run-control-plane"
	@echo "  make run-runtime"

test:
	$(UV) run pytest -q

test-runtime:
	cd $(RT) && go test ./internal/permission/ -v
	cd $(RT) && go test ./... -count=1

test-cov:
	$(UV) run pytest -q --cov=app --cov-report=term-missing

lint:
	$(UV) run ruff check .
	$(UV) run ruff format --check .
	cd $(RT) && go vet ./...

fmt:
	$(UV) run ruff check . --fix
	$(UV) run ruff format .
	cd $(RT) && gofmt -w .

swagger:
	@echo "Control-plane: http://localhost:8000/docs"
	@echo "Runtime:       http://localhost:8081/swagger/"

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
