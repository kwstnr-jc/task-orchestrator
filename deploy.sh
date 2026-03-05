#!/bin/bash
set -euo pipefail

# Deployment script for Task Orchestrator on Hetzner VPS
# Pulls latest image, stops old container, starts new one, verifies health

COMPOSE_FILE="${COMPOSE_FILE:-.}"
REGISTRY="ghcr.io/kwstnr-jc/task-orchestrator"
IMAGE_TAG="${IMAGE_TAG:-latest}"
FULL_IMAGE="$REGISTRY:$IMAGE_TAG"

echo "🚀 Deploying $FULL_IMAGE"

# Authenticate with GHCR (assumes token in ~/.docker/config.json or env)
# or you can use: docker login -u <username> -p <token> ghcr.io

# Pull latest image
echo "📦 Pulling image..."
docker pull "$FULL_IMAGE"

# Stop and remove old container (if exists)
echo "🛑 Stopping old container..."
docker stop orchestrator 2>/dev/null || true
docker rm orchestrator 2>/dev/null || true

# Load environment
if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

# Start new container
echo "🟢 Starting new container..."
docker run -d \
  --name orchestrator \
  --restart unless-stopped \
  --env-file .env \
  -p 127.0.0.1:8080:8080 \
  -v /var/lib/orchestrator/data:/app/data \
  "$FULL_IMAGE"

# Wait for container to be healthy
echo "⏳ Waiting for container to be ready..."
for i in {1..30}; do
  if docker exec orchestrator curl -sf http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "✅ Container is healthy"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ Container failed to become healthy"
    docker logs orchestrator
    exit 1
  fi
  sleep 2
done

echo "✨ Deployment successful!"
