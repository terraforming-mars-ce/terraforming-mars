Highlights since v5.

## Added
- Mobile: the game is playable on phones in landscape. A phone layout
  with a top bar, resource and status rails and a bottom dock opens
  full-screen Hand, Actions, Projects, Tableau, Players and Log/Chat
  screens. Cards scale as one unit at any height.
- Mobile: touch controls for the board. Pinch to zoom, two-finger pan,
  tap a hex for its tooltip, and select-then-confirm tile placement
  with the camera framing the valid hexes.
- Mobile: phone layouts for card-play prompts, selection overlays,
  starting selection (in steps), production, the corporation showcase,
  the end game, the lobby and the menu pages, plus a slide-in menu
  drawer. Phones in portrait get a rotate prompt.
- Mobile: an install bar on the landing page offers installing the
  game as an app (Chrome install dialog, or Add to Home Screen steps
  on iOS and Firefox), with an Install app item in the menus.
- Mobile: a low graphics tier for phones and weak GPUs with smaller
  planet textures, a procedural starfield and fewer effects, and
  recovery after WebGL context loss.
- Bots: in-app bot players run through MCP with personas and spend
  limits. Bots react in chat, remember rivals, write an end-game recap,
  and can be retried by the host. A bot inspector shows plans, calls
  and grudges with board highlights.
- Chat emotes, typing indicators, player reactions and a bot thoughts
  toggle.
- Card engine: payment quotes from the server drive every payment,
  including metals, substitutes and card storage sources. New
  decisions for card reveals, effect copies, resource removal, card
  receipts, action reuse and trade track steps.
- Colonies: trade fleets, colony tile additions, activation resources
  and inactive colonies.
- Board: Mars surface relief, a climate that follows the global
  parameters with frozen oceans and frost, pine groves and shoreline
  meadows, and a rebuilt nuclear zone with an on-screen blast.
- Host-paced corporation and prelude showcases at game start.
- Card browser: card size toggle, consistent search and sort, and
  settings kept in the URL.
- Music playback controls with track titles, a loading progress bar,
  and a lobby sound when players become ready.

## Changed
- Card descriptions are stored and shown as labeled sections.
- Corporation first actions run at the start of the player's turn.
- Unknown routes redirect to the main menu.
- Notifications on phones appear at the top centre. The info style is
  darker on all screens.

## Fixed
- Award funding is limited to the awards selected for the game.
- Wild tags are counted by context, and requirement lenience applies
  to global parameter steps.
- Popovers stay inside the viewport and close on Escape and outside
  taps.
- City foundations, volcano tiles and lake growth render correctly.
- Moholes play the construction sound when placed.
- Busy game buttons show a spinner, and the showcase no longer
  activates cards on its own.

## Build
- Make is replaced by just: a root justfile with backend and frontend
  modules. `just check` matches CI, and Go tools are pinned with
  `go tool`.
- GLB models are compressed and planet textures come in several sizes.
- Playwright checks for the phone layout.
