---
id: 000624
status: codecomplete
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '29b6575827d40782b82a67ce143ac4eab604a045' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T03:07:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 2.05
---

# fleet inventory --json reports merge intents (producer for pair#848)

## Problem

## Spec

**`sdlc fleet inventory --json` reports merge intents (producer for pair#848).** After a restart, Couch's recovery plan
must re-enter each landing a restart interrupted, once. sdlc already owns merge-intent records (validated v1 files,
lock-based liveness, exact re-entry command construction), but `fleet inventory` doesn't expose them, so Couch would
have to read sdlc's state files itself. pair#848's plan (pair `workshop/plans/000848-orphaned-merge-recovery-plan.md`,
TL-approved 10-25) makes inventory the single source.

Add per-row `merge_intents[]`: for each intent, its state (live, orphaned, terminal, stale), the slot and exact member
root it belongs to (dependencies included), its physical identity (for de-duplication across worktrees sharing one
Git common dir; group shared stores once), the exact re-entry command, and a carried `note` (e.g. queue paused). Each
row also states read quality (complete, partial, unreadable), and a corrupt record without a trustworthy root appears
once as a repository-level inspection diagnostic. Additive JSON fields only; an older Couch ignores them.

Transition: no intent format change; the new reader reads previous-build v1 files through the existing preserving
codec, pinned by legacy fixtures. A capability field lets Couch tell "no merge intents" from "this sdlc can't report
them" (it must then say upgrade and rerun, never claim none).

## Done when

- Inventory JSON carries merge intents with state, root, identity, re-entry command, note and read quality; tests
  over a stateful fake for live, orphaned, terminal, paused, unreadable, shared-store duplicates and corrupt-root
  records, plus legacy v1 fixtures.
- One inventory call per fleet; bounded output; the capability field is present and documented in `--help`.

## Plan

- [x] Implement and verify the shared report, grouped read-only inventory adapter and additive wire contract in [the durable plan](../plans/000624-fleet-merge-intent-inventory-plan.md), after hub approval and change-code.
- [ ] Qualify private real-store/executable conformance, recorded pre-close check and close review; land and hand the final producer contract to pair:5.

## Log

### 2026-10-25
- 2026-10-25: closed — Post-SHIP fix removes the context.Background compatibility wrapper flagged by TestVerbContextsReachTrackerReads and threads existing caller contexts through shared state/landing-status observation without changing legacy signal owners. Focused race including context guard, native SIGINT/SIGTERM production phases, private real-lock/store conformance and existing status tests passes 16.843s; catalog passes 1.057s, diff-check clean. Fresh recorded pre-close retained on pushed head 9229a4df3ed0f94b78810614b5b37cd6858dbc51 must pass before landing; prior incomplete suite is not reused. Separate scratch alarm filed unclaimed #631, train-barrier failures owned by #622/#625, per hub. Plan bypass acknowledges only outstanding suite/landing process checkboxes. (fixed-then-ship); review verdict: SHIP
- 2026-10-25: closed — Round-two BR-1 through BR-4 verified addressed. BR-5 native public-executable SIGINT/SIGTERM controls reproduce surviving Git and pipe children before the fix in vantage, collection and intents; all six pass after inventory uses the existing CLI signal owner and bounded Git throughout. Broad fleet/intent/README race plus unchanged CLI lock-unwind and net signal controls pass 87.873s, catalog 1.939s. Private real-lock/store conformance, frozen v1, strict/additive algebra, 8/16 growth and 1 MiB boundary controls pass. Recorded pre-close retained on pushed head 5724aaba9338a40537db4841a9e11e762fc73bd8 must pass before landing. Plan bypass acknowledges remaining suite/close/landing process checkboxes only. Consumer #848 remains plan-only until producer lands, per pair:5. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-25: flow upgraded quick → full — 627 added lines in code files (limit 100); an earlier round of this close already ran the full review

Claimed and entered start-plan from main after #617 landed (Git merge954c8b922af48d0fcc1a5bc148ccefb57be73c2d). No retained check or merge runner remains. Read the existing fleet collection, common-directory identity, slot/dependency rows, state/landing shared intent observer and preserving lock-based v1 store. Read pair#848's approved plan through its issue branch without changing that peer checkout; pair:5 is working elsewhere. Its consumer remains parked until this producer lands.

Pair:5 authenticated receipt bba6d841-4e31-400d-a24a-71d4b4648b3e confirms capabilities.merge_intents=1; row quality complete/partial/unreadable with error; exact canonical record identity and Root-only association; complete+[] after terminal removal, not new tombstones; malformed/legacy/mixed-store controls. The existing observer silently skips non-ENOENT open errors and drops earlier observations on a later branch error: the plan closes that uncertainty class through the same reader, rather than a second classifier (ARCH-DRY/ARCH-SECURE). Group by the inventory's existing physical common-dir identity, not a publication URL or branch name.

Durable plan drafted and checkpointed for ops:0 review before change-code. It includes the visible-state table, one shared classifier/command owner, one store read per group, bounded1MiB JSON matching the consumer, counted8/16 growth controls, preserving v1 fixtures and private real-lock conformance. Proposed rollout is private executable proof plus the hub's normal post-land primary rebuild; no live staging, pause, sweep or merge-recovery execution. No implementation or test edits and no new runtime state/schema.

### 2026-10-25 09:01 PDT — Parked pending #631 and primary rebuild

Narrow re-close returned SHIP with zero findings and codecomplete evidence c4e094bb7df4daa1e72efd7c1a9d6eccd05986ae. CI is green; no source changes after the tested context fix. Fresh pre-close attempt3ac11e1917634f6e2a09509bfe116141 on9229a4df3ed0f94b78810614b5b37cd6858dbc51 acquired06:18:43 PDT after2473s queueing and ended06:48:51. All eight shards reached1800.4s; 537 listed tests never ran, while the other79 packages passed. Saved failed log SHA25622e2c601a592355514b27b52b7a10eb18ecf79f176083c69dc66f9fd6c83f3ea, common-dir sdlc/test-stats/failed/3ac11e1917634f6e2a09509bfe116141.log. Samples span shared scratch RemoveAll, Git Wait, transcript WalkDir and review fixtures. Both known #622/#625 train barriers also recurred. No pass/resume or new flake registration is claimed.

At09:00 current load was27.16/27.38/35.33; this is later host context, not a measurement during the failed run. The runner survived the Couch restart and completed; no owned suite/review/merge runner remains.

Authenticated ops:0 receipt2c7e41d6-1e44-4f29-b920-218aea8deda5 directs a hold: #631's owner measured cleanup repeatedly retrying roughly21 undeletable read-only module caches (77MB each), from roughly270 calls per shard. This is the hub's diagnosis, not independently reproduced here. Park #624/PR440 until #631 lands AND ops:0 rebuilds ariadne:0, then run one fresh pre-close on this retained candidate and proceed through central merge. No retry meanwhile. Do not resume either incomplete attempt.

## Revisions

### 2026-10-25 — Clarify terminal observation and freeze consumer vocabulary

The existing merge-intent lifecycle removes terminal records; inventory reports a completed empty read rather than inventing a terminal state or tombstone. Pair:5 agreed. Keep schema_version1 and unchanged v1 intents, with an explicit capability distinguishing an old unsupported producer from complete empty evidence. The new row-quality vocabulary describes enumeration completeness independently of each record's live/orphaned/stale/unreadable state.

### 2026-10-25 — Hub approved implementation

Authenticated ops:0 receipt565af516-d8bc-4ef8-b8c8-46042ce3b16e approves the durable plan and explicitly includes the existing reader's lost-partial-result bug. Its non-ENOENT failure and partial-observation controls are required tests. Private executable/store qualification plus the hub's post-land primary rebuild is approved; no live staging. Proceed to change-code before implementation.

### 2026-10-25 — Shared reader and wire implementation checkpoint

change-code exited0, inferred quick within the design shell; no flow pin, estimate or bypass. Implementation is in progress, not codecomplete. Added additive capability/row-quality/report types and the grouped command adapter, preserving the existing report/classifier through a type alias. Pure placement is indexed by common-dir/root and deduplicates physical identities; output is encoded before the consumer1MiB bound, with zero output on overflow or pre-publication cancellation.

RED/GREEN evidence: initial new wire cases failed unknown capabilities, then legacy/quality/type controls passed. TestFleetMergeIntentsReaderRetainsEarlierObservations reproduced losing the earlier observation when the second branch probe failed; fixed in the shared observer, with failed probes inspection-only and no command. TestIntentOpenFailureIsNotAbsence reproduced non-ENOENT disappearance; fixed. TestFleetMergeIntentsCanonicalPhysicalIdentity reproduced a logical-path identity through a symlinked store; shared observation now canonicalizes the physical record. TestIntentNonregularEntriesAreInspectionEvidence reproduced silently dropping symlink/directory entries; they now surface without intentional traversal. TestMergeIntentCapabilityRejectsNullAndToleratesFutureMetadata reproduced accepting null support; fixed.

Focused shared reader, lifecycle and new command cases passed30.039s (go test ./cmd/sdlc with TestFleetMergeIntents/TestMergeIntent). Stateful public-command controls cover live → orphaned, pause note, terminal removal, stale branch, mixed valid/corrupt and isolated unreadable stores. Pure placement tests cover physical duplicates/conflicts/unmatched roots and8/16 counted row/record lookups. Current internal fleet/landingline wire, intent and existing JSON controls passed0.178s/0.435s. No targeted runner remains, no full suite or live store/queue mutation has run. Still required: complete adversarial wire/legacy fixtures, explicit store-read/branch-probe growth controls, private real executable conformance and consumer fixture, docs, race/catalog, recorded pre-close and mandatory close/merge.

### 2026-10-25 — Implementation qualification before recorded verification

TestFleetMergeIntentsPrivateExecutableConformance passed through a separate built SDLC process against a private linked-worktree store, declared independent dependency clone and private WF_SDLC_QUEUE_ROOT. It observes real owner flock live, interruption orphan, literal pause note, mixed valid/corrupt partial, terminal removal complete empty and deleted-branch stale; every inventory read leaves intent/queue bytes and directory entries unchanged. TestFleetMergeIntentsLegacyV1Fixture pins the optional-workspace/unknown-nested v1 contract without claiming a live capture. Consumer #848 remains parked for this producer, per pair:5; final new-Couch decoder qualification belongs after this landing, with the retained reproducible fixture.

Growth controls at8/16 assert one store read despite more linked rows, one local branch probe per intent (N+1 total Git calls), indexed root lookup counts and linear partial-payload size. The size regression first failed9820→28284 bytes; details are now emitted once in diagnostics with a short row pointer. Missing observations and repository-discovery failures are explicitly unavailable; root mismatch, unreadable-with-values, null support, malformed types and duplicate keys refuse. A rejected symlink's diagnostic stays on the entry, not its target. Existing repository/path-first diagnostic ordering is shared through the extracted fleet.SortDiagnostics helper; the old end-to-end fixture now expects the additional unavailable-merge-store diagnostic for its malformed repository.

Broad fleet/intent race run passed33.470s, internal wire/intent races1.231s/1.280s; catalog passed0.983s before the last tiny sorting/hint changes and will rerun on the final tracked sources. Both existing status callers retain partial results. A new RED control caught the old unreadable hint advising removal after a Git probe failure; changing it to inspect the named source and rerun passed4.820s with existing status controls. No targeted runner remains. Next: final catalog, checkpoint/push, recorded pre-close suite and close review, then central merge; no full-suite pass or SHIP is claimed yet.

Final tracked-source catalog passed0.770s and diff-check is clean. Additional pure controls for multiple branches on one root, ambiguous roots and held corruption passed0.422s. Pair:5 has the final contract and retained fixture paths, with an offer of a private capture if a decoder already exists; no stacking or consumer implementation is requested. The consumer implementation was explicitly parked in its authenticated reply, so pending new-consumer qualification is stated, not claimed. All implementation changes are checkpointed before PR and the queued recorded check. No targeted/background runner remains at this checkpoint.

Pair:5 authenticated receipt b0834d96-7d4f-4709-8ad1-67bb48508d4e confirms no new consumer decoder exists yet and no private capture is needed before landing. It accepts the retained real-executable/store and literal v1 fixtures as producer proof; #848 integrates and qualifies the final wire after this prerequisite lands. PR #440 is open. Ops:0 retired measurement-refusal waivers after #603; close will omit actual flags and record the new confidence outcome normally. No retained check or review is active before queueing.

Round 1 returned REWORK (BR-1 through BR-4). The first close preparation correctly refused a changed #616 adoption before review; a fresh preparation dispatched the review. Cancelled only the still-queued owned pre-close lease-273107732 after the verdict, with no suite started. TestFleetMergeIntentsUsesSharedStoreDespiteBrokenMember and TestFleetMergeIntentsCancelsActiveGitProbe reproduced the two production defects RED, then passed with stateful/growth controls0.779s: observe via canonical common dir, reuse bounded cancellable Git and stop later probes. FuzzJSONContractAlgebra reproduced both existing seed failures; updating its oracle to additive inventory/rows while retaining strict policies/duplicate rejection is under verification. README coverage and the class-level lessons are updated. No SHIP or full-suite pass claimed.

Round-one fixes qualified: full internal fleet/landingline packages with race passed16.169s/5.163s (including the formerly failing algebra seeds); broad command fleet/intent race passed37.091s including the new production-path controls and private executable conformance. Tracked-source catalog passed1.085s. No test remains running; checkpointing all fixes, review sidecars and lessons before pushing and requeueing final-head verification.

### 2026-10-25 — Re-plan: native inventory signal ownership (approval pending)

Round two independently reproduced BR-1/2/3 fixes and cleared BR-4, but native SIGTERM exits fleet inventory before its isolated Git child is reaped (BR-5, same cancellation family). This is the third observer boundary adjustment, so stop and request the consolidated boundary rather than another incremental patch. The simplest alternative is retain default signal exit and let Git inherit the terminal group; it fails because owned subprocesses can survive CLI termination and cancellation/deadline semantics remain unproved.

Proposed single owner: add only fleet inventory to executeCLI's existing SIGINT/SIGTERM context ownership and apply the existing bounded landing Git runner to all its production Git phases: vantage normalization, inventory collection and intent branch/store probes. No new signal handler, runner, resource, runtime state or broad legacy-command signal migration. fleet policy creates none of the new isolated probes and retains its behavior; state/landing-status keep their existing observer entry behavior. Existing review/net and managed landing/local-check signal owners remain their authorities. The inventory command returns nonzero with no JSON on interruption only after the active Git group is terminated/reaped; no later probe starts. Native public-executable SIGINT and SIGTERM tests block each production phase with positive readiness, include a child retaining pipes, and prove CLI/child cleanup, zero stdout and no later probe. Reuse private repositories/queue root and existing binary/cleanup helpers. Targeted/race/catalog, final-head queued suite, close review and central landing remain required.

Re-plan approved by authenticated ops:0 receipt811b88c2-65e0-4ca1-976c-1fd4713400ee. Cancelled only still-queued lease-928955512; its runner exited without starting a suite. Native public executable control TestFleetMergeIntentsCLISignalsJoinAllGitPhases failed all six phase/signal cases before the fix, witnessing both Git and a pipe-holding child outlive CLI default-signal exit. The same six passed11.103s after inventory joined the existing executeCLI owner and all inventory Git phases reused bounded Git. No new handler/runner/state or other legacy-command changes. Broad race/catalog and fresh final-head queued verification remain before close/landing.

Approved native-signal fix qualification: broad fleet/intent/README race plus existing CLI lock-unwind and net native-signal controls passed87.873s; catalog passed1.939s, diff-check clean. All owned target runners are finished. Next checkpoint/push, fresh recorded pre-close check and round-three full close; no other legacy command changed, and no full-suite/SHIP claimed.

Round-three close returned SHIP with all five findings addressed, no new findings, exact measured actual and codecomplete evidence. Recorded pre-close attempt d27b2c86f3f1abbc6864ec1451e9738c on5724aaba9338a40537db4841a9e11e762fc73bd8 then failed after30m4s (queue1541s): TestVerbContextsReachTrackerReads found my context.Background compatibility wrapper; two landing-train registration barriers belong to #622/#625; shard3 alarm sampled TestSysVObservationPreservesCallerValuesWithIndependentBound/true in sweepStaleTemp/os.RemoveAll, leaving22 tests unrun. This is an incomplete failed suite, never a resumable pass.

Authenticated ops:0 receipt d2ab6974-89ff-4d6c-9a9d-1783efd0025b directs context fix, unclaimed scratch incident and fresh pre-close. Filed #631 with stack, exact failed log SHA256 bdb8d31e93054c9ec57539e5917a86f1479f71b5441aaa203ff09c7190873462 and05:30 load15.49/20.83/29.62. Root /Volumes/TestScratch/tmp is inferred from the24-byte root and unchanged runtime setup; exact removed basename and time inside RemoveAll were not recorded. No fake liveness, live-root cleanup or flake registration used in diagnosis.

Removed the context-free compatibility wrapper, consolidating on the context-taking shared reader and threading existing state/landing-status contexts; no other legacy signal owner or runner changed. Focused race/context guard/status/native/private conformance passed16.843s. Next: tracked catalog, checkpoint/push, fresh queued pre-close and narrow re-close for the post-SHIP fix, then central merge.
