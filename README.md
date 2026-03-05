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


## Deployment

### Quick Start (Local Testing)

```bash
cp .env.example .env
# Edit .env with Auth0 credentials
docker compose up
# Frontend: http://localhost:5173
# API: http://localhost:8080
```

### Production Deployment

Full automation via GitHub Actions:

1. **Feature branch** → GitHub CI runs (lint, test, build)
2. **Merge to `main`** → Auto-deploys to `tasks.kwstnr.ch`
3. **Git tag** (v*) → Tags image, deploys with version number

**Setup required:**
- [Auth0 free tier](https://auth0.com) with GitHub connection
- Hetzner VPS with Docker
- GitHub secrets: `DEPLOY_KEY`, `DEPLOY_HOST`, `DEPLOY_USER`

See [DEPLOYMENT.md](./DEPLOYMENT.md) for full setup guide.

### Workflow

```bash
# Feature work
git checkout -b feat/my-feature
# ... make changes ...
git push origin feat/my-feature
# → GitHub CI runs on PR

# Merge
git checkout main
git merge feat/my-feature
git push origin main
# → CI passes, auto-deploys in ~5-10 min

# Or release
make release v=1.0.0
# → Tags and deploys with version number
```
