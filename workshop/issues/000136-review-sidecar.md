---
id: 000136
status: done
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-29
estimate_hours: 0.59
started: 2026-06-29T15:42:51-07:00
actual_hours: 0.48
---

# sdlc boundary review sidecar

## Problem

`sdlc` boundary reviews can produce more output than an interactive agent can
reliably keep in scrollback. During pair#81 retro, pair#72 boundary-review
output was effectively a transient terminal artifact: the close/milestone gate
could make a decision, but the agent did not have a stable file to reopen for
details after truncation or context compaction.

Boundary reviews are workflow evidence. They should be persisted as first-class
sidecar artifacts alongside the issue/plan record instead of existing only as
TTY output and commit trailers.

## Spec

`sdlc` boundary-review paths should write the full review transcript to a
durable sidecar file under `workshop/plans/`, following the existing durable
planning artifact convention.

- Milestone review sidecars use
  `workshop/plans/NNNNNN-slug-m2-review.md` for milestone `M2`.
- Issue-close review sidecars use an equally predictable name, such as
  `workshop/plans/NNNNNN-slug-close-review.md`.
- If a review is re-run for the same boundary, the command must not silently
  destroy prior evidence. Either append a timestamped revision section or choose
  a deterministic collision-safe suffix; the chosen behavior must be documented.
- The sidecar includes enough metadata to orient a fresh reader:
  issue id/title, repo identity, issue file path, boundary kind, milestone when
  present, base/head SHAs, command invoked, reviewer agent/model if known,
  timestamp, verdict, and the full review body.
- Terminal output from `sdlc close`, `sdlc milestone-close`, and the underlying
  boundary-review judge should print a compact verdict/summary plus the sidecar
  path, rather than relying on the terminal as the only durable surface for the
  full review.
- Existing behavior remains intact: `Review-Verdict:` trailers, close gates,
  issue log expectations, and failure behavior continue to work.
- The sidecar is intended for agents to read after the gate runs, including
  after scrollback loss, context compaction, or a follow-up session.

## Done when

- Boundary review output is persisted under `workshop/plans/` for both
  milestone close and full issue close.
- The naming convention covers milestone and issue-close boundaries and handles
  re-runs without silent overwrite.
- The terminal prints the verdict and sidecar path in a compact form.
- Tests cover path naming, required metadata, preserved review body, and
  existing verdict/trailer behavior.
- Help or atlas documentation tells agents where to find the sidecar.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.25 impl=0.3
design-buffer: 0.15
total: 0.59
```

A new single-concern module (`reviewsidecar.go`: pure render + path + atomic IO
writer) plus additive wiring into the shared boundary-review dispatch and 2 call
sites, ~150 lines of tests, and an atlas note. Design is largely pre-resolved by
the durable plan (decisions D1–D5), so reduced design + +15% buffer; impl is the
v3.1 40%-scaled greenfield range.

## Plan

Detailed design + TDD task breakdown: `workshop/plans/000136-review-sidecar-plan.md`.

- [x] Locate the close/milestone-close boundary review dispatch and output path.
- [x] Define sidecar naming, re-run semantics, and metadata fields.
- [x] Persist the full review body atomically under `workshop/plans/`.
- [x] Adjust CLI output to show compact verdict plus sidecar path.
- [x] Add regression tests for milestone close, issue close, and trailer/gate preservation.
- [x] Update help or atlas docs with the sidecar convention.

## Log

### 2026-06-26

- Created from pair#81 retro point 6: boundary reviews need a durable sidecar
  file, e.g. `workshop/plans/NNNNNN-slug-m2-review.md`, so agents can read the
  full review after the gate runs.

### 2026-06-29
- 2026-06-29: closed — go test ./cmd/sdlc/... all pass. New reviewsidecar_test.go: pure path-naming (close+milestone) + render metadata/revision + create-then-append (prior evidence preserved, no overwrite) + atomic no-temp-leak + missing-issue error. Integration: TestRunCloseWithReview_IssueClose_Dispatches asserts the close sidecar file+body+metadata; TestDispatchBoundaryReview_WritesMilestoneSidecar pins -m1-review.md via shared dispatch. Existing close/milestone tests stay green (additive, non-fatal write). This close itself dogfoods the feature → writes workshop/plans/000136-review-sidecar-close-review.md.; review verdict: FIX-THEN-SHIP

Implemented per `workshop/plans/000136-review-sidecar-plan.md`. New
`cmd/sdlc/reviewsidecar.go` — pure `sidecarMeta` + `sidecarPath` +
`renderReviewEntry` (zero-IO unit tests) behind a thin `writeReviewSidecar`
(atomic temp+rename). Wired into the **single shared** `dispatchBoundaryReview`
(milestoneclose.go), so both `sdlc close` and `sdlc milestone-close` persist the
full transcript to `workshop/plans/NNNNNN-slug-{close|m<x>}-review.md` with a
metadata header (issue/repo/issue-file/boundary/milestone/window/command/reviewer/
timestamp/verdict) + body. Re-runs append a `## Re-review` section (never
overwrite). Strictly additive: trailers/log-annotation/verdict/gates unchanged,
write is non-fatal; full body still prints to stdout + a compact `review sidecar:
<path>` line (D3 — keeps the in-session gate intact).

Discoveries:
- `judge.Verdict` is `type Verdict string` with values = the labels (`"SHIP"` …),
  so `string(verdict)` yields the label directly — no `.String()` method (the one
  edge the plan-quality judge flagged to verify).
- Reused `issueTitleFromContent` (changecode.go) for the title — ARCH-DRY.
- `atomicWriteFile` is the first temp+rename writer in `cmd/sdlc` (only prior
  `os.Rename` uses are archive-moves / lock graveyard) — duplicates nothing.

Verification: `go test ./cmd/sdlc/...` all pass. New `reviewsidecar_test.go`
(path naming close+milestone, render metadata completeness + revision heading,
create-then-append preserving prior evidence, no temp-file leak, missing-issue
error). Integration: `TestRunCloseWithReview_IssueClose_Dispatches` now asserts
the close sidecar file + body + metadata; new
`TestDispatchBoundaryReview_WritesMilestoneSidecar` pins the `-m1-review.md` path
through the shared dispatch. Existing close/milestone tests stay green
(existing-behavior-intact).

Boundary review (the close dogfooded the feature → wrote
`workshop/plans/000136-review-sidecar-close-review.md`): verdict **FIX-THEN-SHIP**,
two Important findings, both fixed before the boundary:
- **Reviewer cell empty in the default invocation** — `writeReviewSidecar` recorded
  the raw `--agent` flag (`""` by default) instead of the resolved dispatch agent.
  Fixed: `dispatchBoundaryReview` now threads `string(opts.Agent)` into the write;
  pinned by a non-empty-reviewer assertion in the close integration test. (The
  already-written sidecar's cell was hand-corrected to `claude`, the real reviewer.)
- **D4 no-write untested** — added a no-sidecar `os.Stat` assertion to the
  `--no-judge` skip test.
Minor `repoIdentity` triplication finding deliberately not taken (the other two
sites don't share the basename-only shape — Simplicity-First); see plan
`## Revisions`. Re-review happens at the pre-merge judges.
