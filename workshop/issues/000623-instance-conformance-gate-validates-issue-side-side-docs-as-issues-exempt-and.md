---
id: 000623
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'a7944929497ceb1b0ade8b2b4ea012eac6cd8f56' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T12:00:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Instance-conformance gate validates issue .side/ side docs as issues; exempt and validate by type

## Problem

## Spec

**The instance-conformance gate validates side docs under an issue's `.side/` folder as issues (parley.nvim:1, 10-25).**
parley.nvim#440's merge (PR #354) was refused: the gate checks every `.md` under `workshop/issues/`, including
`workshop/issues/000440-….side/upstream-neovim-draft.md`, a `type: side-doc` file, and demands an issue's status,
Spec, Plan and Done when. Merge's own archive step already moves an issue's `.side/` folder with it, so side docs
there are an intended layout; the gate is the inconsistency.

Exempt files under `<issue>.side/` from Issue validation, and validate them by their own `type:` (side-doc
frontmatter: type, parent) instead. Use the same notion of "this file belongs to issue N's side folder" that the
archive step uses (one shared helper, not a second path rule).

## Done when

- A repo with an issue `.side/` side doc passes the gate; a malformed side doc (missing type or parent) fails with a
  side-doc message; a malformed issue still fails as before. Tests for all three.
- The archive step and the gate share the side-folder rule.
- **Plan side, same gap:** merge's archive step refuses a `workshop/plans/<issue>-…-plan.side/` folder whose plan document doesn't exist ('active or conflicting archive artifact'), leaving it on main after the issue archives (parley.nvim#440: its upstream-neovim-draft.md stayed in workshop/plans while the archived issue's link points under history/plans). A side folder archives with its issue whether or not a plan document exists; test both. Then archive parley.nvim's leftover file.
- **Duplicate-ID check, same gap (pair#858, 10-25):** `sdlc merge` refused `#000858 would be claimed by 2 files after merge: workshop/issues/000858-….md and workshop/issues/000858-….side`. `cmd/sdlc/issueids.go` (refIDSpace/renderClashes) feeds non-recursive `git ls-tree --name-only` output into `cmd/sdlc/internal/issue/scaffold.go` PathsByID, which parses the `.side` directory name as another issue; LocalPathsByID also admits directory names. ID scanning ignores `<issue>.side/` entries through the same shared side-folder rule as the gate and the archive step; test with an issue plus its `.side/` folder.

## Plan

- [ ] Approve [the shared side-folder design](../plans/000623-side-folder-rule-plan.md), then enter change-code.
- [ ] Implement shared ownership, typed conformance, ID-space exclusions and complete archive/recovery coverage with red-first tests.
- [ ] Verify serial targeted tests and owner-coordinated parley.nvim leftover repair; obtain hub authorization for recorded-check/landing phase.
- [ ] Complete required verification, close review and landing.

## Log

### 2026-10-25

### 2026-10-25 — claimed; design checkpoint

Recovered on clean resting main-slot1 after #560's completed landing; no old runner is resumed. Claimed #623 and started its independent branch from current main. Read the issue and all three consumer families; a read-only audit found existing sidedoc helpers but no side-doc vocabulary model, plus document-dependent folder discovery across archive variants. Full design is warranted rather than a <=100-line quick patch that leaves archive/recovery siblings inconsistent. Existing Folder excludes a directory path itself; preserve that contract and add shared owner projection. Plan is ready for TL approval; no implementation or tests started. Machine budget is targeted only, one process at a time, count1; a full recorded check/merge waits for hub authorization. No new estimate before plan-quality.
