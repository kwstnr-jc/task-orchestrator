# Task Orchestrator

A kanban board hub for a personal automation system.

## Architecture

```
                          +------------------+
                          |   React SPA      |
                          |  (Vite + TS)     |
                          +--------+---------+
                                   |
                           HttpOnly Cookie
                                   |
                          +--------v---------+
   +----------+  Bearer   |    Go API        |
   | Mac Mini +----------->    (chi router)  |
   | Worker   |   JWT     |    :8080         |
   +----+-----+           +----+--------+----+
        |                      |        |
        |                      v        v
        |               +------+--+ +---+----+
        +-------------->|Postgres | | Auth0   |
          via SQS       |  :5432  | | (GH)   |
                        +---------+ +--------+
```

## Quick Start

```bash
cp .env.example .env
# Fill in Auth0 credentials in .env
docker compose up
```

- **Frontend**: http://localhost:5173
- **API**: http://localhost:8080

## Auth0 Setup

1. Create an Auth0 application (Regular Web Application)
2. Enable the **GitHub** social connection
3. Set callback URL: `http://localhost:8080/api/auth/callback`
4. Set logout URL: `http://localhost:5173`
5. Copy **Domain**, **Client ID**, **Client Secret** into `.env`

## Deployment (Hetzner VPS + Caddy)

1. Build and push the Docker image:
   ```bash
   docker build -t ghcr.io/kwstnr-jc/task-orchestrator:latest .
   docker push ghcr.io/kwstnr-jc/task-orchestrator:latest
   ```

2. On the VPS, pull and run:
   ```bash
   docker pull ghcr.io/kwstnr-jc/task-orchestrator:latest
   docker run -d --name orchestrator \
     --env-file .env \
     -p 8080:8080 \
     ghcr.io/kwstnr-jc/task-orchestrator:latest
   ```

3. Install Caddy and copy the `Caddyfile`:
   ```bash
   sudo cp Caddyfile /etc/caddy/Caddyfile
   sudo systemctl reload caddy
   ```

4. Run Postgres (or use a managed DB):
   ```bash
   docker run -d --name postgres \
     -e POSTGRES_USER=orchestrator \
     -e POSTGRES_PASSWORD=<strong-password> \
     -e POSTGRES_DB=orchestrator \
     -v pgdata:/var/lib/postgresql/data \
     -p 5432:5432 \
     postgres:16-alpine
   ```

5. Run migrations:
   ```bash
   psql $DATABASE_URL -f db/migrations/001_init.sql
   ```

## Environment Variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | Postgres connection string |
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | Auth0 app client ID |
| `AUTH0_CLIENT_SECRET` | Auth0 app client secret |
| `AUTH0_CALLBACK_URL` | OAuth callback URL |
| `SESSION_SECRET` | 32-byte hex string for cookie signing |
| `COOKIE_DOMAIN` | Domain for session cookie |
| `ALLOWED_USERS` | Comma-separated GitHub usernames |
| `ALLOWED_MACHINES` | Comma-separated machine client IDs |
| `AWS_REGION` | AWS region for SQS |
| `SQS_QUEUE_URL` | SQS queue URL for task dispatch |
| `PORT` | API server port (default: 8080) |
| `CORS_ORIGIN` | Allowed CORS origin |

## Development

```bash
make dev          # Start all services
make test         # Run Go tests
make sqlc         # Regenerate sqlc code
make lint         # Lint Go + TypeScript
make migrate      # Run DB migrations
make clean        # Tear down containers and volumes
```
