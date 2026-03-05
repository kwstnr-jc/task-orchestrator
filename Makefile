.PHONY: dev build test migrate sqlc lint clean

dev:
	docker compose up --build

build:
	docker build -t task-orchestrator .

test:
	cd api && go test ./...

migrate:
	docker compose exec postgres psql -U orchestrator -d orchestrator -f /docker-entrypoint-initdb.d/001_init.sql

sqlc:
	cd api && sqlc generate

lint:
	cd api && golangci-lint run ./...
	cd frontend && npx eslint src/

clean:
	docker compose down -v
	rm -rf frontend/dist api/tmp
