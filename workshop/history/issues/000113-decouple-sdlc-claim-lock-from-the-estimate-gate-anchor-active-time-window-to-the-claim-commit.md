---
id: 000113
status: done
deps: []
github_issue:
created: 2026-06-17
updated: 2026-06-17
estimate_hours: 4
actual_hours: 0.52
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

Detail in `workshop/plans/000113-decouple-claim-from-estimate-anchor-window-plan.md`.

- [x] A1 — drop the `→ working` estimate guard from `claim`/`set-status` (`setstatus.go`); invert the two estimate tests
- [x] A2 — extract `issue.CheckEstimate` (pure) out of `CheckStructural`; move estimate test cases (ARCH-DRY)
- [x] A3 — `change-code` dedicated estimate gate + `--no-estimate` flag, reusing `CheckEstimate` (pure `estimateRefusal`, ARCH-PURE)
- [x] A4 — `start-plan` non-blocking estimate nudge (`estimateNudge` pure + `issueEstimate` IO seam)
- [x] A5 — reconcile help text (`claim`/`root`/`set-status`/`change-code`/`start-plan`/`issue`)
- [x] B1 — `gitx.WorkingTransitionISO` (working-transition commit time, cap-bounded)
- [x] B2 — `computeActual` window-start = `windowStart(parentISO, wtISO)` (pure helper, ARCH-PURE)
- [x] B3 — tests: `TestWorkingTransitionISO` (throwaway repo) + `TestWindowStart` (table)
- [x] C1 — AGENTS.md §2 claim-early prose
- [x] C2 — brainstorming SKILL.md claim-early offer
- [x] V — `go build/test/vet` green + manual claim/change-code/actual spot-checks + atlas reconciled

## Log

### 2026-06-17
- 2026-06-17: closed — go build/vet/test ./cmd/sdlc/... green — new tests CheckEstimate, EstimateRefusal, EstimateNudge, WorkingTransitionISO (throwaway repo), WindowStart; inverted the 2 estimate-guard tests. Binary smoke (throwaway repo): claim w/o estimate flips (no refusal); change-code w/o estimate exits 1; --no-estimate + valid estimate both pass; actual override path runs cleanly. Atlas issue-lifecycle + sdlc-binary reconciled; AGENTS.base.md (source) + brainstorming SKILL carry claim-early prose.; review verdict: SHIP

- Claimed; set `estimate_hours: 4`. `start-plan` delivered ARCH-DRY + ARCH-PURE.
- Read claim/setstatus/changecode/startplan/gitx/activetime/actual. **Finding:**
  `change-code` already enforces `estimate_hours` via `CheckStructural`→`checkEstimate`
  (only caller). So Part A = remove the premature `claim` guard + give the existing
  check its own `--no-estimate` bypass. Validated `git log -G'^status: *working'`
  finds the working-transition commit (tested on #52). Durable plan written.
- `change-code` plan-quality judge: **VERDICT INFO** (passing), 3 non-blocking
  findings. Decisions: (1) defer the close-side estimate backstop for the rare
  pure-doc `claim→close` path (noted in plan Non-goals); (2) pin B2 seam —
  `issuesDir` from env, `strconv.Atoi(issueNum)` for `locateIssueFile`; (3)
  widening `firstISO` to claim-time also widens `DiscoverWindowIssues` peer
  membership (attribution change, not just minutes) — intended, re-derive Peers.
  Branch `000113-…` created in-place.
- **Implemented A+B+C.** A: removed `checkTransitionGuards` Guard 2; extracted
  pure `issue.CheckEstimate`; `change-code` gained a dedicated estimate gate via
  pure `estimateRefusal` + `--no-estimate`; `start-plan` prints `estimateNudge`;
  6 help files + scaffold reconciled. B: `gitx.WorkingTransitionISO`
  (`git log -G'^status: *working' --reverse`, cap-bounded) + pure `windowStart`
  (earlier-of-two) wired into `computeActual`. C: AGENTS.md §2 claim-early bullet
  + brainstorming SKILL.md step 2.
- **Verified.** `go build/vet/test ./cmd/sdlc/...` green (new tests:
  `TestCheckEstimate`, `TestEstimateRefusal`, `TestEstimateNudge`,
  `TestWorkingTransitionISO`, `TestWindowStart`; inverted:
  `…WorkingNoLongerNeedsEstimate`, `…OpenWithoutEstimateStillFlips`). Binary
  smoke (throwaway repo): claim w/o estimate flips (no refusal); change-code w/o
  estimate refuses (exit 1); `--no-estimate` + valid estimate both pass; `actual`
  override path runs cleanly. Atlas (issue-lifecycle, sdlc-binary) reconciled.
