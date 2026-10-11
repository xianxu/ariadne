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
  byte-identical. It is empty on plan-quality rounds and on never-ran
  rounds.
- **BoundaryRounds → Timeline**:
  `BuildTimeline(rounds []BoundaryRound, issueCommits []time.Time, since, until time.Time) (Timeline, source string)`.
  - `BoundaryRound` is a small activetime-local view: Boundary, Dispatched,
    Finalized, Finalizing. activetime must not import gatestate; the caller
    maps between them.
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
  - stamps `reviewResult.Dispatched = nowRFC3339()` just before
    `judge.Dispatch`, including on a failed dispatch, because the reviewer
    did start;
  - `persistBoundaryRound` copies it onto the round unless the round never
    ran.
  - The clock is the package's `nowRFC3339` seam, faked in tests the way
    `persistBoundaryRound`'s timestamp already is.
- **computeActual** reads the gate ledger through `readBoundaryGateLedger`
  (a missing file means no rounds, so `split_source = no-review`), maps the
  rounds to `[]activetime.BoundaryRound`, builds the timeline and splits. It
  gains:
  - `actualResult.Build/Converge/Source/Rounds/Recipe`;
  - a `milestone` parameter, which restricts the timeline to that
    milestone's phases (D7).

  The fakes are the existing real-git fixtures plus a gate ledger written
  into the test's temp `workshop/plans`.
- **appendCalibrationRow** gets the split from `closeFlags`:
  - set by `adoptOmittedActual`;
  - `manual` when `--actual` was typed.

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

- [ ] Write table tests (no IO), each asserting phases and
      Build/Converge/Tail:
  - (a) one boundary, one passing round, all before since/until → all build;
  - (b) the #321 shape: REWORK at d1, fix commits, then the final round at
    d2 → converge = d1 → last commit ≤ d2;
  - (c) M1 then M2: M1's tail (lastCommit ≤ M1 final dispatch, M1 finalize]
    is excluded, and M2's build starts at M1's finalize;
  - (d) an open boundary at measurement time (REWORK pending) → converge to
    `until`;
  - (e) a round with no `Dispatched` → its finalize time is the cut, and the
    source is `finalize`;
  - (f) no rounds → all build, source `no-review`;
  - (g) a segment straddling a cut splits by time;
  - (h) the invariant Build + Converge + Tail == ΣAlloc over random
    segments.
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

- [ ] Write failing tests:
  - computeActual over a temp repo with a gate ledger reports
    Build/Converge/Source, and Hours == (Build+Converge)/60 to 2 dp;
  - computeActual with milestone `M2` reports only M2's phases;
  - milestone-close with `--actual` omitted adopts the increment and prints
    the adopt line naming the milestone window. Today it exits 1.
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

- [ ] Write failing tests:
  - `FormatRow`/`ParseRows` round-trip the five new columns;
  - 10-, 20- and 23-column rows still parse, with `SplitSource` reading as
    `-`;
  - `UpgradeHeader` extends the header;
  - `ExcludedFromCalibration` is true for `Estimate == 0`;
  - the ratio cell is `-` when the estimate is 0;
  - the close OK line for a quick close reads `no estimate — not a
    calibration point` plus build/converge;
  - a full-flow est-0 row no longer becomes drift's "latest".
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

- [ ] Write failing tests:
  - `driftSample` ignores rows without a split (`split_source ∈ {-, manual}`)
    and rows with build 0;
  - the ratio used is est/build;
  - 5 new-rule rows all >2× over build → a warning;
  - old rows interleaved do not count toward n.
- [ ] Implement `LedgerRow.BuildRatio()` and have `calibratable` require a
      split. Run (PASS) and commit.

### Task 8: `sdlc converge-report`

**Files:** create `internal/estimate/converge.go` and `converge_test.go`,
`cmd/sdlc/convergereport.go`, `helptext/converge-report.md`, a main.go
registration, and an entry in the verb catalog.

- [ ] Write failing tests:
  - `ConvergeFactors` groups by (model, flow, recipe), dedupes with
    `NewestPerIssue`, skips legacy rows, and computes k = Σconverge/Σbuild
    with n;
  - the CLI renders a table and, with zero split rows, says
    `no rows with a build/converge split yet (ariadne#234; backfill #323)`.
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
