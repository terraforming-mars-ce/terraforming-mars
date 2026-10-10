# Deploying Open Mars

One container, a reverse proxy in front. That's it.

## The container

`ghcr.io/openmars-app/openmars` is one Go server on port 3001. It runs the game logic and the bug report API, holds the WebSocket connections, and serves the React app. Everything is on one origin, so there's nothing to point the frontend at.

The image is built from the repo root with the root `Dockerfile`.

## Reverse proxy

Send every path for your hostname to the container on port 3001. TLS usually ends at the proxy.

WebSocket connections (`/ws`) stay open for the duration of a game, so set proxy read/send timeouts to something high (3600s works). If your proxy drops idle connections after 60 seconds, players will get disconnected constantly.

## Environment variables

| Variable | Required | Default | What it does |
|----------|----------|---------|--------------|
| `OPENMARS_LOG_LEVEL` | No | `info` | Log verbosity. Options: `debug`, `info`, `warn`, `error` |
| `OPENMARS_ADDR` | No | `:3001` | Address the server listens on, e.g. `:3000` |
| `OPENMARS_WEB_DIR` | No | `web` | Directory with the built frontend, relative to the working directory. The image has it at `/app/web`. When the directory is missing, only the API is served |
| `OPENMARS_SERVER_ALIAS` | No | `local` | This server's permanent id on a gateway that fronts several servers (lowercase a-z, 0-9, -). It appears in shared links |
| `OPENMARS_SERVER_NAME` | No | the alias | Display name on that gateway, e.g. `EU 1` |

The game works fine without any environment variables. The bug report feature has its own set of optional env vars below.

### Bug reporting

All optional. Without these, the game runs normally but the in-game bug report button shows "not available".

| Variable | Required | Default | What it does |
|----------|----------|---------|--------------|
| `GITHUB_APP_ID` | No | _(unset)_ | GitHub App ID for creating bug report issues |
| `GITHUB_INSTALLATION_ID` | No | _(unset)_ | GitHub App installation ID. Without this, bug reporting is disabled entirely |
| `GITHUB_PRIVATE_KEY_PATH` | No | `./private-key.pem` | Path to the GitHub App private key file. Must be readable inside the container |
| `GITHUB_REPO_OWNER` | No | `openmars-app` | GitHub repo owner where issues get created |
| `GITHUB_REPO_NAME` | No | `openmars` | GitHub repo name where issues get created |
| `OPENMARS_REPO_PATH` | No | _(unset)_ | Path to source code for Claude analysis. The Docker image already sets this to `/repo`, so you only need to set it when running locally (e.g. `OPENMARS_REPO_PATH=./` from the repo root) |
| `CLAUDE_CODE_OAUTH_TOKEN` | No | _(unset)_ | OAuth token for Claude Code CLI. Without this (and without a mounted `~/.claude` directory), Claude analysis is skipped. Bug reports still get created -- they just won't have AI-generated code analysis |

The backend logs which capabilities are active on startup:

```
Bug report service initialized  {"github_app": true, "claude": true}
```

If something is misconfigured you'll see `false` and a warning explaining why.

## Bug report capabilities

The bug report feature has two independent capabilities that degrade gracefully:

| Capability | What you need | What happens without it |
|------------|---------------|------------------------|
| **GitHub App** | `GITHUB_INSTALLATION_ID` + private key file | Bug reporting is completely disabled. The UI shows "not available" |
| **Claude** | `CLAUDE_CODE_OAUTH_TOKEN` + `OPENMARS_REPO_PATH` pointing to source code | Bug reports still get created, but without AI analysis. The issue body has the player's description and game state, but no code-level analysis |

To get the OAuth token, run `claude setup-token` on any machine with Claude Code installed. It gives you a long-lived token (valid for about a year).

## Deployment guides

- [Docker](docker/README.md)
- [Kubernetes](kubernetes/README.md)
