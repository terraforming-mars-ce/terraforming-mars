#!/usr/bin/env bash
set -euo pipefail

PI_HOST="${PI_HOST:-mhm@ssh.mh-hemma.rackaracka.net}"
PI_COMPOSE_DIR="${PI_COMPOSE_DIR:-/home/mhm/terraforming-mars}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

BACKEND_IMAGE="ghcr.io/terraforming-mars-ce/terraforming-mars-backend:latest"
FRONTEND_IMAGE="ghcr.io/terraforming-mars-ce/terraforming-mars-frontend:latest"

GIT_DESCRIBE=$(git -C "$PROJECT_ROOT" describe --tags --always)
if echo "$GIT_DESCRIBE" | grep -q '-'; then
    TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    BUILD_VERSION="${GIT_DESCRIBE}_${TIMESTAMP}-local"
else
    BUILD_VERSION="${GIT_DESCRIBE}-local"
fi

usage() {
    echo "Usage: $0 [--build-only]"
    echo ""
    echo "  (no flags)    Build and deploy to Pi"
    echo "  --build-only  Build images without deploying"
    echo ""
    echo "Environment variables:"
    echo "  PI_HOST         SSH target (default: mhm@ssh.mh-hemma.rackaracka.net)"
    echo "  PI_COMPOSE_DIR  Remote compose dir (default: /home/mhm/terraforming-mars)"
}

build() {
    echo "==> Build version: $BUILD_VERSION"

    echo "==> Building backend image (linux/arm64)..."
    docker buildx build \
        --platform linux/arm64 \
        --load \
        --build-arg BUILD_VERSION="$BUILD_VERSION" \
        -t "$BACKEND_IMAGE" \
        -f "$PROJECT_ROOT/backend/Dockerfile" \
        "$PROJECT_ROOT"

    echo "==> Building frontend image (linux/arm64)..."
    docker buildx build \
        --platform linux/arm64 \
        --load \
        --build-arg BUILD_VERSION="$BUILD_VERSION" \
        -t "$FRONTEND_IMAGE" \
        -f "$PROJECT_ROOT/frontend/Dockerfile" \
        "$PROJECT_ROOT/frontend"

    echo "==> Images built"
}

deploy() {
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf '$tmp_dir'" EXIT

    echo "==> Saving images..."
    docker save "$BACKEND_IMAGE" | gzip > "$tmp_dir/backend.tar.gz"
    docker save "$FRONTEND_IMAGE" | gzip > "$tmp_dir/frontend.tar.gz"

    echo "==> Transferring images to $PI_HOST..."
    rsync -az --progress "$tmp_dir/backend.tar.gz" "$tmp_dir/frontend.tar.gz" "$PI_HOST:/tmp/"

    echo "==> Loading images and restarting on Pi..."
    ssh "$PI_HOST" bash <<REMOTE
set -euo pipefail

echo "    Loading backend image..."
docker load < /tmp/backend.tar.gz

echo "    Loading frontend image..."
docker load < /tmp/frontend.tar.gz

echo "    Cleaning up transfer files..."
rm -f /tmp/backend.tar.gz /tmp/frontend.tar.gz

echo "    Restarting containers..."
cd $PI_COMPOSE_DIR
docker compose up -d --force-recreate backend frontend

echo "    Pruning old images..."
docker image prune -f
REMOTE

    echo "==> Verifying containers are running..."
    sleep 3

    ssh "$PI_HOST" bash <<'HEALTHCHECK'
set -euo pipefail
for container in tm-backend tm-frontend; do
    status=$(docker inspect "$container" --format '{{.State.Status}}')
    if [ "$status" = "running" ]; then
        echo "    $container: running"
    else
        echo "    $container: $status (expected running)"
        exit 1
    fi
done
HEALTHCHECK

    echo "==> Deploy complete"
}

case "${1:-}" in
    --help|-h)
        usage
        exit 0
        ;;
    --build-only)
        build
        ;;
    "")
        build
        deploy
        ;;
    *)
        echo "Unknown flag: $1"
        usage
        exit 1
        ;;
esac
