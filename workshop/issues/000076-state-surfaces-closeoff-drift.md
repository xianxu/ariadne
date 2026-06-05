---
id: 000076
status: done
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-04
estimate_hours: 2
actual_hours: 0.24
---

# sdlc state surfaces done-but-open close-off drift

## Problem

`sdlc` gates the **entry** to each workflow stage (claim, change-code, close,
merge all refuse without evidence) but never surfaces a **stale exit**: an issue
whose work has shipped yet was never formally closed. The work is done; the
bookkeeping isn't — exactly the close-off hygiene drift `sdlc` exists to kill.

Concrete specimen (#51, found 2026-06-03): the in-place-branch workflow shipped
2026-05-31 (`ed4b550`) and became the live default everyone used for a week, but
the issue sat `status: open`. The only thing holding it open was a *deferred-
validation* checkbox ("live dogfood — best done as the first post-bootstrap
branch-based task"). That validation happened — repeatedly, automatically (every
`change-code → pr → merge` since) — but **nothing triggered closing the loop**.
A one-off ad-hoc scan on 2026-06-03 found this isn't unique: several open/working
issues have merged work + near-complete plans.

`sdlc state` already exists for "structural drift detection" (warn-only) and
already has `detectDrift` + `DriftFinding` (e.g. it flags "working with N plan
items, none ticked"). This is the natural home for the close-off check — the
inverse of "started but no progress": **finished but never closed**.

## Spec

Add a drift check to `detectDrift` (cmd/sdlc/state.go) that flags an open/working
issue as a **close-off candidate** when its work appears shipped but its status
hasn't flipped. Warn-only, like every other `state` finding — **surface a
candidate for a human glance; never auto-close** (closing carries actual/verified
judgment a heuristic can't supply).

The signal (the ad-hoc scan showed the naive version is too noisy — a bare
`git log --grep "#N"` count catches *filing*/`claim`/`issue-sync` commits, not
just implementation — so it must be refined):

- **status is `open` or `working`** (skip done/wontfix/blocked/punt — blocked is
  deliberately parked; punt/wontfix are decisions, not drift).
- **plan-completion is high** — all, or all-but-one, `## Plan` checkboxes ticked
  (`PlanTicked`/`PlanTotal` are already on `IssueState`). A lone trailing
  unchecked item is the #51 pattern (a deferred-validation line).
- **the work landed on main** — distinguish real shipping from bookkeeping. Best
  signal: a **merged PR** referencing the issue (`gh pr list --state merged
  --search "#N"`, or the issue's own `Review-Verdict:` close trailer present in
  `git log main`). Weaker fallback: ≥1 commit on main whose subject matches the
  `#N Mx:` / `#N:` *work* convention (§12), excluding `claim`/`issue-sync`/`file
  issue` subjects.

Emit e.g.: `⚠ #51 looks done — merged work on main + plan 13/14 — close it?
(sdlc close --issue 51)`.

Open questions — RESOLVED at change-code (2026-06-04):
- **gh dependency → drop it.** The "best signal" (merged PR referencing #N) is
  already observable gh-free: a merged PR lands the work-convention commits
  (`#N Mx:` / `#N:`) onto main, and milestone-close/close commits carry the same
  `#N`-anchored subjects. So a subject-anchored `git log <main>` scan *is* the
  merged-work signal — no network, degrades by construction (no main ref → no
  finding, never hard-fails). Avoids `gh` flakiness for marginal gain (Simplicity
  First). `issue-sync`/`claim` commits never anchor `#N` in their subject
  (`issue-sync: update issues`), so they're excluded for free; `file issue` /
  `ticket` / `close` *do* anchor `#N` and are name-denylisted.
- **Threshold → `PlanTicked >= 1 && PlanTotal - PlanTicked <= 1`** (all, or
  all-but-one, with at least one tick). The `>= 1` floor keeps this disjoint from
  the existing "working, none ticked" info finding (no contradictory double-flag)
  and rejects the degenerate 1-item/0-ticked case. #51 (13/14) qualifies; a
  fresh 0/1 issue does not.

## Done when

- `sdlc state` lists close-off candidates as warn-only `DriftFinding`s, with the
  issue ref + the close command to run.
- The check distinguishes shipped work from bookkeeping commits (no false
  positive from a `claim`/`issue-sync`-only issue), and degrades gracefully when
  `gh`/network is unavailable.
- A test pins: a high-plan-completion open issue with merged work → flagged; a
  freshly-claimed open issue with only a filing commit → NOT flagged; a `done`
  issue → never flagged.

## Plan

- [x] Single-pass: surface close-off candidates in `sdlc state`.
  - **ARCH-PURE — pure classifier:** `gitx.IsShippedWorkSubject(issueNum, subject)
    bool` — subject-anchored (`^#N($|[^0-9])`, the same anchor `CommitWindow`
    uses) minus a bookkeeping denylist (`file issue`/`ticket`/`claim`/`close`).
    Pure, table-tested in `gitx`.
  - **ARCH-PURE — thin IO probe (injected seam):** `gitx.ShippedWorkOnMain(
    issueNum) (sha, subject string, shipped bool)` runs `git log <origin/main|main>`
    via the package `run` shim and applies the classifier; tested in `gitx` by
    overriding `run`. `detectDrift` takes a `shipProbe` func param (production
    wires `gitx.ShippedWorkOnMain`; `state_test` passes a fake) so the drift logic
    is tested without git. Wiring is covered both ends so the pure helper can't be
    silently un-wired (lessons.md #72).
  - **ARCH-DRY:** reuse the `CommitWindow` subject-anchor pattern and the
    `recentCommits` main-ref resolution (origin/main → main → none); extend the
    existing `detectDrift`/`DriftFinding`/`renderProse` structures, add nothing
    parallel.
  - close-off finding in `detectDrift` for status `open`/`working` when
    `PlanTicked >= 1 && PlanTotal-PlanTicked <= 1 && shipped`; warn-only; message
    carries the issue ref + `sdlc close --issue N`.
  - Tests: high-completion open + shipped → flagged; freshly-claimed open +
    filing-only → NOT flagged; `done` → never flagged; classifier table.
  - Atlas note if the surface warrants.

## Relationships

- Inverse of the existing `detectDrift` "working, none ticked" check (no-progress)
  → this is the all-progress-no-close end of the same spectrum.
- Motivated by #51's retroactive close (2026-06-03). The deeper hygiene principle
  is in `workshop/lessons.md` — a deferred-validation checkbox with no close-the-
  loop trigger is how done work stays open.
- Composes with #69's binary-owned close: once a candidate is surfaced, `sdlc
  close` does the reviewed close.

## Log

### 2026-06-03

Filed after closing #51 (which had drifted done-but-open for a week). Operator:
"that's why we started with `sdlc` — there's some lack of hygiene tracking
closing off issues." `sdlc state` gates entries; this makes it also surface stale
exits. Grounded in the existing `detectDrift`/`DriftFinding` structure.

### 2026-06-04 — implemented (single-pass)
- 2026-06-04: closed — go test ./... all green; dogfood: synthetic 2/2 issue for shipped #82 → warn close-off finding with real commit d468fbc3 + "sdlc close --issue 82"; real-repo scan flags no false positives (#52/#76 no-progress). gitx classifier table + run-shim wiring tests + state close-off table pin flagged/unshipped/0-1/done-never.; review verdict: SHIP

- **ARCH-PURE seam.** `gitx.IsShippedWorkSubject(num, subject)` is the pure
  classifier (subject-anchored `^#N($|[^0-9])` minus a `file
  issue`/`ticket`/`claim`/`close` denylist, whole-token so `close-off` ≠
  `close`); `gitx.ShippedWorkOnMain(num)` is the thin IO probe (git log over
  `MainRef` with a `--grep` prefilter). `detectDrift` takes a `shipProbe` func —
  production wires `gitx.ShippedWorkOnMain` in `runState`, `state_test` fakes it.
  `closeOffFinding` is pure; flags `open`/`working` with `PlanTicked >= 1 &&
  PlanTotal-PlanTicked <= 1 && shipped`.
- **ARCH-DRY.** Resolved the plan-quality DRY flag for real: extracted
  `gitx.MainRef()` (origin/main → main → "") and routed *both* `recentCommits`
  and the probe through it; reused `CommitWindow`'s subject-anchor + `--grep`
  pattern; extended the existing `detectDrift`/`DriftFinding`/`renderProse`
  rather than adding parallel surface. Did **not** reuse milestoneclose's
  `shortSHA` (it does a `git rev-parse` round-trip — pointless since the probe
  hands back a full SHA; used a pure `abbrevSHA` to keep `closeOffFinding` pure).
- **gh dropped** (resolved open Q): a merged PR lands the `#N Mx:` work commits
  on main, so the gh-free subject scan *is* the merged-work signal; no main ref →
  not-shipped (degrades, never hard-fails).
- **Dogfood caught a real bug** the unit tests missed: the probe was first called
  with the *padded* ID (`000082`) but commit subjects use `#82` → zero matches.
  Fixed by unpadding before the probe/close-hint; the test fake now keys on the
  unpadded number, so dropping `unpadID` fails the test (guards the wiring,
  lessons.md #72). Verified live: synthetic 2/2 issue for the already-shipped #82
  → `⚠ #000082 looks done — plan 2/2 + shipped work on main (d468fbc3 "#82 M3…")
  — close it? (sdlc close --issue 82)`; real-repo scan flags nothing (no false
  positives on the no-progress #52/#76).
- **Tests:** `gitx` classifier table (11 cases) + `ShippedWorkOnMain` run-shim
  wiring (work-wins-over-bookkeeping / filing-only / no-main-ref); `state`
  close-off table (flagged / unshipped / 0-1 pre-filter / done-never). Full suite
  green. Atlas: new "Drift checks" section in `sdlc-binary.md`.
