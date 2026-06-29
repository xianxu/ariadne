# Global Commit Boundary Active Time Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prevent long-running issue commit spans from absorbing activity that is closer to intervening commits for other issues.

**Architecture:** Keep attribution as a pure transformation in `cmd/sdlc/internal/activetime` (ARCH-PURE): loaded events, spans, and commits enter the pure core; CLI commands only render results. Replace the current target-enclosing segment interpretation with a global commit-boundary claimant model, reusing existing event/span gap math and commit issue parsing (ARCH-DRY). The purpose is not a blunt cap; caps/warnings only surface remaining suspicious attribution when boundary evidence is insufficient (ARCH-PURPOSE).

**Tech Stack:** Go, pure unit tests in `cmd/sdlc/internal/activetime`, CLI renderer tests in `cmd/sdlc`, SDLC atlas/helptext docs.

---

## Implementation Contract

Preserve the existing pure-core / thin-IO split (ARCH-PURE):

- IO stays in `loadEventsWithFiles` and `loadWindowCommits`.
- `loadWindowCommits` already loads every commit in the window. Keep that shape,
  but change commit issue parsing to use an all-issue ref pattern for commit
  subjects, not only `issuePattern(opts.Issues)`. Commits for intervening issues
  must become claimants even when the caller seeded only the primary issue.
  Commits with no issue refs remain neutral temporal boundaries.
- Add source/session identity at the loader boundary by extending `Event` and
  `TaskSpan` with `Source string`, populated from the JSONL file path. Tests that
  construct values directly may leave it empty; pure helpers treat empty source
  as one synthetic source.
- Pure attribution lives in `segment.go`. `Compute` remains the orchestrator:
  load evidence, call pure helpers, sum `Result.PerIssue`, compute warnings.
- CLI code only renders `Result.Segments`, `Result.PerIssue`, and
  `Result.Warnings`. Warnings are visible but do not change exit status.

Pure helper seam:

```go
type ActivityRun struct {
    Start, End time.Time
    Active     float64
    Mentions   map[string]int
    Source     string
}

type AttributionWarning struct {
    Issue  string
    Start  time.Time
    End    time.Time
    Active float64
    Share  float64
    Reason string
}

func activityRuns(events []Event, spans []TaskSpan, thresholdMin int) []ActivityRun
func claimActivityRuns(runs []ActivityRun, commits []Commit, commitWeight, prefixWeight float64) []Segment
func attributionWarnings(segs []Segment, perIssue map[string]float64) []AttributionWarning
```

`activityRuns` emits claimable intervals per source/session: capped event gaps
between consecutive events in the same source plus full task spans in that same
source. Overlaps are unioned within one source only, never across sources.

`claimActivityRuns` chooses one claimant per run:

- find the closest previous and next issue-referenced commits around the run;
- distance to the next commit is `next.Time - run.End`;
- distance to the previous commit is `run.Start - prev.Time`;
- choose the smaller non-negative distance; on exact tie choose the next commit;
- commits without issue refs cut spans but never claim;
- use mention/unattributed fallback only when no issue-referenced boundary exists
  on either side of the run.

When a commit claimant exists, mentions do not allocate to other issues. At
`commitWeight == 1.0` the full run goes to the claimant issue refs. At
`commitWeight < 1.0`, the commit-weighted share goes to claimant issue refs and
the remaining share goes to `UnattributedKey`; mention fallback is reserved for
the no-plausible-boundary case. This intentionally narrows the old
`attributeSegment` behavior because baseline-v3 showed cross-issue mentions are
the contamination source.

Warning thresholds are constants in `compute.go` next to the warning function:

```go
const suspiciousSpanMin = 120.0
const suspiciousShare = 0.50
```

Emit a warning when one segment/run contributes more than 50% of an issue total
and that segment's wall span exceeds 120 minutes, or when attribution falls back
to mentions/unattributed because no plausible issue boundary exists. Warning
text includes the issue key, segment start/end, minutes, share, and reason.

## Core Concepts

