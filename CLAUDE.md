# Task Orchestrator

## Architecture

- **API**: Go (Chi router, pgx for Postgres, no ORM)
- **Frontend**: React + TypeScript + Tailwind, Vite dev server
- **Database**: PostgreSQL (Neon in production, Docker locally)
- **Infra**: Docker Compose for local dev, GHCR + Watchtower for deploy

## Commands

- `make dev` — docker compose up (postgres + api + frontend)
- `make test` — run Go and frontend tests
- `make test-api` / `make test-frontend` — run individually
- `make lint` — go vet + golangci-lint + tsc
- `make build` — Docker production image

## Database & Migrations

Migrations auto-run on API startup. No manual steps needed.

**When changing the schema:**
1. Create a new numbered SQL file in `api/internal/db/migrations/` (e.g., `003_add_foo.sql`)
2. Also copy it to `db/migrations/` (used by Docker Compose for fresh postgres init)
3. The Go binary embeds `api/internal/db/migrations/` and applies pending ones via a `schema_migrations` tracking table
4. Migrations must be idempotent-safe — they run once and are tracked by filename

## Testing

- Go handler tests use mock store interfaces (`TaskStore`, `ProjectStore`, `UserStore` in `api/internal/db/store.go`)
- Migration and store integration tests use testcontainers (spins up a real Postgres in Docker)
- Frontend tests use Vitest + React Testing Library
- CI runs on PR to main: lint, test (unit + integration), type-check, build, Docker build
- Integration tests are skipped with `-short` flag if needed

## Before Pushing

**All of the following must pass locally before committing/pushing:**

```bash
cd api && go vet ./... && go test -race ./...
cd frontend && npx tsc --noEmit && npm test
```

Or simply: `make lint && make test`

This matches what CI runs — if it fails in CI, it should have been caught locally first.

## Key Conventions

- No ORM — raw SQL with pgx, queries in `api/internal/db/db.go`
- Store interfaces for testability, concrete `*Store` wraps pgxpool
- `DEV_MODE=true` bypasses Auth0 (injects fake dev user)
- Bearer JWT for machine-to-machine auth (daily-digest → API)
- Mobile-first responsive design: `lg:` breakpoint (1024px) separates mobile/desktop layouts
