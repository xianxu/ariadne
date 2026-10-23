---
id: 000557
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 1.43
collaboration-mode: collaboration
card_mirror: 'c81582bd533e695c20dbf4387d01018412d17e4c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T17:29:49-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# start-plan's weave refresh regenerates seeded workflows on the resting branch, then refuses

## Problem

## Spec

`sdlc start-plan`'s catch-up bumps a peer repo's ariadne pin and runs `weave refresh`, which regenerates seeded files
(here `.github/workflows/merge-check.yml`: ariadne's SIGPIPE/drain fix in the compiler-provenance step, two
`grep -Eq` lines becoming `grep -E ... >/dev/null` plus a drain comment) **on the slot's resting branch**. start-plan
then refuses, because the resting branch has uncommitted tracked changes beyond the issue's details. Every slot that
refreshes against the new ariadne hits this:
- pair:8 (pair#809, 10-22 13:1x): bump ariadne 2dc46703 → 5785330c; saved the patch, restored the file, re-ran.
- parley.nvim:3 (parley.nvim#409, 10-22): "weave-generated merge-check workflow drift" kept in a named stash.

Resting branches only fast-forward (#284), so the refresh must not leave tracked changes there, and seeded workflow
files are local-gates configuration (#309): their update has to reach main as its own change the operator verifies,
not ride along on an issue branch.

Fix: start-plan's catch-up never leaves regenerated tracked files on a resting branch. When a refresh would change a
tracked seeded file, it restores the file, reports "seeded file X is behind ariadne; file or land the weave refresh
as its own change", and continues (or the refresh runs where it can't dirty the resting branch). Plus a way for the
pending regeneration to reach each consumer's main as one local-gates change.

## Done when

- A start-plan whose catch-up would regenerate a tracked seeded workflow leaves the resting branch clean, proceeds,
  and names the pending seeded-file update (test).
- The pending update can be produced as one reviewable change for the consumer repo (documented, tested).

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only; calibration is stale, so this is provisional.

Extending existing Go/FS/refresh seams, no novel stack/library needed. Per-primitive design uses the thorough-spec ×0.2 discount; implementation is v2.1 ×0.4. Familiarity ×1.0, design buffer +15%. The four implementation units include their colocated/native tests; review and docs are separate.

| Unit | Primitive and source hours | Discounted design | Scaled impl |
|---|---|---:|---:|
| Pure policy/classifier | smaller Go 0.30 / 0.50 | 0.06 | 0.20 |
| FS currentness/report deferral | smaller Go 0.30 / 0.50 | 0.06 | 0.20 |
| One compile-policy wiring refactor | cross-cutting refactor 0.40 / 0.50 | 0.08 | 0.20 |
| Native refresh/start-plan/regeneration proof | smaller Go 0.30 / 0.50 | 0.06 | 0.20 |
| README/help/atlas | docs 0.10 / 0.20 | 0.02 | 0.08 |
| One mandatory close review | milestone review 0.10 / 0.50 | 0.02 | 0.20 |

Design 0.30 ×1.15 + impl 1.08 = 1.425h, rounded to 1.43h. Suite queue wait is not focused work and is recorded separately, never hidden as estimated implementation.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: cross-cutting-refactor design=0.08 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.02 impl=0.20
design-buffer: 0.15
total: 1.43
```

## Plan

Proposed design: [planning seed deferral](../plans/000557-planning-seed-deferral-plan.md). Hub approved the seeded-output scope; entering change-code before implementation.

- [ ] Add a planning-only publication policy that defers every Seed / SeedOnce destination and existing protected settings before any output effects.
- [ ] Wire start-plan catch-up to that policy; report stale seed paths and the separate consumer regeneration action.
- [ ] Prove native catch-up cleanliness and explicit consumer-branch regeneration, document the operator-verified PR route, then complete suite/review/landing.

## Log

### 2026-10-22

### 2026-10-22 — claim and design investigation

Claimed in ariadne:2 and started planning from fresh main. #531's merge runner completed with a passing authoritative suite (2302 SDLC tests, 76 other packages), PR396 merged, and this slot clean on main-slot2 before the new claim. This ariadne checkout did not reproduce consumer seed drift because its own start-plan succeeded.

Root cause verified: startPlanBranch calls catchUpBeforeWork while on the resting branch; installed weave runs refresh --defer-settings. Refresh fast-forwards captured host/substrate targets and then compilePreparedSettings publishes normal actions. Only .claude/settings.json is deferred. plan.Seed performs an atomic content/mode replacement and is consumer-owned, outside managed identity/displacement recovery. plan.SeedOnce can also adopt a tracked entrypoint. Restoring files after refresh is unsafe on partial failure or a kill; prevent these actions before publication instead (ARCH-ORDER / SECURE).

Read-only bounded code trace identified existing OutputCurrent, DeferredOutput, settings deferral and real-Git refresh fixtures for reuse (ARCH-DRY). Verified both source commits named by the issue with Git: 2dc46703996c42f4b0916515d4572cf439dd2a32 and 5785330cda6bf31d533e4282aecc8f3f9cd632a7. Their actual workflow diff is the two quiet grep compiler-identity checks replaced with full-input grep plus a drain comment; capture that diff in the native regression rather than inventing a similar incident.

Proposal: a new explicit weave refresh --planning policy defers all Seed and SeedOnce actions regardless of tracking, plus existing protected settings; normal compile and explicit refresh retain their current regeneration semantics. Compare before publication, retain DeferredOutput evidence, name each stale destination and direct regeneration onto its own consumer issue branch with weave compile. That separate local-gates PR requires the operator's label; no automatic label, consumer branch, commit or PR creation. Plan explains that this covers the whole seeded-output class, not an unspecified global rollback for other tracked writer families; requesting hub confirmation of that scope. No source edits or implementation tests yet.

### 2026-10-22 — plan gate and estimate

Plan-quality round 1 CLEAN, no findings. Only after that acceptance, derived provisional Method-A v3.1 estimate 1.43h from the shared grammar and local calibration; recording card estimate through its setter before retrying change-code.

## Revisions

### 2026-10-22 — proposed prevention at publication boundary

Propose deferring seed / seed-once publication before effects, instead of post-refresh restoration. Preserve ordinary regeneration and route pending updates through a dedicated consumer issue and operator-verified PR. The existing Done when remains the acceptance contract; await hub review of the seeded-output scope.

### 2026-10-22 — scope approved

Hub approval received and authenticated (submitted delivery b2a58869-feeb-45f0-b41c-973fb0167e90). Suppress every Seed/SeedOnce before planning publication, retain protected-settings deferral, and use ordinary weave compile on a dedicated consumer issue branch for the operator-verified workflow PR. Wider Ship / retirement / .gitignore protection is a known follow-up, to file only on observed dirtying evidence. No stash, live consumer mutation, automatic PR or self-label. Proceeding to change-code with the full-flow proposal; no source edits yet.

