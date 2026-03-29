# Assets - Static Game Data

Contains card definitions and static game data loaded by the backend at startup.

## Card Database

Authoritative source for all card definitions: corporations, project cards, and prelude cards with behaviors, costs, requirements, and effects.

### Card Structure

```json
{
  "id": "B07",
  "name": "PhoboLog",
  "type": "corporation",
  "cost": 0,
  "description": "Effect description text",
  "pack": "base-game",
  "tags": ["space"],
  "behaviors": [...]
}
```

**Card Types:** `corporation`, `active`, `automated`, `event`, `prelude`

**Card Packs:** `base-game`, `corporate-era`, `prelude`, `venus-next`, `colonies`, `turmoil`

### Behavior System

Each card has a `behaviors` array. Each behavior contains:
- `triggers`: When the behavior activates
- `inputs`: Resources consumed (costs)
- `outputs`: Resources produced (effects)
- `choices`: Alternative options (A OR B)

The JSON uses a flat format with a `type` field that determines which fields are valid. At load time, each condition is categorized into a typed Go struct (see `internal/game/shared/CLAUDE.md`).

### Trigger Types

| Trigger | Description |
|---------|-------------|
| `auto` | Applies immediately when card is played |
| `auto-corporation-start` | Applies once when corporation is selected |
| `manual` | Player-activated action (blue cards) |
| `auto` + `condition` | Passive effect triggered by game events |

### Output Type Categories

Each output `type` belongs to a category that determines which fields are valid:

| Category | Types | Valid Fields |
|----------|-------|-------------|
| Basic Resource | `credit`, `steel`, `titanium`, `plant`, `energy`, `heat` | `per`, `variableAmount`, `targetRestriction`, `maxTrigger` |
| Production | `credit-production` ... `heat-production`, `any-production` | `per`, `variableAmount` |
| Tile Placement | `city-placement`, `greenery-placement`, `ocean-placement`, `volcano-placement`, `tile-placement`, `land-claim` | `tileRestrictions`, `tileType` |
| Global Parameter | `temperature`, `oxygen`, `ocean`, `venus`, `tr`, `global-parameter` | `per` |
| Card Operation | `card-draw`, `card-take`, `card-peek`, `card-buy`, `card-discard` | `selectors`, `variableAmount` |
| Card Storage | `microbe`, `animal`, `floater`, `science`, `asteroid`, `fighter`, `disease`, `card-resource` | `selectors`, `per`, `variableAmount` |
| Effect | `discount`, `payment-substitute`, `value-modifier`, `defense`, `global-parameter-lenience`, `ignore-global-requirements`, `ocean-adjacency-bonus`, `action-reuse`, `effect`, `tag` | `selectors`, `temporary` |
| Colony | `colony-tile`, `colony-count`, `colony-bonus`, `colony-track-step` | `allowDuplicatePlayerColony` |
| Tile Modification | `tile-destruction`, `tile-replacement` | `tileType` |
| Misc | `extra-actions`, `bonus-tags`, `world-tree-tile`, `award-fund`, `trade` | `per`, `selectors` |

All types share: `type`, `amount`, `target`.

## Adding New Cards

1. Add card JSON to `terraforming_mars_cards.json`
2. Use existing output types and fields from the table above
3. Run `make test` — the validation test checks all field combinations are valid
4. For new resource types, see `internal/game/shared/CLAUDE.md` for how to add a category

Most cards (90%+) can be added via JSON only without Go code changes.
