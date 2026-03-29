# Test Directory

Tests mirror the `internal/` structure. Place tests in the matching subdirectory:

- `action/card_packs/` - Card and corporation tests, organized by card pack
- `action/card_effects/` - Card effect application tests (resource, steal, etc.)
- `action/colony/` - Colony system tests (trade, build)
- `action/core/` - Core game mechanics
- `action/play_card/` - Generic card play tests (any-player targeting, etc.)
- `action/tiles/` - Tile placement tests
- `game/` - Game state and domain logic tests
- `events/` - Event system tests

Use `testutil/` helpers for game setup and assertions.
