# Build vs Converge Actuals Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split every measured actual into **build** (up to a boundary's first
review dispatch) and **converge** (the review-and-fix loop after it). Record
both on the calibration row. Have milestone-close derive its increment, and
have drift read the estimate against build.

**Architecture:** Each boundary round gets a dispatch stamp. A pure
`Timeline` (built from the gate ledger's rounds and the issue's `#N` commit
times) labels every instant of the measured window as build, converge or tail.
A pure `Split` applies those labels to one `activetime.Compute` result by
cutting segments in proportion to time. Runs are contiguous intervals, so the
cut is exact in minutes. Tails are dropped, so the adopted actual is
`build + converge` by construction.

**Tech Stack:** Go (`cmd/sdlc`, `internal/activetime`, `internal/gatestate`,
`internal/estimate`), the brain calibration ledger TSV.

---

## Decisions (TL, 2026-10-10)

- **D1. The cut rule.** Measurement runs before dispatch, and the window ends
  at the last `#N` commit.
  - For each boundary *b*, **build** runs from the window start, or from the
    previous boundary's finalize, to *b*'s **first** round's dispatch.
  - **Converge** runs from that first dispatch to **the last `#N` commit at or
    before *b*'s final (finalizing) round's dispatch**.
  - **(tailStart_b, finalize_b] is the tail.** It is uncounted at **every**
    boundary, which is the operator's 10-10 consistency rule.
  - A boundary still open at measurement time (it was reviewed, but REWORK is
    pending) is converge from its first dispatch to the window end.
  - **Precedence (PQ-2):** for each boundary,
    - build ends at `min(firstDispatch, tailStart)`;
    - converge is `[firstDispatch, tailStart]` when
      `tailStart > firstDispatch`, and empty otherwise;
    - the tail is `(min(firstDispatch, tailStart), finalize]` minus
      converge.

    A boundary that passes on its first round therefore has no converge, and
    its pre-dispatch verification after the last commit is tail. That is the
    same as today's single-boundary rule, where the window ends at the last
    commit. Phases never overlap: `BuildTimeline` emits them in time order and
    clamps each to its predecessor's end.
- **D2. The actual itself changes (Q1).** The adopted actual is `build +
  converge`, and that sum holds exactly.
  - **Milestone issues' actuals drop compared with today's rule.** Today each
    earlier milestone's tail (its final test run and review wait) sits inside
    the window, because later commits extend the window past it.
  - Single-boundary issues are unchanged: their only tail is already after
    the window end.
  - Old rows are left as written. They stay labelled `split_source: -`, and
    M2's drift reads only rows that carry a split, so old-rule and new-rule
    actuals never mix in drift.
- **D3. Proportional cut.** A segment's claimant and mentions are aggregated
  over its whole run, so re-attributing per side of a cut would need
  event-level work and could change claimants. Instead, `Split` scales each
  segment's `Alloc[issue]` by the time fraction on each side. This keeps one
  `Compute` result as the single source (ARCH-DRY with `sdlc actual`), and the
  invariant build + converge == actual holds by construction.
- **D4. The backfill is out (Q2).** It is filed as ariadne#323. Historical
  actuals cannot be reproduced, because slot cwds were not recorded.
- **D5. E3: an est-0 row is not a calibration point.**
  - `ExcludedFromCalibration` also returns true for `Estimate == 0`. That
    covers quick rows and the 25 full-flow est-0 rows that today can become
    "latest" and switch drift off.
  - The ratio column is written as `-` when the estimate is 0.
  - The close line says `no estimate — not a calibration point` instead of
    `ratio 0.0×`.
  - The estimate column stays numeric, because `ParseRows` skips a row whose
    estimate fails to parse, and throughput needs those rows.
- **D6. Columns (appended, indices 23–27):**
  - `build_hours`, `converge_hours`, `boundary_rounds`, `split_source`,
    `review_recipe`.
  - `review_recipe` is added beyond the spec's four, because k per recipe
    needs it on the row.
  - `split_source` values:
    - `dispatch`: every cut came from a stamped dispatch;
    - `finalize`: at least one cut fell back to the round's finalize time,
      for rounds written before the stamp;
    - `no-review`: no boundary rounds, so everything is build;
    - `manual`: a hand-typed `--actual`, so no split;
    - `-`: a legacy row.
