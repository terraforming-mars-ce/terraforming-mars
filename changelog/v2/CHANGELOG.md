## Added
- Venus Next as a per-game setting: a clickable 3D Venus with its own
  tiles, a Venus gauge, camera travel between Mars and Venus, and the
  Venus corporations.
- Preludes. Game start is split into corporation, prelude and starting
  card steps, with corporation and prelude effects applied in order.
- Claude bot players: convert a seat to a bot, and an MCP server for
  playing from Claude Code.
- Spectators, in-game chat and a player colour picker.
- The host can kick players and end the game.
- An experimental card pack with The World Tree and Front of the Line.
- Global parameter track bonuses and the Flooding card's steal.
- Space and Enter hotkeys, alternating background music, and flowers
  and birds on living greenery.
- Redesigned cards with descriptions, and a redesigned admin tools
  window.

## Fixed
- The Mars board matches the official Tharsis layout.
- Greenery placed by cards must be next to your own tiles.
- 13 cards were missing their negative VP conditions.
- Event card tags no longer count toward tag requirements.
- The ocean limit adjusts when other tiles take ocean spaces.
- The deck is shuffled at game creation.
- Ocean placement triggers passive effects such as Arctic Algae.
- Card costs that were modelled as negative outputs are now inputs,
  and cards with negative outputs check that you can afford them.
- Action-phase-only operations are rejected during production.
- Unselected preludes are removed from the game.
- Audio unlocks on the first click or tap.
- Spectators see standard projects, milestones and awards.
- The lobby keeps a stable player order.

## Build
- npm is replaced by bun.
- Large assets are stored in Git LFS.
- A Raspberry Pi deploy script builds versioned images.
