---
id: 000504
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours: 2.142
collaboration-mode: collaboration
card_mirror: '65234802c1ac89d997a640594b15b146b04e079b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T09:56:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 1.56
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

- Migration input envelope: only plain `parent: ../file.md` forms emitted by tl-lead are accepted; other parent forms refuse before effects with the file named. Append missing ignore lines, then require real Git ignore checks for every destination before copying. Preserve ordinary document business-owner metadata. Refuse side-tree `.gitignore` files at every depth; recheck resulting effective Git policy before switching and each deletion.
- Live-recovery boundary: copies are verified before the root registration switches. The new tree may then be edited and locked; cleanup checks unchanged old-source snapshots and confirms prior rename on retry. Keep old leftovers quiet between runs. Migration Couch calls are bounded to five seconds; list observations and Git checks are batched, with distinct registration conflicts named and retained.

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
- 2026-10-20: closed — Affected tl-lead, sidedoc and golint packages pass; real scratch Couch conformance passes. Reference-target matrix checks two moving source coordinates, six target classes and five consumers; private overlay removal reproduces root-link and empty-directory failures. Directory collision refuses before effects. Prior nested-policy, parent, effective-ignore and transition-outcome regressions pass. Fresh recorded suite on 339ca087 queued/running, NOT yet passed; tl-worker allows review overlap. Green suite and SHIP precede staging; TL/operator live migration evidence and stage clear precede merge.; review verdict: SHIP

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

- Boundary review round 1: REWORK, BR-1 (Critical) accepted parent syntax exceeded link transform support; BR-2 (Important) later negations defeated pattern-presence ignore checks. Targeted tests reproduced both classes, plus business-owner metadata rejection. TL approved plain-parent-only refusal (delivery 8587784a-497c-46c5-b96a-781a4dc99c0f) and simple Git check-ignore verification, explicitly declining structural YAML preservation or owned ignore-block management. Added pre-effect parent validation, typed session-only routing checks, and effective destination ignore checks before copies. No new journal or state machine mechanism. The queued pre-fix suite was canceled before admission; requeue the corrected code.

- Review fixes verified: `go test ./cmd/tl-lead ./pkg/sidedoc ./pkg/golint -count=1` passed (15.158s, 0.138s, 0.803s); scratch `TL_LEAD_LIVE=1 go test ./cmd/tl-lead -run 'TestSessionSync.*Conformance' -count=1 -v` passed (2.809s). Parent refusal tests failed on every formerly accepted unsupported form before the guard; business-owner preservation failed before routing was scoped; real Git-negation tests failed before effective-ignore checks.

- Boundary round 2: REWORK, BR-1/BR-2 verified addressed. BR-3 found copied nested ignore policy could expose destinations; BR-4 found the phase tracker did not grant production effects. Added all-depth source/retry policy refusals and fresh pre-switch/per-delete Git checks. Replaced the tracker with a pure per-effect grant model consumed by production, with confirmed copy/binding counts, verification, failure and uncertain-outcome stops. Eight nested-policy cases failed before guards; outcome mutation made TestSessionMoveTransition fail and restoring it passed. Affected packages passed (tl-lead 29.721s, sidedoc 0.140s, golint 0.752s); scratch Couch conformance passed (3.052s). Canceled the old queued suite and will requeue on the pushed correction; full suite and live migrations remain pending.

- Boundary round 3: REWORK, BR-3/BR-4 verified addressed; BR-5 reproduced deleted side-root link targets. Added exact side-root mapping and explicit source-directory observations/copies, preserving empty targets and refusing directory/file collisions before effects. Production matrix tests cover two source coordinates, six target classes and all five reference consumers; a private overlay removing mapping/directory planning reproduced broken root links and missing empty targets. Fixed matrix passed (37.059s), collision refusal passed (0.240s); affected packages passed (88.891s, 0.149s, 0.754s) and scratch Couch conformance passed (4.936s). Previous queued suite canceled; fresh recorded suite required on this correction. No live hubs touched.

- Final close round 4 returned SHIP with all five findings addressed; codecomplete bound to 746ba0177d2e (card mirror 855aca23e), measured actual 1.56h. Recorded full suite on 339ca08719fb439537a062638e467c4921952c50 passed: packages plus all eight shards, 1,865 cmd/sdlc tests and 75 other packages, zero failures, wall 694.3s; log sha256 prefix 5f1d07e30a28. `sdlc stage --build` staged 855aca23ec953c6e5e5ba402099b6936c2bb887a on ariadne:0; all seven binaries report that vcs.revision and vcs.modified=false. Readiness message to hub was confirmed submitted. Waiting for TL/operator live migration evidence for ops:0 and ops-echo:0; worker has touched no live hub files. Stage stays held pending that evidence; merge remains held. This checkpoint changes evidence only, not staged code.

