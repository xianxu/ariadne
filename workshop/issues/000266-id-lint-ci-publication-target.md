---
id: 000266
status: working
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '7f5071598ad7b1faefee00433da038600b42afd0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-28T21:16:53-07:00
flow: {kind: quick, provenance: inferred, spec: "63a81d39", done: "029ec9a1"}
---

# sdlc id lint cannot resolve publication target in CI checkouts

## Problem

Consumer merge-check CI fails on `40-duplicate-issue-id.sh` in every PR run
(seen in parley.nvim runs 36474779448, 36491104572 and 36511435873):

```
id lint COULD NOT RUN: /home/runner/work/parley.nvim/parley.nvim has a fetched
issue tracker but its publication target is unusable: configure main to track
one named remote/main: resolve publication target: git config: exit status 1
```

`actions/checkout` leaves a detached PR checkout with no local `main` tracking
a remote. The #252 lint wants a publication target it doesn't need just to read
IDs, and it exits 2 ("a check that did not look must not report clean"). So
consumer CI is red whatever the change is, which also blocked #241's "real
CI green" evidence (weave install and compile passed).

## Spec

A read-only lint should derive what it reads from the fetched tracker ref
without requiring publication configuration, or the seeded CI should configure
the tracking it needs. Pick one after reading the lint's resolution path.

**Decision: the lint takes its remote explicitly.** The failure path is
`cardlessAdditions` → `tracker.RepositoryForCheckout` → `RepositoryFor`, which
resolves `branch.main.remote`/`.merge` (the *publication* target, needed to
write cards) just to learn which remote's `issue-tracker` to read. The CI script
already fixes that answer: it reads the published id space from `origin/main`
and exits 0 when there is no `origin`. So `sdlc issue lint-ids` gains
`--remote NAME`: when given, the tracker is opened directly on that remote
(`tracker.NewRepository`), no resting-branch configuration consulted; the
script passes `--remote origin`, so trunk and cards come from one remote
(ARCH-DRY). Without the flag the operator path is unchanged. Still fails
closed: a marked repository whose named remote carries no tracker exits 2.

Rejected: having the seeded CI write `branch.main.*` config — it mutates the
runner to satisfy a read that doesn't need publication, and every consumer's
CI would have to carry it.

## Done when

- parley.nvim merge-check passes `40-duplicate-issue-id.sh` on a PR run.
- A regression test covers a detached CI-style checkout with a fetched tracker.

## Plan

- [x] Regression test: CI-shaped checkout (fresh `git init`, fetched
      `refs/remotes/origin/*` incl. `issue-tracker`, detached HEAD, no local
      `main`) running the real `40-duplicate-issue-id.sh` → exit 0 for carded
      details, 1 for cardless; confirm it fails (exit 2) before the fix.
- [x] `lint-ids --remote`; script passes `--remote origin`.
- [ ] Ship; re-run parley.nvim merge-check on a PR once it has the new script.

## Revisions

- 2026-09-28 — Done-when clause 1 (parley.nvim PR run passes) and Plan box 3 are
  verified **after merge**, not at close: parley's `40-duplicate-issue-id.sh`
  symlinks into ariadne and its CI clones ariadne `main`, so no pre-merge run can
  exercise this change. Close attests clause 2; clause 1 is logged post-merge.

## Log

### 2026-09-28
- Reproduced: `TestDuplicateIDCheckRunsInADetachedCICheckout` drives the real
  `40-duplicate-issue-id.sh` in a fresh-init, fetched, detached checkout; before
  the fix it exits 2 with the exact CI message ("publication target is
  unusable: configure main to track one named remote/main").
- Fixed via `lint-ids --remote`; the test passes (carded → 0, hand-made → 1).
- Close review BR-1: the fail-closed path (origin without a tracker → 2) had no
  test. Added to the same test; mutation-checked (a tolerant snapshot read turns it red).
- Full `cmd/sdlc` suite: green except `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`,
  which reads a plan archived to history on main (#210), unrelated. The suite takes
  ~19 min, so it needs `-timeout 30m`; the default 10m looks like a hang.
- parley.nvim's `40-duplicate-issue-id.sh` symlinks into ariadne, and its CI clones
  ariadne main, so the parley PR-run evidence can only come after merge.
