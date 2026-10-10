Highlights since v7. Supersedes v7.0.1, whose image build was
cancelled. Releases now use vMAJOR.MINOR.PATCH.

## Added
- Server switching inside the game: on desktop the menu footer's
  Change server opens a popover; on phones it is in the menu drawer
  and opens a full-screen server selector. Each server shows its
  status, version and ping, and picking one reloads into it.
- When a server stops responding during a game, players behind the
  gateway get a Lost connection dialog with Retry and Switch server
  instead of a broken main menu.
- The loading screen names what it is loading (starfield, Mars,
  models) instead of only "Loading".
- The menu version line shows the current server, for example
  "EU 1 · v7.0.2".

## Changed
- Gateway server picker restyled to match the game menu, with
  bundled Orbitron fonts and a phone layout with a top bar and list.
- The picker lists every server in order and re-checks every 5
  seconds: online servers are green, servers that do not answer are
  red and disabled, and servers without a valid /api/v1/meta are
  hidden.
- First visits go straight to the first listed server and fall back
  to the next one if it is down. Only an explicit pick or a shared
  link is remembered.
- Phones no longer show a menu footer; View cards, Change server and
  Feedback are in the menu drawer.

## Fixed
- openmars.app hung in browsers after a gateway deploy. Ubuntu's
  Caddy breaks its HTTP/3 listener on reload, so the gateway now
  serves HTTP/1.1 and HTTP/2 only.
- Menu chrome (View cards, footer, hamburger) appears together with
  the landing page instead of popping in after the loading fade.
- On desktop, Create lobby sits at the bottom of the left column of
  the Create game form again.

## Build
- Image builds run only for vX.Y.Z tags. The Deploy workflow and the
  gateway deploy script accept only latest or vX.Y.Z.
- Pushing a version tag opens a draft GitHub release with the tag
  message as notes.
- Docker build stages run natively on the build machine; only the
  runtime stage is built per architecture, so the arm64 images no
  longer take 40+ minutes under emulation.
- just dev --proxy runs the gateway on :4000 in front of the hot
  reloading frontend; the dev server no longer opens a browser tab.
