# Backend - Terraforming Mars API Server

This document provides guidance for working with the backend API server.

## Overview

Go-based REST and WebSocket API server implementing the Terraforming Mars board game logic. Provides real-time multiplayer game state synchronization and enforces game rules.

## Go Coding Standards

**IMPORTANT**: This backend follows idiomatic Go practices and community standards. For comprehensive Go coding guidelines, see **[go.instructions.md](./go.instructions.md)**.

Key standards include:
- Follow Effective Go, Go Code Review Comments, and Google's Go Style Guide
- Write simple, clear, and idiomatic Go code
- Use proper naming conventions (mixedCaps, avoid underscores)
- Check all errors immediately
- Keep the happy path left-aligned (minimize indentation, return early)
- Document all exported symbols
- Use `gofmt` and `goimports` for formatting
- **CRITICAL**: Each `.go` file must have exactly ONE `package` declaration
- **CRITICAL**: NO unnecessary comments - code should be self-documenting. Avoid logic comments that just restate what the code does. Only add comments when truly necessary (complex algorithms, non-obvious business rules, or required doc comments for exported symbols).

For detailed guidance on naming, error handling, concurrency, API design, testing, and more, consult `go.instructions.md`.

## Server Restart Policy

**CRITICAL**: NEVER restart the backend server yourself. ALWAYS ask the user if you think a restart is needed.

- **Normal Mode**: User runs `make backend` or `make run` with **Air hot reload** - server automatically restarts on code changes
- **Watch Mode Active**: Code changes (Go files, JSON assets) trigger instant automatic reload
- **No Manual Restarts**: You should NEVER execute restart commands
- **If Restart Seems Needed**: Ask user "Should I restart the backend?" (they'll confirm or explain why it's not needed)

The user's development environment handles all server lifecycle management. Your role is to write code, not manage processes.

## Architecture

### Clean Architecture Layers

The backend follows clean architecture principles with strict separation of concerns:

**Domain Layer** (`internal/game/`)

- Core business entities: Game (containing all state), Player, GlobalParameters, Board, Deck
- Subpackages: player/, board/, deck/, shared/, global_parameters/
- Value objects in shared/: Resources, Production, Tile, HexPosition
- Domain events defined in `internal/events/`
- Private fields with public accessor methods
- Zero external dependencies

**Action Layer** (`internal/action/`)

- Single-responsibility actions executing business logic (~100-200 lines each)
- BaseAction provides common dependencies (GameRepository, CardRegistry, logger)
- Main actions modify game state (JoinGameAction, PlayCardAction)
- Query actions for read operations (GetGameAction, ListGamesAction)
- Admin actions for game management (SetResourcesAction)
- Depends on domain types via GameRepository

**Infrastructure Layer** (`internal/game/`)

- GameRepository manages collection of active games
- Game contains all state: Players, Board, Deck, GlobalParameters
- State methods publish events via injected EventBus
- Private fields enforce encapsulation
- No separate subdomain repositories - all accessed via Game

**Presentation Layer** (`internal/delivery/`)

- HTTP endpoints delegate to actions (`http/`)
- WebSocket handlers delegate to actions (`websocket/`)
- DTOs for external communication (`dto/`)
- Request/response mapping
- Depends on action layer, not repositories directly

**Card System** (`internal/cards/` + `internal/game/cards/`)

- `internal/cards/`: Card registry and JSON loader
- `internal/game/cards/`: Card behavior logic, validation, requirement checking (NO state mutation)

**Bug Report Service** (`internal/service/bugreport/` + `internal/delivery/http/bugreport_handler.go`)

- Submits bug reports to GitHub Issues via `POST /api/v1/bugs`
- Status and individual report retrieval via `GET /api/v1/bugs/status` and `GET /api/v1/bugs/{id}`

**Event System** (`internal/events/`)

- Type-safe event bus for pub/sub
- Domain event definitions (TemperatureChanged, ResourcesChanged, TilePlaced, etc.)
- CardEffectSubscriber for passive card effects
- Event-driven architecture decoupling actions from effects

### Directory Structure