| Name | Lives in | Status |
|------|----------|--------|
| `ActivityRun` | `cmd/sdlc/internal/activetime/segment.go` | new |
| `CommitBoundary` | `cmd/sdlc/internal/activetime/segment.go` | new |
| `AttributionWarning` | `cmd/sdlc/internal/activetime/compute.go` | new |
| `Segment` | `cmd/sdlc/internal/activetime/segment.go` | modified |

**ActivityRun** — one claimable unit of transcript work: event gaps and task spans that should be attributed together.
- **Relationships:** N:1 with transcript source/session; N:N with commit boundaries as candidates; 1:1 with one final attribution decision.
- **DRY rationale:** Reuses `activeMinutesUnion` for minutes; avoids per-command reimplementation of time math.
- **Future extensions:** Can carry source/session identity if later attribution needs per-harness confidence.

**CommitBoundary** — a commit timestamp that cuts the activity timeline; issue refs on the commit make it a claimant for nearby activity.
- **Relationships:** 1:N from a commit to zero or more issue refs; N:1 from activity runs to the selected boundary when the boundary has claimant refs.
- **DRY rationale:** Existing `Commit` already carries `Time`, `SHA`, `Subject`, and `Issues`; use that shape and share the all-issue-ref regexp with `gitx.DiscoverWindowIssues` or a small activetime-local equivalent rather than inventing a second subject parser.
- **Future extensions:** Add confidence/distance metadata for warnings.

**AttributionWarning** — visible diagnostic when attribution remains suspect.
- **Relationships:** N:1 with a `Result`; may reference a `Segment`/boundary and issue key.
- **DRY rationale:** One warning surface consumed by both `sdlc active-time` and `sdlc actual`/close suggestion.
- **Future extensions:** Thresholds can widen from dominant-boundary warnings to low-actual/single-commit warnings from #127.

**Segment** — rendered row in `sdlc active-time`.
- **Relationships:** Existing renderer expects segments; keep compatibility by rendering attributed activity runs as segments.
- **DRY rationale:** Preserve one table renderer and one per-issue sum path.
- **Future extensions:** Add fields such as `Warning` or `BoundaryDistance` without changing loaders.

## Integration Points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| Transcript/git loaders | `cmd/sdlc/internal/activetime/compute.go`, `event.go`, `commit.go` | modified | filesystem + git |
| Active-time CLI renderer | `cmd/sdlc/activetime.go` | modified | stdout/stderr |
| Actual close suggestion | `cmd/sdlc/actual.go`, `cmd/sdlc/close.go` | modified | user-facing warnings |

**Transcript/git loaders** — continue loading evidence only; no attribution logic moves into IO. The transcript loader adds source/session identity to `Event` and `TaskSpan`; the git loader keeps returning all commits in the window with all commit-subject issue refs parsed by `uniqueRefs`.
- **Injected into:** `Compute`, which passes loaded slices to pure attribution.
- **Future extensions:** Add more transcript sources without touching attribution.

**Active-time CLI renderer** — prints segments, totals, and warnings.
- **Injected into:** none; consumes `activetime.Result`.
- **Future extensions:** Render confidence/debug fields behind flags.

**Actual close suggestion** — surfaces contaminated-number warnings when `sdlc actual` or close computes a suggested actual.
- **Injected into:** no new external service; consumes `activetime.Result.Warnings`.
- **Future extensions:** Close can refuse extremely suspect measured actuals if future evidence warrants.

## Chunk 1: Pure Boundary Attribution

### Task 1: Pin the contamination fixture

**Files:**
- Modify: `cmd/sdlc/internal/activetime/segment_test.go`

- [ ] **Step 1: Write failing test**

Add `TestBuildSegments_GlobalBoundariesPreventLongIssueAbsorption` with commits:

```text
00:00 #1 c11
00:20 #2 c21
00:40 #2 c22
01:00 #3 c31
01:20 #2 c23
01:40 #3 c32
03:00 #1 c12
```

and activity events/runs near the issue 2/3 commits. Assert issue 1 does not receive the intervening issue 2/3 activity just because `c12` closes a long span.

