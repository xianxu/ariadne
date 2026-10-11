---
id: 000234
status: codecomplete
deps: [ariadne#231]
github_issue:
created: 2026-09-17
updated: 2026-10-10
estimate_hours: 2.7
card_mirror: '6d4e595109af391fa534bb06b99329e87fe431a9' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T17:41:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.38
---

# Split actuals at the first boundary review: build time vs converge time

## Problem

Calibration compares an estimate against an actual that it cannot, by
construction, predict all of. The actual runs from claim to close, so it includes
every boundary-review round and the fixes each one forces. An estimate prices the
work up to "I believe it's done", and nobody budgets time to fix bugs they have
not written yet. Operator, 2026-09-17:

> don't worry about estimate. all we need is to add a calibration factor, e.g. at
> your current skill level, the amount of bugs you will write. no one budget time
> to fix bugs they write, so estimate necessarily not accounting for those closing
> rounds. … maybe we should be tracking time to boundary review and boundary
> review itself. and really the estimation quality is time to boundary review.

#231 is the example: 6.33h estimated. Its measured actual passed 9.6h during M2,
by which point M1 had taken three boundary rounds and M2 four. The overrun
is mostly convergence, not forecasting error, but today's est/actual ratio cannot
tell the two apart.

## Spec

Split every issue's actual at its first boundary-review dispatch into two
measured quantities:

1. **build**: active time from the window start (`started:` or claim) to the
   dispatch of the first boundary review of each boundary, summed over
   boundaries. This is what an estimate predicts: design and implementation up to
   "done as far as I can tell". Plan-quality rounds fall inside it, as design
   convergence.
2. **converge**: active time from each boundary's first review dispatch to that
   boundary's finalize, summed. This is the review-and-fix loop: the bugs written
   plus the gate's appetite. It is not estimated per issue. It is calibrated as
   a factor `k = converge / build`, tracked per model, per flow (#231's full or
   quick) and per review recipe.

For milestone issues, build and converge alternate (M1 build, M1 converge, M2
build, …), so each boundary contributes its own segments.

**Estimate quality becomes est/build.** Drift and calibration read est against
build. k gets its own report. Quick-flow issues, which have no estimate,
still contribute converge data, and that is the evidence #231's Done-when defers
("judged after adoption on gate rounds").

**The cut point.** `close` and `milestone-close` stamp each boundary round's
DISPATCH time on the gate-ledger round (`gatestate.Round` today records only its
persist time, after the review). The cut for a boundary is its first round's
dispatch.

**Converge in two parts, if cheap:** review-wait (the judge subprocess's
wall-clock, already counted as active time) and fix (the rest). A slow judge and a
buggy author are different problems with different fixes.

**Ledger.** New columns appended after #231's flow columns (the column contract
is append-only and positional): `build_hours`, `converge_hours`,
`boundary_rounds`, and `split_source` (which evidence supplied the cut).

### Backfill: recompute the accumulated calibration data

The data to split history already exists, so do a pass over the calibration
ledger (526 rows at filing: parley.nvim 129, pair 125, ariadne 103, tools 63,
metis 58, and smaller repos). Cut-point evidence per boundary, best first:

1. a stamped dispatch time (rows closed after this issue);
2. the boundary gate ledger's first round (`*-close-gate.md`, since #194;
   timestamp minus the review's duration where recorded). At filing: ariadne 20,
   pair 54, parley.nvim 33, kbench 2 ledgers;
