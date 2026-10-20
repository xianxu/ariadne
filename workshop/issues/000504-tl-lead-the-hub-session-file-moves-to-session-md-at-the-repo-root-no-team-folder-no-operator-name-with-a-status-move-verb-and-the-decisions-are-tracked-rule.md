---
id: 000504
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours: 2.142
collaboration-mode: collaboration
card_mirror: '1a07649ec4b92327cbc556fe510652cec158cf69' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T09:56:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# tl-lead: the hub session file moves to session.md at the repo root (no team folder, no operator name), with a status move verb and the decisions-are-tracked rule

## Problem

The TL session file lives at `teams/<team>/tl-status-<operator>.md` (side docs in `tl-status-<operator>.side/`). Since
introduce-team settled on one team per ops repo, the team folder adds nothing (the repo is the team), and the
operator's name adds nothing either (the file is gitignored, so it's per machine and per operator already).
Operator, 10-20 (ops side doc `2026-10-20-03-session-file-everywhere.md`): move it to the repo root.

Also stated there, and worth writing into tl-lead: the session file and its side docs are the **ephemeral TL/human
interface**; durable decisions that workers can act on live in the issue system (tracked, managed by sdlc), so a
decision made in a side doc gets copied into the tracked artifact (issue, project Log, charter) when it's made.

## Spec

- tl-lead's default hub session file becomes **`session.md` at the hub repo's root** (gitignored), side docs in
  **`session.side/`**; `tl-lead --help` and every `tl-lead status …` default follow. The `operator:` frontmatter
  stays (defaulting from git's user.name) for tooling that wants it.
- **`tl-lead status move [OLD]`**: one-shot migration that renames the file and its side docs, rewrites the side docs'
  parent links and the status file's links into them, updates Couch's registration (`couch --session-sync rename`),
  and updates `.gitignore`. Refuses under the operator's lock; dry run by default.
- tl-lead rule: decisions get copied from side docs into tracked artifacts when made (TL vs worker: workers take
  instructions from the issue system, with occasional draft-pane or TL input; the TL multiplexes and keeps the
  temporal picture in the session file).
- Check parley's attention finder still lists side docs under the new parent (adjust in parley.nvim if not).

## Done when

- Tests: `tl-lead status move` on a fixture renames file and side docs, rewrites links, and refuses under a lock.
- `tl-lead --help` documents the new location and the decisions rule.
- **Last step, manual:** migrate the two live hubs, ops:0 (`teams/devinfra/tl-status-xian-xu.md`) and ops-echo:0
  (`teams/echo/tl-status-xian-xu.md`), with the operator at a quiet moment.

## Plan

- [x] Implement shared omitted-FILE defaults while preserving explicit paths.
- [x] Implement pure copy-then-switch planning and fixture link rebasing.
- [x] Implement guarded copying, Couch registration switch, verification and retryable cleanup.
- [x] Update help/README/atlas with session.md, session.side/ and the decisions rule; verify Parley attention compatibility.
- [ ] Run focused checks, recorded suite and the SDLC boundary review.
- [ ] TL and operator migrate both live hubs at a quiet moment; worker does not run live migrations.

Durable plan: [session file move](../plans/000504-session-file-move-plan.md).

## Log

### 2026-10-20

- Claimed in ariadne:4 and entered planning with `sdlc start-plan`. No implementation edits or live hub mutations.
- TL accepted unambiguous legacy discovery, optional FILE defaults, compatibility and lock/collision/symlink refusals. TL directed copy-then-switch without a recovery journal. Sources remain until copies and registration switch are verified; after switch, cleanup failures use the new session and rerun cleanup. No editor lease exists, so apply requires a quiet window and rechecks observed locks (ARCH-ORDER).
- Reuse `SessionSync.Rename` to preserve registry identity, baselines and owed events; reuse sidedoc link/index transformations (ARCH-DRY). Do not mark unread edits seen during migration.
- Read-only delegated Parley check: root-parent side-doc discovery is path agnostic (`lua/parley/attention.lua`, `attention_model.lua`, `document_move_plan.lua`); `tests/integration/attention_finder_spec.lua` covers root discovery and nested documents on subsequent opens. No Parley change indicated. Tests not yet run.
- Complete plan drafted for TL approval before change-code. Manual migration remains in Done when and will block completion until the TL/operator supplies evidence.

- Implementation: shared optional-FILE defaults on all existing status leaves; pure relocation/index transforms; copy/switch/verify/cleanup with identity-preserving Couch rename, source registration prerequisite and no journal. Snapshot IO was split into status_move_snapshot.go to keep the executor focused.
- Verification: `go test ./cmd/tl-lead ./pkg/sidedoc -count=1` passed; real scratch `TL_LEAD_LIVE=1 go test ./cmd/tl-lead -run 'TestSessionSync.*Conformance' -count=1 -v` passed, including complete migration, identity and unread-edit preservation. Static before/after corpus compares the entire tree, and every effect boundary has an interrupted retry. Late-source and changed-destination tests caught and now guard final-cleanup loss/reporting gaps.
- Isolated Parley checkout at its returned HEAD 11b39d8d ran `make -f Makefile.parley test-spec SPEC=ui/attention_finder` successfully, including existing root-parent and nested side-doc discovery. No Parley changes. Temporary checkout and harness state removed after the run.
- TL requested stage only after SHIP and green suite, then TL/operator live migration, evidence, clear and merge. This conflicts with current manual pre-close Done when; clarification requested to move live acceptance to a post-close/pre-merge hold. Couch reported recipient unavailable on both clarification sends; no delivery receipt was created. No live files touched and no stage performed.

## Revisions

### 2026-10-20 — planning decisions from the TL

- Added the executable plan and explicit defaults/compatibility semantics without changing the existing Spec or Done when. Chose copy-then-switch instead of a recovery journal at the TL's direction. Live migrations remain the last manual acceptance step, owned by the TL and operator.


## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is marked stale by estimate-source, so these rates remain provisional.

Decomposition: issue/spec checkpoint, shared status-argument defaults, pure relocation planner, filesystem/Couch copy-switch integration, operator docs, and one close review. Thorough approved design applies the v2.1 ×0.2 design discount and +15% buffer. Existing Go/Cobra/SessionSync/sidedoc libraries cover the stack; the planner's base design is halved for reuse before the spec discount. Familiarity is 1.0 (existing Go CLI patterns); v3.1 implementation rates are 40% of v2 ranges. The integration uses the top of the API primitive range for interruption/reconciliation fixtures. No new Lua feature or peer implementation is budgeted. Derived estimate is approximately 2.14 focused ship-hours, including verification and the manual acceptance coordination; elapsed waiting is not focused time.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.1 impl=0.08
item: smaller-go-module design=0.04 impl=0.16
item: greenfield-go-module design=0.1 impl=0.28
item: api-integration design=0.4 impl=0.6
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.02 impl=0.16
design-buffer: 0.15
total: 2.142
```

Estimate advisory mapping: the integration primitive includes executor guards, stateful fake, interruption/reconciliation tests and scratch conformance. The docs primitive includes Parley evidence and operator migration notes; the review primitive includes recorded-suite launch/results, staging checks, close review and active migration-evidence coordination. Rates are idle-removed ship wall-clock, including AI/subagent execution spans; no parallelism discount. These allocations remain uncertain at the stale calibration, rather than excluding verification from measured work.