Specific assertions:
- `#1` receives only the run adjacent to `c12`, not runs between `c21`/`c22`,
  `c22`/`c31`, `c31`/`c23`, or `c23`/`c32`.
- `#2` receives runs nearest to `c22` and `c23`.
- `#3` receives runs nearest to `c31` and `c32`.
- total allocated minutes equals total run active minutes.

- [ ] **Step 2: Verify RED**

Run: `go test ./cmd/sdlc/internal/activetime -run TestBuildSegments_GlobalBoundariesPreventLongIssueAbsorption -count=1`
Expected: FAIL under current target-enclosing segment behavior if the fixture is expressed in the existing segment model.

- [ ] **Step 3: Add pure attribution helper**

Introduce focused helpers in `segment.go`:
- `activityRuns(events []Event, spans []TaskSpan, thresholdMin int) []ActivityRun`
- `claimActivityRuns(runs []ActivityRun, commits []Commit, commitWeight, prefixWeight float64) []Segment`

Keep `activeMinutesUnion` as the minute math source. Candidate selection uses nearest plausible issue-referenced commit boundary, preferring the next commit on ties. Commits without relevant refs remain neutral temporal cut points so long spans cannot silently bridge across them.

Replace `buildSegments`' target-enclosing segment loop with a wrapper around the
new helpers so existing callers and renderers still consume `[]Segment`.
Replace or narrow `attributeSegment` so commit-boundary attribution never gives
mention share to other issues when a plausible issue commit exists. Keep one
shared allocation helper for `claimActivityRuns` and the no-commit fallback
(ARCH-DRY); do not leave old and new weighting rules in parallel.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./cmd/sdlc/internal/activetime -run 'TestBuildSegments_GlobalBoundariesPreventLongIssueAbsorption|TestBuildSegments' -count=1`
Expected: PASS.

### Task 2: Preserve overlapping parallel issue work

**Files:**
- Modify: `cmd/sdlc/internal/activetime/segment_test.go`

- [ ] **Step 1: Write failing test**

Add a fixture with two overlapping task spans or session runs that are nearest to different issue commits. Assert each issue receives its own 15 minutes, not one globally unioned 15-minute bucket.

Construct the fixture with two sources represented by `Event.Source`:
- source A: `[10:00,10:15]`, nearest next commit `#8` at `10:16`;
- source B: `[10:00,10:15]`, nearest next commit `#9` at `10:17`.

Assert `#8 == 15`, `#9 == 15`, and `TotalActive == 30`. This deliberately
differs from `activeMinutesUnion`'s intra-source behavior, where overlapping
intervals inside one source still union to wall-clock.

- [ ] **Step 2: Verify RED**

Run: `go test ./cmd/sdlc/internal/activetime -run TestBuildSegments_ParallelRunsCanAttributeToDifferentIssues -count=1`
Expected: FAIL if global union collapses overlapping sessions.

- [ ] **Step 3: Implement without global union**

Ensure activity runs retain source/run identity until after attribution. Union only inside a single run where needed to avoid double-counting that run's own spans.

Add `TestActivityRunsUnionWithinSourceOnly`, proving two overlapping intervals
in the same source union, while equal intervals in two sources remain two runs.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./cmd/sdlc/internal/activetime -run TestBuildSegments_ParallelRunsCanAttributeToDifferentIssues -count=1`
Expected: PASS.

### Task 3: Pin boundary selection edge cases

**Files:**
- Modify: `cmd/sdlc/internal/activetime/segment_test.go`

- [ ] **Step 1: Write failing tests**

Add tests for the selector behavior before implementation:
- `TestClaimActivityRuns_PrefersNextCommitOnTie`: equidistant previous/next
  issue commits choose the next commit.
- `TestClaimActivityRuns_NeutralCommitCutsButDoesNotClaim`: a no-issue commit
  between a run and a later target commit prevents a rendered segment from
  spanning across it, but receives no allocation.
- `TestClaimActivityRuns_PreviousCommitFallback`: when only a previous
  issue-referenced commit exists, it claims the run.
- `TestClaimActivityRuns_MentionFallbackOnlyWithoutIssueBoundary`: mention
  fallback is used only when there is no previous or next issue-referenced
  boundary.
- `TestClaimActivityRuns_BoundarySuppressesMentionAllocation`: with
  `commitWeight < 1.0`, a run mentioning `#9` but claimed by a nearby `#8`
  commit allocates the commit-weighted share to `#8`, zero to `#9`, and the
  remainder to `UnattributedKey`.

