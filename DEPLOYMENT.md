# Deployment Guide

## Overview

Task Orchestrator uses **full CI/CD automation**:

1. **Feature branches**: lint, test, build (no deploy)
2. **Merge to `main`**: auto-deploy to production
3. **Git tags** (`v*`): build, push to GHCR, deploy with health checks

## Prerequisites

### Hetzner VPS Setup

```bash
# SSH into VPS
ssh root@your-vps-ip

# Install Docker + Docker Compose
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Clone repo (or pull latest in existing clone)
git clone https://github.com/kwstnr-jc/task-orchestrator.git ~/task-orchestrator
cd ~/task-orchestrator

# Create data directory
mkdir -p /var/lib/orchestrator/data
chmod 755 /var/lib/orchestrator/data

# Copy and edit env file
cp .env.example .env
# Edit .env with Auth0 credentials, DB URL, allowed users
```

### GitHub Secrets

Go to **Settings → Secrets and variables → Actions** on the repo:

| Secret | Value |
|--------|-------|
| `DEPLOY_KEY` | SSH private key (4096-bit RSA) for deploy user |
| `DEPLOY_HOST` | VPS hostname/IP (e.g., `vps.example.com`) |
| `DEPLOY_USER` | VPS deploy user (e.g., `deploy`) |

**Generate deploy key (on your machine):**
```bash
ssh-keygen -t rsa -b 4096 -f deploy_key -N ""
# Copy contents of deploy_key to DEPLOY_KEY secret
# Add deploy_key.pub to VPS ~/.ssh/authorized_keys
```

### VPS Deploy User

```bash
# On VPS, as root
useradd -m -s /bin/bash deploy
mkdir -p /home/deploy/.ssh
# Paste deploy_key.pub content into authorized_keys
echo "YOUR_PUBLIC_KEY_HERE" >> /home/deploy/.ssh/authorized_keys
chmod 700 /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys

# Give docker permissions
usermod -aG docker deploy

# Create task directory
sudo -u deploy mkdir -p /home/deploy/task-orchestrator
cd /home/deploy/task-orchestrator
sudo -u deploy git clone https://github.com/kwstnr-jc/task-orchestrator.git .
```

### Caddy Setup

```bash
# Install Caddy
sudo apt install -y caddy

# Copy Caddyfile
sudo cp Caddyfile /etc/caddy/Caddyfile

# Edit with your domain
sudo nano /etc/caddy/Caddyfile

# Reload
sudo systemctl reload caddy
sudo systemctl enable caddy
```

---

## Workflows

### 1. Feature Development

```bash
git checkout -b feat/my-feature
# Make changes
git push origin feat/my-feature
```

**GitHub Actions:**
- ✅ Lints Go + TypeScript
- ✅ Runs tests
- ✅ Builds Docker image (doesn't push)
- 🚫 Does NOT deploy

**PR merges only after CI passes** (branch protection enforced).

### 2. Merge to Main (Auto-Deploy)

```bash
git checkout main
git merge feat/my-feature
git push origin main
```

**GitHub Actions:**
1. CI runs (same checks as PR)
2. Builds + pushes to `ghcr.io/kwstnr-jc/task-orchestrator:latest`
3. SSH into Hetzner, runs `deploy.sh`:
   - Pulls new image
   - Stops old container
   - Starts new container with `.env`
   - Health check (waits for `/api/auth/me`)
4. Verifies service is healthy

**Timeline:** ~5-10 minutes from push to live.

### 3. Semantic Versioning (Optional)

For stable releases, tag with semver:

```bash
make release v=1.0.0
```

This:
- Creates annotated tag `v1.0.0`
- Pushes to GitHub
- Triggers **same deploy flow** as main

**Image tags created:**
- `ghcr.io/kwstnr-jc/task-orchestrator:v1.0.0`
- `ghcr.io/kwstnr-jc/task-orchestrator:1.0`
- `ghcr.io/kwstnr-jc/task-orchestrator:latest` (if main)

---

## Rollback

If deployment breaks:

```bash
# Via GitHub: Revert commit + push (auto-deploy will kick in)
git revert HEAD
git push origin main

# Or manual on VPS:
ssh deploy@vps-ip "cd ~/task-orchestrator && bash deploy.sh" # Re-runs old image
```

---

## Manual Deployment

If you need to deploy without GitHub:

```bash
# Locally: build and push image
make build
make push

# Or on VPS directly:
ssh deploy@vps-ip "cd ~/task-orchestrator && bash deploy.sh"
```

---

## Production Checklist

- [ ] Auth0 app created with GitHub connection
- [ ] `.env` file on VPS with all secrets
- [ ] `DEPLOY_KEY`, `DEPLOY_HOST`, `DEPLOY_USER` secrets in GitHub
- [ ] Caddyfile configured with your domain
- [ ] SSL certificate auto-renewed by Caddy
- [ ] Database backups configured (Neon auto-backups)
- [ ] Monitoring/alerts set up (optional)

---

## Watchtower (Auto-Pull)

The `docker-compose.prod.yml` includes **Watchtower**, which:
- Checks GHCR every 60 seconds for new images
- Auto-pulls and restarts if new version is found
- Alternative to manual `deploy.sh` trigger

To enable:
```bash
docker compose -f docker-compose.prod.yml up -d watchtower
```

Set `WATCHTOWER_NOTIFICATION_URL` to Slack/Discord webhook for notifications.

---

## Troubleshooting

**Deployment fails health check:**
```bash
ssh deploy@vps-ip
docker logs orchestrator
```

**Image won't pull:**
```bash
# Verify GHCR credentials
docker login -u <github-username> -p <github-token> ghcr.io
```

**No auto-deploy after push to main:**
- Check GitHub Actions tab for workflow errors
- Verify branch protection rules don't block
- Check deploy secrets are set

---

## Next Steps

- Set up monitoring (Prometheus, Grafana, or simpler: uptimerobot.com)
- Add Discord notifications for deploy status
- Consider database backup retention policy
