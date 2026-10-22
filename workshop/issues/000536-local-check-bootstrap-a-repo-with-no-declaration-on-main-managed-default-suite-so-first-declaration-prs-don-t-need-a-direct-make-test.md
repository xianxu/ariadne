---
id: 000536
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c1215e05f18b29153f053c814dceb4477a3954ac' # card fields mirrored from issue-cards; edit via sdlc
---

# local-check: bootstrap a repo with no declaration on main (managed default suite), so first-declaration PRs don't need a direct make test

## Problem

## Spec

From ops-echo:0 (10-21): `sdlc local-check run` refuses in echo because origin/main has no `.sdlc/local-checks.json`
yet, and echo#1 is the PR that adds echo's first declaration (a bootstrap case). echo:1 therefore ran `make test`
directly; it was tracked as an unmanaged suite, and when its owner exited it left a tombstone that held machine-wide
suite capacity until the :0 steward reconciled it (op 65dbdc31, 18:30).

Fix: a supported bootstrap path. When main has **no** declaration, `local-check run` uses the built-in default suite
(`make test`, the same default `effectiveLocalChecks` already injects) under a normal managed lease and records it,
or, for a PR that adds the first declaration, uses that PR's declaration for its own evidence only, clearly labelled
as bootstrap. Never weaker than the default; refuse only when there is no Makefile `test` target.

## Done when

- In a repo with no declaration on main, `sdlc local-check run --name suite` runs `make test` under a managed lease
  and records the result (test).
- The first-declaration PR can produce recorded evidence that its own merge accepts (test), and no unmanaged run is
  needed.

## Plan

- [ ]

## Log

### 2026-10-21