- [ ] **Step 2: Verify RED**

Run: `go test ./cmd/sdlc/internal/activetime -run 'TestClaimActivityRuns|TestActivityRuns' -count=1`
Expected: FAIL until the selector and source-aware run construction exist.

- [ ] **Step 3: Implement selector**

Keep the selector pure and local to `segment.go`. It should accept a run and the
already-loaded commit slice; it must not call git, read files, or parse subjects.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./cmd/sdlc/internal/activetime -run 'TestClaimActivityRuns|TestActivityRuns|TestBuildSegments' -count=1`
Expected: PASS.

### Task 4: Make intervening issue commits real claimants

**Files:**
- Modify: `cmd/sdlc/internal/activetime/commit.go`
- Modify: `cmd/sdlc/internal/activetime/commit_test.go`
- Modify: `cmd/sdlc/internal/activetime/compute_test.go`

- [ ] **Step 1: Write failing loader tests**

Add `TestLoadWindowCommitsParsesAllIssueRefsForClaimants`: fake git log output
contains `#1`, `#2`, `#3`, and a no-ref commit; call the same loader path that
`Compute` uses with `Options{Issues: []string{"1"}}`; assert commits for `#2`
and `#3` still carry `Issues`.

Add `TestComputeDiscoversInterveningIssueClaimants`: seed only issue `#1` in
`Options.Issues`, fake commits in the `c11/c21/c22/c31/c23/c32/c12` shape, and
assert `Result.PerIssue` includes `#2` and `#3` allocations while `#1` does not
claim their runs.

- [ ] **Step 2: Verify RED**

Run: `go test ./cmd/sdlc/internal/activetime -run 'TestLoadWindowCommitsParsesAllIssueRefsForClaimants|TestComputeDiscoversInterveningIssueClaimants' -count=1`
Expected: FAIL while commit issue parsing is limited to `opts.Issues`.

- [ ] **Step 3: Implement all-ref commit parsing**

Introduce one all-issue-ref pattern for commit subjects. Prefer sharing the
regexp shape with `gitx.DiscoverWindowIssues` if a clean package boundary exists;
otherwise keep an activetime-local helper and add the loader test above as the
drift guard. Event mention parsing may still use `issuePattern(opts.Issues)` for
fallback mentions; commit claimants must not.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./cmd/sdlc/internal/activetime -run 'TestLoadWindowCommitsParsesAllIssueRefsForClaimants|TestComputeDiscoversInterveningIssueClaimants' -count=1`
Expected: PASS.

## Chunk 2: Warning Surface

### Task 5: Add suspicious attribution warnings

**Files:**
- Modify: `cmd/sdlc/internal/activetime/compute.go`
- Modify: `cmd/sdlc/internal/activetime/compute_test.go`
- Modify: `cmd/sdlc/activetime.go`
- Modify: `cmd/sdlc/actual.go`
- Modify: `cmd/sdlc/activetime_test.go`
- Modify: `cmd/sdlc/actual_test.go`

- [ ] **Step 1: Write failing pure test**

Add `TestComputeDominantBoundaryWarning` with one issue receiving >50% of its total from a run/boundary spanning more than the configured suspicious gap threshold. Assert `Result.Warnings` contains the issue, segment, and reason.

Also add `TestComputeMentionFallbackWarning` for the no-boundary fallback path.

- [ ] **Step 2: Verify RED**

Run: `go test ./cmd/sdlc/internal/activetime -run TestComputeDominantBoundaryWarning -count=1`
Expected: FAIL because `Result.Warnings` does not exist.

- [ ] **Step 3: Implement warning computation**

Add `Warnings []AttributionWarning` to `Result`. Compute warnings after segments are built and per-issue totals are known. Start with conservative thresholds: one segment/run contributes >50% of an issue total and spans >2h, or a run has no plausible boundary and falls back to mentions.

- [ ] **Step 4: Render warnings**

Print warnings in `sdlc active-time` after the segment table. In `sdlc actual`, include warning text in `actualResult.Detail` or a new field rendered before the close suggestion.

Expected output contract:
- active-time stdout includes a `# attribution warnings` section after totals.
- actual stderr includes `attribution warning:` lines before `→ close with:`.
- warning presence does not change exit code.

