Highlights since v4.

## Added
- Procedural cities with distinct card layouts, including Capital's
  spire and the expanded Urbanized Area grid.
- Recessed lakes with sandy shores, connected ground materials, and
  denser forests. Tile appearances persist across players and replays.
- Planet atmospheres, animated solar lighting, and a revised skybox.
- Expanded card inspection with drag-to-play and placement previews.
- Configurable game creation, lobby chat, and milestone/award status.
- Available-action indicators and reproducible seeded game setup.

## Changed
- Refreshed card frames, corporation artwork, and resource icons.
- Unified menus and popovers with keyboard-accessible controls.
- Added construction smoke and varied construction sounds; revised
  card sounds and reduced UI and terraforming effect volumes.
- Replaced ambient music with a shuffled five-track soundtrack.
- Moved terrain generation into a worker and retained city geometry
  in GPU batches. Replaced triangle-based planet picking.

## Fixed
- Search For Life now awards 3 VP for science stored on that card.
- Forced colony placement now advances initial selection correctly.
- Joining waits for confirmed player identity before saving a session.
- Extra prelude choices are disabled once the selection limit is met.
- Card dragging no longer activates the board underneath it.
- Historical logs no longer replay live action feedback.
- Phobos cities sit on the moon surface; transparent materials receive
  planet haze correctly.
- City ground follows circular footprints and blends at shared edges
  without painting opposite tiles or exposing Mars between materials.

## Build
- Added a Git LFS asset catalog with generated runtime media and
  validation, including Docker and CI integration.
- Upgraded to Go 1.27.1, Node 26.10.0, Bun 1.4.2, Vite 8, and TypeScript 7.
- Added landscape verification and backend architecture checks.
- Consolidated backend domain packages and switched logging to slog.
