# CLAUDE.md

Digital implementation of Terraforming Mars board game with real-time multiplayer and 3D game view. WebSocket multiplayer with Go backend and React frontend.

## Quick Start

```bash
make run         # Run both frontend (3000) and backend (3001) with hot reload
make help        # Show all available commands
```

## Essential Commands

### Development
```bash
make frontend    # React dev server (port 3000)
make backend     # Go backend with Air hot reload (port 3001)
make dev-setup   # Set up environment (go mod tidy + bun install)
```

### Testing
```bash
make test         # Run all backend tests
make test-verbose # Verbose test output
make test-coverage# Generate coverage report
```

### Code Quality
```bash
make lint         # Run all linters (Go fmt + oxlint)
make format       # Format all code (Go + TypeScript)
make generate     # Generate TypeScript types from Go structs
```

### Pre-Commit

**CRITICAL**: Before any `git add`, `git commit`, `git push`, or creating a PR, run:

```bash
make prepare-for-commit
```

Fix all errors before proceeding with git operations. **Always run this before `make pr` or `gh pr create`.**

### Build
```bash
make build        # Build both frontend and backend
make clean        # Clean build artifacts
```

## Adding New Game Features

**CRITICAL**: Always check `docs/TERRAFORMING_MARS_RULES.md` first for any task involving game mechanics, rules, or card effects.

1. **Define domain types** in `backend/internal/game/` with `json:` and `ts:` tags
2. **Create action** in `backend/internal/action/` extending BaseAction
3. **Wire handlers** (HTTP or WebSocket) to delegate to action
4. **Generate types**: Run `make generate`
5. **Frontend**: Import generated types, implement UI
6. **Format and lint**: Run `make format` and `make lint`

## Type Safety Bridge

Go structs generate TypeScript interfaces via `tygo`. Use `tstype:` tags to override generated types (e.g., string literal unions for discriminated unions). Note: `ts:` tags are **ignored** by tygo.

```bash
make generate     # After any Go type changes
```

See `backend/CLAUDE.md` for Go type tagging and `frontend/CLAUDE.md` for consuming generated types.

## Important Notes

### Development Workflow
- Both servers run with hot reload (`make run`)
- Type generation: Go changes → `make generate` → React implementation
- State flow: All changes originate from Go backend via WebSocket
- No client-side game logic (prevents desync)

### Frontend App-Phase State Machine

The frontend tracks "what screen the user is on" in a single discriminated-union `AppPhase` (`frontend/src/stores/appPhaseStore.ts`) rather than composing multiple boolean / enum flags across stores. Only `useGameInitialization` (bootstrapping) and `useGameTransitions` (lifecycle) write to it; components read `phase.kind` and selector helpers. See `frontend/CLAUDE.md` § "App Phase State Machine" for the legal transitions, the persistent `<SpaceBackground>` invariant, and the rules around resetting world-ready flags between games.

### Test Creation
- Always write tests for new backend features
- Test files go in `backend/test/` directory

### Card Database
See `backend/assets/CLAUDE.md` for card behavior documentation.

### Hex Coordinate System
Cube coordinates (q, r, s) where q + r + s = 0. Utilities in `frontend/src/utils/`.

### Energy/Power Reference
When working with energy, it's referenced as `power.png` in assets.

### MCP Server
An MCP server at `mcp-server/` lets Claude Code play the game via WebSocket. Run `make mcp-setup` to install, then restart Claude Code to pick up `.mcp.json`. Exposes tools: `connect_to_game`, `get_game_state`, `play_card`, `use_card_action`, `standard_project`, `convert_resources`, `skip_action`, `select_tile`, `select_starting_choices`, `confirm_cards`, `claim_milestone`, `fund_award`, `start_game`, `wait_for_turn`.

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
