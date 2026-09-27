# Game State Management - Critical Rules

This document contains critical reminders for Claude Code when working with game state in `/internal/game/`.

## Broadcasting Rule

**CRITICAL**: Every state mutation MUST result in a broadcast to WebSocket clients.

### Current Pattern (Manual Broadcasting)

When any state changes, we currently publish TWO events:
1. **Domain event** - Business logic event (e.g., `TemperatureChangedEvent`, `ResourcesChangedEvent`)
2. **BroadcastEvent** - Signals to Broadcaster to send game state updates to clients

```go
// Example from global_parameters.go
events.Publish(gp.eventBus, events.TemperatureChangedEvent{
    GameID:    gp.gameID,
    OldValue:  oldTemp,
    NewValue:  gp.temperature,
    ChangedBy: ctx.Value("playerID").(string),
    Timestamp: time.Now(),
})

events.Publish(gp.eventBus, events.BroadcastEvent{
    GameID: gp.gameID,
})
```

### Problem with Current Pattern

- **50+ duplicate BroadcastEvent calls** across the codebase
- **Easy to forget** - New state changes might not broadcast
- **Not DRY** - Same BroadcastEvent code repeated everywhere
- **Maintenance burden** - Every state method needs two Publish calls

### Proposed Solution: Automatic Broadcasting

**Idea**: Modify `EventBus.Publish()` to automatically broadcast game state after ANY event is published.

**Implementation Options:**

#### Option A - Simplest (Recommended)

Modify `internal/events/event_bus.go` to inject gameID at creation and auto-broadcast:

```go
type EventBusImpl struct {
    subscriptions map[SubscriptionID]*subscription
    nextID        uint64
    mutex         sync.RWMutex
    logger        *zap.Logger
    gameID        string              // NEW: Game ID for auto-broadcasting
    broadcaster   BroadcastFunc       // NEW: Callback for broadcasting
}

type BroadcastFunc func(gameID string, playerIDs []string)

func NewEventBus(gameID string, broadcaster BroadcastFunc) *EventBusImpl {
    return &EventBusImpl{
        subscriptions: make(map[SubscriptionID]*subscription),
        nextID:        1,
        logger:        logger.Get(),
        gameID:        gameID,
        broadcaster:   broadcaster,
    }
}

func Publish[T any](eb *EventBusImpl, event T) {
    eb.mutex.RLock()
    defer eb.mutex.RUnlock()

    // Execute event handlers (existing logic)
    eventType := fmt.Sprintf("%T", event)
    // ... execute handlers ...

    // NEW: Auto-broadcast after event processing
    if eb.broadcaster != nil {
        // Extract PlayerIDs if event has them (e.g., BroadcastEvent pattern)
        var playerIDs []string
        if broadcastable, ok := any(event).(interface{ GetPlayerIDs() []string }); ok {
            playerIDs = broadcastable.GetPlayerIDs()
        }
        eb.broadcaster(eb.gameID, playerIDs)
    }
}
```

**Benefits:**
- ✅ Remove ALL manual `BroadcastEvent` publishing
- ✅ Impossible to forget broadcasting
- ✅ EventBus per game already - gameID known at creation
- ✅ Cleaner state methods - only publish domain events

#### Option B - Configurable

Add a flag to enable/disable auto-broadcasting:

```go
type EventBusImpl struct {
    // ... existing fields ...
    autoBroadcast bool
}

func NewEventBus(gameID string, broadcaster BroadcastFunc, autoBroadcast bool) *EventBusImpl {
    // ...
}
```

**Benefits:**
- Allows opt-out for internal-only events if needed
- More flexible but adds complexity

#### Option C - Event Metadata

Create an interface that events can implement:

```go
type BroadcastableEvent interface {
    ShouldBroadcast() bool
    GetPlayerIDs() []string  // nil = all players
}
```

Only auto-broadcast for events implementing this interface.

**Benefits:**
- Fine-grained control per event type
- Most complex, possibly over-engineered

### Recommendation

**Use Option A** because:
1. All current state changes already broadcast (checked 50+ instances)
2. Real-time multiplayer game = all state changes ARE broadcasted
3. Simplest implementation, easiest to maintain
4. Per-game EventBus already provides perfect isolation

### Migration Steps

If implementing automatic broadcasting:

1. Modify `EventBus` to accept gameID and broadcaster callback during creation
2. Update `EventBus.Publish()` to auto-broadcast after event handlers execute
3. Update `Game.NewGame()` to pass gameID when creating EventBus
4. Remove all manual `events.Publish(eventBus, BroadcastEvent{...})` calls (50+ instances)
5. Remove `BroadcastEvent` type (no longer needed as explicit event)
6. Test thoroughly - every state change must still broadcast

### Current State Management Locations

State mutation happens in these packages:
- `/internal/game/game.go` - Game-level state (status, phase, generation)
- `/internal/game/global_parameters/` - Temperature, oxygen, oceans
- `/internal/game/board/` - Tile placement
- `/internal/game/player/` - Player resources, cards, production
- `/internal/game/deck/` - Deck operations

All of these currently publish `BroadcastEvent` after state changes.

## Architecture Context

### EventBus Per Game

Each `Game` instance has its own `EventBus`:
```go
type Game struct {
    id       string
    eventBus *events.EventBusImpl  // Per-game event bus
    // ...
}
```

This means:
- ✅ No cross-game event pollution
- ✅ GameID is known at EventBus creation time
- ✅ Perfect isolation for automatic broadcasting

### Broadcaster Pattern

The `Broadcaster` subscribes to events and sends WebSocket updates:

```go
// internal/delivery/websocket/broadcaster.go
func (b *Broadcaster) OnBroadcastEvent(event events.BroadcastEvent) {
    game, _ := b.gameRepo.Get(ctx, event.GameID)

    // Send personalized game state to each player
    for _, playerID := range playerIDs {
        gameDto := dto.ToGameDto(game, b.cardRegistry, playerID)
        b.hub.SendToPlayer(game.ID(), playerID, message)
    }
}
```

With automatic broadcasting, this would become a callback injected into EventBus.

## Summary

**Remember**: Any time you add new state management, it MUST broadcast to clients. The proposed automatic broadcasting pattern eliminates the manual burden and prevents mistakes.