- 2026-10-20 live evidence from TL (receipt 5aa41abe-e9eb-4838-9834-8eef6ca43eae verified submitted): ops:0 switch succeeded; session.md is live, registration identity and owed events preserved, operator submits from the new file. Apply inside the agent sandbox could not reach Couch and remained silent for over five minutes. An outside-sandbox migration of 196 docs took about twenty minutes with per-document lookups. Cleanup then refused an edited destination, although editing the switched file is expected. TL reports leftovers teams/devinfra/tl-status-xian-xu.md plus nine nested .side entries and an old-path registration. ops-echo migration waits. Reopened to working; merge remains held. No worker inspection or mutation of live hubs has been performed.

- Approved live-recovery implementation: migration-only five-second Couch command budget plus 250ms output draining; typed list snapshots replace per-document lookup/status; complete NUL-separated Git membership batches replace whole-tree checks per file. Copies are verified once before switching; confirmed root rename authorizes source-only cleanup. Cleanup retry confirms Couch's previous-path rename, skips live destination byte/lock scans, and retains distinct binding conflicts. New live-edit/lock and retry fixtures failed before correction; counted 8/16-doc fixtures reject per-document probes and bound Git processes to N+12. Real scratch conformance now interrupts after root rename, edits live text and retries source cleanup, preserving identity and unread edits (all three conformance tests passed, 1.367s). New strict list path tests reproduced malformed canonical-path acceptance and pass after validation. Scoped session-move-batched-reads lint prevents the mechanical API regression. Former cleanup-destination refusal assertion is superseded by the approved live ownership contract.
- Estimate-quality advisory: original 2.142h remains the historical whole-issue baseline, not a forecast of remaining recovery work. Approved live recovery replaces the integration/verification assumptions; actual will be remeasured on re-close, never copied from the former 1.56h close.

- Recovery verification: affected packages passed (tl-lead 17.451s, sidedoc 0.073s, golint 0.578s). Cleanup dry-run regression failed because output still promised copying into the live tree; rendering now lists unchanged-old removals only and allows the new tree to stay open. Its targeted rerun passed. The refreshed full suite, fresh close review and both TL/operator checks are still pending; former green-suite/SHIP evidence does not cover this code.

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

### 2026-10-20 10:54 PDT — boundary corrections, TL-approved narrow input envelope

BR-1: accept only the literal plain parent form produced by tl-lead. Quoted, commented, tagged, aliased, block and quoted-key forms refuse before effects with their source file. This supersedes the initial implicit broader-parent acceptance; no structural YAML preservation is added. Ordinary non-session-sync business `owner`/`operator` metadata remains untouched.

BR-2: retain append-only-if-missing ignore lines, then require real `git check-ignore` success on every destination before copying. Existing negations are not rewritten or hidden; refusal keeps sources and copies untouched, though missing ignore additions can remain. The TL approved this policy. Current Spec and runtime help now state both boundaries.

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


### 2026-10-20 11:10 PDT — BR-3/BR-4 publication and effect authority

Every published destination must remain ignored under the resulting Git policy.
Refuse any `.gitignore` in the source side tree or existing destination tree,
at every depth, before effects; these files can change the policy after a
pre-copy check. Freshly rescan destinations and run real Git checks before
switching and before each deletion, catching late policy files and exposed
paths. No prospective Git policy simulator or managed ignore block is added.

The earlier phase-only tracker is superseded by invocation-local
`sessionMoveState`: counts of confirmed copies and bindings, whole-move
verification, and cleanup progress. `transitionMove` consumes the observed
success, failure or unconfirmed result of its authorized effect and grants
exactly one next effect. Production dispatches only that grant. Failure or
uncertainty stops authority; retry reconstructs from actual disk/Couch state,
without a journal. Pure sequence tests independently assert all copies precede
switches, all binding checks precede verification, verification precedes every
deletion, and interruptions at every boundary grant no later effects. Existing
production interruption fixtures exercise the same dispatch loop.

The green-suite, SHIP, staging and operator/TL live-migration landing holds remain.


### 2026-10-20 — BR-5 preserve reference targets including directories

The relocation map covers the old session, side-tree root, every descendant
file or directory, and leaves stationary targets in place. Link rebasing uses
that same map for inline links, images, reference definitions, plain parents
and discussions paths. Source directory observations are now explicit: the
plan retains empty directories too, checks file/directory destination
collisions before effects, and recreates directory targets before switching.
Real Git checks include destination directories. This supersedes the previous
implicit assumption that copying files recreated every referenced directory.
Tests enumerate target kinds and consumers, including populated and empty
side roots and descendant directories. The full-suite and manual live-migration
holds remain unchanged.


