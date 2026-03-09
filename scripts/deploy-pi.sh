#!/usr/bin/env bash
set -euo pipefail

PI_HOST="${PI_HOST:-mhm@ssh.mh-hemma.rackaracka.net}"
PI_COMPOSE_DIR="${PI_COMPOSE_DIR:-/home/mhm/terraforming-mars}"
SKIP_BUILD="${SKIP_BUILD:-0}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
STAGING_DIR="$PROJECT_ROOT/.deploy-staging"

BACKEND_IMAGE_BASE="ghcr.io/rackaracka123/terraforming-mars-backend"
FRONTEND_IMAGE_BASE="ghcr.io/rackaracka123/terraforming-mars-frontend"

GIT_DESCRIBE=$(git -C "$PROJECT_ROOT" describe --tags --always)
if echo "$GIT_DESCRIBE" | grep -q '-'; then
    TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    BUILD_VERSION="${GIT_DESCRIBE}_${TIMESTAMP}-local"
else
    BUILD_VERSION="${GIT_DESCRIBE}-local"
fi

BACKEND_IMAGE="${BACKEND_IMAGE_BASE}:latest"
FRONTEND_IMAGE="${FRONTEND_IMAGE_BASE}:latest"

usage() {
    echo "Usage: $0 [--build-only]"
    echo ""
    echo "  (no flags)    Build and deploy to Pi via Docker"
    echo "  --build-only  Build without deploying"
    echo ""
    echo "Environment variables:"
    echo "  PI_HOST         SSH target (default: mhm@ssh.mh-hemma.rackaracka.net)"
    echo "  PI_COMPOSE_DIR  Remote compose dir (default: /home/mhm/terraforming-mars)"
    echo "  SKIP_BUILD=1    Skip build, deploy pre-built artifacts"
}

build() {
    echo "==> Build version: $BUILD_VERSION"

    echo "==> Building backend (linux/arm64)..."
    cd "$PROJECT_ROOT/backend"
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
        -ldflags="-w -s -X main.Version=${BUILD_VERSION}" \
        -o bin/server-arm64 cmd/server/main.go
    echo "    backend/bin/server-arm64 ready"

    echo "==> Building frontend..."
    cd "$PROJECT_ROOT/frontend"
    VITE_APP_VERSION="$BUILD_VERSION" bun run build
    echo "    frontend/build/ ready"
}

stage() {
    echo "==> Staging deploy artifacts..."
    rm -rf "$STAGING_DIR"
    mkdir -p "$STAGING_DIR/backend" "$STAGING_DIR/frontend"

    cp "$PROJECT_ROOT/backend/bin/server-arm64" "$STAGING_DIR/backend/server"
    cp -r "$PROJECT_ROOT/backend/assets" "$STAGING_DIR/backend/assets"

    cp -r "$PROJECT_ROOT/frontend/build" "$STAGING_DIR/frontend/build"
    cp "$PROJECT_ROOT/frontend/nginx.conf" "$STAGING_DIR/frontend/nginx.conf"
    cp "$PROJECT_ROOT/frontend/docker-entrypoint.sh" "$STAGING_DIR/frontend/docker-entrypoint.sh"

    cat > "$STAGING_DIR/backend/Dockerfile" <<'DOCKERFILE'
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY server .
COPY assets ./assets
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/api/health || exit 1
CMD ["./server"]
DOCKERFILE

    cat > "$STAGING_DIR/frontend/Dockerfile" <<'DOCKERFILE'
FROM nginx:alpine
COPY build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
DOCKERFILE

    echo "    staged to .deploy-staging/"
}

deploy() {
    echo "==> Uploading to $PI_HOST..."
    rsync -az --delete "$STAGING_DIR/" "$PI_HOST:/tmp/tm-deploy/"

    echo "==> Building Docker images on Pi (version: $BUILD_VERSION)..."
    ssh "$PI_HOST" bash <<REMOTE
set -euo pipefail

echo "    Building backend image..."
docker build -t $BACKEND_IMAGE /tmp/tm-deploy/backend/

echo "    Building frontend image..."
docker build -t $FRONTEND_IMAGE /tmp/tm-deploy/frontend/

echo "    Restarting containers..."
cd $PI_COMPOSE_DIR
docker compose up -d --force-recreate backend frontend

echo "    Cleaning up..."
rm -rf /tmp/tm-deploy
docker image prune -f
REMOTE

    echo "==> Waiting for health checks..."
    sleep 5

    if ssh "$PI_HOST" "docker inspect tm-backend --format '{{.State.Health.Status}}'" | grep -q healthy; then
        echo "    backend: healthy"
    else
        echo "    backend: waiting..."
        sleep 10
        ssh "$PI_HOST" "docker inspect tm-backend --format '{{.State.Health.Status}}'"
    fi

    if ssh "$PI_HOST" "docker inspect tm-frontend --format '{{.State.Health.Status}}'" | grep -q healthy; then
        echo "    frontend: healthy"
    else
        echo "    frontend: waiting..."
        sleep 10
        ssh "$PI_HOST" "docker inspect tm-frontend --format '{{.State.Health.Status}}'"
    fi

    rm -rf "$STAGING_DIR"
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
        if [ "$SKIP_BUILD" = "1" ]; then
            echo "==> SKIP_BUILD=1, skipping build phase"
        else
            build
        fi
        stage
        deploy
        ;;
    *)
        echo "Unknown flag: $1"
        usage
        exit 1
        ;;
esac
