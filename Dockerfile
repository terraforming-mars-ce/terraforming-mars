# One image per game server: the Go server serves the API, the WebSocket and the
# built frontend. Build stages run natively on the build machine; only the
# runtime stage is per architecture.

FROM --platform=$BUILDPLATFORM oven/bun:1.4.2 AS bun

FROM --platform=$BUILDPLATFORM node:26.10.0-bookworm-slim AS deps
COPY --from=bun /usr/local/bin/bun /usr/local/bin/bun
WORKDIR /build/frontend
COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile

# Game content (~360 MB) depends only on the originals and the pipeline, so this
# layer is reused from the cache by releases that change neither
FROM deps AS assets
COPY frontend/scripts/assets/ ./scripts/assets/
COPY assets/ /build/assets/
RUN bun run assets

FROM deps AS web
ARG BUILD_VERSION="localbuild"
ENV VITE_APP_VERSION=${BUILD_VERSION}
COPY frontend/ ./
COPY --from=assets /build/frontend/public/ ./public/
COPY --from=assets /build/frontend/src/assets/generated/ ./src/assets/generated/
RUN node_modules/.bin/vite build && \
    find build -type f \( -name '*.js' -o -name '*.css' -o -name '*.html' -o -name '*.json' \
        -o -name '*.svg' -o -name '*.map' \) -exec gzip -k -9 {} +

FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine AS server
WORKDIR /build
RUN apk add --no-cache git ca-certificates tzdata
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
ARG BUILD_VERSION="localbuild"
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags="-w -s -X main.Version=${BUILD_VERSION}" \
    -o server \
    cmd/server/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata bash curl libstdc++ libgcc && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

RUN curl -fsSL https://claude.ai/install.sh | bash && \
    chmod 755 /root && \
    ln -s /root/.local/bin/claude /usr/local/bin/claude

COPY --from=server /build/server .
COPY --from=server /build/assets ./assets
COPY --from=web /build/frontend/build ./web

# Source code for Claude bug report analysis
COPY backend/internal/ /repo/backend/internal/
COPY backend/cmd/ /repo/backend/cmd/
COPY backend/assets/ /repo/backend/assets/
COPY frontend/src/ /repo/frontend/src/

ENV OPENMARS_REPO_PATH=/repo
ENV SHELL=/bin/bash

RUN chown -R appuser:appuser /app /repo
USER appuser

EXPOSE 3001

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD addr="${OPENMARS_ADDR:-:3001}"; wget --quiet --tries=1 --output-document=/dev/null "http://localhost:${addr##*:}/api/v1/health" || exit 1

CMD ["./server"]