```
backend/
├── cmd/
│   └── server/            # Main server with dependency injection
├── internal/
│   ├── action/            # Business logic actions (ONLY place for state mutation)
│   ├── cards/             # Card registry and loader
│   ├── delivery/          # Presentation layer
│   │   ├── dto/           # Data Transfer Objects and mappers
│   │   ├── http/          # HTTP handlers
│   │   └── websocket/     # WebSocket hub, handlers, broadcaster
│   ├── events/            # Event bus and domain events
│   ├── game/              # Game state and domain types
│   │   ├── board/         # Board and Tile types
│   │   ├── cards/         # Card behavior logic (NO state mutation)
│   │   ├── deck/          # Deck management
│   │   ├── global_parameters/  # Temperature, oxygen, oceans
│   │   ├── player/        # Player entity and components
│   │   └── shared/        # Shared types (Resources, HexPosition, etc.)
│   ├── logger/            # Structured logging
│   └── middleware/        # HTTP middleware
├── test/                  # Test suite (mirrors internal/ structure)
├── assets/                # Static game data (JSON card definitions)
└── docs/                  # Architecture documentation
```

## Development Workflow

### Running the Server

```bash
# From project root
make backend              # Hot reload via Air (port 3001)
make run                  # Run both frontend and backend

# Direct commands (from backend/)
go run cmd/server/main.go
air                       # Hot reload with Air
```

### Testing

```bash
# From project root
make test                 # Run all backend tests
make test-verbose         # Detailed test output
make test-coverage        # Generate coverage report

# From backend/
go test ./test/...        # All tests
go test ./test/action/    # Specific package
go test -json ./test/...  # JSON output for parsing
```

**Test Location**: Tests live in `test/` directory, mirroring `internal/` structure. Example: `test/action/confirm_production_cards_test.go` tests `internal/action/confirm_production_cards.go`.

### Code Quality

```bash
# From project root
make lint-backend         # Run Go formatting
make format               # Format all code

# From backend/
make format               # Run gofmt
go fmt ./...              # Direct formatting
```

### Type Generation

Generate TypeScript types for frontend consumption:

```bash
# From project root
make generate             # Generate types from Go structs

# From backend/
tygo generate             # Direct tygo command
```

Add `ts:` tags to structs for type generation:

```go
type Player struct {
    ID       string `json:"id" ts:"string"`
    Credits  int    `json:"credits" ts:"number"`
}
```

## Key Development Patterns

### Adding New Game Operations

1. **Create action** in `internal/action/`:
   - Extend `BaseAction` struct
   - Implement `Execute()` method with clear parameters
   - Validate inputs and call session repositories
   - Return explicit result type or error

2. **Create WebSocket handler** (if needed) in `internal/delivery/websocket/handler/`:
   - Parse incoming WebSocket message
   - Call the action's Execute() method
   - SessionManager handles broadcasting

3. **Create HTTP handler** (if needed) in `internal/delivery/http/`:
   - Parse HTTP request
   - Call the action's Execute() method
   - Map result to DTO and respond

4. **Add message/request types** to frontend types if needed

### Implementing Card Effects

**For passive effects** (event-driven):

1. Define behavior in card JSON with triggers and outputs
2. Ensure Game state methods publish relevant domain events
3. CardEffectSubscriber automatically subscribes on card play
4. No manual action code needed for passive effects

See `docs/EVENT_SYSTEM.md` for complete event system documentation.

**For immediate effects**:

1. Implement logic in card effect handler
2. Call via action when card is played
3. Action updates Game via state methods
4. Game publishes events, Broadcaster sends updates to clients

### Game Repository Pattern

- **Single Source of Truth**: Game contains all state (Players, Board, Deck, GlobalParameters)
- **Encapsulation**: Private fields with public accessor methods
- **Event Integration**: State methods automatically publish domain events
- **GameRepository**: Manages collection of active Game instances
- **Access Pattern**: `game := gameRepo.Get(gameID)` → `player := game.GetPlayer(playerID)`

### Action Layer Rules

- **Single Responsibility**: Each action performs ONE operation (~100-200 lines)
- **Extend BaseAction**: Use common dependencies (GameRepository, CardRegistry, logger)
- **Actions do only what they say**: Don't manually check for passive card effects
- **Call Game state methods**: Game methods publish events automatically
- **Broadcaster handles WebSocket updates**: Subscribes to BroadcastEvent
- **Event system handles passive effects**: CardEffectSubscriber triggers effects via events

### State Ownership and Encapsulation

The architecture follows clear ownership boundaries for game state:

**Game Repository Owns:**
- Game-wide state (status, phase, generation, current turn)
- Player-specific phase state (ProductionPhase, SelectStartingCardsPhase, PendingCardSelection, PendingCardDrawSelection, PendingTileSelection, PendingTileSelectionQueue, ForcedFirstAction)
- Global parameters (temperature, oxygen, oceans)
- Game configuration and settings

