---
id: 000092
status: done
deps: []
github_issue:
created: 2026-06-11
updated: 2026-06-29
estimate_hours: 5.95
started: 2026-06-29T13:23:59-07:00
actual_hours: 2.86
---

# active-time-v3: a long inter-commit gap creates one fat segment that absorbs cross-session activity and over-attributes it to the segment's issue

## Problem

The active-time-v3 engine (`cmd/sdlc/internal/activetime`, #110) is
segment-anchored: it cuts the timeline at every commit in the window and
attributes each segment's active-minutes to that segment-ending commit's issue
refs (at `--commit-weight 1.0`, **100%** by commit). The per-event 15-min gap cap
(`activeMinutes`) correctly
stops a single idle gap from counting as hours — but it does **not** bound a
*segment's* width. When an issue goes a long calendar time between two of its
commits, that one segment spans the whole gap and **vacuums up every transcript
event in the watched dirs (repo + brain) during that span** — including unrelated
other-session activity and brain autosave/sync — and pins all of it on the
segment's issue.

### Concrete evidence (nous#48, the issue that surfaced this)

`sdlc actual --issue 48` reported **14.47h**, vs the per-milestone measurements
of M1 1.11h + M2 4.85h. The v3 per-segment table shows why — one row dominates:

```
33  2026-06-08 16:55 → 2026-06-11 10:03   779.5 min   3fa9414 #48 M2: certify real-Microsoft   #48=779.5m
```

That single segment is **779.5 min ≈ 13h, attributed 100% to #48 — ~90% of #48's
measured total (870 min).** It spans the **2.7-day gap** between the last 06-08
commit (`ee6a387`, 16:55) and the first 06-11 commit (`3fa9414`, 10:03), i.e. the
stretch where the operator stepped away. The properly-anchored, short-span #48
work segments sum to only ~1.4h; the genuine focused effort was ~4h. The other
~10h is **cross-session contamination** absorbed by the fat segment: an
interrupted multi-hour resume session plus unrelated `nous`/`brain` transcript
activity over those 2.7 days, none of it filtered because v3 can't tell a
06-09 brain-autosave event from #48 work — it only sees "an event inside the
#48-anchored segment."

### Why it matters

`sdlc close` treats the v3 number as the authoritative `actual_hours` for velocity
calibration (#68 built v3 precisely so actuals are *measured, not guessed*). A 3–4×
over-count on any issue whose work straddles a multi-day pause silently poisons the
calibration baseline — and "work spanning a pause" is common (review waits,
operator-collaborative steps, weekends). The failure is invisible: the number
*looks* measured. nous#48 only caught it because the operator questioned a
suspiciously high close number; most won't.

The script's own docstring already flags the adjacent gap: *"Parallel-session
dedup not yet implemented for v3 (rare in practice for the operator)."* This is
that gap biting in practice.

## Spec

Replace the target-issue enclosing-span model with **global commit-boundary
attribution**.

The input evidence is limited and noisy:

1. transcript activity intervals (15-minute-capped event gaps plus task spans),
   with overlapping chat sessions allowed to count as separate issue work; and
2. commit timestamps with issue refs, which are the best available "work ended
   for this issue" signal.

The desired model is: each activity run is claimed by the nearest plausible
commit boundary, with a bias toward the next commit because commits usually close
the work that preceded them. A long-running issue's later commit must not claim
activity that is closer to intervening commits for other issues.

Using the running example:

```text
issue 1: c11..       ..              ......c12
issue 2:      c21..c22.....c23
issue 3:              c31........c32
```

The effective boundary timeline is:

```text
c11 | c21 | c22 | c31 | c23 | c32 | c12
```

Activity between `c21` and `c22` should primarily belong to issue 2, not issue 1,
even though issue 1's `c11 → c12` span encloses it. If three sessions produce
three overlapping 15-minute activity runs for three different issues, the per-
issue calibration ledger should hand out 15 minutes to each issue. Global
operator capacity/concurrency is a separate model; #92 is about per-issue actual
work, not reducing issue-minutes to one wall-clock union.

Algorithm shape:

1. Build activity runs from transcript events/spans. Preserve overlapping runs
   from different sessions as separate claimable work units; union only where the
   same run/session would otherwise double-count itself.
2. Build a global boundary list from all commits in the active measurement
   window. Commits with issue refs are claimants; commits without relevant issue
   refs still cut time as neutral boundaries and can trigger diagnostics when
   activity has no plausible claimant.
3. For each activity run, choose candidate commits near the run:
   - next commit after the run (preferred);
   - previous commit before the run (fallback / tie participant);
   - mention-weighted fallback only when no plausible commit boundary exists.
4. Assign each run to the nearest plausible boundary, preferring the next commit
   on ties. If there are more commit claimants than activity runs, some commits
   get zero. A commit is boundary evidence, not proof that time existed.
5. Surface a visible warning when a run/segment still spans an implausibly large
   gap or one boundary dominates an issue total. A cap may be useful as a safety
   warning, but it is not the primary remedy; if a cap would fire, the attribution
   algorithm probably failed to find the right boundary.

Keep the cheap behavioral mitigation in SDLC guidance: commit more often, because
more boundaries give the attribution model better evidence.

## Done when

- A single multi-day inter-commit gap can no longer attribute intervening
  issue-referenced activity to the gap's issue; re-running v3 over the nous#48
  window attributes ≈ the focused-effort figure (~4h), not 14.47h.
- The fix is validated against the recorded human baselines (`baseline-v3.md`) to
  confirm it doesn't regress normal single-session attribution (the v3 method was
  tuned to within ~5% of human numbers — keep that).
- `sdlc actual` / v3 surfaces a visible flag when attribution remains suspicious
  (one boundary/run dominates an issue total, or a large gap had no plausible
  intervening boundary).
- SDLC guidance notes the commit-frequency mitigation.

## Estimate

P50 5.95h. The work is primarily a pure attribution-model change with synthetic
fixtures plus documentation. The estimate uses v3.1's scaled implementation
hours; it may be revised after the detailed plan resolves whether the CLI needs a
new warning surface.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec          design=0.5 impl=0.1
item: pensive             design=1.0 impl=0.1
item: greenfield-go-module design=1.0 impl=1.2
item: smaller-go-module   design=0.2 impl=0.8
item: atlas-docs          design=0.2 impl=0.1
item: milestone-review    design=0.0 impl=0.3
design-buffer: 0.15
total: 5.95
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

- [x] Reproduce the contamination with a synthetic fixture matching the `c11/c21/c22/c31/c23/c32/c12` shape; assert the old target-issue segment lets issue 1 claim intervening issue 2/3 activity.
- [x] Introduce a pure global-boundary attribution model: activity runs are claimable work units; commit refs are boundaries/claimants; nearest plausible boundary wins with next-commit bias.
- [x] Preserve overlapping sessions as separate issue work when they attribute to different issues; do not globally union them into one operator-capacity ledger.
- [x] Validate the nous#48 window and the `baseline-v3.md` human-baselined sessions; confirm the fix removes the 14.47h artifact without regressing normal sessions.
- [x] Add suspicious-attribution warnings to `sdlc active-time` / `sdlc actual` output.
- [x] Document the commit-frequency mitigation in the SDLC guidance.

## Revisions

### 2026-06-29 — reframed from segment cap to global boundary attribution

Reason: discussion clarified that the issue is not dormant time itself, and a
hard segment cap is only a symptom detector. The real failure is letting a target
issue's long enclosing span claim activity that is closer to intervening commits
for other issues.

Delta: replace the primary remedy with global commit-boundary attribution. Keep
warnings for suspicious oversized attribution, but treat caps as fallback
signals, not the algorithm. Preserve overlapping parallel sessions as separate
per-issue work; global operator concurrency belongs to a separate capacity model.

## Log

### 2026-06-11

Filed from nous#48 (shim(oauth) Microsoft n=2-real). That issue's close measured
14.47h; the operator questioned it; investigation traced ~90% of the figure to one
2.7-day-spanning segment (table above) and corrected nous#48's `actual_hours` to a
labeled judgment of 4.0h. This issue is the upstream fix so the artifact doesn't
silently recur. Source: `cmd/sdlc/internal/activetime/segment.go`
(`attributeSegment` / `buildSegments`); the package doc notes
"parallel-session dedup not yet implemented" as a known gap.

### 2026-06-29
- 2026-06-29: closed — go test ./cmd/sdlc/... -count=1 passed; git diff --check passed; second boundary-review REWORK fixes route no-commit fallback through source-scoped activity runs, preserve overlapping fallback sessions, render fallback segments, and warn on unattributed fallback allocations; review verdict: FIX-THEN-SHIP
- 2026-06-29: closed — go test ./cmd/sdlc/... -count=1 passed; git diff --check passed; boundary-review REWORK fixes added neutral no-ref delimiter behavior and no-commit fallback warnings; historical baseline-v3 and nous#48 transcript windows were attempted but unavailable locally as documented in the issue log; review verdict: REWORK
- 2026-06-29: closed — go test ./cmd/sdlc/... -count=1 passed; git diff --check passed; baseline-v3 and nous#48 live commands were attempted but local historical transcript windows returned 0 events, so synthetic #92 claimant tests plus the updated attribution golden carry the regression evidence; review verdict: REWORK

- Claimed #92 and reframed the spec after operator review. The agreed design:
  collapse issue commits into one boundary timeline and attribute activity runs
  to nearest plausible commit boundaries instead of bounding the target issue's
  long span with a blunt cap.
- Refined the durable plan after the `change-code` plan-quality gate flagged
  vague seams: the plan now names source-aware activity runs, boundary selection
  edge cases, warning thresholds, and concrete validation commands.
- Refined again after the gate caught two purpose-level gaps: commit issue refs
  must be parsed for all intervening issue claimants even when the caller seeds
  one primary issue, and mention fallback must not allocate share when a commit
  boundary exists. The plan now includes the recorded baseline-v3 command and
  expected values instead of treating the in-repo golden as a substitute.
- Implemented the global-boundary attribution model in `cmd/sdlc/internal/activetime`:
  source-aware activity runs, all-issue commit claimants, nearest boundary
  selection, mention fallback only without a boundary, and attribution warnings
  rendered by `sdlc active-time` / `sdlc actual`. Focused verification passed:
  `go test ./cmd/sdlc/internal/activetime ./cmd/sdlc -count=1`.
- Baseline/live validation caveat: the recorded baseline-v3 command and
  `sdlc actual --issue 48` in `/Users/xianxu/workspace/nous` are both telemetry
  unavailable locally now (0 transcript events in those historical windows),
  despite the transcript dirs existing. Automated coverage therefore carries the
  regression evidence: synthetic #92 claimant tests plus the updated attribution
  golden.
- Addressed boundary-review REWORK findings: neutral no-ref commits now cut off
  previous/next issue claimants without claiming the run; no-commit mention
  fallback now produces attribution warnings; the unused commit-loader pattern
  parameter was removed; the `Segment` comment and durable plan were reconciled
  with the implemented `Commit` boundary role. Exact historical validation
  results: baseline-v3 command used `~/.claude/projects/-Users-xianxu-workspace-nous`,
  `~/.claude/projects/-Users-xianxu-workspace-brain`, `~/workspace/nous`,
  `2026-05-07T16:54:00Z..2026-05-08T05:13:00Z` and returned 0 events / 23
  commits; nous#48 check used `/Users/xianxu/workspace/nous` and returned no
  measurable activity for #48.
- Addressed the second boundary-review REWORK findings: no-commit measured
  windows now use the same source-scoped activity-run attribution path as
  commit-backed windows, so overlapping fallback sessions are preserved and
  rendered as segments; unattributed fallback allocations now emit explicit
  warnings. Verification passed: `go test ./cmd/sdlc/... -count=1` and
  `git diff --check`.
- Addressed boundary-review FIX-THEN-SHIP notes before merge: the durable plan
  checklist is now ticked, `AttributionWarning` carries `time.Time` boundaries
  and formats at the CLI boundary, and the exact
  `TestBuildSegments_GlobalBoundariesPreventLongIssueAbsorption` fixture exists.
  Verification passed: `go test ./cmd/sdlc/... -count=1` and `git diff --check`.
