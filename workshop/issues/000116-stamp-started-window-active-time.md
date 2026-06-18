---
id: 000116
status: working
deps: []
github_issue:
created: 2026-06-17
updated: 2026-06-18
estimate_hours: 1.4
---

# Stamp started: at claim and window active-time from engagement start

## Problem

`sdlc actual` (active-time-v3) windows on commits: the window starts at the
first `#N`-subject commit's parent (`internal/gitx/window.go`). So all the
operator-attention-heavy DESIGN time — brainstorm, spec, plan, plan-reviews —
that happens *before the first code commit* is excluded from the measure.

Live evidence (the same data that motivated #112):
- **#111** actual measured **0.35h**, but the close note records the true figure
  as ~2h: *"design+brainstorm+4 reviews predate the first commit and both M1+M2
  were implemented by delegated subagents, so most effort is outside the operator
  transcript window."*
- **#110** close note: *"pre-first-commit planning/plan-review + background
  review-waits are outside the commit window, so true effort was higher."*

This is a systematic *under*-measurement of operator-attention. On its own it's a
fidelity bug; for #117 it is the upgrade that makes the calibration loop
trustworthy — #117's auto-calibration (mechanism 3) scores estimate against `sdlc
actual`, and a truncated actual is garbage data. #117 ships first and stamps such
rows `window-trusted: no`; **this issue is what flips new rows to trusted.** Not a
hard blocker for #117 (the trust flag handles its absence), but the reason #117's
data is worthless until this lands.

## Spec

Stamp a precise `started:` (ISO-8601) timestamp into the issue frontmatter at
`sdlc claim`, and window active-time from it.

- **Write side — `sdlc claim`:** on the open→working flip (and only then, once),
  write `started: <ISO>` to the frontmatter. Use the claim commit's timestamp as
  the source. Idempotent: never overwrite an existing `started:` (re-claim /
  re-sync must not move the anchor). The claim-at-start-of-engagement convention
  (claim → brainstorm → plan → code, already the #113 posture) keeps the window
  *tight* (bounds cross-issue bleed, cf. #092) yet *complete*. Gap-truncation
  (15-min cap, `activeMinutes`) already makes an earlier anchor safe from
  dormant-time inflation — a long idle gap before the first commit costs at most
  15 min, not the whole dormancy.
- **Read side — `sdlc actual` / `internal/activetime`:** window `SinceISO` from
  `started:` when present; **fall back** to the current first-`#N`-commit-parent
  anchor when absent (back-compat for issues claimed before this lands, and for
  the no-`started:` path generally). The attribution rule (commit-anchored,
  weight 1.0) is unchanged — only the window's left edge moves.

**Known residual (explicitly out of scope here):** windowing from `started:`
fixes *design-time* truncation, but delegated-subagent work still lives in a
separate transcript file outside the operator's session dir, so it is still not
counted as operator-attention. That is *correct* for an operator-attention
measure (you weren't attending), and #112's prose model already accounts for the
escalation/supervision tail. Measuring delegated-subagent attention is a
different problem; not addressed here.

## Done when

- `sdlc claim` stamps an idempotent `started:` ISO on the open→working flip;
  re-claim does not move it. Unit-tested.
- `sdlc actual` / `internal/activetime` windows from `started:` when present and
  falls back to the commit-parent anchor when absent. Unit-tested both paths.
- Demonstrated on a real claimed issue (e.g. #112, claimed at engagement start
  2026-06-17): the pre-first-commit design attention that #111 lost is now inside
  the window.
- Atlas note on the windowing change (`atlas/` active-time/actual surface).

## Revisions

### 2026-06-18 — premise reconciled to the #113 mechanism
**Reason:** code reading found the Problem's premise is partly stale. `sdlc actual`
(`computeActual`, `actual.go`) ALREADY pulls the window-start back to the claim:
the #113 block calls `gitx.WorkingTransitionISO` (the open→working transition
commit's `%aI` date) and `windowStart` takes the *earlier* of (commit-parent,
claim). So claim-anchored windowing exists — but via a **git-log heuristic** that
is explicitly "best-effort — a locate/parse miss just keeps the commit-based
start" (a silent fallback that drops design time).

**Delta (operator chose "harden it"):** #116 is reframed from *fix the excluded
window* to *make the anchor explicit + robust*:
- **Write side:** stamp `started:` (local-offset RFC3339, to stay lexically
  comparable with git's `%aI` that `windowStart` compares) at the open→working
  flip — in the shared `applyStatus` so both `claim` and `set-status working`
  stamp it. Idempotent (never overwrite). New code injects a clock seam rather
  than calling `time.Now()` directly (controllable-time).
- **Read side:** the anchor preference becomes `started:` (explicit) →
  `WorkingTransitionISO` (#113 legacy heuristic, fallback) → commit-parent. So
  `started:` supersedes the heuristic when present; nothing regresses for legacy
  issues without it.

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: smaller-go-module   design=0.1 impl=0.3
item: smaller-go-module   design=0.1 impl=0.3
item: atlas-docs          design=0.1 impl=0.2
item: milestone-review    design=0.0 impl=0.2
design-buffer: 0.30
total: 1.4
```

design 0.3 ×1.30 = 0.39 · impl 1.0 ×1.0 · total ≈ 1.4 (v2 build-effort; single-pass).

## Plan

- [x] Stamp `started:` in `applyStatus` on open→working (clock seam, idempotent) + test
- [x] `computeActual` prefers `started:` over `WorkingTransitionISO`, then commit-parent + test
- [x] Reconcile atlas (active-time/actual windowing) + verify on a real claimed issue

## Log

### 2026-06-17
Carved out of #112's "Companion (actual-side)" section during the #112 brainstorm
(the estimate↔actual-coherence discussion). Upgrades #117's calibration rows from
`window-trusted: no` → `yes` (not a hard blocker — #117 trust-flags around it).
Independent Go change. Work order this session: #117 → #116.
