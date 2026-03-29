# Card Pack Tests

Tests are organized by the card's `pack` field in the card database JSON.

## File naming conventions

- **Project card tests:** `play_card_<pack>_test.go` (e.g., `play_card_base_test.go`, `play_card_colonies_test.go`)
- **Corporation tests:** `corporations_<pack>_test.go` (e.g., `corporations_colonies_test.go`)

## Rules

- All card tests go in the file matching their card pack. Do not create standalone files for individual cards.
- Check the card's `pack` field in `backend/assets/terraforming_mars_cards.json` to determine the correct file.
- All files use package `card_packs_test`.
