#!/usr/bin/env bash
set -euo pipefail

PI_HOST="${PI_HOST:-mhm@ssh.mh-hemma.rackaracka.net}"
PI_COMPOSE_DIR="${PI_COMPOSE_DIR:-/home/mhm/terraforming-mars}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

APP_IMAGE="ghcr.io/terraforming-mars-ce/terraforming-mars:latest"

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
    echo "  --build-only  Build the image without deploying"
    echo ""
    echo "Environment variables:"
    echo "  PI_HOST         SSH target (default: mhm@ssh.mh-hemma.rackaracka.net)"
    echo "  PI_COMPOSE_DIR  Remote compose dir (default: /home/mhm/terraforming-mars)"
}

build() {
    echo "==> Build version: $BUILD_VERSION"

    echo "==> Building app image (linux/arm64)..."
    docker buildx build \
        --platform linux/arm64 \
        --load \
        --build-arg BUILD_VERSION="$BUILD_VERSION" \
        -t "$APP_IMAGE" \
        -f "$PROJECT_ROOT/Dockerfile" \
        "$PROJECT_ROOT"

    echo "==> Image built"
}

deploy() {
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf '$tmp_dir'" EXIT

    echo "==> Saving image..."
    docker save "$APP_IMAGE" | gzip > "$tmp_dir/app.tar.gz"

    echo "==> Transferring image to $PI_HOST..."
    rsync -az --progress "$tmp_dir/app.tar.gz" "$PI_HOST:/tmp/"

    echo "==> Loading image and restarting on Pi..."
    ssh "$PI_HOST" bash <<REMOTE
set -euo pipefail

echo "    Loading app image..."
docker load < /tmp/app.tar.gz

echo "    Cleaning up transfer file..."
rm -f /tmp/app.tar.gz

echo "    Restarting the app..."
cd $PI_COMPOSE_DIR
docker compose up -d --force-recreate app

echo "    Pruning old images..."
docker image prune -f
REMOTE

    echo "==> Verifying containers are running..."
    sleep 3

    ssh "$PI_HOST" bash <<'HEALTHCHECK'
set -euo pipefail
for container in tm-app; do
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
