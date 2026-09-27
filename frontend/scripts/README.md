# Asset tools

Approved masters live in `assets/original/`, with export settings in the root asset catalog and profiles. Install frontend dependencies with Bun and fetch source media with `git lfs pull`. The image exporter runs under Node 24+; Bun remains the package manager and test runner.

- `make assets`: incrementally generate hashed runtime files and the shared registry.
- `make assets-check`: validate source coverage, card identities, and generated files.
- `make assets-test`: run exporter tests.
- `make assets-preview`: generate `output/asset-preview/index.html` for local review.

`make frontend`, `make run`, production builds, and direct frontend start/build/typecheck scripts prepare assets automatically. Vite watches originals and settings during development. Generated output is ignored; edit sources and export settings instead. CI and Docker generate assets before building the application.
