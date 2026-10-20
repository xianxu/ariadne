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
- Pre-stage verification: focused tests, scratch Couch conformance, Parley attention compatibility, recorded green suite and SHIP from the boundary review. Close review may overlap a queued suite under tl-worker; staging and landing require its pass.
- **Post-close, pre-merge hold:** after green suite and SHIP, stage the final build with `sdlc stage --build`.
  The TL and operator migrate ops:0 (`teams/devinfra/tl-status-xian-xu.md`) and ops-echo:0
  (`teams/echo/tl-status-xian-xu.md`) at a quiet moment. The worker waits for TL evidence, appends it to the Log,
  clears the stage, and only then merges. Closing does not release this landing hold.

## Plan

- [x] Implement shared omitted-FILE defaults while preserving explicit paths.
- [x] Implement pure copy-then-switch planning and fixture link rebasing.
- [x] Implement guarded copying, Couch registration switch, verification and retryable cleanup.
- [x] Update help/README/atlas with session.md, session.side/ and the decisions rule; verify Parley attention compatibility.
- [x] Launch the recorded full suite on the pushed head and retain the pre-stage hold until it passes; the close verb owns the boundary review.
The post-close migration is tracked in Done when and Landing hold below, not as a pre-close Plan checkbox.

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

- Drain checkpoint: implementation committed at `bb7847138bd2e5cc6c8beccdeceb890e885354cb`, published as [PR #351](https://github.com/xianxu/ariadne/pull/351). Final combined `TL_LEAD_LIVE=1 go test ./cmd/tl-lead ./pkg/sidedoc -count=1` passed (tl-lead 25.859s; sidedoc 0.131s). No source tests are running.
- Recorded suite was launched on that pushed head, but remained queued behind the machine-wide pause for Couch drain `69bb7fc3-a125-42c6-9900-5814613b8eba` (operator shutdown/reboot/RAM-disk fix). Canceled this slot's queued suite with `sdlc local-check cancel --name suite`; no full-suite pass exists. Never treat the canceled attempt as validation.
- Stop state for restart: issue remains working; no close review, SHIP, stage, live migration or merge. Resume with `sdlc state`, rerun the recorded suite when the fleet resumes, and obtain TL approval of the Done-when ordering revision before closing. TL wants green suite + SHIP, then `sdlc stage --build`, TL/operator migration of ops:0 and ops-echo:0, evidence appended here, stage cleared and merge. Existing Done when still requires manual migration before close because the clarification has not been delivered/approved. Ordinary Couch sends currently refuse while draining; no approval inferred from silence.

- Drain acknowledgement attempted only after the checkpoint reached PR #351. Couch refused `--drain --ack 69bb7fc3-a125-42c6-9900-5814613b8eba` with `ack does not name current drain/member`; the board still lists this slot as checkpoint requested. No ack was recorded. Work is stopped with a clean checkpoint; operator/supervisor must reconcile membership or deliver the current drain notice. No automatic retry or separate drained report.

- Resumed after restart. TL approved revising live migration to a post-close/pre-merge hold (delivery 530ea8ef-34e8-4475-874c-6f495ed9e4d3, canonical submitted receipt verified). Recorded suite and SHIP precede staging; no live migration will be run by the worker.

- Post-restart verification refinement: a targeted regression reproduced quoted/tagged YAML session types bypassing the nested-session registration prerequisite. Consolidated metadata classification through the typed YAML parser for root and nested documents; plain, single/double-quoted and explicit string-tag forms now all refuse before effects when unregistered. `go test ./cmd/tl-lead ./pkg/sidedoc -count=1` passed after the fix. Canceled the pre-fix recorded suite on 0f1b7d37 (still awaiting admission; no pass), and will record a fresh run on the corrected pushed head.

- Added the owner-scoped `session-move-raw-fields` Go lint catalog rule (#504) so the snapshot reader cannot return to the raw YAML reader; alias/shadowing controls avoid affecting valid consumers elsewhere. `go test ./cmd/tl-lead ./pkg/sidedoc ./pkg/golint -count=1` passed (10.672s, 0.094s, 0.838s). The lint atlas catalog is synchronized.

- Final suite launched on adec0f4c6c0e (PR #351); machine gate reports FIFO position 2 behind ariadne:3 and parley.nvim:2. No pass yet. Per tl-worker explicit review/suite overlap permission, start close review while the suite queues. Pre-stage verification wording and Plan now state launch plus hold, not a false suite-pass claim. Green suite and SHIP remain required before staging, and all live migration evidence remains required before merge. Stage dry-run only; no build staged.

## Landing hold

Pending: green suite and SHIP → final build staged → TL/operator migrate both live hubs → TL evidence appended here → stage cleared → merge. Worker must not merge before the TL returns that evidence.

## Revisions

### 2026-10-20 10:40 PDT — TL-approved acceptance ordering

Reason: keep the stage window short and stage only the final SHIP build.
Superseded Done when: “Last step, manual: migrate the two live hubs … with the operator at a quiet moment,” previously interpreted as required before close. The manual migration remains required before merge, but is now explicitly a post-close hold. Current Done when and the Plan were synchronized: the pre-close Plan checkbox covers the recorded suite; close owns its review, and manual migration lives under Landing hold.


### 2026-10-20 — planning decisions from the TL

- Added the executable plan and explicit defaults/compatibility semantics without changing the existing Spec or Done when. Chose copy-then-switch instead of a recovery journal at the TL's direction. Live migrations remain the last manual acceptance step, owned by the TL and operator.


### 2026-10-20 — verification scheduling clarification

`tl-worker --help` explicitly permits close review to overlap a queued recorded suite when the final tested code must pass before landing and review fixes requeue the suite. Use that permission for this queued run; the approved green-suite-and-SHIP requirement still precedes staging. No acceptance was waived and no pending test result is claimed as a pass.

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
