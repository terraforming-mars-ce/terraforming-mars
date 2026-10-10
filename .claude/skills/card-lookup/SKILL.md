---
name: card-lookup
description: Look up Terraforming Mars card data and corporation card data. Use this skill when the user asks about specific cards, card effects, card costs, corporation abilities, or needs to find cards by type, tag, or pack. Provides jq commands for efficient card data retrieval without loading the entire JSON file.
---

# Card Lookup

This skill provides efficient access to the Terraforming Mars card database.

## Card Data Location

The card database is located at:
```
backend/assets/cards.json
```

This file contains 453 cards and should NOT be read directly (too large). Always use `jq` commands to query specific data.

## Card Structure

Each card has the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier (e.g., "001", "B01" for corporations) |
| `name` | string | Card name |
| `type` | string | One of: `automated`, `active`, `event`, `corporation`, `prelude` |
| `cost` | number | Megacredit cost to play |
| `description` | array | Ordered `{type, text}` sections; `type` is `generic`, `effect`, `action`, or `requirement`. Labels are generated, so `text` has no prefix |
| `pack` | string | Expansion pack the card belongs to |
| `tags` | array | Card tags (optional) |
| `requirements` | array | Play requirements (optional) |
| `behaviors` | array | Card effects/triggers (optional) |
| `vpConditions` | array | Victory point conditions (optional) |
| `resourceStorage` | object | Resource storage info (optional) |

### Card Types
- `automated` - One-time effect cards (green border)
- `active` - Ongoing effect cards (blue border)
- `event` - One-time event cards (red border)
- `corporation` - Starting corporation cards
- `prelude` - Prelude expansion cards

### Tags
Available tags: `animal`, `building`, `city`, `earth`, `jovian`, `microbe`, `plant`, `power`, `science`, `space`, `venus`, `wild`

### Packs
Available packs: `base-game`, `corporate-era`, `prelude`, `venus-next`, `colonies`, `turmoil`, `promo`

## jq Commands

### Find a card by name (case-insensitive)
```bash
jq '.[] | select(.name | test("Mining"; "i"))' backend/assets/cards.json
```

### Find a card by exact name
```bash
jq '.[] | select(.name == "Mining Rights")' backend/assets/cards.json
```

### Find a card by ID
```bash
jq '.[] | select(.id == "001")' backend/assets/cards.json
```

### List all corporations
```bash
jq '[.[] | select(.type == "corporation")] | .[] | {id, name, description}' backend/assets/cards.json
```

### Find corporation by name
```bash
jq '.[] | select(.type == "corporation") | select(.name | test("Ecoline"; "i"))' backend/assets/cards.json
```

### Find cards by type
```bash
jq '[.[] | select(.type == "event")] | length' backend/assets/cards.json  # Count
jq '[.[] | select(.type == "event")] | .[0:5]' backend/assets/cards.json  # First 5
```

### Find cards by tag
```bash
jq '[.[] | select(.tags != null) | select(.tags | contains(["science"]))]' backend/assets/cards.json
```

### Find cards by pack
```bash
jq '[.[] | select(.pack == "venus-next")] | .[] | {name, type, cost}' backend/assets/cards.json
```

### Find cards by cost range
```bash
jq '[.[] | select(.cost >= 20 and .cost <= 30)] | .[] | {name, cost, type}' backend/assets/cards.json
```

### Find cards with specific requirements
```bash
jq '[.[] | select(.requirements != null) | select(.requirements[] | .type == "oxygen")]' backend/assets/cards.json
```

### List all card names (quick reference)
```bash
jq '[.[] | .name] | sort' backend/assets/cards.json
```

### Get card summary (name, type, cost, tags)
```bash
jq '.[] | select(.name | test("Birds"; "i")) | {name, type, cost, tags, description}' backend/assets/cards.json
```

## Usage Guidelines

1. Always use `jq` to query the card database - never read the entire file
2. For partial name matches, use `test("pattern"; "i")` for case-insensitive search
3. Pipe results to `head` or use array slicing `.[0:N]` when expecting many results
4. Use `{field1, field2}` projection to limit output to relevant fields
