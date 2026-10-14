---
id: 000355
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '5100b38eed7731940c5f3673945f3cc6434493d8' # card fields mirrored from issue-cards; edit via sdlc
---

# dev-aliases: build to a temp file and rename into place, so a rebuilding alias never leaves the binary missing

## Problem

Each dev-alias function (`construct/dev-aliases.sh`) rebuilds its owner binary on every call as `rm -f bin/X && go build -o bin/X ./cmd/X`, then runs it. On 2026-10-13 the operator's .zshrc switched every interactive shell to these functions with a fixed workspace, so any `couch` call anywhere deletes `~/workspace/pair/bin/couch` for the few seconds of the build; other processes calling that path meanwhile (non-interactive Codex shells via PATH/links, the hub by full path) fail with 'no such file or directory'. The TL hit exactly that.

## Spec

Build to a temp file beside the target (`bin/.X.tmp.$$`) and `mv -f` it over `bin/X` (an atomic rename on the same filesystem), so the old binary stays until the new one replaces it. Concurrent alias calls must not corrupt each other (unique temp names; last rename wins). Keep the dev-prepare step (#299) and stdout-to-stderr behaviour; don't add a second staleness scheme (go's build cache already makes rebuilds fast).

## Done when

- The emitted functions build to a temp file and rename; a test runs two concurrent alias builds while a third process repeatedly execs the target path and never sees it missing; dev-aliases.test.sh passes under bash 3.2.

## Plan

- [ ]

## Log

### 2026-10-13
