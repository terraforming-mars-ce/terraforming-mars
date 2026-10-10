---
name: release-creator
description: "Write the release notes for a new Open Mars version (changelog/<tag>/CHANGELOG.md for developers and CHANGELOG-USER.md for players) and tag it with `just release`. Use when the user wants to tag a new version, cut a release, write a changelog or release notes, or asks what's new for players. Triggers on phrases like 'tag a release', 'create v7.2.0', 'cut a release', 'release notes for', 'changelog for', 'time we tagged'."
allowed-tools:
  - Bash
  - Read
  - Edit
  - Write
  - Grep
  - Glob
---

# Release Creator

Release notes live only in `changelog/<tag>/`, two files per version:

- `CHANGELOG.md` is for developers. It is the GitHub release body.
- `CHANGELOG-USER.md` is for players. It is the only changelog the game shows (main menu and in-game menu, "What's new" after an update, the update pill).

Tag messages carry no notes, only the subject `Open Mars vX.Y.Z`. `just release` and the Release workflow refuse a tag unless both files exist and parse (`backend/internal/changelog`). Never write notes anywhere else, so nothing can drift.

## Process

1. **Find the previous tag.** `git tag --list 'v*' --sort=-v:refname | head -5`. Pick the highest existing version. Ask the user if ambiguous. Patch for fixes only, minor for new features, major for a big player-facing update.
2. **List commits since then.** `git log --oneline <prev-tag>..origin/main`. Skim subjects.
3. **Read the substantive commits in detail.** For anything that touches gameplay, cards, UI, security, persisted data or public API, `git show --stat <sha>` and read the commit body and the diff. Find what a player actually sees: which button, screen, card or corporation, and what it did before and after.
4. **Group by theme, not chronology.** Developer sections in order: Security, Added, Changed, Fixed, Build. Player sections in order: Major update(s), New, Improved, Fixed. Drop sections that have nothing in them.
5. **Draft both files** in the formats below. Show them to the user verbatim. Wait for approval. Never tag without approval.
6. **Validate.** `just changelog-check vX.Y.Z`.
7. **Open a PR.** On a new branch (never commit to `main`), commit `changelog/vX.Y.Z/`, run `just prepare-for-commit`, push the branch and open a PR. Follow the repo's PR rules.
8. **Tag after the PR is merged.** `just release vX.Y.Z` on an up-to-date `main` checks both files, creates the annotated tag and pushes it. That push starts the Release workflow (images and a draft GitHub release), which is visible state, so ask the user before running it, or let them run it.
9. **GitHub release.** The workflow creates it as a draft with `CHANGELOG.md` as the body. Publishing is a separate, explicit step for the user.

## Developer notes: CHANGELOG.md

```markdown
Highlights since v7.0.2.

## Added
- TM_ADDR sets the listen address (default :3001), and TM_WEB_DIR the
  frontend directory.

## Changed
- Each game server is now a single image. The Go server serves the
  API, the WebSocket and the built frontend on one port.

## Fixed
- assets-check runs the asset pipeline first, so it no longer fails
  on a clean checkout.

## Build
- Release builds pull every image through mirror.gcr.io.
```

- No subject line; the folder name is the version.
- Optional one-paragraph intro, e.g. "Highlights since vA.B.C." or "Security release. Upgrade is recommended." No commit counts.
- Headings are exactly `## Security`, `## Added`, `## Changed`, `## Fixed`, `## Build`, in that order, never empty.
- `- ` bullets. Hard-wrap around 72 characters; wrapped lines are indented two spaces. Inline `code` is the only markup.

### Developer voice

- Concrete. Name fields, env vars, endpoints, files, error codes. Not "improved security" but "fix authentication bypass via client-controlled operationName".
- Lead with the effect, then the mechanism. "The bottom bar uses the server's conversion availability, so its buttons match what the server accepts" beats "Read resourceConversions in BottomResourceBar".
- Active voice. One thought per bullet. If a bullet runs more than three wrapped lines, split it.

### Developer sections

- **Security** comes first whenever present. Add the framing line "Security release. Upgrade is recommended." or "This release contains a security fix." as the intro. Per fix, name the vulnerability shape, what an attacker could do, and the mitigation.
- **Added** is for genuinely new functionality. Refactors and renames go in Changed.
- **Changed** is for behaviour or contract changes that are not fixes: different defaults, renamed fields, removed images or env vars. State the contract change so operators know what to update.
- **Fixed** is for bugs. Lead with the failure mode, not the patch.
- **Build** is for CI, Docker, tooling and deploys. Skip if cosmetic.

## Player notes: CHANGELOG-USER.md

Players do not care how it was built. They care what is different the next time they play. Keep it short.

```markdown
## Major update: Multiple servers

Pick between servers like EU 1 and EU 2. Your choice is remembered.

- Game links take friends to the right server.
- Switch anytime with Change server in the menu.

## New
- Aridor, Arklight and Stormcraft corporations.
- Stormcraft can pay for Heat to temperature with floaters.

## Improved
- Payments show exactly what you pay, including steel, titanium and
  card resources.

## Fixed
- Plants to greenery and Heat to temperature are no longer greyed out
  when you can afford them.
```

- No subject line; the folder name is the version.
- `## Major update: <Name>` sections come first, only for big features (a new expansion, a new way to play, a new platform). Each has a one- or two-sentence intro paragraph, then three or four bullets at most on what it means at the table.
- A major update may show one image, on its own line between the intro and the bullets: `![Trading with a colony](colonies.png)`. The file (`.png`, `.jpg`, `.jpeg` or `.webp`) sits in the same `changelog/<tag>/` folder and is stored in Git LFS. Alt text is required. Prefer a real in-game screenshot, cropped to the feature and around 1600px wide at most. Only major updates get images; the developer notes have none.
- Then `## New`, `## Improved`, `## Fixed`, in that order, never empty.
- `- ` bullets, one line where possible and two at most. Wrapped lines are indented two spaces. No markup.
- A release with nothing a player would notice is a single line and no sections: `Behind-the-scenes improvements. Nothing changes in your games.`
- A tag whose images never shipped gets no `CHANGELOG-USER.md`; its changes go into the next version that ships. (`just release` still requires the file for every new tag.)

### Player voice

- Use the game's own words: card, corporation, tile and button names exactly as shown in the game (Helion, Ecoline, Plants to greenery, Change server, M€, TR, VP).
- Features lead with the thing: "Stormcraft can pay for Heat to temperature with floaters", "Aridor, Arklight and Stormcraft corporations".
- Fixes say what the player saw, never the cause: "Search For Life gives 3 VP when it holds science", "Dragging a card no longer moves the board". "Fixed a bug where…" is fine when there is no shorter way.
- Leave out everything a player cannot see: refactors, data fields, endpoints, servers, Docker, CI, performance internals. If only those changed, use the one-line release.
- Plain, friendly and short. Second person ("you", "your") is fine.

## Don'ts (both files)

- No em dashes. Hyphen, comma, period, or rephrase.
- No emojis.
- No marketing vocabulary: comprehensive, robust, seamless, modern, powerful, elegant, leverages, exciting.
- No triplets ("fast, reliable, and scalable").
- No commit hashes, no "this PR", no "this commit", no co-author trailers.
- No issue links unless the issue carries non-obvious context (CVE id, upstream advisory).

## Permissions

- It writes files under `changelog/` on a feature branch and opens a PR only after the user approves the drafts.
- It does NOT run `just release` (which tags and pushes) without explicit user permission.
- It does NOT publish a draft release without explicit user permission.
- It does NOT amend, move, or delete existing tags.
