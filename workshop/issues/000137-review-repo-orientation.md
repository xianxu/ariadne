---
id: 000137
status: done
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-29
estimate_hours: 0.53
started: 2026-06-29T20:56:16-07:00
actual_hours: 0.27
---

# sdlc boundary review repo orientation

## Problem

Boundary review prompts can misorient a fresh reviewer about which repository is
under review. During pair#81 retro, review context for pair work was observed to
refer to an `ariadne#...` issue shape, even though the operating repository was
`pair`.

This is risky because the boundary reviewer is intentionally fresh-context. If
the prompt names or implies the wrong repo, the reviewer can inspect the wrong
tracker, apply ariadne base-repo assumptions to a downstream repo, or report
findings against the wrong issue surface.

## Spec

Tighten `sdlc` boundary-review orientation so every close and milestone review
prompt identifies the current repo and issue owner explicitly and accurately.

- The prompt must include the repository under review, derived from the current
  git root/remote/cwd context rather than a hardcoded `ariadne` label.
- The prompt must include enough concrete anchors for a fresh reviewer:
  repo slug/name, repo root path, issue reference such as `pair#72`, issue file
  path, boundary kind, milestone when present, base SHA, and head SHA.
- The prompt must distinguish base-repo work from downstream/peer repo work. A
  reviewer operating in `pair` should be told that `pair` is the reviewed repo;
  ariadne should appear only when it is actually the reviewed repo or explicitly
  relevant as a dependency.
- Existing boundary-review behavior remains intact: verdict format, trailers,
  gates, and side effects should not change except for clearer orientation.
- This should apply consistently to `sdlc close`, `sdlc milestone-close`, and
  the underlying boundary-review judge/prompt construction path.

## Done when

- Boundary review prompts name the current reviewed repo accurately for ariadne
  and for a downstream repo fixture.
- Prompt text includes repo root, issue file, issue reference, base/head SHAs,
  boundary kind, and milestone when applicable.
- Tests fail if prompt construction falls back to a hardcoded `ariadne#N`
  reference for non-ariadne repos.
- Existing verdict/trailer/gate tests continue to pass.
- Help, atlas, or prompt comments document the repo-orientation contract.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.1 impl=0.15
item: smaller-go-module   design=0.1 impl=0.15
design-buffer: 0.15
total: 0.53
```

Two extends: (1) pure judge prompt — `PromptInput` orientation fields,
`CodeReviewBody` signature, `code-review.md` header; (2) cmd/sdlc `boundaryOrientation`
derivation + dispatch wiring + dropping the hardcoded `ariadne#` ref. Design
pre-resolved by the durable plan → reduced design + +15% buffer; impl at the v3.1
40%-scaled smaller-go-module top.

Detailed design + TDD breakdown: `workshop/plans/000137-review-repo-orientation-plan.md`.

## Plan

- [x] Locate boundary-review prompt construction for close and milestone-close.
- [x] Define a single repo-orientation data structure shared by boundary prompts.
- [x] Derive repo slug/name from the active git root/remote/cwd context.
- [x] Render explicit issue and repo anchors in all boundary-review prompts.
- [x] Add tests for ariadne and downstream-repo prompt orientation.
- [x] Update documentation or inline prompt comments for the orientation contract.

## Log

### 2026-06-26

- Created from pair#81 retro point 7: boundary review instructions need to
  orient the reviewer to the repo being operated on, especially for downstream
  repos like `pair`.

### 2026-06-29
- 2026-06-29: closed — go test ./cmd/sdlc/... all pass. New TestBoundaryOrientation: temp repo named pair → IssueRef pair#72, base-vs-downstream via construct/base.manifest, asserts it NEVER falls back to ariadne# for a non-ariadne root. CodeReviewBody test asserts every orientation anchor (repo/root/issue-file/boundary/note) renders + no {{ placeholder left. close+milestone dispatch integration tests now assert the derived <repo>#69 ref end-to-end (and reject a hardcoded ariadne#69). Also fixed the #136 sidecar-H1 ariadne hardcode (plan-quality gate shadow-sweep). internal/judge stays pure (orientation derived in cmd/sdlc, passed as strings).; review verdict: unknown

Implemented per `workshop/plans/000137-review-repo-orientation-plan.md`.
- **Pure (internal/judge):** `PromptInput` gains `Repo/RepoRoot/IssueFile/Boundary/
  RepoNote`; `CodeReviewBody(in PromptInput)` substitutes them into `code-review.md`'s
  rewritten orientation header (names the repo under review + a base-vs-downstream
  note). Preserves the `<unknown>` ref fallback.
- **IO (cmd/sdlc):** new `orientation.go` — `boundaryOrientation` derives repo
  name/root + the `<repo>#N[ Mx]` ref + issue file + boundary + note from the live
  git context (base detected via `construct/base.manifest`). Consolidated the
  repo-name derivation into `repoNameAndRoot` (`repoIdentity` routes through it —
  resolves the #136 triplication note). Wired once into the shared
  `boundaryReviewDispatchOptions`; removed the `IssueRef` field from
  `boundaryReviewParams` + its 3 hardcoded `ariadne#N` assignments.
- **Shadow-sweep (plan-quality gate FAILURE → fixed):** the gate caught that
  `reviewsidecar.go`'s sidecar H1 *also* hardcoded `ariadne#%d` (a #136 bug) — the
  same misorientation persisted into the durable artifact. Fixed to render from
  `m.Repo`. The gate also corrected the test-update surface; both folded into the
  revised plan (`## Revisions`).

Verification: `go test ./cmd/sdlc/...` all pass. New `TestBoundaryOrientation`
(temp repo named `pair` → `pair#72`, never `ariadne#`; base vs downstream via
`base.manifest`); `CodeReviewBody` test asserts every anchor renders + no `{{`
left; the close/milestone dispatch integration tests now assert the derived
`<repo>#69` ref (and that it never hardcodes `ariadne#69`). The close of this
issue dispatches a real review in ariadne → the prompt + sidecar should read
`ariadne#137` (the base repo, correctly).

- Boundary review **dogfooded #137 on itself**: the close-review sidecar H1 reads
  `# Boundary Review — ariadne#137` (derived from the live repo, no longer
  hardcoded). Verdict was FIX-THEN-SHIP (parsed `unknown` — prose verdict format,
  as with #143; the full body was captured durably by the #136 sidecar). Real
  finding fixed before the boundary: the **dry-run path** (`close.go:742`) omitted
  `IssueNum`, so `sdlc close --dry-run` would render `<repo>#0 (file: )` — the same
  misorientation, on the preview surface (the dispatch path already had `IssueNum`
  from #136). Added `IssueNum: f.Issue` to the dry-run literal + tightened
  `TestRunCloseWithReview_DryRunPrintsPairAgentCommand` to assert the derived ref
  and reject `#0`.