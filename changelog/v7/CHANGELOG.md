Highlights since v6.

## Added
- Gateway: a new proxy/ service puts several independent game servers
  behind one address. Players pick a server on first visit (EU 1,
  EU 2), the choice is remembered per tab and per browser, and the
  chosen server's own frontend boots straight from that server.
  Assets, API and WebSocket traffic never pass through the gateway.
- Shared game links name their server (?s=eu-1), so a friend who opens
  one lands on the right server.
- Menu footer shows the current server with a Change server link when
  playing through the gateway.
- Backend: GET /api/v1/meta returns the server's alias, name and
  version. Set them with TM_SERVER_ALIAS (lowercase a-z, 0-9, -;
  default local) and TM_SERVER_NAME.

## Changed
- The frontend resolves API, WebSocket and asset URLs against the
  origin it was loaded from, not the page origin.
- Saved game sessions and active-tab keys in localStorage are scoped
  per server when playing through the gateway.
- Frontend nginx allows cross-origin reads of static files and serves
  index.html and runtime-config.js with Cache-Control: no-cache.

## Fixed
- assets-check runs the asset pipeline first, so it no longer fails
  on a clean checkout.

## Build
- Release tags only build images. Images are now tagged with the
  release tag (v7) as well as latest, and a proxy image is published.
- Deploys are manual: the Deploy workflow takes a target (gateway or
  pi) and a version, and checks SSH host keys against stored secrets.
- proxy/deploy/ holds the gateway VM setup: a systemd unit, Caddy
  config for openmars.app and a restricted deploy script.