- [ ] **Step 5: Verify GREEN**

Run: `go test ./cmd/sdlc/internal/activetime ./cmd/sdlc -run 'TestComputeDominantBoundaryWarning|TestRunActiveTime' -count=1`
Expected: PASS.

## Chunk 3: End-To-End Validation And Docs

### Task 6: Validate against fixtures and baselines

**Files:**
- Modify: `cmd/sdlc/internal/activetime/parity_test.go` if fixture-level E2E coverage is needed.
- Modify: `workshop/issues/000092-*.md`

- [ ] **Step 1: Run focused and package tests**

Run: `go test ./cmd/sdlc/internal/activetime ./cmd/sdlc -count=1`
Expected: PASS.

- [ ] **Step 2: Run live/baseline checks**

Run `sdlc active-time` or `go test` fixtures against:
- the synthetic #92 contamination case;
- the recorded baseline-v3 portfolio window from
  `/Users/xianxu/workspace/brain/data/life/42shots/velocity/baseline-v3.md`;
- the existing golden parity fixture (`TestAttributionGolden`) as an additional
  in-repo regression guard;
- the real nous#48 window if the local transcript dirs are present.

Concrete commands:

```bash
go test ./cmd/sdlc/internal/activetime -run 'TestBuildSegments_GlobalBoundariesPreventLongIssueAbsorption|TestAttributionGolden' -count=1
go test ./cmd/sdlc/internal/activetime ./cmd/sdlc -count=1
```

Baseline-v3 command, translated from the recorded Python invocation to the
in-binary engine:

```bash
sdlc active-time \
  --dir ~/.claude/projects/-Users-xianxu-workspace-nous \
  --dir ~/.claude/projects/-Users-xianxu-workspace-brain \
  --git-repo ~/workspace/nous \
  --since 2026-05-07T16:54:00Z --until 2026-05-08T05:13:00Z \
  --issue 8 --issue 10 --issue 11 --issue 4 --issue 3 \
  --commit-weight 1.0 --threshold-min 15 --include-assistant
```

Expected per baseline-v3: `#11 ≈ 4.28h`, `#4 ≈ 2.61h`, `#10 ≈ 0.45h`,
`#8 ≈ 0.45h`, `#3 ≈ 0.07h`. A small drift from the Go port / transcript
availability must be logged in the issue; a large drift blocks close.

For the #48 live check, use `sdlc actual --issue 48` from the nous repo if
`/Users/xianxu/workspace/nous` and matching transcript sources are available; log
the measured value or the unavailability. The baseline-v3 command above is not
optional for close unless the recorded transcript dirs are unavailable, in which
case log that fact and the exact missing paths.

### Task 7: Update docs

**Files:**
- Modify: `cmd/sdlc/helptext/actual.md`
- Modify: `cmd/sdlc/helptext/close.md`
- Modify: `atlas/workflow/sdlc-binary.md`
- Modify: `atlas/workflow/ledger-landscape.md`
- Modify: `workshop/issues/000092-*.md`

- [ ] **Step 1: Document attribution model**

Describe global commit-boundary attribution, overlapping issue work, and suspicious attribution warnings. Include the mitigation: commit more often to provide boundaries.

- [ ] **Step 2: Final verification**

Run:

```bash
go test ./cmd/sdlc/... -count=1
git diff --check
```

Expected: both pass.
