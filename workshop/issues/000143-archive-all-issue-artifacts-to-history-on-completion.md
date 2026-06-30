---
id: 000143
status: working
deps: []
github_issue:
created: 2026-06-29
updated: 2026-06-29
estimate_hours: 0.59
started: 2026-06-29T17:31:30-07:00
---

# Archive all issue artifacts to history on completion

## Problem

On issue completion, the archive sweep in `sdlc merge` (`archiveDoneIssuesInDir`,
`cmd/sdlc/merge.go:548`) and `sdlc push` (`archiveDoneIssues`,
`cmd/sdlc/push.go:468`) moves only `workshop/issues/NNNNNN-*.md` to
`workshop/history/`. Artifacts that share the issue's 6-digit id prefix — the
durable plan (`workshop/plans/NNNNNN-slug-plan.md`) and boundary-review sidecars
(`workshop/plans/NNNNNN-slug-{close,m<x>}-review.md`, #136) — are left behind in
`workshop/plans/` and accumulate indefinitely.

Observed at the #136 close: the plan + the close-review sidecar stayed in
`workshop/plans/` while the issue moved to history. The atlas already *claims*
plans are "Archived with issue" (`atlas/workflow/artifact-hierarchy.md`,
`atlas/workflow/sdlc-binary.md`) — so today's behavior contradicts the documented
lifecycle.

## Spec

When an issue reaches terminal status and is archived, move **all** workshop
artifacts sharing its 6-digit id prefix to `workshop/history/`, not just the
issue file.

- Extend the archive sweep so that, for each archived issue id `NNNNNN`, it also
  moves `workshop/plans/NNNNNN-*` (the plan + every review sidecar) to
  `workshop/history/`.
- Apply to **both** publish paths — `sdlc merge` (`archiveDoneIssuesInDir`) and
  `sdlc push` (`archiveDoneIssues`). Prefer a shared per-issue artifact-sweep
  helper so the logic lives in one place (`ARCH-DRY`), not duplicated across the
  two callers.
- Archive plans **only** for issues actually being archived (keyed by the terminal
  issue's id) — never for still-open issues.
- Moved plan/sidecar files must be included in the archive commit (the
  `preparedArchiveMove` / `archiveAddArgs` git-add path), so they land tracked in
  history rather than as dangling working-tree deletions.
- An issue with no plan/sidecar archives cleanly — the glob simply matches nothing,
  no error.
- Filenames preserved flat in `workshop/history/`, consistent with current issue
  archival.
- Honor the `WF_PLANS_DIR` override (default `workshop/plans`).

## Done when

- `sdlc merge` and `sdlc push`, on archiving a done issue, also move
  `workshop/plans/NNNNNN-*` (plan + review sidecars) into `workshop/history/`,
  committed in the same archive commit.
- An issue with no plan archives with no error (only the issue file moves).
- Plans for issues NOT being archived are left untouched.
- The atlas "Archived with issue" claim becomes accurate.
- Tests cover: (a) issue with plan + ≥1 sidecar → all moved + committed; (b) issue
  with no plan → only the issue file moves; (c) mixed batch (one done, one open) →
  only the done issue's plan is moved.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.1 impl=0.15
item: smaller-go-module   design=0.15 impl=0.15
design-buffer: 0.15
total: 0.59
```

Two extends of existing Go modules: (1) the forward archive path — a shared
`archivePlanArtifacts` mover wired into push + merge + flags; (2) the push
interrupted-archive recovery parser (`preparedArchiveMoves`) — a fiddlier refactor
(source-tagged halves, deferred terminal gate). Design pre-resolved by the durable
plan → reduced design + +15% buffer; impl at the v3.1 40%-scaled smaller-go-module
top.

Detailed design + TDD breakdown: `workshop/plans/000143-archive-issue-artifacts-plan.md`.

## Plan

- [x] Read `archiveDoneIssues` (push.go:468) + `archiveDoneIssuesInDir`
      (merge.go:548) + `preparedArchiveMove`/`archiveAddArgs` to find the single seam.
- [x] Add a helper that, given an archived issue's id, globs `workshop/plans/NNNNNN-*`
      and returns those as additional `preparedArchiveMove`s.
- [x] Wire it into both the push and merge archive paths (one helper, two callers).
- [x] Ensure moved plan/sidecar paths are included in the archive commit.
- [x] Tests: plan+sidecar moved+committed; no-plan no-op; mixed-batch isolation.
- [x] Verify the atlas "archived with issue" lines are now accurate (adjust wording
      if needed).

## Log

### 2026-06-29

- Created from the #136 close observation: the durable plan + the dogfooded
  close-review sidecar stayed in `workshop/plans/` after #136 archived to
  `workshop/history/`. The merge/push archive sweep (`archiveDoneIssues` /
  `archiveDoneIssuesInDir`) moves only the issue file. Generalize to every
  `NNNNNN-`-prefixed workshop artifact so behavior matches the atlas's documented
  "archived with issue" lifecycle.

- Implemented per `workshop/plans/000143-archive-issue-artifacts-plan.md`. New
  `archivePlanArtifacts` (push.go) globs `plansDir/NNNNNN-*` by the archived
  issue's id prefix and `os.Rename`s each to history, recording moves with the
  caller's path convention (push: cwd-relative; merge: rename under mainPath,
  record mainPath-relative). Wired into `archiveDoneIssues` + `archiveDoneIssuesInDir`
  after the issue move, so plan moves ride the existing `archiveAddArgs` → commit
  (ARCH-DRY). Added `--plans-dir` to both commands.
- Recovery: `preparedArchiveMoves` refactored to source-tag each `half`
  (`srcIsPlan`), defer the terminal-frontmatter check to finalization, and apply
  it only to issue halves — plan artifacts have no frontmatter, so their
  id-prefixed plans-dir source is the membership proof. Prevents a mid-archive
  `sdlc push` crash from flagging moved plans as "unrelated" and stranding
  recovery (Root Cause). A non-terminal issue pair still refuses BOTH halves
  (preserved behavior, pinned by the existing reject test).
- Verification: `go test ./cmd/sdlc/...` all pass. New `archiveartifacts_test.go`:
  `archivePlanArtifacts` moves id-prefixed artifacts + leaves unrelated + asserts
  both halves stage via `archiveAddArgs` (the committed contract); no-plan no-op;
  `archiveDoneIssues` sweeps a done issue's plan+sidecar while leaving an open
  issue's plan; recovery reconstructs an issue move + a plan move with `other`
  empty. Existing archive/recovery tests stay green (the deferred-terminal refactor
  preserves issue behavior). The **merge of this issue dogfoods the feature** —
  `000143-*-plan.md` + `000143-*-close-review.md` should land in history.