### 2026-10-20 — revised live-recovery plan (awaiting TL approval)

The live check invalidates the former assumption that destination bytes remain
unchanged through cleanup. The switch is the ownership boundary: the new tree
may be edited and locked once live. No editor lease or journal is added.

The simplest alternative is operator-only manual deletion. It does not fulfill
the verb's promised retryable cleanup and would leave the same scale and
reachability defects for ops-echo, so retain copy-then-switch with these corrected
boundaries (ARCH-PURPOSE, ARCH-ORDER, ARCH-CONSTRAINTS):

| Entity | Lives in | Change |
| --- | --- | --- |
| registration snapshot decoder | cmd/tl-lead/status_move_registry.go | typed complete persisted list, keyed by canonical path, no per-doc lookup |
| session move plan/state | cmd/tl-lead/status_move.go | pre-switch copying vs confirmed post-switch cleanup; source snapshot is cleanup authority |
| migration Couch runner | cmd/tl-lead/sessionsync.go | bounded single subprocess, no retry; deadline distinguishes read refusal from unconfirmed mutation |
| snapshot/executor | status_move_snapshot.go, status_move_apply.go | batch registration observations at phase boundaries; source-only cleanup |

- [ ] Add failing regression fixtures for denied/hung Couch before effects,
  counting calls at N=8 and 2N=16, and edits/locks on the new session and side
  docs immediately after switch and between cleanup steps. Include interrupted
  post-switch retry with only old leftovers and a changed live destination.
- [ ] Migration starts with one typed `couch --session-sync list --json` snapshot.
  Missing/bad replies are errors, never absence. Preserve ID and registered
  fallback owner; local session routing remains read from the document. Use
  list snapshots at fixed phase boundaries, not lookup/status per document.
  Check all relevant old/new `.lock`/`.clobbered` paths locally; no active
  editor lease is claimed. Perform necessary Rename mutations only for actual
  registered pairs, using Couch's idempotent previous-path check on retry.
- [ ] Bound each migration Couch invocation to five seconds plus bounded pipe
  draining. A preflight denial/timeout reports cannot reach Couch before local
  writes; a mutation timeout says outcome unconfirmed and keeps old copies.
  Do not try to bypass a sandbox or retry silently. Tests use an isolated
  helper subprocess; no live sockets or hub files.
- [ ] Verify copies before switching. Once root registration switches last,
  destination edits are normal: do not compare live bytes, restore metadata,
  write destinations, or reject live locks during cleanup. Compare each old
  source against the captured source bytes immediately before its deletion.
  Source edits/locks/late files still preserve old files. On rerun, confirm
  the prior root rename with Couch before entering cleanup-only behavior;
  an unrelated destination cannot authorize deletion. Never forget or replace
  a distinct occupied registration automatically; report the conflicting IDs.
  Across separate invocations the no-journal design cannot prove that an old
  leftover was never edited; the quiet-window rule applies to old leftovers.
- [ ] Sweep external-operation growth: batch Git ignore checks with complete
  output membership validation at phase boundaries, and one destination check
  per deletion; avoid rescanning/checking all N files for each of N steps.
  Test counts bound lookups/status to zero, list calls to a fixed phase count,
  renames to registered document count, and Git checks to O(N).
- [ ] Update help/atlas and the current input/maintenance contract. Record
  regression failures, package and scratch real-Couch conformance passes.
  Push, queue a new recorded suite, re-close for fresh SHIP, then restage only
  when green. TL/operator retry ops:0 cleanup and migrate ops-echo, returning
  evidence before clear-stage and merge.

| Failure/state | Visible state | Recovery | Possible loss |
| --- | --- | --- | --- |
| Preflight cannot reach Couch | old still in use, no local effects | rerun outside blocked sandbox | none |
| Partial copies or unconfirmed rename | keep copies; switch unknown | rerun identical command to reconcile with Couch | none known; no cleanup authority |
| Root switch confirmed, live destination edited | new in use, old cleanup pending | remove unchanged old copies; preserve live tree | none in live tree |
| Source changed/locked during cleanup or unrelated registration | new in use; named old copy retained | resolve old source/conflict, rerun in old-tree quiet window | no automatic deletion of observed source change |

The prior same-bytes destination requirement applies only before switching.
The live-acceptance hold remains operator/TL-owned. This is a replacement plan
for the same mechanism, not an incremental addition of recovery machinery.


### 2026-10-20 — revised live-recovery plan approved

TL approved the replacement plan (delivery 04e0afed-a17c-448d-8436-aa3ffc9324b2), including the quiet-old-tree limit. Keep the fix bounded: no journal, no automatic disposal of a registration conflict. After re-close and green suite, restage; the TL/operator checks are ops:0 leftover cleanup and ops-echo full migration.
