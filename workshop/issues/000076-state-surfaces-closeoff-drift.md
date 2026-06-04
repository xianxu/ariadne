---
id: 000076
status: working
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-04
estimate_hours: 2
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

Open questions to resolve at change-code:
- gh dependency: the merged-PR check needs `gh`/network. Degrade gracefully when
  absent (fall back to the commit-subject heuristic; never hard-fail `state`).
- Threshold: all-ticked only, or all-but-one? (#51 was all-but-one — the trailing
  item was the deferred dogfood.) Lean all-but-one, since the trailing item is
  often the exact thing that drifts.

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

- [ ] Single-pass: extend `detectDrift` with the close-off-candidate check +
  the work-vs-bookkeeping commit/PR discriminator (a pure helper over
  `IssueState` + a `git log`/`gh` probe, kept behind the existing thin IO seam so
  it's testable with a fake); render it in `renderProse`; tests for flagged /
  not-flagged (filing-only) / never-on-done; atlas note if the surface warrants.

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
