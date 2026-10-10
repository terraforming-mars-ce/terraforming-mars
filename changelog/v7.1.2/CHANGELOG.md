Fixes since v7.1.0. Supersedes v7.1.1, whose images never built
because Docker Hub rate-limited the release runners.

## Changed
- Player data carries resourceConversions with the state of the heat
  and plant conversions.

## Fixed
- The bottom resource bar and the mobile dock and actions screen
  decided on their own whether heat and plant conversions were
  possible. They now use the availability, errors and effective cost
  the server computes, so the buttons match what the server accepts.
- Conversion filtering is removed from the standard project
  availability checks.

## Build
- Release builds pull every image (base images, BuildKit, QEMU)
  through Google's public Docker Hub mirror, mirror.gcr.io, so
  anonymous Docker Hub rate limits and timeouts on shared runners no
  longer fail releases.
