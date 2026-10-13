# #344 `sdlc stage :0`: Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A slot puts its HEAD on `:0` as a detached commit for the operator's smoke test, optionally built, and keeps its branch. It closes and lands from its own slot as usual.

**Architecture:** `sdlc stage [:N]` reuses `sdlc move`'s observation and verification:
- `workspace.Resolve` and `observeMoveSide` read both slots, and a pure `checkStage` decides.
- The facts are re-observed and compared before the one switch, `git switch --detach <HEAD>` in the target. Then `workspace.Resolve` verifies the result.
- A per-worktree record (`git rev-parse --git-path sdlc-staged`) names the staged commit. A detached target is safe to leave when a ref contains its HEAD or it is the recorded commit. That is the rule `stage` (replace), `stage --clear` and `move` (into a staged `:0`) share, so a later normal use returns `:0` without losing an operator's unreferenced commit.
- `--build` runs the target's `make build`, then reports each `bin/*` Go binary's `vcs.revision`, read by `debug/buildinfo` with no exec, against the staged commit.

**Tech Stack:** Go (`cmd/sdlc`), git porcelain via `execGitRunner`, `pkg/workspace`, `debug/buildinfo`.

---

## Scope

| Spec item | Disposition |
|---|---|
| Stage HEAD on `:0` detached; the slot keeps its branch | `sdlc stage [address]` (default `:0`) |
| `--build`; print the staged commit and the binaries' `vcs.revision` | `make build` in the target; a revision mismatch or `vcs.modified` warns |
| Refuse when `:0` is dirty | `checkStage`, the same uncommitted-change and operation rules as move |
| A second stage replaces the first, printing what it replaced | the detached target is leavable (the record or a containing ref) |
| `--clear`, or any later normal use, returns `:0` to rest | `stage --clear`; `move` accepts a leavable detached target and clears the record. Landing never needs it. |
| Retire `move :0` from the smoke procedure | the atlas (workspace-branching) and the move/stage help. The xx-tl skill lives in ops (`~/workspace/ops/skills-draft/xx-tl`), which workers must not edit, so the hub is told what to change. |

The source slot may be dirty: only its HEAD is staged, and uncommitted changes are reported as not staged. Different repositories refuse, as in move. Slots share one object store, so nothing is fetched.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `stageFacts` / `checkStage` / `checkUnstage` | `cmd/sdlc/stageplan.go` | new |
| `moveSide.Leavable` | `cmd/sdlc/moveplan.go` | modified |
| `checkMove` (accepts a leavable detached target) | `cmd/sdlc/moveplan.go` | modified |
| `buildRevision` (BuildInfo → revision, modified) | `cmd/sdlc/stagebuild.go` | new |

- **checkStage**: every refusal of `stage`; nil means the switch may run. Unit-tested in `stageplan_test.go` on literal facts, with no git.
  - **Refuses:** a different repository; the same slot; no source HEAD; a target Git operation or tracked change; a target on a non-resting branch; a detached target that is not leavable; target untracked files the commit would overwrite (`untrackedCollisions`, reused).
  - **Relationships:** 1:1 with one stage run. Its facts embed two `moveSide`s (ARCH-DRY: one slot observation for both verbs).
  - **Future extensions:** a staged commit from another clone would add a fetch step here; refused today.
- **checkUnstage**: the `--clear` rules. On rest, there is nothing to clear. On a branch, it refuses. Detached, it must be leavable and clean.
- **Leavable**: "switching away loses nothing". A ref (`refs/heads`, `refs/remotes`) contains the detached HEAD, or it is the commit sdlc recorded staging. The record covers a slot that amends after staging.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `runStage` / `runUnstage` | `cmd/sdlc/stage.go` | new | `git switch`, `workspace.Resolve` |
| staged record (`readStaged`/`writeStaged`/`removeStaged`) | `cmd/sdlc/stage.go` | new | a file in the target worktree's git dir |
| `stageBuild` | `cmd/sdlc/stagebuild.go` | new | `make build`; `buildinfo.ReadFile` (a var, so tests fake it) |

Lifecycle (ARCH lifecycle):
- **Created:** the staged record is written by `stage`.
- **Last needed:** by the next `stage`, `--clear` or `move` into that slot.
- **Removed:** by `--clear` and by `move`.
- **Size:** one line per worktree, overwritten in place, never growing.

Real-git tests run on temp fleets (`procedureFixture`), with no network and no real build. The build step is covered with a fake `make build` (a Makefile that writes a file into `bin/`) and a faked revision reader.

## Tasks

- [ ] **1. Pure rules.** `stageplan.go`: `stageFacts`, `checkStage` and `checkUnstage`. `moveSide.Leavable`, and `checkMove` accepting a leavable detached target. `stageplan_test.go` gets one row per refusal and per acceptance; `moveplan_test.go` gets a staged target accepted and an unreferenced detached target refused.
- [ ] **2. Observe and switch.**
  - `observeMoveSide` sets `Leavable` (for-each-ref `--contains`, or the record).
  - `stage.go`:
    - observe, `checkStage`, re-observe and compare, `git switch --detach`, verify;
    - write the record, and print the staged commit (and what it replaced);
    - a dirty source warns.
  - `stage --clear` switches the target to its resting branch and removes the record.
  - `move` removes the target's record after a verified move.
  - Real-git tests (`stage_test.go`):
    - stage, and the slot keeps its branch;
    - the slot commits again and re-stages, replacing the first, which is printed;
    - a dirty `:0` refuses and changes nothing;
    - `--clear` returns `:0` to rest;
    - move into a staged `:0` works;
    - an operator's unreferenced detached commit refuses replace and clear;
    - the stage survives the slot amending after staging (the record).
- [ ] **3. Build.** `stagebuild.go`: `make build` in the target (output to stderr; a failure leaves the stage and reports it), then the `bin/*` revisions. A mismatch or `vcs.modified` warns. Tests: a fake Makefile writes `bin/tool`, a faked reader returns the staged commit (matches) or another one (warns), and a failing build reports and keeps the stage.
- [ ] **4. Docs.**
  - `helptext/stage.md`, and a pointer from `helptext/move.md`.
  - atlas `workspace-branching.md`: a "Stage this branch on :0" section; "Move this branch to :N" drops the smoke-test framing.
  - `sdlc-binary.md` verb list.
  - Tell the hub the xx-tl change: the manual detached procedure becomes `sdlc stage`.
