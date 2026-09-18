---
id: 000234
status: open
deps: [ariadne#231]
github_issue:
created: 2026-09-17
updated: 2026-09-17
estimate_hours:
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

## Plan

- [ ] Design: the dispatch stamp, the split function and its milestone-issue
      summing, and the column contract after #231 M3.
- [ ] Stamp dispatch time on boundary rounds; ledger columns; drift reads est
      against build; the k report.
- [ ] Backfill: cut-point precedence, a dry-run on a sample for operator review,
      then the full ledger; brain `velocity/SKILL.md`.

## Log

### 2026-09-17

Filed from #231's review loop. #231 depends-on relationship: this issue appends
its columns after #231 M3's flow columns and reads the per-round `Recipe` stamp
#231 added to the boundary ledger, which is what makes k per recipe computable.
The counts above were taken at filing and will move.
