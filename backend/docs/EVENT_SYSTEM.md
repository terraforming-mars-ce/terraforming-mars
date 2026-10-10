# Event System Architecture

## Overview

The Open Mars backend uses an event-driven architecture to decouple game state changes from WebSocket broadcasting and passive card effects. This document explains how the event system works and how to use it correctly.

## Core Concepts

### Event Types

The system uses two categories of events:

**1. Domain Events** - Represent specific game state changes
```go
type TemperatureChangedEvent struct {
    GameID   string
    OldValue int
    NewValue int
    Steps    int
}

type ResourcesChangedEvent struct {
    GameID    string
    PlayerID  string
    Resources shared.Resources
}
```

**2. BroadcastEvent** - Signals that clients need game state updates
```go
type BroadcastEvent struct {
    GameID    string
    PlayerIDs []string  // nil = broadcast to all players
}
```

### EventBus

Type-safe publish-subscribe system for all events:

```go
type EventBusImpl struct {
    subscribers map[SubscriptionID]*subscription
    mu          sync.RWMutex
}

// Subscribe to specific event type
func Subscribe[T any](eb *EventBusImpl, handler EventHandler[T]) SubscriptionID

// Publish event to all subscribers
func Publish[T any](eb *EventBusImpl, event T)
```

## Architecture Flow

### Complete Message Flow

```
1. Client Action
   └─> WebSocket Hub receives message

2. Handler Delegation
   └─> Manager routes to WebSocket Handler
       └─> Handler calls Action.Execute()

3. State Mutation
   └─> Action calls Game state methods
       └─> Game methods update private fields
           └─> Game publishes domain events (TemperatureChanged, ResourcesChanged, etc.)
               └─> Game publishes BroadcastEvent

4. Event Broadcasting
   └─> EventBus notifies all subscribers
       ├─> Broadcaster receives BroadcastEvent
       │   └─> Fetches game state from GameRepository
       │       └─> Creates personalized DTO for each player
       │           └─> Sends to WebSocket clients
       │
       └─> CardEffectSubscriber receives domain events
           └─> Triggers passive card effects
               └─> Effects may update game state (loop back to step 3)
```

### Visual Diagram

```
┌────────────────────────────────────────────────────────────┐
│ ACTION LAYER                                                │
│ ┌────────────────────────────────────────────────────────┐ │
│ │ Action.Execute()                                       │ │
│ │   ├─> Validate inputs                                 │ │
│ │   ├─> Fetch game from GameRepository                  │ │
│ │   └─> Call game state methods                         │ │
│ └────────────────────────────────────────────────────────┘ │
└─────────────────────────┬──────────────────────────────────┘
                          │
                          ▼
┌────────────────────────────────────────────────────────────┐
│ DOMAIN LAYER                                                │
│ ┌────────────────────────────────────────────────────────┐ │
│ │ Game State Method (e.g., IncreaseTemperature)         │ │
│ │   ├─> Lock mutex                                      │ │
│ │   ├─> Update private fields                           │ │
│ │   ├─> Unlock mutex                                    │ │
│ │   ├─> Publish domain event (TemperatureChanged)       │ │
│ │   └─> Publish BroadcastEvent                          │ │
│ └────────────────────────────────────────────────────────┘ │
└─────────────────────────┬──────────────────────────────────┘
                          │
                          ▼
┌────────────────────────────────────────────────────────────┐
│ EVENT BUS                                                   │
│ Notifies all subscribers of published events               │
└─────────────┬─────────────────────────────┬────────────────┘
              │                             │
              ▼                             ▼
┌─────────────────────────┐   ┌─────────────────────────────┐
│ BROADCASTER             │   │ CARD EFFECT SUBSCRIBER      │
│ ├─> On BroadcastEvent   │   │ ├─> On TemperatureChanged   │
│ ├─> Fetch game state    │   │ ├─> Find passive effects    │
│ ├─> Create DTOs         │   │ └─> Trigger effect actions  │
│ └─> Send to clients     │   └─────────────────────────────┘
└─────────────────────────┘
```

## Implementation Patterns

### Pattern 1: Game State Method with Events

All Game state mutation methods follow this pattern:

```go
func (g *Game) IncreaseTemperature(ctx context.Context, steps int) {
    // CRITICAL: Capture values while holding lock
    var oldTemp, newTemp int

    g.mu.Lock()
    oldTemp = g.globalParameters.temperature
    g.globalParameters.temperature += steps
    newTemp = g.globalParameters.temperature
    g.mu.Unlock()

    // CRITICAL: Publish events AFTER releasing lock
    // This prevents deadlocks

    // 1. Publish domain event for passive card effects
    events.Publish(g.eventBus, events.TemperatureChangedEvent{
        GameID:   g.id,
        OldValue: oldTemp,
        NewValue: newTemp,
        Steps:    steps,
    })

    // 2. Publish broadcast event for client updates
    events.Publish(g.eventBus, events.BroadcastEvent{
        GameID:    g.id,
        PlayerIDs: nil,  // nil = all players
    })
}
```

