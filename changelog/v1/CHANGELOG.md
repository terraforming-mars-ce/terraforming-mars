## Added
- 3D tiles for the mining tile (cave entrance and animated cart), the
  reserved area fence, living greenery and the Nuclear Zone with a
  mushroom cloud on placement.
- Starting card selection on a single screen.
- Milestone and award popovers show every player's progress and
  scores.
- The lobby sets the maximum number of players, up to 10.
- In-game bug reporting to GitHub with Claude analysis.
- A turn banner, stable player colours and tile tooltips.

## Fixed
- Triggered effect notifications reach every player.
- The settings button works in the lobby and waiting phases.
- Placing a tile next to oceans pays 2 M€ per adjacent ocean.
- Discarded project cards are recycled back into the deck.
- Game logs survive a page reload, and the log popover scrolls to the
  newest entry.
- The first player follows the turn order.
- Resource removal, TR gains from global parameters and the Mining
  Rights passive effect.
- Spectate mode layout and player tile colours.

## Build
- nginx proxies the API and WebSocket in production, supports relative
  API URLs and never caches runtime-config.js.
