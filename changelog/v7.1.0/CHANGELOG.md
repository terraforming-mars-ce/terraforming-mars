One image per game server. Highlights since v7.0.2.

## Added
- TM_ADDR sets the listen address (default :3001), and TM_WEB_DIR the
  frontend directory (default web, /app/web in the image). Without that
  directory only the API is served.
- Static files are served with gzip variants, immutable caching for
  /assets/*, no-cache for the page shell and cross-origin access for
  the openmars.app gateway. Large files such as the 130 MB skybox get a
  10 minute write deadline instead of the 15 second API timeout.

## Changed
- Each game server is now a single image,
  ghcr.io/terraforming-mars-ce/terraforming-mars. The Go server serves
  the API, the WebSocket and the built frontend on one port (3001 by
  default). The terraforming-mars-backend and terraforming-mars-frontend
  images are no longer published.
- A reverse proxy now sends every path to the one container. The
  separate /api, /ws and / routing is gone.
- API_URL and runtime-config.js are removed: the frontend always talks
  to the server it was loaded from.

## Build
- The game content (about 360 MB) builds in its own Docker stage and is
  reused from the cache unless assets/ or the asset pipeline change.
- Release images each keep their own build cache.
- just backend serve runs the single app locally; just proxy member
  uses it for the dev gateway.
