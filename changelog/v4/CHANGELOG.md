## Added
- JSON-driven map system with ten Mars maps and host-side selection in
  the lobby.
- Final greenery phase: convert remaining plants before scoring.
- Lobby settings overlay covering map, packs, max players,
  dev/demo/random-buy toggles, and Claude token.
- "Buy 1 Random" production-card option.
- In-game card browser overlay accessible from the hamburger menu.
- Hellas and Elysium expansion milestones and awards, picked randomly
  per game.
- Aridor, Arklight, and Stormcraft corporations.
- Storage payment substitutes (Stormcraft floaters pay for heat
  conversions).
- Endgame Score-projection graph and Global Parameters as % completed.
- Tile bonuses for production, oxygen, temperature, and ocean
  placement.
- Merged single-pass ocean renderer with capsule SDF and adjacent-tile
  water bodies.
- Counter-based production phase animation with sound.
- Colony count requirements on cards.
- Admin Tools entry in the dev-mode hamburger menu.

## Changed
- App phase consolidated into a single discriminated-union AppPhase
  store; transitionStore removed.
- SpaceBackground stays mounted across menu, lobby, and game.
- Game settings edited from the lobby via update-game-settings instead
  of at create time.
- TotalVP computed server-side and shared via SnapshotEnricher.
- colony-tile renamed to colony; Colonies component extracted from
  Game.
- Random buy draws from the deck rather than the available pool.
- PlayerSelector simplified from multi-select to single-select
  dropdown.

## Fixed
- Production-phase animation rewritten on a counter-based state
  machine.
- Ocean lighting computed in local space so highlights stay correct as
  Mars orbits.
- 3D hit detection on celestial bodies uses a sphere hitbox with hover
  cursor.
- Colony marker position floored to colony count so trade rewards stay
  correct.
- Per-colony VP scoring counts colonies via the new Colonies
  component.
- Destroy/steal effects can be skipped when no opponent holds the
  resource.
- Endgame graph view sits below the top menu bar so SCORE/GRAPHS stay
  clickable.
- Poseidon's init phase waits for colony selection before
  auto-confirming.
- World Tree VP counts only adjacent greeneries.
- Wild tag tooltips on game cards and tags popover.

## Build
- Build workflows trigger on v* tag pushes.