**Player Repository Owns:**
- Player identity (ID, name, gameID)
- Corporation selection
- Cards (hand and played cards)
- Resources, production, terraform rating, victory points
- Turn state (passed, available actions, connection status)
- Player effects, actions, requirement modifiers

**Why Phase State Lives in Game:**
- Phase state is transient - exists only during specific game phases
- Game controls phase transitions and needs atomic access to all players' phase states
- Cleaner separation: Player represents persistent player state, Game manages workflow state
- Simplifies phase transition logic (e.g., checking if all players completed starting selection)

**Access Pattern:**
```go
// ✅ CORRECT: Access phase state via Game
game, _ := gameRepo.Get(gameID)
productionPhase := game.GetProductionPhase(playerID)
game.SetProductionPhase(ctx, playerID, phase)

// ❌ WRONG: Phase state not on Player
player := game.GetPlayer(playerID)
productionPhase := player.ProductionPhase() // This method doesn't exist
```

## Data Flow

### WebSocket Message Flow

```
Client → WebSocket Connection → Hub.HandleMessage()
                                       ↓
                                 Manager.RouteMessage()
                                       ↓
                             WebSocket Handler.Handle()
                                       ↓
                                  Action.Execute()
                                       ↓
                            Game State Updates + Events
                                       ↓
                           EventBus → Broadcaster
                                       ↓
                              All Clients Updated
```

### Game State Synchronization

1. Action performs business logic via Execute() method
2. Game state methods update state and publish events
3. EventBus notifies subscribers (Broadcaster, passive effects, etc.)
4. Broadcaster automatically sends updates on BroadcastEvent
5. Broadcaster fetches complete game state from GameRepository
6. Clients receive personalized state update via WebSocket

## Type System Integration

### Go to TypeScript

DTO structs in `internal/delivery/dto/` generate TypeScript interfaces via `tygo generate`.

**Tag behavior:**
- `json:"fieldName"` — controls the TypeScript field name
- `json:"...,omitempty"` — makes the field optional (`?`)
- `tstype:"'a' | 'b'"` — overrides the generated TypeScript type (used for discriminated unions)
- `//tygo:emit <code>` — emits literal TypeScript before a struct (used for union type aliases)
- `ts:"..."` — **ignored by tygo**, do not use

### Behavior Condition System

Card behaviors use a discriminated union pattern. See `internal/game/shared/CLAUDE.md` for the full type system.

The DTO layer maps typed Go conditions to category-specific DTO structs (`BasicResourceConditionDto`, `EffectConditionDto`, etc.) which generate a TypeScript discriminated union (`ResourceCondition`) with per-category literal types on the `type` field.

### Keeping Types in Sync

1. Modify Go structs in `internal/game/` or subpackages
2. Update corresponding DTOs in `internal/delivery/dto/`
3. Update DTO mappers in `internal/delivery/dto/mapper_*.go`
4. Run `make generate` from project root
5. Frontend automatically gets updated types

## Important Notes

### Event-Driven Architecture

**CRITICAL**: Actions should NOT manually check for passive card effects. The event system handles this automatically:

```go
// ✅ CORRECT
func (a *ConvertHeatToTemperatureAction) Execute(...) {
    game.GlobalParameters().IncreaseTemperature(ctx, steps)  // Publishes TemperatureChangedEvent
    // CardEffectSubscriber automatically triggers passive effects
    return result, nil
}

// ❌ WRONG
func (a *ConvertHeatToTemperatureAction) Execute(...) {
    game.GlobalParameters().IncreaseTemperature(ctx, steps)
    // Don't manually loop through cards to check effects
    for _, card := range player.PlayedCards().Cards() { ... }
    return result, nil
}
```

### State Management

- No timeouts or arbitrary delays
- Implement proper event handling
- Design deterministic state transitions
- Use proper synchronization when needed

### Event-Driven Patterns for Game Logic

Use the EventBus for "do this when that happens" scenarios. Instead of checking conditions in actions or polling for state changes, subscribe to domain events and react automatically. This pattern is used for generational event tracking, passive card effects, and state synchronization.

**Pattern: Subscribe to Domain Events in Game Initialization**

For tracking or reacting to game state changes, subscribe in the Game constructor:

```go
func NewGame(...) *Game {
    g := &Game{...}
    g.subscribeToGameEvents()
    return g
}

func (g *Game) subscribeToGameEvents() {
    // Track TR raises for generational events
    events.Subscribe(g.eventBus, func(e events.TerraformRatingChangedEvent) {
        if e.NewRating > e.OldRating {
            p, err := g.GetPlayer(e.PlayerID)
            if err != nil {
                return
            }
            p.GenerationalEvents().Increment(shared.GenerationalEventTRRaise)
        }
    })

    // Track tile placements
    events.Subscribe(g.eventBus, func(e events.TilePlacedEvent) {
        p, err := g.GetPlayer(e.PlayerID)
        if err != nil {
            return
        }
        switch e.TileType {
        case "ocean":
            p.GenerationalEvents().Increment(shared.GenerationalEventOceanPlacement)
        case "city":
            p.GenerationalEvents().Increment(shared.GenerationalEventCityPlacement)
        case "greenery":
            p.GenerationalEvents().Increment(shared.GenerationalEventGreeneryPlacement)
        }
    })

    // Clear per-generation state on generation advance
    events.Subscribe(g.eventBus, func(e events.GenerationAdvancedEvent) {
        for _, p := range g.GetAllPlayers() {
            p.GenerationalEvents().Clear()
        }
    })
}
```

**When to Use Event Subscriptions**

✅ Tracking player activities within a generation (TR raises, tile placements)
✅ Resetting per-generation state when generation advances
✅ Triggering passive card effects when game state changes
✅ Broadcasting state updates to clients

**When NOT to Use Event Subscriptions**

❌ Validating requirements before an action (use state calculator instead)
❌ Computing costs or discounts (use RequirementModifierCalculator)
❌ Direct action-to-action communication (use the action layer pattern)

### State Calculator Pattern

The state calculator (`internal/action/state_calculator.go`) determines whether cards, card actions, and standard projects are available to the player. It computes errors, costs, and metadata for each entity, which the frontend uses to enable/disable UI elements.

**When to Extend the State Calculator**

Extend the state calculator when adding new requirements that must be validated before an action can be taken:
- New global parameter checks (like Venus track)
- New player state requirements (like generational events)
- New resource or production requirements
- New tile placement availability checks

**Adding a New Requirement Type**

1. Define error codes in `internal/game/player/state_error_codes.go`:

```go
const (
    ErrorCodeMyNewRequirement StateErrorCode = "my-new-requirement-not-met"
)
```

2. Create a validation function in `state_calculator.go`:

```go
func validateMyNewRequirement(
    behavior shared.CardBehavior,
    p *player.Player,
) []player.StateError {
    // Return empty slice if no requirements to check
    if len(behavior.MyNewRequirements) == 0 {
        return nil
    }

    var errors []player.StateError
    for _, req := range behavior.MyNewRequirements {
        // Validate against player state
        if !req.IsSatisfied(p) {
            errors = append(errors, player.StateError{
                Code:     player.ErrorCodeMyNewRequirement,
                Category: player.ErrorCategoryRequirement,
                Message:  formatMyNewRequirementError(req),
            })
        }
    }
    return errors
}
```

3. Call the validation function from the appropriate calculator:

```go
// For card actions (manual triggers on played cards)
func CalculatePlayerCardActionState(...) player.EntityState {
    // ... existing validations ...
    errors = append(errors, validateMyNewRequirement(behavior, p)...)
    // ...
}

// For playing cards from hand
func CalculatePlayerCardState(...) player.EntityState {
    // ... existing validations ...
    errors = append(errors, validateMyNewRequirement(card, p, g)...)
    // ...
}

// For standard projects
func CalculatePlayerStandardProjectState(...) player.EntityState {
    // ... existing validations ...
}
```

**Existing Validation Functions**

| Function | Purpose | Used By |
|----------|---------|---------|
| `validatePhase` | Check game phase is action phase | Card play |
| `validateNoActiveTileSelection` | Block actions during tile selection | Cards, actions, projects |
| `validateAffordabilityWithSubstitutes` | Check resource costs with Helion-style substitutes | Card play |
| `validateRequirements` | Check global parameters, tags, resources, production | Card play |
| `validateProductionOutputs` | Check player has production for negative outputs | Card play |
| `validateTileOutputs` | Check board has available tile placements | Card play |
| `validateBehaviorTileOutputs` | Check tile availability for action outputs | Card actions |
| `validateGenerationalEventRequirements` | Check generational events (TR raise, tile placement) | Card actions |

**Error Categories**

Use appropriate error categories for proper frontend display:

