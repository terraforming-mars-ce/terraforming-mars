Rebrand to Open Mars, release notes in the repo and an in-game update
notice. Self-hosters must rename their environment variables.

## Added
- changelog/<tag>/ holds CHANGELOG.md for developers and
  CHANGELOG-USER.md for players, backfilled for every earlier tag.
  `just changelog-check vX.Y.Z` validates both.
- GET /api/v1/changelog serves the player notes, and
  GET /api/v1/changelog/{version}/{file} the images they reference.
  OPENMARS_CHANGELOG_DIR sets the directory (default changelog,
  /app/changelog in the image).
- The frontend compares its version with /api/v1/meta on load, every
  5 minutes, on tab focus and after reconnects, and offers a reload
  when they differ. A Changelog dialog is in the main and in-game
  menus and opens once after an update.

## Changed
- Renamed to Open Mars. The repo is openmars-app/openmars and the
  images are ghcr.io/openmars-app/openmars and
  ghcr.io/openmars-app/openmars-proxy.
- Every TM_* environment variable is now OPENMARS_*, for example
  OPENMARS_ADDR, OPENMARS_WEB_DIR, OPENMARS_SERVER_ALIAS and
  OPENMARS_SERVERS. The old names are ignored.
- The gateway runs as openmars-gateway.service with its config in
  /etc/openmars-gateway, deployed by
  /usr/local/bin/openmars-gateway-deploy. Re-run setup-vm.sh on the
  gateway VM before the next deploy.
- The Pi deploy runs from /home/mhm/openmars and reads
  OPENMARS_VERSION.
- Browser storage keys use an openmars. prefix, so saved player
  names, sessions and settings reset once.
- The Go module is openmars and the game data files dropped their
  terraforming_mars_ prefix (cards.json, maps.json and so on).

## Build
- Releases are tagged with `just release vX.Y.Z`, which requires both
  changelog files. The Release workflow runs only on tag pushes,
  checks the changelog before building anything and drafts the
  GitHub release from both files.
