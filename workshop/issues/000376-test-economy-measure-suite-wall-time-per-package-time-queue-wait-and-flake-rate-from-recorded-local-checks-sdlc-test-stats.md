---
id: 000376
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: 'c35033a5e677d323d0fa19ec7f0be14d1783dafc' # card fields mirrored from issue-cards; edit via sdlc
---

# Test economy: measure suite wall time, per-package time, queue wait and flake rate from recorded local checks (sdlc test-stats)

## Problem

The operator (2026-10-14) made test health and performance a train ("test economy" in the ops TL status file): find a sustainable pattern to keep pair's and ariadne's tests fast and trustworthy. On 10-14 the suite queue set the fleet's pace: ariadne:3's recorded suite waited ~50 min for its turn; pair#528 spent most of the day waiting for or rerunning ~10-minute suites; flakes (#369's two, pair#423, pair#554, the #514 class) cost whole extra runs. Nobody can say where the time goes, because nothing aggregates it.

## Spec

`sdlc test-stats [--since 7d] [--repo …] [--json]`, read-only, from what already exists: recorded local checks (#345/#332 `.git/sdlc/local-checks/<sha>/`), suite logs (`go test -json` / shard summaries), and the suite lock's ticket log (#357) for queue waits. Report per repo: runs, wall time (p50/p90), time per package or shard (top N), queue wait (p50/p90), and **flakes**: tests that failed and then passed on the same head, or failed on main in a baseline run, with counts. It's the first car of the test-economy train: it says where to aim targeted runs (#328) and minimization sweeps (#327).

## Done when

- `sdlc test-stats --since 7d` on this machine prints the per-repo table for ariadne and pair, with the top slow packages and the flake list (tests on recorded fixtures for the parsers).

## Plan

- [ ]

## Log

### 2026-10-14