- `ErrorCategoryPhase` - Wrong game phase
- `ErrorCategoryTurn` - Not the player's turn
- `ErrorCategoryCost` - Insufficient resources to pay costs
- `ErrorCategoryRequirement` - Game requirement not met (temperature, tags, etc.)
- `ErrorCategoryAvailability` - Required resource/placement not available
- `ErrorCategoryAchievement` - Milestone/award already claimed/funded
- `ErrorCategoryInput` - Insufficient input resources for action
- `ErrorCategoryConfiguration` - Invalid configuration (unknown project type, etc.)

### Logging Guidelines

Production-quality logging. Every log line should earn its place at its level.

**General Rules**
- **No emoji** in log messages — keep logs clean and professional
- **No "successfully"** — if it's logged at Info, success is implied. Write "Game created" not "Game created successfully"
- Log messages should be concise, descriptive, and use sentence case (e.g., "Game created")
- No decorative prefixes, symbols, or formatting in messages
- Use structured fields (zap key-value pairs) for IDs, counts, and contextual data — not string interpolation
- **No duplicate logging** — don't log the same error at multiple layers (action + handler + middleware). Let the outermost layer handle it

**Info Level — One per operation**
- Each action/operation gets exactly **one** Info log: the **final success summary**
- Examples: "Game created successfully", "Card played successfully", "Player joined game successfully"
- Also appropriate for: server start/stop, phase transitions (entering action phase, game ended)
- **Never** log intermediate steps at Info (e.g., "Removing card from hand", "Deducting payment", "Processing behaviors")

**Debug Level — Everything else**
- Intermediate action steps (card removed from hand, payment deducted, behaviors processed)
- "Attempting to..." or "Starting..." logs that precede the operation
- Resource changes (added credits, increased production, etc.)
- Corporation/card processing internals
- HTTP request/response logging for successful requests
- WebSocket connection established/closed
- Query operations (get game, list cards, list games)
- Handler entry/exit, message routing
- Initialization details, configuration, setup steps
- Bot/service internals (health checks, dispatching)

**Warn Level**
- Expected failure cases that callers should handle (game not found, player not found)
- Invalid client input that was rejected
- Recoverable errors or degraded functionality

**Error Level**
- Unexpected failures: database errors, serialization failures, internal invariant violations
- Errors that indicate bugs or system problems, not user mistakes

**Pattern for new actions:**
```go
func (a *MyAction) Execute(ctx context.Context, ...) (*Result, error) {
    a.logger.Debug("Starting operation", zap.String("game_id", gameID))

    // ... intermediate steps, all logged at Debug ...
    a.logger.Debug("Intermediate step completed", zap.Int("count", n))

    // One Info log at the end
    a.logger.Info("Operation completed successfully", zap.String("game_id", gameID))
    return result, nil
}
```

## Testing Guidelines

### Test Organization

- Tests mirror `internal/` structure in `test/`
- Use table-driven tests for multiple scenarios
- Mock external dependencies via interfaces
- Test business logic in isolation

### Test Patterns

```go
func TestPlayerService_DoAction(t *testing.T) {
    tests := []struct {
        name     string
        input    Input
        expected Expected
        wantErr  bool
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Common Tasks

### Adding a New Domain Type

1. Create struct in `internal/game/` or subpackage with `json:` and `ts:` tags
2. Add private fields with public accessor methods
3. Add Game methods to access/modify the new type
4. Create action in `internal/action/` extending BaseAction
5. Add DTOs in `internal/delivery/dto/`
6. Add mappers in `internal/delivery/dto/mapper.go`
7. Run `make generate` to sync TypeScript types

### Adding a New Game Rule

1. Check `TERRAFORMING_MARS_RULES.md` in project root
2. Define types in `internal/game/` or subpackages
3. Create action in `internal/action/` with validation logic
4. Update Game methods if new state access needed
5. Create tests in `test/action/`
6. Update relevant HTTP or WebSocket handlers to call action

### Debugging

- Use structured logger for consistent output
- Check `EventBus` for event flow
- Inspect WebSocket messages in browser DevTools
- Use `go test -json` for parseable test output

## Dependencies

### Core Libraries

- **gorilla/websocket**: WebSocket communication
- **go-chi/chi**: HTTP routing and middleware
- **tygo**: TypeScript type generation

### Development Tools

- **Air**: Hot reload for development

## Related Documentation

- **Project Root CLAUDE.md**: Project overview, commands, and cross-cutting workflows
- **frontend/CLAUDE.md**: Frontend architecture, components, and patterns
- **assets/CLAUDE.md**: Card database documentation (behavior types, output formats)
- **docs/EVENT_SYSTEM.md**: Event-driven architecture and broadcasting
- **TERRAFORMING_MARS_RULES.md**: Complete game rules reference
