---
id: 000113
status: open
deps: []
github_issue:
created: 2026-06-17
updated: 2026-06-17
estimate_hours:
---

# Decouple sdlc claim (lock) from the estimate gate; anchor active-time window to the claim commit

## Problem

Two coupled defects, surfaced while refining estimate/actual calibration (#112):

1. **`sdlc claim` bundles the estimate guard onto its original job.** Claim's
   real purpose is a *lock*: flip `open → working` + broadcast to origin/main so
   peer agents don't collide. But it also requires `estimate_hours` — and at claim
   time the operator often *can't* estimate yet (brainstorm/design hasn't
   happened). So claiming EARLY — which is exactly what you want — is blocked by a
   premature estimate demand.

2. **active-time's window starts too late.** `gitx.CommitWindow` anchors the
   window at the parent of the first `#N`-subject commit, so operator-attention-
   heavy DESIGN time (brainstorm / spec / plan / reviews) *before* the first code
   commit is excluded from `sdlc actual` — systematically under-measuring the
   scarce resource (operator attention). Observed live: #111 measured 0.35h though
   design + four reviews were the bulk of the attention.

These are linked: if claim becomes a cheap early lock (no estimate), you claim at
the *start* of engagement, and **the claim commit's git timestamp becomes a
precise, free window-start anchor** — no new frontmatter, design attention
captured. The scarce resource is operator attention (AI execution is flat-
subsidized, immaterial now); this makes `sdlc actual` measure it honestly.

## Spec

### A. Decouple `claim` from the estimate; relocate the gate (option (a))

- `sdlc claim` reverts to its original role — flip `open → working` + broadcast
  the lock — and **no longer requires/sets `estimate_hours`**. Claim becomes cheap
  ⇒ claim early (a side-benefit: the lock reserves the issue sooner → less peer
  collision, the original point of the broadcast).
- `estimate_hours` is **collected at `sdlc start-plan`** (post-design, where it's
  knowable) and **hard-required at `sdlc change-code`** — the universal gate every
  implementation passes (planned *or* atomic single-pass), so no worked issue
  escapes without an estimate. `change-code` refuses without `estimate_hours`
  (alongside its structural + plan-quality gates), honoring the per-gate
  `--no-<gate>` bypass convention.
- Net: the estimate guard MOVES `claim → change-code` (naturally set at
  `start-plan`). (Edge: a rare no-`change-code` issue — pure-doc work — would skip
  the gate; `sdlc close` can backstop "estimate present" for that path.)

### B. Anchor active-time's window to the claim (working-transition) commit

- Replace the window-start in `gitx.CommitWindow` (parent-of-first-`#N`-subject-
  commit) with **the git timestamp of the commit that flipped the issue to
  `working`** (the claim's issue-sync commit touching the issue file). Git already
  records this precisely — no new frontmatter field.
- Mechanism: locate the issue file's `status: working` transition commit; use its
  commit time as the window's left edge. Bounded by the existing `WindowCapDays`
  sanity cap; gap-truncation (15-min) keeps a dormant claim→work gap from
  inflating the actual.
- Effect: design attention after claim (brainstorm/spec/plan/reviews) lands
  in-window ⇒ `sdlc actual` reflects operator attention truthfully.

### C. Claim-early nudge (workflow prose)

- AGENTS.md workflow + the brainstorming flow: when brainstorming an *existing*
  issue (or as soon as an idea crystallizes into one), **offer to `sdlc claim` it
  first**, so the window anchors at engagement start. For pure pre-issue
  exploration, claim once the issue is created (`sdlc issue new` then claim).

## Done when

- `sdlc claim` flips `open → working` + broadcasts with **no estimate demanded**;
  runnable at brainstorm start. Its `--help` no longer mentions the estimate.
- `estimate_hours` is **required at `sdlc change-code`** (refuses without it),
  prompted/collected at `sdlc start-plan`; the estimate-guard tests move from
  claim → change-code. Help reconciled: `claim`, `start-plan`, `change-code`,
  and the `estimate_hours` field doc in `sdlc issue --help` ("set at start-plan,
  required by change-code — not at claim").
- active-time (`cmd/sdlc/internal/activetime` + `gitx.CommitWindow`) anchors the
  window-start to the issue's `working`-transition commit timestamp; unit-tested
  (a fixture where design commits precede the first `#N` commit → they're
  in-window).
- A claim-early instruction lands in AGENTS.md + the brainstorming flow.
- `go build ./... && go test ./cmd/sdlc/... && go vet` green.

## Scope / non-goals

- The estimation **model** (turning estimate inputs into attention-complexity:
  fragmentation, delegation-escalation) is **#112** — out of scope here. This
  issue only ensures estimates are collected at the right gate and the actual is
  anchored consistently with the estimate.
- **Agent-driven backfill for late claims** (the agent identifying the in-session
  brainstorm span when claim happened late) — deferred. Claim-early is the norm;
  the commit-anchor + gap-truncation cover the common case.
- **AI-cost modeling** — out (deferred; flat subsidized rates, per #112).
- Related: #112 (estimation model — companion), #092 (segment over-attribution —
  a related window concern), `baseline-v3.md` (the actual method doc).

## Plan

- [ ]

## Log

### 2026-06-17