3. the review sidecar (`*-close-review.md` / `*-mX-review.md`, since #136). At
   filing: ariadne 97, pair 155, parley.nvim 108, kbench 32;
4. the first `Review-Verdict:` trailer commit for the boundary. At filing:
   ariadne 239, pair 309, parley.nvim 194, tools 84, kbench 35.

The split itself re-runs the v3 active-time engine over the same window
(`sdlc active-time` already prints per-segment timestamps) and cuts at the chosen
point. A row with no recoverable cut (before the binary owned the boundary
review, #69, or a repo with no review evidence) is marked `split_source: none`,
never guessed. Original estimate and actual values are left untouched; the pass
only fills the new columns.

This is heavy data (AGENTS.md §10). Build it as a scripted pass with a dry-run
that prints the proposed split table, run it first on a small sample the operator
checks, then the whole ledger. Document the method in brain's `velocity/SKILL.md`.

## Done when

- `close` and `milestone-close` stamp the dispatch time on every boundary round,
  and a pure function splits a window's active-time segments at a boundary's cut
  into build and converge, summing per boundary for milestone issues.
- Calibration rows carry `build_hours`, `converge_hours`, `boundary_rounds` and
  `split_source`, appended after #231's flow columns. Older rows keep parsing.
- Drift reads est against build, and a report shows k per model, flow and review
  recipe.
- The backfill has filled the split for every historical row with recoverable
  evidence and marked the rest `none`. It ran as a dry-run on a sample first, its
  cut-point precedence is unit-tested, and the method is documented in brain's
  `velocity/SKILL.md`.

## Estimate

> *Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` (stale; provisional) against the durable plan. Design was largely done in planning, so the buffer is 0.15; impl is 40% of the v2 ranges.*

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.3  impl=0.32
item: smaller-go-module      design=0.0  impl=0.16
item: smaller-go-module      design=0.1  impl=0.2
item: smaller-go-module      design=0.1  impl=0.2
item: atlas-docs             design=0.1  impl=0.08
item: milestone-review       design=0.0  impl=0.2
item: smaller-go-module      design=0.05 impl=0.12
item: smaller-go-module      design=0.1  impl=0.16
item: atlas-docs             design=0.1  impl=0.08
item: milestone-review       design=0.0  impl=0.2
design-buffer: 0.15
total: 2.70
```

The M1 rows are the timeline/split (greenfield), the dispatch stamp, the
computeActual + E1 wiring, the ledger columns + E3, the atlas and the review.
The M2 rows are drift, the converge-report verb, the docs and the review.

## Plan

Durable plan: `workshop/plans/000234-build-converge-split-plan.md`.

- [x] M1 — dispatch stamp on boundary rounds; pure timeline + split; actual =
      build + converge (tails excluded at every boundary); milestone-close
      derives its increment (E1); five appended ledger columns; est-0 rows
      excluded from calibration (E3)
- [x] M2 — drift reads est vs build (split rows only); `sdlc converge-report`
      (k per model/flow/recipe); brain `velocity/SKILL.md`

## Revisions

- 2026-10-10 (TL Q1/Q2, project folds). The **backfill moves out** to
  ariadne#323, because historical actuals can't be reproduced: slot cwds were
  not recorded. Old rows keep `split_source -`, and drift reads only rows
  that have a split.
  - **The actual itself changes:** the tail is uncounted at every boundary,
    and build + converge == actual. Milestone issues' actuals drop versus
    today's rule.
  - Folded in: E1 (milestone-close derives its increment) and E3 (est-0 rows
    are not calibration points).
  - Added the column `review_recipe`, because k per recipe needs it on the
    row.
  - The cut is the first round's *dispatch*. Converge ends at the last `#N`
    commit at or before the final round's dispatch, which is the operator's
    consistent end point.

## Log

### 2026-09-17

Filed from #231's review loop. #231 depends-on relationship: this issue appends
its columns after #231 M3's flow columns and reads the per-round `Recipe` stamp
#231 added to the boundary ledger, which is what makes k per recipe computable.
The counts above were taken at filing and will move.

### 2026-10-10
- 2026-10-10: closed — Fix window after the SHIP close: merged origin/main (#183 landed; lessons.md union, no code conflict) and added a split table case pinning that #183's --fixed-then-ship round extends converge (last finalizing round wins). activetime + close/actual targeted tests pass. (fixed-then-ship); review verdict: SHIP
- 2026-10-10: closed — Re-close after folding the close-review Minor (convergereport reuses estimate.Dash; no behaviour change). Prior evidence stands: make test passes except the sandbox-only processgroup ps test; estimate + converge-report + calibration-row tests pass.; review verdict: SHIP
- 2026-10-10: closed — make test: all cmd/sdlc shards pass; only processgroup TestCancellationKillsDescendants fails, sandbox-only (ps fork denied; passes unsandboxed). M1 (closed SHIP r2): dispatch stamp, pure timeline+split (table + 500-case property), computeActualFor split + E1 milestone increment (live: M1 adopted 0.95h with 0.08h converge), ledger cols 23-27 + E3. M2: drift reads est/build on split rows only (TestDrift_EstimateAgainstBuild), estimate.ConvergeFactors + sdlc converge-report (TestConvergeFactors, TestConvergeReport; live: "no rows with a split yet" until this close), brain velocity/SKILL.md + atlas. --no-project: the TL-owned ops project tracks #234 as one task row; its tick is the TL's.; review verdict: SHIP
- 2026-10-10: closed M1 — Round 2 after fixing BR-1 + all 5 Minors (plans dir threaded through computeActualFor/the close seam + TestComputeActualReadsTheGivenPlansDir; cutTimes shared parse rule + parity case; REWORK fixture TestComputeActualCountsConvergeAfterRework 20m build/30m converge; converge-ends-at-commit property; godoc; growth note). Targeted activetime/estimate/close-actual tests pass; make test passed before the fixes (only the sandbox-only processgroup ps test fails). --no-project: the TL-owned ops project tracks #234 as one task row, with no per-milestone detail block.; review verdict: SHIP

- ariadne:3 claimed (measurement train, after #317/#321). The project folds
  apply: E1 (milestone-close derives its increment), E3 (no-estimate closes
  stop writing est 0.00 / ratio 0× rows), and the operator decision of 10-10
  (the verification tail stays uncounted; one consistent end point per
  boundary).
- Surface facts (read for the design):
  - close and milestone-close measure *before* they dispatch the review, and
    the window ends at the last `#N` commit (`gitx.CommitWindow`'s lastISO).
    So the final round's review is already outside the actual.
  - No dispatch time is stored; `gatestate.Round.Timestamp` is the finalize
    time. One `*-close-gate.md` per issue holds the rounds, tagged by
    `boundary`.
  - Activity runs are contiguous (`Active == End−Start`), so a
    time-proportional split of a segment's allocation is exact in minutes.
  - milestone-close refuses to adopt (`adoptOmittedActual`), so its
    `--actual` is hand-typed today.
  - 25 full-flow, non-upgraded est-0 rows pass `calibratable()`, and one of
    them as "latest" disables drift.
- Draft design (pending TL answers on (Q1) whether milestone tails leave the
  actual too, and (Q2) whether the backfill splits out to its own issue):
  - build = window start (or the previous boundary's finalize) → the
    boundary's first dispatch;
  - converge = first dispatch → the last `#N` commit before that boundary's
    final dispatch;
  - the tail is uncounted at every boundary;
  - build + converge == actual, as a time-proportional split of one
    `Compute` result.
- M1 implementation:
  - The dispatch stamp is `Round.Dispatched`, set through the `dispatchNow`
    seam.
  - The pure `activetime.BuildTimeline`/`Split` live in `split.go`, with a
    table test plus a 500-case property test (ordered, contiguous, conserving).
  - `Result.IssueCommitTimes` added.
  - `computeActualFor` splits, and `--milestone` was added to `sdlc actual`.
  - `adoptOmittedActual` adopts the milestone increment (E1), and the
    deviation check compares a milestone value with its own increment.
  - Ledger block 23–27, plus `rowSplitMeta`/`rowSplitSource` (PQ-1) and E3.
  - Help and atlas updated.
- Bug caught by the integration test: measuring a *past* milestone tagged the
  trailing build (after its finalize) as that milestone's, which gave M1 all
  of M2's work. The trailing build now goes to `current` only while `current`
  is unfinalized.
- M1 closed: SHIP in round 2.
  - Round 1 was FIX-THEN-SHIP with BR-1 (the split ignored `--plans-dir`)
    and 5 Minors; all were fixed.
  - The advisory Minor from round 2 (a second round filter) was fixed
    afterwards, with a lesson.
  - The derived increment was adopted: 0.75h → 0.95h, because round 1's
    fixes counted as converge (0.08h). The feature measured its own loop.
  - `--no-project`: the ops project tracks #234 as one row.
- M2 implementation:
  - drift reads est/build on split rows only (`calibratable` requires a split
    and positive build, via `BuildRatio`);
  - `estimate.ConvergeFactors` and `sdlc converge-report`;
  - brain `velocity/SKILL.md` documents columns 23–27 (brain auto-commits);
  - atlas: verb table, file map, drift/k.
  - Live: `sdlc converge-report` reports no split rows yet, which is correct
    until this close writes the first one.

