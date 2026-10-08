# Terraforming Mars task runner. Run `just` to list every recipe.
# Backend and frontend recipes live in modules: `just backend <recipe>`, `just frontend <recipe>`.

set shell := ["bash", "-euo", "pipefail", "-c"]

mod backend
mod frontend

[private]
default:
    @just --list --unsorted --list-submodules

[group('setup')]
[doc('Install Go modules and frontend dependencies')]
deps: backend::deps frontend::deps

[group('dev')]
[doc('Run the backend (:3001, hot reload) and frontend (:3000) together; Ctrl-C or either one exiting stops both')]
dev:
    #!/usr/bin/env bash
    set -uo pipefail
    # Each server runs in its own process group so shutdown reaches its children
    # (air's rebuilt binary, vite under bun). A server still running five seconds
    # after SIGTERM is killed.
    groups=()
    stop() {
        trap - EXIT INT TERM
        for group in "${groups[@]}"; do kill -TERM -- "-$group" 2>/dev/null; done
        for _ in $(seq 50); do
            alive=false
            for group in "${groups[@]}"; do kill -0 -- "-$group" 2>/dev/null && alive=true; done
            $alive || return
            sleep 0.1
        done
        for group in "${groups[@]}"; do kill -KILL -- "-$group" 2>/dev/null; done
    }
    trap stop EXIT
    trap 'exit 130' INT TERM
    setsid {{ just_executable() }} backend dev &
    groups+=($!)
    setsid {{ just_executable() }} frontend dev &
    groups+=($!)
    wait -n

[group('dev')]
[doc('Stop stray dev servers: anything on :3000/:3001 plus air and vite processes from this repo')]
kill:
    #!/usr/bin/env bash
    set -uo pipefail
    pids=$( (lsof -t -i:3000 -i:3001 -sTCP:LISTEN 2>/dev/null; pgrep -f "{{ justfile_directory() }}.*(air|vite)" 2>/dev/null) | sort -u)
    if [ -z "$pids" ]; then
        echo "No dev servers running."
        exit 0
    fi
    echo "Stopping: $(echo $pids)"
    kill -TERM $pids 2>/dev/null
    sleep 2
    remaining=$(for pid in $pids; do kill -0 "$pid" 2>/dev/null && echo "$pid"; done)
    if [ -n "$remaining" ]; then
        kill -KILL $remaining 2>/dev/null
    fi

[group('check')]
[doc('Fail if any backend or frontend source is not formatted')]
format-check: backend::format-check frontend::format-check

[group('check')]
[doc('Lint backend and frontend')]
lint: backend::lint frontend::lint

[group('check')]
[doc('Type-check the frontend')]
typecheck: frontend::typecheck

[group('check')]
[doc('Run backend tests')]
test: backend::test

[group('check')]
[doc('Fail if the committed generated TypeScript types are out of date')]
generate-check: backend::generate-check

[group('check')]
[doc('Run every check CI runs, without modifying files')]
check: backend::check frontend::check

[group('fix')]
[doc('Format backend and frontend source')]
format: backend::format frontend::format

[group('fix')]
[doc('Format, regenerate types, then run every check CI runs')]
prepare-for-commit:
    {{ just_executable() }} format
    {{ just_executable() }} generate
    {{ just_executable() }} check

[group('codegen')]
[doc('Generate frontend TypeScript types from the backend DTOs')]
generate: backend::generate

[group('build')]
[doc('Build the backend binary and the frontend bundle')]
build: backend::build frontend::build

[group('build')]
[doc('Remove backend and frontend build output')]
clean: backend::clean frontend::clean

[group('deploy')]
[doc('Build and deploy to the Raspberry Pi (see scripts/deploy-pi.sh)')]
deploy-pi:
    ./scripts/deploy-pi.sh
