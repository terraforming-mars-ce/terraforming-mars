FROM golang:1.24-alpine AS builder
WORKDIR /build

RUN apk add --no-cache git ca-certificates tzdata
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .

ARG BUILD_VERSION="localbuild"

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -extldflags \"-static\" -X main.Version=${BUILD_VERSION}" \
    -a -installsuffix cgo \
    -o server \
    cmd/server/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata bash unzip curl libstdc++ libgcc && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

RUN curl -fsSL https://bun.sh/install | bash && \
    ln -s /root/.bun/bin/bun /usr/local/bin/bun && \
    ln -s /root/.bun/bin/bunx /usr/local/bin/bunx && \
    bun install -g @anthropic-ai/claude-code

COPY --from=builder /build/server .
COPY --from=builder /build/assets ./assets

# Source code for Claude bug report analysis
COPY backend/internal/ /repo/backend/internal/
COPY backend/cmd/ /repo/backend/cmd/
COPY backend/assets/ /repo/backend/assets/
COPY frontend/src/ /repo/frontend/src/

ENV TM_REPO_PATH=/repo
ENV SHELL=/bin/bash

RUN chown -R appuser:appuser /app /repo
USER appuser

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/api/health || exit 1

CMD ["./server"]
