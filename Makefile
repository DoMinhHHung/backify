CP := services/control-plane
UV := uv --directory $(CP)

.PHONY: infra-up infra-down infra-logs install migrate-up migrate-down run test test-cov lint fmt clean

infra-up:
	docker compose up -d --wait

infra-down:
	docker compose down -v

infra-logs:
	docker compose logs -f

install:
	$(UV) sync --extra dev

migrate-up:
	cd $(CP) && uv run alembic upgrade head

migrate-down:
	cd $(CP) && uv run alembic downgrade -1

run:
	cd $(CP) && uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

test:
	$(UV) run pytest -q

test-cov:
	$(UV) run pytest -q --cov=app --cov-report=term-missing

lint:
	$(UV) run ruff check .
	$(UV) run ruff format --check .

fmt:
	$(UV) run ruff check . --fix
	$(UV) run ruff format .

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
