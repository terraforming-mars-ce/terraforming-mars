FROM oven/bun:1.4.2 AS builder
WORKDIR /build/proxy

COPY proxy/package.json proxy/bun.lock ./
RUN bun install --frozen-lockfile
COPY proxy/ ./
COPY assets/original/app/ /build/assets/original/app/
COPY frontend/public/manifest.json /build/frontend/public/manifest.json

RUN bun run build

FROM nginx:alpine
WORKDIR /app

COPY --from=builder /build/proxy/dist /usr/share/nginx/html
COPY proxy/nginx.conf /etc/nginx/conf.d/default.conf
COPY proxy/docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /usr/share/nginx/html && \
    chown -R appuser:appuser /var/cache/nginx && \
    chown -R appuser:appuser /var/log/nginx && \
    chown -R appuser:appuser /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown -R appuser:appuser /var/run/nginx.pid
USER appuser

EXPOSE 4000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --output-document=/dev/null http://127.0.0.1:4000/health || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
