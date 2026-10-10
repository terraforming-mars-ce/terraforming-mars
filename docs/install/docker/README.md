# Docker

## Building the image

The Dockerfile expects to be built from the repo root. The game's art and audio are in Git LFS, so pull them first.

```bash
git lfs pull
docker build -t openmars .
```

## Running the game

Minimal setup -- just the game, no bug reporting:

```bash
docker run -d --name openmars -p 3001:3001 openmars
```

Or use the example compose file in this directory:

```bash
docker compose up -d
```

Game is at `http://localhost:3001`. The same server answers the page, the API and the WebSocket.

## Running with bug reporting

The GitHub App private key should be bind-mounted into the container as a file, not passed as an env var. The `:ro` flag makes it read-only inside the container.

```bash
docker run -d --name openmars \
  -p 3001:3001 \
  -v /path/to/private-key.pem:/etc/secrets/github/private-key.pem:ro \
  -e GITHUB_APP_ID=<app-id> \
  -e GITHUB_INSTALLATION_ID=<installation-id> \
  -e GITHUB_PRIVATE_KEY_PATH=/etc/secrets/github/private-key.pem \
  -e GITHUB_REPO_OWNER=<owner> \
  -e GITHUB_REPO_NAME=<repo> \
  -e CLAUDE_CODE_OAUTH_TOKEN=<token> \
  openmars
```

If you're on Fedora or another SELinux system, you may need `:z` instead of `:ro` on the volume mount so the container process can actually read the file.

## Running behind a reverse proxy

Point the proxy at port 3001 for every path. See the [overview](../README.md) for the WebSocket timeouts it needs.

## Note on the private key

Bind-mounting keeps the key out of the image and out of environment variables, but the file still lives on the host disk. For a more locked-down setup, consider Kubernetes secrets (mounted as tmpfs) or Docker secrets in Swarm mode.
