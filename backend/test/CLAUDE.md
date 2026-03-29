# Test Directory

Tests mirror the `internal/` structure. Place tests in the matching subdirectory:

- `action/behavior/` - Behavior condition tests by category (basic resource, production, tile placement, effects, etc.)
- `action/card_packs/` - Card and corporation integration tests, organized by card pack
- `action/card_effects/` - Passive effects, temporary effects, choice requirements
- `action/colony/` - Colony system tests (trade, build)
- `action/core/` - Core game mechanics (state calculator, variable amounts, generational events)
- `action/play_card/` - Card play flow tests (any-player targeting, card play action)
- `action/tiles/` - Tile placement bonus and special tile tests
- `action/payment/` - Payment modifier tests (discounts, value modifiers)
- `game/` - Game state and domain logic tests
- `events/` - Event system tests

Use `testutil/` helpers for game setup and assertions.

## Behavior Tests (`action/behavior/`)

One file per condition category. Tests exercise each field on the category struct using constructed behaviors (not card-based). See `internal/game/shared/CLAUDE.md` for the category definitions.

| File | Category | Fields Tested |
|------|----------|---------------|
| `basic_resource_test.go` | BasicResourceCondition | PaymentAllowed, TargetRestriction, steal/any-player targets |
| `production_test.go` | ProductionCondition | Per (tag scaling, integer division), any-player reduction |
| `tile_placement_test.go` | TilePlacementCondition | TileRestrictions (adjacency, onTileType, boardTags), TileType |
| `global_parameter_test.go` | GlobalParameterCondition | Temperature/oxygen/venus increases, TR with Per |
| `card_operation_test.go` | CardOperationCondition | Optional discard, all-opponents draw |
| `card_storage_test.go` | CardStorageCondition | Self-card/any-card targeting, steal-from-any-card, Selectors |
| `effect_test.go` | EffectCondition | Discount selectors, payment substitute, value modifier, lenience |
| `colony_test.go` | ColonyCondition | Colony placement, AllowDuplicatePlayerColony |
| `tile_modification_test.go` | TileModificationCondition | Tile destruction, replacement with TileType |
| `misc_test.go` | MiscCondition | Extra actions, bonus tags with Per |
| `per_condition_test.go` | Per field (cross-cutting) | City tile, tag, card resource, colony count, location |
