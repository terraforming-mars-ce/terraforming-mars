# Docker

## Building the images

Both Dockerfiles expect to be built from the repo root.

```bash
docker build -f backend/Dockerfile -t tm-backend .
docker build -f frontend/Dockerfile -t tm-frontend frontend/
```

## Running the game

Minimal setup -- just the game, no bug reporting:

```bash
docker run -d --name tm-backend -p 3001:3001 tm-backend
docker run -d --name tm-frontend -p 3000:3000 tm-frontend
```

Or use the example compose file in this directory:

```bash
docker compose up -d
```

Game is at `http://localhost:3000`. The frontend talks to the backend through the browser at `localhost:3001`, so `API_URL` is set to point there.

## Running with bug reporting

The GitHub App private key should be bind-mounted into the container as a file, not passed as an env var. The `:ro` flag makes it read-only inside the container.

```bash
docker run -d --name tm-backend \
  -p 3001:3001 \
  -v /path/to/private-key.pem:/etc/secrets/github/private-key.pem:ro \
  -e GITHUB_APP_ID=<app-id> \
  -e GITHUB_INSTALLATION_ID=<installation-id> \
  -e GITHUB_PRIVATE_KEY_PATH=/etc/secrets/github/private-key.pem \
  -e GITHUB_REPO_OWNER=<owner> \
  -e GITHUB_REPO_NAME=<repo> \
  -e CLAUDE_CODE_OAUTH_TOKEN=<token> \
  tm-backend
```

If you're on Fedora or another SELinux system, you may need `:z` instead of `:ro` on the volume mount so the container process can actually read the file.

## Running behind a reverse proxy

When both containers sit behind a proxy on the same hostname, don't set `API_URL` at all. The default (`/api/v1`) uses the browser's current origin for both HTTP and WebSocket connections.

See the [overview](../README.md) for the required proxy routing rules.

## Note on the private key

Bind-mounting keeps the key out of the image and out of environment variables, but the file still lives on the host disk. For a more locked-down setup, consider Kubernetes secrets (mounted as tmpfs) or Docker secrets in Swarm mode.
