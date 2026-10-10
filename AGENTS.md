# CLAUDE.md

Open Mars is a digital board game based on Terraforming Mars, with real-time multiplayer and 3D game view. WebSocket multiplayer with Go backend and React frontend.

## Commands

Tasks run through [just](https://just.systems). The root `justfile` loads three modules: `backend/justfile`, `frontend/justfile` and `proxy/justfile` (the gateway). Run `just` to list every recipe. Call a module recipe as `just backend test` or `just frontend lint`; inside `backend/` or `frontend/`, plain `just <recipe>` runs that module's recipe.

### Development
```bash
just deps          # Install Go modules and frontend dependencies
just dev           # Run backend (3001, Air hot reload) and frontend (3000) together
just dev --proxy   # Same, plus the gateway (4000) in front of the frontend
just backend dev   # Backend only
just frontend dev  # Frontend only
just kill          # Stop stray dev servers on 3000/3001/4000
```

### Testing
```bash
just test                                          # Run all backend tests
just backend test -v -run TestKick ./test/action/... # Args replace the default ./test/... pattern
just backend coverage                              # Coverage report (backend/coverage.html)
just backend test-race                             # Race detector
```

### Code Quality
```bash
just check         # Everything CI runs; never modifies files
just lint          # go vet + errcheck + golangci-lint, and oxlint
just format        # Format Go and TypeScript
just generate      # Generate TypeScript types from Go structs (formats the output)
```

Go tools (air, tygo, errcheck, golangci-lint) are pinned in `backend/go.mod` and `backend/golangci-lint.mod` and run through `go tool`, so nothing needs installing globally.

### Pre-Commit

**CRITICAL**: Before any `git add`, `git commit`, `git push`, or creating a PR, run:

```bash
just prepare-for-commit
```

It formats, regenerates types, then runs `just check` — the same checks CI runs. Fix all errors before proceeding with git operations, including before `gh pr create`.

### Build
```bash
just build         # Build the backend binary and the frontend bundle
just clean         # Remove build output
```

### Releases

Release notes live in `changelog/<tag>/CHANGELOG.md`, the only place they are written. The GitHub release body and the in-game changelog (`GET /api/v1/changelog`) both read that file; tag messages carry no notes.

1. Add `changelog/vX.Y.Z/CHANGELOG.md` in a PR and merge it.
2. On an up-to-date `main`, run `just release vX.Y.Z`. It checks the changelog, creates the tag and pushes it.
3. The Release workflow fails before building anything if the changelog is missing or malformed, then drafts the GitHub release from it and pushes the images.

Format (enforced by `backend/internal/changelog`, `just changelog-check vX.Y.Z` and a backend test over every file):
- An optional one-paragraph intro.
- `## Security`, `## Added`, `## Changed`, `## Fixed`, `## Build`, in that order, each optional and never empty.
- `- ` bullets; wrapped lines are indented two spaces. Inline `code` is the only markup.

## Adding New Game Features

1. **Define domain types** in `backend/internal/game/` with `json:` and `ts:` tags
2. **Create action** in `backend/internal/action/` extending BaseAction
3. **Wire handlers** (HTTP or WebSocket) to delegate to action
4. **Generate types**: Run `just generate`
5. **Frontend**: Import generated types, implement UI
6. **Check**: Run `just prepare-for-commit`

## Type Safety Bridge

Go structs generate TypeScript interfaces via `tygo`. Use `tstype:` tags to override generated types (e.g., string literal unions for discriminated unions). Note: `ts:` tags are **ignored** by tygo.

```bash
just generate     # After any Go type changes
```

See `backend/CLAUDE.md` for Go type tagging and `frontend/CLAUDE.md` for consuming generated types.

## Important Notes

### Development Workflow
- Both servers run with hot reload (`just dev`)
- Type generation: Go changes → `just generate` → React implementation
- State flow: All changes originate from Go backend via WebSocket
- No client-side game logic (prevents desync)

### Frontend App-Phase State Machine

The frontend tracks "what screen the user is on" in a single discriminated-union `AppPhase` (`frontend/src/stores/appPhaseStore.ts`) rather than composing multiple boolean / enum flags across stores. Only `useGameInitialization` (bootstrapping) and `useGameTransitions` (lifecycle) write to it; components read `phase.kind` and selector helpers. See `frontend/CLAUDE.md` § "App Phase State Machine" for the legal transitions, the persistent `<SpaceBackground>` invariant, and the rules around resetting world-ready flags between games.

### Mobile Layout
Phones use a separate compact layout (landscape-only in-game). Any UI change must keep it working: see `frontend/CLAUDE.md` § "Mobile / compact layout".

### Test Creation
- Always write tests for new backend features
- Test files go in `backend/test/` directory

### Card Database
See `backend/assets/CLAUDE.md` for card behavior documentation.

### Hex Coordinate System
Cube coordinates (q, r, s) where q + r + s = 0. Utilities in `frontend/src/utils/`.

### Energy/Power Reference
When working with energy, it's referenced as `power.png` in assets.

### Bots
In-app bot players live in `backend/internal/service/bot/`. A bot plays through typed MCP tools served in-process on a loopback port (no shell access): an Opus planner thinks ahead between turns, Sonnet plays the turn, and Haiku writes reactions and the end-game recap. Bot game access runs on the WebSocket hub goroutine via `Hub.Do`. Personas and the strategy guide are in `backend/assets/bot/`.

## Important Instruction Reminders

- Do what has been asked; nothing more, nothing less
- NEVER create files unless absolutely necessary
- ALWAYS prefer editing existing files over creating new ones
- NEVER proactively create documentation files (only when explicitly requested)
- **No backwards compatibility, ever.** Always break APIs forward — never keep deprecated fields, message types, shims, or compatibility branches. Move only forward. If a name, route, or message type changes, delete the old one outright
- Write tests for new backend features
- **NEVER use deprecated code or comments** - Remove deprecated fields, functions, and comments entirely
- **NEVER push directly to main** - Always create a separate feature branch and open a pull request

## Active Technologies
- Go 1.21+ (backend), TypeScript 5.x (frontend) + gorilla/websocket, chi router, React 18, Tailwind CSS v4 (001-generational-events)
- In-memory game state (no persistence required for generational events) (001-generational-events)
- Go 1.21+, TypeScript 5.x + None (custom diff computation) (001-game-state-repo)
- In-memory (map-based, per-game isolation) (001-game-state-repo)
- TypeScript 5.x (React 18) + React 18, Tailwind CSS v4, existing BehaviorSection component system (001-full-card-view)
- N/A (no persistence needed) (001-full-card-view)
- TypeScript 5.x (React 18), Go 1.21+ (backend DTO change) + React 18, Tailwind CSS v4, existing BehaviorSection component system (001-full-card-view)

## Recent Changes
- 001-generational-events: Added Go 1.21+ (backend), TypeScript 5.x (frontend) + gorilla/websocket, chi router, React 18, Tailwind CSS v4
