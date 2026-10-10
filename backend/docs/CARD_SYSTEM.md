# Card System Architecture

This document explains the complete card effect system architecture in the Open Mars backend.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Principles](#architecture-principles)
3. [Directory Structure](#directory-structure)
4. [Card Data Model](#card-data-model)
5. [Effect Application Flow](#effect-application-flow)
6. [Validation System](#validation-system)
7. [Adding New Cards](#adding-new-cards)
8. [Examples](#examples)

## Overview

The card system is designed to be **JSON-driven**, meaning **90%+ of cards can be added by only editing JSON**, with no Go code changes required. The system uses a declarative behavior model where cards specify:
- **Triggers**: When the effect activates (auto, manual, conditional)
- **Inputs**: What the player must spend
- **Outputs**: What the player/game receives
- **Choices**: Multiple options for complex cards

## Architecture Principles

### 1. **Actions Mutate State**
- **ONLY** actions in `/internal/action/` may mutate game state
- Actions call game state methods: `player.Resources().AddCredits()`, `game.GlobalParameters().IncreaseTemperature()`
- All other code provides helpers, parses data, or subscribes to events

### 2. **Separation of Concerns**
- **`/internal/cards/`**: Card data outside game context (registry, JSON loading, JSON validation)
- **`/internal/game/cards/`**: Card-related game logic (behavior helpers, gameplay validation, NO state mutation)
- **`/internal/action/`**: Applies card effects to game state

### 3. **Event-Driven Passive Effects**
- Passive effects (conditional triggers) subscribe to domain events
- When event matches trigger condition, effect outputs are applied
- No manual polling or checking - fully event-driven

### 4. **Data-Driven, Not Code-Driven**
- Card behaviors defined in JSON
- Generic application logic handles all standard patterns
- Only truly unique card mechanics require custom code (rare)

## Directory Structure

```
backend/
├── assets/
│   └── cards.json    # Card definitions (manually edited)
├── internal/
│   ├── cards/                           # Card data (outside game)
│   │   ├── registry.go                  # Card lookup by ID
│   │   ├── loader.go                    # JSON loading
│   │   └── card_json_validator.go       # JSON structure validation
│   │
│   ├── game/
│   │   └── cards/                       # Card game logic (inside game)
│   │       ├── card_types.go            # Card data structures
│   │       ├── card_validator.go        # Gameplay validation
│   │       ├── behavior_helpers.go      # Trigger type checking
│   │       ├── corporation_processor.go # Corporation setup
│   │       └── effect_subscriber.go     # Event-driven passive effects
│   │
│   └── action/                          # State mutation
│       ├── play_card.go                 # Applies immediate effects
│       └── use_card_action.go           # Executes manual actions
```

## Card Data Model

### Card Structure

Cards are defined in `/backend/assets/cards.json`:

```json
{
  "id": "142",
  "name": "Deep Well Heating",
  "type": "automated",
  "cost": 13,
  "description": [
    {"type": "requirement", "text": "3 science tags."},
    {"type": "generic", "text": "Raise temperature 1 step. Increase energy production 1 step."}
  ],
  "tags": ["power", "building"],
  "requirements": {
    "tags": {"science": {"min": 3}}
  },
  "behaviors": [
    {
      "triggers": [{"type": "auto"}],
      "outputs": [
        {"type": "energy-production", "amount": 1, "target": "self-player"},
        {"type": "temperature", "amount": 1, "target": "none"}
      ]
    }
  ]
}
```

### Behavior Components

**Triggers** - When the behavior activates:
- `auto`: Immediately when card is played
- `manual`: Player-activated action (blue cards)
- `auto` with `condition`: Event-driven passive effect

**Inputs** - Resources player must spend:
```json
"inputs": [
  {"type": "energy", "amount": 4, "target": "self-player"}
]
```

**Outputs** - Effects applied to game:
```json
"outputs": [
  {"type": "steel", "amount": 2, "target": "self-player"},
  {"type": "oxygen", "amount": 1, "target": "none"}
]
```

**Output Types**:
- Resources: `credits`, `steel`, `titanium`, `plants`, `energy`, `heat`
- Production: `credits-production`, `steel-production`, `titanium-production`, `plants-production`, `energy-production`, `heat-production`
- Global Parameters: `temperature`, `oxygen`, `ocean`
- Other: `tr` (terraform rating), `cards` (draw cards), `tile-placement`

**Targets**:
- `self-player`: Affects the player who played the card
- `any-player`: Can affect any player (opponent targeting)
- `self-card`: Affects resource storage on the card itself
- `none`: No target (global parameters)

**Choices** - For cards with multiple options:
```json
"choices": [
  {
    "id": "choice-1",
    "description": "Gain 2 steel",
    "outputs": [{"type": "steel", "amount": 2}]
  },
  {
    "id": "choice-2",
    "description": "Gain 3 titanium",
    "outputs": [{"type": "titanium", "amount": 3}]
  }
]
```

## Effect Application Flow

### 1. Immediate Effects (Auto Trigger)

```
Player plays card
    ↓
PlayCardAction.Execute()
    ↓
Get card from CardRegistry
    ↓
Iterate card.Behaviors
    ↓
For each behavior with "auto" trigger:
    ↓
For each output in behavior:
    ↓
ACTION applies to game state:
  - player.Resources().AddCredits()
  - player.Production().IncreaseSteel()
  - game.GlobalParameters().IncreaseTemperature()
    ↓
Game methods publish domain events
    ↓
Broadcaster sends updates to clients
```

**Code Pattern**:
```go
// In PlayCardAction.Execute()
card := a.cardRegistry.GetCard(cardID)

for _, behavior := range card.Behaviors {
    if hasAutoTrigger(behavior) && !hasCondition(behavior) {
        // ACTION applies outputs
        for _, output := range behavior.Outputs {
            switch output.Type {
            case "credits":
                player.Resources().AddCredits(ctx, output.Amount)
            case "temperature":
                game.GlobalParameters().IncreaseTemperature(ctx, output.Amount)
            // ... all resource/production/global parameter types
            }
        }
    }
}
```

### 2. Manual Actions (Blue Cards)

```
Player plays blue card
    ↓
PlayCardAction.Execute()
    ↓
Extract behaviors with "manual" trigger
    ↓
Register CardAction to Player.Actions()
    ↓
[Later] Player clicks card action button
    ↓
UseCardAction.Execute()
    ↓
Get action from player.Actions().FindByCardID()
    ↓
Validate inputs available
    ↓
ACTION applies outputs to game state
    ↓
Events published → Broadcaster → Clients
```

**Code Pattern**:
```go
// When card is played
if hasManualTrigger(behavior) {
    player.Actions().Add(CardAction{
        CardID: card.ID,
        Behavior: behavior,
    })
}

// When player uses action (UseCardAction)
action := player.Actions().FindByCardID(cardID)

// Validate inputs
for _, input := range action.Behavior.Inputs {
    if !player.Resources().Has(input.Type, input.Amount) {
        return ErrInsufficientResources
    }
}

// Apply inputs (spend resources)
for _, input := range action.Behavior.Inputs {
    player.Resources().Subtract(ctx, input.Type, input.Amount)
}

// Apply outputs (gain resources, raise global parameters)
for _, output := range action.Behavior.Outputs {
    // ... same switch as immediate effects
}
```

### 3. Passive Effects (Conditional Triggers)

```
Player plays card with conditional trigger
    ↓
PlayCardAction.Execute()
    ↓
Extract behaviors with condition
    ↓
Register CardEffect to Player.Effects()
    ↓
[Later] Event occurs (city placed, greenery placed, etc.)
    ↓
Game publishes domain event (TilePlacedEvent)
    ↓
CardEffectSubscriber receives event
    ↓
Check all players' Effects() for matching triggers
    ↓
For matching effects:
    ↓
ACTION applies outputs to game state
    ↓
Events published → Broadcaster → Clients
```

**Example Card** - "Urbanized Area" (Gain 2 MC when any city is placed):
```json
{
  "behaviors": [{
    "triggers": [{
      "type": "auto",
      "condition": {
        "type": "city-placed",
        "location": "anywhere",
        "target": "any-player"
      }
    }],
    "outputs": [
      {"type": "credits", "amount": 2, "target": "self-player"}
    ]
  }]
}
```

**Code Pattern**:
```go
// CardEffectSubscriber
func (s *EffectSubscriber) OnTilePlaced(event TilePlacedEvent) {
    if event.TileType != "city" {
        return
    }

    game := s.gameRepo.Get(event.GameID)

    for _, player := range game.Players() {
        for _, effect := range player.Effects().List() {
            if matchesCityPlacedCondition(effect, event) {
                // Apply effect outputs via ACTION pattern
                for _, output := range effect.Behavior.Outputs {
                    switch output.Type {
                    case "credits":
                        player.Resources().AddCredits(ctx, output.Amount)
                    // ...
                    }
                }
            }
        }
    }
}
```

## Validation System

### Two-Level Validation

**1. JSON Structure Validation** (`card_json_validator.go`)
- Validates card JSON at load time
- Checks: Valid resource types, trigger types, condition types
- Ensures: Required fields present, amounts are reasonable
- **Location**: `/internal/cards/card_json_validator.go`
- **When**: Card JSON is loaded from file

**2. Gameplay Validation** (`card_validator.go`)
- Validates if card CAN BE PLAYED in game context
- Checks: Player has resources to pay cost, requirements met
- Requirements: Temperature, oxygen, ocean levels, tags, production
- **Location**: `/internal/game/cards/card_validator.go`
- **When**: Player attempts to play card (PlayCardAction)

### Validation Flow

```
Card JSON loaded
    ↓
card_json_validator validates structure
    ↓
If invalid: Error logged, card not added to registry
    ↓
If valid: Card added to CardRegistry
    ↓
[Later] Player attempts to play card
    ↓
card_validator validates gameplay requirements
    ↓
If requirements not met: Return error to player
    ↓
If valid: PlayCardAction applies effects
```

## Adding New Cards

### Standard Cards (90%+ of cases)

**Just edit JSON** - no Go code required!

1. Open `/backend/assets/cards.json`
2. Add card definition with behaviors
3. Run `just generate` to sync TypeScript types
4. Test by playing the card

**Example - Automated Card**:
```json
{
  "id": "999",
  "name": "Advanced Alloys",
  "type": "automated",
  "cost": 9,
  "description": [
    {"type": "generic", "text": "Increase steel production 1 step. Increase titanium production 1 step."}
  ],
  "tags": ["science"],
  "behaviors": [{
    "triggers": [{"type": "auto"}],
    "outputs": [
      {"type": "steel-production", "amount": 1, "target": "self-player"},
      {"type": "titanium-production", "amount": 1, "target": "self-player"}
    ]
  }]
}
```

**Example - Blue Card (Manual Action)**:
```json
{
  "id": "1000",
  "name": "Energy Converter",
  "type": "active",
  "cost": 5,
  "description": [
    {"type": "action", "text": "Spend 4 energy to gain 2 steel and raise oxygen 1 step."}
  ],
  "tags": ["power", "building"],
  "behaviors": [{
    "triggers": [{"type": "manual"}],
    "inputs": [
      {"type": "energy", "amount": 4, "target": "self-player"}
    ],
    "outputs": [
      {"type": "steel", "amount": 2, "target": "self-player"},
      {"type": "oxygen", "amount": 1, "target": "none"}
    ]
  }]
}
```

**Example - Passive Effect**:
```json
{
  "id": "1001",
  "name": "Architect Guild",
  "type": "automated",
  "cost": 8,
  "description": [
    {"type": "effect", "text": "Gain 2 M€ whenever you place a city tile."}
  ],
  "tags": ["building"],
  "behaviors": [{
    "triggers": [{
      "type": "auto",
      "condition": {
        "type": "city-placed",
        "location": "anywhere",
        "target": "self-player"
      }
    }],
    "outputs": [
      {"type": "credits", "amount": 2, "target": "self-player"}
    ]
  }]
}
```

### Complex Cards (5-10% of cases)

Some cards have unique mechanics not covered by the standard behavior system. These require custom code:

**Examples of complex cards**:
- **Ants**: Removes resources from other players' cards
- **Predators**: Complex targeting and resource stealing
- **Research**: Player draws cards and chooses which to keep
- **Special Tiles**: Tiles with unique adjacency bonuses

**When custom code is needed**:
1. Add behavior to JSON with `"custom": true` flag
2. Implement custom logic in action or helper function
3. Document the custom behavior

## Examples

### Example 1: Simple Resource Card

```json
{
  "id": "001",
  "name": "Supplier",
  "type": "automated",
  "cost": 10,
  "behaviors": [{
    "triggers": [{"type": "auto"}],
    "outputs": [
      {"type": "energy-production", "amount": 1},
      {"type": "credits", "amount": 4}
    ]
  }]
}
```

**Effect**: When played, increase energy production by 1 and gain 4 credits.

### Example 2: Blue Card with Action

```json
{
  "id": "002",
  "name": "Steelworks",
  "type": "active",
  "cost": 15,
  "behaviors": [{
    "triggers": [{"type": "manual"}],
    "inputs": [{"type": "energy", "amount": 4}],
    "outputs": [
      {"type": "steel", "amount": 2},
      {"type": "oxygen", "amount": 1}
    ]
  }]
}
```

**Effect**: Action - Spend 4 energy to gain 2 steel and raise oxygen 1 step.

### Example 3: Passive Effect

```json
{
  "id": "003",
  "name": "Rover Construction",
  "type": "automated",
  "cost": 8,
  "behaviors": [{
    "triggers": [{
      "type": "auto",
      "condition": {
        "type": "city-placed",
        "location": "anywhere",
        "target": "any-player"
      }
    }],
    "outputs": [{"type": "credits", "amount": 2}]
  }]
}
```

**Effect**: Gain 2 MC whenever any player places a city tile (including yourself).

### Example 4: Card with Requirements

```json
{
  "id": "004",
  "name": "Water Import From Europa",
  "type": "event",
  "cost": 25,
  "requirements": {
    "tags": {"jovian": {"min": 1}}
  },
  "behaviors": [{
    "triggers": [{"type": "auto"}],
    "outputs": [
      {"type": "ocean", "amount": 1},
      {"type": "tr", "amount": 1}
    ]
  }]
}
```

**Effect**: Requires 1 Jovian tag. Place an ocean tile and gain 1 TR.

### Example 5: Card with Choices

```json
{
  "id": "005",
  "name": "Designed Microorganisms",
  "type": "automated",
  "cost": 11,
  "behaviors": [{
    "triggers": [{"type": "auto"}],
    "choices": [
      {
        "id": "plant-production",
        "description": "Increase plant production 2 steps",
        "outputs": [{"type": "plants-production", "amount": 2}]
      },
      {
        "id": "add-microbes",
        "description": "Add 3 microbes to another card",
        "outputs": [{"type": "microbes", "amount": 3, "target": "other-card"}]
      }
    ]
  }]
}
```

**Effect**: Choose one: Increase plant production 2 steps OR add 3 microbes to another card.

## Summary

The card system is designed to be:
- **JSON-driven**: Most cards require only JSON edits
- **Type-safe**: Card structures validated at load and play time
- **Event-driven**: Passive effects trigger automatically
- **Maintainable**: Clear separation between data, helpers, and state mutation
- **Extensible**: Easy to add new cards, rare custom code needed

**Key takeaway**: To add a new card, edit JSON. The system handles the rest automatically!
