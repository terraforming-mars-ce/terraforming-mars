# Deploying Terraforming Mars

Two containers, a reverse proxy in front. That's it.

## Containers

**Backend** (`ghcr.io/rackaracka123/terraforming-mars-backend`) -- Go server on port 3001. Handles game logic, WebSocket connections, and the bug report API.

**Frontend** (`ghcr.io/rackaracka123/terraforming-mars-frontend`) -- Nginx serving the React app on port 3000. Static files with a runtime config injected at startup.

Both images are built from the repo root. See `backend/Dockerfile` and `frontend/Dockerfile`.

## Reverse proxy

You need something in front that routes traffic to the right container. The frontend uses relative URLs (`/api/v1`), so both services must be behind the same hostname.

Required routing:

| Path | Destination | Notes |
|------|-------------|-------|
| `/api/*` | backend:3001 | REST API |
| `/ws` | backend:3001 | WebSocket -- needs long timeouts |
| `/` | frontend:3000 | Everything else |

WebSocket connections stay open for the duration of a game, so set proxy read/send timeouts to something high (3600s works). If your proxy drops idle connections after 60 seconds, players will get disconnected constantly.

## Environment variables

### Backend

| Variable | Required | Default | What it does |
|----------|----------|---------|--------------|
| `TM_LOG_LEVEL` | No | `info` | Log verbosity. Options: `debug`, `info`, `warn`, `error` |

The game works fine without any environment variables. The bug report feature has its own set of optional env vars below.

#### Bug reporting

All optional. Without these, the game runs normally but the in-game bug report button shows "not available".

| Variable | Required | Default | What it does |
|----------|----------|---------|--------------|
| `GITHUB_APP_ID` | No | _(unset)_ | GitHub App ID for creating bug report issues |
| `GITHUB_INSTALLATION_ID` | No | _(unset)_ | GitHub App installation ID. Without this, bug reporting is disabled entirely |
| `GITHUB_PRIVATE_KEY_PATH` | No | `./private-key.pem` | Path to the GitHub App private key file. Must be readable inside the container |
| `GITHUB_REPO_OWNER` | No | `rackaracka123` | GitHub repo owner where issues get created |
| `GITHUB_REPO_NAME` | No | `terraforming-mars` | GitHub repo name where issues get created |
| `TM_REPO_PATH` | No | _(unset)_ | Path to source code for Claude analysis. The Docker image already sets this to `/repo`, so you only need to set it when running locally (e.g. `TM_REPO_PATH=./` from the repo root) |
| `CLAUDE_CODE_OAUTH_TOKEN` | No | _(unset)_ | OAuth token for Claude Code CLI. Without this (and without a mounted `~/.claude` directory), Claude analysis is skipped. Bug reports still get created -- they just won't have AI-generated code analysis |

The backend logs which capabilities are active on startup:

```
Bug report service initialized  {"github_app": true, "claude": true}
```

If something is misconfigured you'll see `false` and a warning explaining why.

### Frontend

| Variable | Required | Default | What it does |
|----------|----------|---------|--------------|
| `API_URL` | No | `/api/v1` | Full API URL. Leave unset when running behind a reverse proxy -- the default relative path works. Set to a complete URL (e.g. `http://localhost:3001/api/v1`) when the backend is on a different host or port |

The default (`/api/v1`) uses `window.location.origin` for both HTTP and WebSocket connections, so TLS, hostname, and port all come from wherever the browser loaded the page. Only set this when you need to point at a different server.

## Bug report capabilities

The bug report feature has two independent capabilities that degrade gracefully:

| Capability | What you need | What happens without it |
|------------|---------------|------------------------|
| **GitHub App** | `GITHUB_INSTALLATION_ID` + private key file | Bug reporting is completely disabled. The UI shows "not available" |
| **Claude** | `CLAUDE_CODE_OAUTH_TOKEN` + `TM_REPO_PATH` pointing to source code | Bug reports still get created, but without AI analysis. The issue body has the player's description and game state, but no code-level analysis |

To get the OAuth token, run `claude setup-token` on any machine with Claude Code installed. It gives you a long-lived token (valid for about a year).

## Deployment guides

- [Docker](docker/README.md)
- [Kubernetes](kubernetes/README.md)