- **D7. E1.** `milestone-close` measures its own increment: the same timeline
  restricted to (previous boundary's finalize, window end]. The milestone
  refusal in `adoptOmittedActual` is removed.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `gatestate.Round.Dispatched` | `cmd/sdlc/internal/gatestate/ledger.go` | modified |
| `activetime.Phase` / `Timeline` | `cmd/sdlc/internal/activetime/split.go` | new |
| `activetime.BuildTimeline` | `cmd/sdlc/internal/activetime/split.go` | new |
| `activetime.Split` / `SplitResult` | `cmd/sdlc/internal/activetime/split.go` | new |
| `activetime.Result.IssueCommitTimes` | `cmd/sdlc/internal/activetime/compute.go` | modified |
| `estimate.LedgerRow` (5 split fields) | `cmd/sdlc/internal/estimate/ledger.go` | modified |
| `estimate.LedgerRow.ExcludedFromCalibration` | `cmd/sdlc/internal/estimate/ledger.go` | modified |
| `estimate.DriftVerdict` (est vs build) | `cmd/sdlc/internal/estimate/drift.go` | modified (M2) |
| `estimate.ConvergeFactors` | `cmd/sdlc/internal/estimate/converge.go` | new (M2) |

- **Round.Dispatched**: an RFC3339 time taken immediately before
  `judge.Dispatch`. It is `omitempty`, so pre-#234 ledgers stay
  byte-identical. It is empty on plan-quality rounds and on never-ran rounds.
  A failed dispatch returns before persist, so no round carries it.
  - **Malformed or hand-edited value:** it does not parse as RFC3339, so the
    cut falls back to the round's finalize time and the source becomes
    `finalize`. A dispatch later than its finalize is treated the same way.
- **BoundaryRounds → Timeline**:
  `BuildTimeline(rounds []BoundaryRound, issueCommits []time.Time, since, until time.Time) (Timeline, source string)`.
  - `BoundaryRound` is a small activetime-local view: Boundary, Dispatched,
    Finalized, Finalizing. activetime must not import gatestate; the caller
    maps between them.
    - `Finalizing` maps from `Round.Reviewed != ""`.
    - The **last** finalizing round per boundary is the final one, so a
      re-close (#183's narrow re-close) extends converge.
  - It returns ordered, non-overlapping `Phase{Start, End, Kind}` with Kind
    one of build, converge or tail, covering [since, until].
  - The fallback (dispatch → finalize time) sets `source = "finalize"`.
  - **Relationships:** 1 Timeline per measurement; N boundaries.
  - **DRY:** the one place D1 is encoded. The issue close, milestone-close
    and `sdlc actual` all read it.
- **Split**: `Split(segs []Segment, issue string, tl Timeline) SplitResult{Build, Converge, Tail float64}`,
  in minutes.
  - For each segment, `Alloc[issue]` is distributed over the overlap of
    [Start, End] with each phase, in proportion to time.
  - A zero-length segment goes wholly to the phase containing Start.
  - Invariant: Build + Converge + Tail == ΣAlloc[issue].
- **Result.IssueCommitTimes**: the author times of window commits whose lead
  names the scope issue (`Commit.Issues` contains `Scope.Issue`). D1 needs
  them for tailStart. They already come from `loadWindowCommits`, so there is
  no new git read.
- **ConvergeFactors (M2)**: rows → k = Σconverge / Σbuild grouped by (model,
  flow, recipe).
  - It reads only `NewestPerIssue` rows with `split_source ∉ {-, manual}` and
    `build > 0`.
  - n per group is reported.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `dispatchBoundaryReview` | `cmd/sdlc/milestoneclose.go` | modified | the judge subprocess (stamps the dispatch clock) |
| `persistBoundaryRound` | `cmd/sdlc/boundaryledger.go` | modified | the gate ledger file |
| `computeActual` | `cmd/sdlc/actual.go` | modified | git, transcripts, gate ledger |
| `adoptOmittedActual` | `cmd/sdlc/close.go` | modified | the close flags |
| `appendCalibrationRow` | `cmd/sdlc/close.go` | modified | the brain ledger TSV |
| `sdlc converge-report` | `cmd/sdlc/convergereport.go` | new (M2) | ledger read |

- **dispatchBoundaryReview**:
  - stamps `reviewResult.Dispatched = dispatchNow()` just before
    `judge.Dispatch`;
  - `persistBoundaryRound` copies it onto the round unless the round never
    ran.
  - `nowRFC3339` is a plain function, so this adds a package var
    `dispatchNow = nowRFC3339` as the test seam. It is swapped in tests the
    same way `gitRun`/`computeActualForCloseFn` are.
- **computeActual** reads the gate ledger through `readBoundaryGateLedger`
  (a missing file means no rounds, so `split_source = no-review`), maps the
  rounds to `[]activetime.BoundaryRound`, builds the timeline and splits. It
  gains:
  - `actualResult.Build/Converge/Source/Rounds/Recipe`;
  - a `milestone` parameter, which restricts the timeline to that
    milestone's phases (D7).

  The fakes are the existing real-git fixtures plus a gate ledger written
  into the test's temp `workshop/plans`.
- **appendCalibrationRow** gets the **minutes** (build, converge) from
  `closeFlags.Split`. That is set by `adoptOmittedActual` from the
  pre-dispatch measurement, or is `manual` when `--actual` was typed.
  - **Round metadata (PQ-1):** it reads `boundary_rounds`, `review_recipe` and
    the cut `split_source` from the **persisted gate ledger at row-write
    time** (`applyClose` runs after `persistBoundaryRound`), so the
    finalizing round is counted.
  - A first-round pass therefore writes `boundary_rounds 1` and the round's
    recipe. Its source is the measurement's (`no-review` before the first
    dispatch), upgraded to `dispatch` once the ledger holds a stamped round.
  - One pure helper, `rowSplitMeta(ledger, boundary)`, computes them, with a
    unit test.

## Chunk 1: M1 — dispatch stamp, timeline, split, columns, E1, E3

- [ ] M1 — stamp, split, columns, derived milestone increment, est-0 exclusion

### Task 1: `Round.Dispatched` stamp

**Files:** `internal/gatestate/ledger.go`, `milestoneclose.go` (`reviewResult`,
`dispatchBoundaryReview`), `boundaryledger.go` (`persistBoundaryRound`), and a
test in `boundaryledger_test.go`.

- [ ] Write a failing test: a fake dispatch with a pinned clock persists a
      round whose `dispatched` equals the pinned dispatch time and differs
      from `timestamp`. A never-ran round has no `dispatched`. A pre-#234
      ledger round-trips byte-identically. Run:
      `go test ./cmd/sdlc/ -run 'TestBoundaryRound.*Dispatch' -count=1`,
      and expect a FAIL.
- [ ] Add `Dispatched string \`yaml:"dispatched,omitempty"\`` (doc comment:
      the review start, #234) and `reviewResult.Dispatched`. Stamp it in
      `dispatchBoundaryReview` before `judge.Dispatch`. In
      `persistBoundaryRound`, set `l.Rounds[last].Dispatched` when
      `!neverRan`.
- [ ] Run the test (PASS) and commit: `#234 M1: boundary rounds stamp their dispatch time`.

### Task 2: pure `BuildTimeline` + `Split`

**Files:** create `internal/activetime/split.go` and `split_test.go`.

- [ ] Tests (no IO).
  - **Strategy:** a table over the D1 shapes: single-round pass; REWORK with
    fix commits; REWORK without fix commits; a single-round M1 then M2; open
    REWORK; undispatched or malformed stamp → finalize source; no rounds.
  - Plus a **property test** over random rounds and segments:
    - phases are ordered and non-overlapping and cover [since, until];
    - Build + Converge + Tail == ΣAlloc;
    - no converge when there is no commit after the first dispatch.
- [ ] Run `go test ./cmd/sdlc/internal/activetime/ -run 'TestBuildTimeline|TestSplit'` (FAIL).
- [ ] Implement `BoundaryRound`, `Phase`, `Timeline`, `BuildTimeline`,
      `Split` and `SplitResult`.
  - Group rounds by Boundary in order of first appearance; `BoundaryAll`
    rounds are skipped.
  - first dispatch = the earliest round's (Dispatched || Finalized);
  - final = the round with Finalizing set;
  - tailStart = the latest issue commit ≤ the final dispatch, else the first
    dispatch;
  - clamp everything to [since, until].
- [ ] Run (PASS) and commit: `#234 M1: pure build/converge timeline and split`.

### Task 3: `Result.IssueCommitTimes`

- [ ] Write a failing test in `commit_test.go`: Compute over a fixture returns
      the times of commits whose lead names `Scope.Issue` and nothing for
      citations. Implement it in `Compute` from the loaded commits. Run
      (PASS) and commit.

### Task 4: `computeActual` splits; E1 milestone increment

**Files:** `actual.go`, `close.go` (`adoptOmittedActual`, `closeFlags`),
`actual_test.go`, `close_adopt_test.go`.

- [ ] Tests.
  - **Strategy:** one real-git fixture with a temp gate ledger drives three
    assertions:
    - the issue-mode split sums to Hours;
    - M2 mode keeps only M2's phases;
    - milestone-close adopts its increment instead of exiting 1.
- [ ] Implement:
  - read the ledger (`readBoundaryGateLedger(plansDir(), …)`) and map the
    rounds;
  - `BuildTimeline(… since=firstISO, until=lastISO)`;
  - `Split(out.Segments, issue, tl)`;
  - set `res.Hours = (Build+Converge)/60` (D2: the actual changes);
  - milestone mode keeps only phases after the previous boundary's finalize;
  - drop the milestone early return in `adoptOmittedActual` and carry the
    split on `closeFlags.Split`;
  - `printActual` shows `build Xh · converge Yh (source)`.
- [ ] Update the recovery catalog, help text (`helptext/actual.md`,
      `milestone-close` help) and atlas for the changed verb surface. This is
      the lessons.md rule that a verb's surface is help + atlas + catalog in
      one commit.
- [ ] Run the targeted tests (PASS) and commit: `#234 M1: actuals split at boundary dispatch; milestone-close derives its increment`.

### Task 5: ledger columns + E3

**Files:** `internal/estimate/ledger.go`, `ledger_test.go`, `close.go`
(`appendCalibrationRow`, the OK line), `close_ledger_test.go`.

- [ ] Tests.
  - **Strategy:** a round-trip table over every historical width (10, 20,
    23, 28) plus `rowSplitMeta` over a ledger whose only round is the
    finalizing one (PQ-1).
  - E3 assertions:
    - est-0 → excluded and ratio `-`;
    - a trailing full est-0 row does not disable drift;
    - the quick close line wording.
  - The ratio column has no programmatic reader beyond `ParseRows`, which
    recomputes it; brain velocity only cites the ledger in prose (checked).
- [ ] Implement: append the header columns and `splitCols = 28`; the block is
      read only when `len(c) >= 28`. Update FormatRow, ParseRows,
      `ExcludedFromCalibration` and the writer.
- [ ] Run (PASS) and commit: `#234 M1: ledger carries build/converge split; est-0 rows are not calibration points`.

### Task 6: M1 close

- [ ] Atlas: `atlas/workflow/sdlc-binary.md` (the active-time paragraph and
      the ledger columns), `atlas/workflow/ledger-landscape.md`.
- [ ] Run `make test`, then `sdlc milestone-close --issue 234 --milestone M1`.
      With E1 live, it derives its own increment.

## Chunk 2: M2 — drift on build, k report, docs

- [ ] M2 — drift reads est vs build; converge-factor report; brain SKILL.md

### Task 7: drift reads est against build

- [ ] Tests.
  - **Strategy:** a drift table mixing legacy and split rows.
  - Only split rows with build > 0 count toward n, and the ratio is
    est/build.
- [ ] Implement `LedgerRow.BuildRatio()` and have `calibratable` require a
      split. Run (PASS) and commit.

### Task 8: `sdlc converge-report`

**Files:** create `internal/estimate/converge.go` and `converge_test.go`,
`cmd/sdlc/convergereport.go`, `helptext/converge-report.md`, a main.go
registration, and an entry in the verb catalog.

- [ ] Tests.
  - **Strategy:** grouping and dedupe over a mixed ledger, plus the empty
    case: `no rows with a build/converge split yet (ariadne#234; backfill
    #323)`.
- [ ] Implement. Run (PASS) and commit.

### Task 9: docs + close

- [ ] Brain `velocity/SKILL.md`: the new columns, the cut rule D1/D2, and the
      fact that old rows are `split_source -` and not comparable. Brain
      auto-commits; edit the file only, no sdlc there.
- [ ] Atlas for the drift and converge-report surface.
- [ ] Run `make test`, then `sdlc close --issue 234`.

## Risks

- **Clock skew between the transcript and git author times.** Both are local
  wall-clock on one machine, and the cuts compare like with like.
- **#183 (`--fixed-then-ship` re-close) lands concurrently.** Its narrow
  re-close is another round on the same boundary, so it is converge by D1
  without code here. Whichever lands second checks this via ops:0.

## Revisions

- 2026-10-10, change-code plan-quality round 1:
  - PQ-1: row metadata is read from the persisted ledger at row-write time.
  - PQ-2: the D1 precedence is stated, and phases never overlap.
  - Minors:
    - `dispatchNow` seam;
    - no stamp on a failed dispatch;
    - Finalizing ← Reviewed, the last one per boundary;
    - a malformed stamp falls back to finalize;
    - the test prose is compressed into strategies plus a property test;
    - the ratio-column consumer sweep is done (none).
