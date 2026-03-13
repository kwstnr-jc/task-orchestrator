.PHONY: dev build test test-api test-frontend migrate lint clean deploy

# ── Development ───────────────────────────────────────
dev:
	docker compose up --build

dev-api:
	cd api && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

# ── Quality ───────────────────────────────────────────
test:
	cd api && go test -race ./...
	cd frontend && npm test

test-api:
	cd api && go test -race -v ./...

test-frontend:
	cd frontend && npm test

lint:
	cd api && go vet ./...
	cd api && golangci-lint run ./... || true
	cd frontend && npx tsc --noEmit

# ── Build ─────────────────────────────────────────────
build:
	docker build -t ghcr.io/kwstnr-jc/task-orchestrator:latest --target production .

build-api:
	cd api && CGO_ENABLED=0 go build -o ../bin/server ./cmd/server

build-frontend:
	cd frontend && npm ci && npm run build

# ── Database ──────────────────────────────────────────
migrate:
	psql "$(DATABASE_URL)" -f db/migrations/001_init.sql

# ── Deployment ────────────────────────────────────────
deploy: build
	docker push ghcr.io/kwstnr-jc/task-orchestrator:latest
	@echo "Image pushed. Watchtower will pick it up, or run: make deploy-manual"

deploy-manual:
	ssh $(DEPLOY_HOST) "cd ~/task-orchestrator && bash deploy.sh"

# Semantic version release: make release v=1.0.0
release:
	@[ -n "$(v)" ] || (echo "Usage: make release v=1.0.0" && exit 1)
	git tag -a "v$(v)" -m "Release v$(v)"
	git push origin "v$(v)"
	@echo "Tagged v$(v) — GitHub Actions will build and deploy"

# ── Cleanup ───────────────────────────────────────────
clean:
	docker compose down -v
	rm -rf frontend/dist frontend/node_modules api/tmp bin/