### Pattern 2: Broadcaster Subscriber

The Broadcaster subscribes to BroadcastEvent and sends personalized game state:

```go
// In internal/delivery/websocket/broadcaster.go
func (b *Broadcaster) OnBroadcastEvent(event events.BroadcastEvent) {
    game, err := b.gameRepo.Get(event.GameID)
    if err != nil {
        b.logger.Error("Failed to get game", "gameID", event.GameID, "error", err)
        return
    }

    if event.PlayerIDs == nil {
        // Broadcast to all players - each gets personalized view
        for _, player := range game.Players() {
            dto := mapper.ToPersonalizedGameDTO(game, player.ID())
            b.sendToPlayer(event.GameID, player.ID(), dto)
        }
    } else {
        // Send to specific players
        for _, playerID := range event.PlayerIDs {
            dto := mapper.ToPersonalizedGameDTO(game, playerID)
            b.sendToPlayer(event.GameID, playerID, dto)
        }
    }
}
```

### Pattern 3: Action Publishing Events

Actions SHOULD NOT manually publish BroadcastEvent - Game methods do this automatically:

```go
// ✅ CORRECT
func (a *StandardProjectAsteroidAction) Execute(
    ctx context.Context,
    gameID string,
    playerID string,
) error {
    game, _ := a.gameRepo.Get(gameID)
    player := game.GetPlayer(playerID)

    // Validate
    if player.Resources().Credits() < 14 {
        return errors.New("insufficient credits")
    }

    // Update state - Game methods publish events automatically
    player.Resources().SubtractCredits(14)
    game.GlobalParameters().IncreaseTemperature(ctx, 1)

    return nil  // Events already published by state methods
}

// ❌ WRONG - Don't manually publish BroadcastEvent
func (a *SomeAction) Execute(ctx context.Context, gameID, playerID string) error {
    game, _ := a.gameRepo.Get(gameID)

    // Update state
    player.Resources().AddCredits(10)

    // ❌ WRONG: Don't do this - state methods already publish
    events.Publish(a.eventBus, events.BroadcastEvent{
        GameID: gameID,
    })

    return nil
}
```

## Personalized DTOs

Each player receives a personalized view of the game state:

```go
func ToPersonalizedGameDTO(game *game.Game, receivingPlayerID string) *GameDTO {
    dto := &GameDTO{
        ID:     game.ID(),
        Status: string(game.Status()),
        // ... other game fields
    }

    // Current player gets full data
    currentPlayer := game.GetPlayer(receivingPlayerID)
    dto.Player = &PlayerDTO{
        ID:    currentPlayer.ID(),
        Name:  currentPlayer.Name(),
        Hand:  currentPlayer.Hand().Cards(),  // Full hand visible
        // ... all player data
    }

    // Other players get limited data
    dto.OtherPlayers = []*OtherPlayerDTO{}
    for _, p := range game.Players() {
        if p.ID() != receivingPlayerID {
            dto.OtherPlayers = append(dto.OtherPlayers, &OtherPlayerDTO{
                ID:       p.ID(),
                Name:     p.Name(),
                HandSize: len(p.Hand().Cards()),  // Count only, no cards
                // ... limited data
            })
        }
    }

    return dto
}
```

## Thread Safety

### Critical Rules

1. **Never publish events while holding a lock**
   ```go
   // ❌ WRONG
   g.mu.Lock()
   g.temperature++
   events.Publish(g.eventBus, SomeEvent{})  // Deadlock risk!
   g.mu.Unlock()

   // ✅ CORRECT
   g.mu.Lock()
   g.temperature++
   newTemp := g.temperature
   g.mu.Unlock()

   events.Publish(g.eventBus, TemperatureChangedEvent{
       NewValue: newTemp,
   })
   ```

2. **Capture values before releasing lock**
   ```go
   g.mu.Lock()
   oldValue := g.someField
   g.someField = newValue
   capturedValue := g.someField
   g.mu.Unlock()

   // Use captured values in events
   events.Publish(g.eventBus, SomeEvent{
       OldValue: oldValue,
       NewValue: capturedValue,
   })
   ```

3. **EventBus is thread-safe**
   - Subscribe/Publish can be called concurrently
   - Internal mutex protects subscriber map

## Event Catalog

### Game State Events

```go
type GenerationChangedEvent struct {
    GameID   string
    OldValue int
    NewValue int
}

type PhaseChangedEvent struct {
    GameID   string
    OldPhase game.GamePhase
    NewPhase game.GamePhase
}

type TurnChangedEvent struct {
    GameID        string
    OldPlayerID   string
    NewPlayerID   string
    GenerationEnd bool
}
```

### Global Parameter Events

```go
type TemperatureChangedEvent struct {
    GameID   string
    OldValue int
    NewValue int
    Steps    int
}

type OxygenChangedEvent struct {
    GameID   string
    OldValue int
    NewValue int
    Steps    int
}

type OceansChangedEvent struct {
    GameID   string
    OldValue int
    NewValue int
    Count    int
}
```

### Player Events

```go
type ResourcesChangedEvent struct {
    GameID    string
    PlayerID  string
    Resources shared.Resources
}

type ProductionChangedEvent struct {
    GameID     string
    PlayerID   string
    Production shared.Production
}

type CardPlayedEvent struct {
    GameID   string
    PlayerID string
    CardID   string
}
```

### Broadcast Events

```go
type BroadcastEvent struct {
    GameID    string
    PlayerIDs []string  // nil = all players, otherwise specific players
}
```

## Passive Card Effects

Card effects subscribe to domain events via CardEffectSubscriber:

```go
// Example: "Gain 2 MC when temperature is raised"
type TemperatureBonusEffect struct {
    gameRepo game.GameRepository
}

func (e *TemperatureBonusEffect) OnTemperatureChanged(event events.TemperatureChangedEvent) {
    game, _ := e.gameRepo.Get(event.GameID)

    // Find players with cards granting temperature bonuses
    for _, player := range game.Players() {
        for _, cardID := range player.PlayedCards().Cards() {
            // Check if card grants temperature bonus
            if hasTemperatureBonus(cardID) {
                player.Resources().AddCredits(2 * event.Steps)
            }
        }
    }
}
```

## Best Practices

### DO

✅ Call Game state methods and let them publish events
✅ Subscribe to specific event types you care about
✅ Publish domain events for game state changes
✅ Publish BroadcastEvent from Game state methods
✅ Release locks before publishing events
✅ Capture values while holding lock for event payloads

### DON'T

❌ Manually publish BroadcastEvent from Actions
❌ Publish events while holding a lock
❌ Directly access private fields without mutex
❌ Create circular event dependencies
❌ Publish events without capturing necessary data

## Debugging Events

### Logging Events

```go
// In Game state method
func (g *Game) SomeStateChange() {
    // ... state update ...

    g.logger.Debug("Publishing domain event",
        "event", "SomeEvent",
        "gameID", g.id,
        "details", someDetails,
    )

    events.Publish(g.eventBus, SomeEvent{...})
}
```

### Tracing Event Flow

1. Check action execution logs
2. Verify Game state method was called
3. Confirm event was published
4. Check Broadcaster received BroadcastEvent
5. Verify DTO was created and sent

## Common Patterns

### Pattern: Targeted Broadcast

Send updates only to specific players:

```go
// Only notify the player who drew cards
events.Publish(g.eventBus, events.BroadcastEvent{
    GameID:    g.id,
    PlayerIDs: []string{playerID},
})
```

### Pattern: Chained Events

One event can trigger actions that publish more events:

```go
// Temperature increase triggers passive effects
events.Publish(g.eventBus, TemperatureChangedEvent{...})
  └─> CardEffectSubscriber
      └─> Triggers passive effect
          └─> Updates resources
              └─> Publishes ResourcesChangedEvent{...}
                  └─> Publishes BroadcastEvent{...}
```

### Pattern: Multiple Events

State methods can publish multiple events:

```go
func (g *Game) PlaceTile(ctx context.Context, tile Tile) {
    // ... update board ...

    // Publish domain event
    events.Publish(g.eventBus, TilePlacedEvent{
        GameID:   g.id,
        Position: tile.Position,
        Type:     tile.Type,
    })

    // If tile placement raised global parameter
    if tile.RaisesOxygen {
        events.Publish(g.eventBus, OxygenChangedEvent{...})
    }

    // Always broadcast state change
    events.Publish(g.eventBus, BroadcastEvent{
        GameID: g.id,
    })
}
```

## Summary

The event system provides:
- **Decoupling**: Actions don't know about WebSocket broadcasting
- **Extensibility**: Add new event subscribers without changing actions
- **Thread Safety**: Clear patterns for lock management
- **Traceability**: Events document all state changes

Key principle: **Game state methods are the single source of truth for events**. Actions update state, Game publishes events, subscribers react.
