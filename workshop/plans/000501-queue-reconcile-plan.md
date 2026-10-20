---
collaboration-mode: discussion
---

# Queue reconciliation implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Execute in this session with superpowers-executing-plans: this is retained #481 context; the hub asked this slot to avoid subagents. Steps use checkbox syntax.

**Current authority:** The 2026-10-20 simplification in Revisions below supersedes the original proposal and its checklists. Implement only that revision.

**Goal:** Let the steward retire one exact unresolved operation after explicitly attesting descendant absence, without losing the evidence or changing any pause owner.

**Architecture:** Serialize reconciliation with the existing control.lock. Validate exact record bytes, original owner incarnation and absence of an operation record, then durably publish an immutable audit bundle before unlinking the tombstone. A pure transition function owns the permitted effects and replay decisions; normal admission remains unchanged.

**Tech Stack:** Go, existing queuegate JSON codecs/flock/processgroup helpers, Cobra, portable private filesystem fixtures.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| ReconcileRequest | cmd/sdlc/internal/queuegate/reconcile_model.go | new |
| ReconciliationRecord | cmd/sdlc/internal/queuegate/reconcile_model.go | new |
| ReconcileState / ReconcileEvent / ReconcileEffect / TransitionReconcile | cmd/sdlc/internal/queuegate/reconcile_model.go | new |

ReconcileRequest binds one safe ID, lowercase 64-hex raw SHA256, nonblank bounded verification note and actor. The note attests that **no descendants remain**; the kernel probe proves only the original owner's incarnation ended. A record owns that request, UTC time and the original bytes' digest; its sibling original.json stores the bytes exactly. There is one audit bundle per reconciled ID during retention. Replays do not rewrite actor, note, time or unknown fields. A different request or record for that ID refuses rather than supersedes the audit.

TransitionReconcile consumes typed, validated observations and returns ordered effects. It neither probes processes nor opens files. Its state/effect enumeration is the single authority used by Store.Reconcile; rejected states emit no destructive effect. Extensions belong here rather than additional CLI booleans.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| Store.Reconcile | cmd/sdlc/internal/queuegate/reconcile_store.go | new | real private-root files, locks, durable audit publication |
| original-owner probe | cmd/sdlc/queue_reconcile.go | new | existing processgroup.StartTime / ErrNoProcess |
| queue reconcile command | cmd/sdlc/queuecmd.go and queue_reconcile.go | modified/new | existing landingline.Authorize, caller identity, Store.Reconcile |

Store remains portable under WF_SDLC_QUEUE_ROOT. The process seam is injected as a typed observation provider; a stateful fake registry holds PID/incarnation/alive/unknown state and can change between attempts. Production uses the existing kernel-backed StartTime implementation, not ps text or topology scans. Real helper-process conformance checks alive, waited-for exit and incarnation mismatch. The existing authority seam supplies operator/:0 admission; actor records OS uid plus existing caller/name information, bounded and JSON-encoded, without pretending that Git user.name is authorization. No credentials or external service is added.

## Scope and alternatives

Recommended: an immutable two-file audit bundle, committed by one directory rename. It preserves original whitespace and opaque future fields without duplicating the operation schema. One base64 JSON envelope would require raising existing queue JSON bounds for otherwise valid 1 MiB records; rewriting a decoded operation loses raw-byte identity. Blind unlink/manual editing and ps-based absence are rejected because neither records operator authority nor proves descendants ended. A new lifetime supervisor could prevent future unmanaged ambiguity but cannot retroactively own these descendants and is explicitly outside #501.

Accept any well-formed unresolved phase=unknown operation with complete PID/incarnation evidence, managed or unmanaged; reject synthetic/unreadable status entries and incomplete identities. A kernel ErrNoProcess or a different nonempty start time proves the original incarnation absent. A matching start time refuses; any probe error or empty result is unknown and refuses. Different-boot automatic cleanup remains status's existing behavior; this command does not broaden it.

The existing gate is the lifetime resource excluding concurrent queue writers during reconciliation. It cannot prove descendant absence after an unmanaged owner's lease has ended; that unavoidable fact is the explicit operator attestation, never a process inventory heuristic.

## State and effects (ARCH-ORDER)

Actions: inspect, reconcile with the exact request, repair evidence, wait for active work to finish.

| Internal state / event | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Exact unknown tombstone, no audit; matching request and original owner absent, operation record absent | ready | reconcile | none |
| Any wrong digest, different request, malformed record, unsafe ID, live/unknown owner, operation record present or held tombstone | refused, named reason | inspect/repair evidence or wait, as named | none |
| Evidence preparation/write/sync fails before publication | not reconciled | retry exact request after storage repair | unpublished temporary audit only |
| Exact audit published, tombstone still present | partially applied | retry exact request, revalidate owner/records and audit, re-sync audit, then unlink | none |
| Exact audit published, tombstone absent; unresolved directory sync/lost reply uncertain | applied, confirmation uncertain | retry exact request and sync directory | none |
| Exact audit and absent tombstone/operation, directory sync confirmed | reconciled / already reconciled | inspect | audit retained for bounded replay period |
| Tombstone and audit both absent, or conflicting/corrupt audit | unknown / refused | inspect/repair evidence | no fabricated successful reconciliation |

Pure model states are Ready, Prepared, Applied and Refused, classified from validated record observations; events are AuditDurable, TombstoneRemoved, DirectoryDurable and IOFailure. Effects are PublishAudit, RemoveExactTombstone, SyncUnresolved and ReturnConfirmed. Ready can only publish evidence; Prepared can only remove after fresh validation; Applied can only confirm after sync. Failure retains partial progress, never rolls back committed evidence. Tests enumerate these permitted and rejected sequences through the production function.

Before any mutation under control.lock: require safe exact filename (no symlink/nonregular file), bounded strict DecodeOperation, ID/unknown phase, complete PID/start identity, matching raw digest; obtain the unresolved file's nonblocking exclusive lock and hold it through retirement; require operations/ID.json absent via Lstat (any record, unreadable result or lease refuses); validate matching audit if present. This also protects crash-replay against ID reuse: any live record or changed tombstone refuses, even if an old audit exists.

An existing cancellation intent is decoded strictly. Preserve any existing interruption evidence unchanged; if intent lacks its retained interruption, use the existing retainInterruption seam before retirement. Leave the intent for normal status reconciliation, which can consume it using that retained evidence. Never modify pause.json, other operations/tombstones, existing interruption bytes or admission bindings. Do not run a full status sweep inside reconcile.

## Audit publication, bounds and end (ARCH-FUNERAL / ARCH-CONSTRAINTS)

Audit path: reconciled/ID/ with original.json (exact original bytes, <= MaxJSON) and receipt.json (strict version-1 metadata, <=16 KiB, including ID/digest/actor/note/time and preserved unknown fields). Both filenames are immutable. Under control.lock, prepare a private .reconcile-* directory, write/fsync both files, fsync that directory, rename to its final ID, then fsync reconciled/. Only after this completes may the exact tombstone be unlinked; fsync unresolved/ before success. An uncertain publication leaves the tombstone intact and returns retry advice. Replay revalidates and re-syncs committed evidence before any deletion. No subprocess or goroutine outlives the call.

Use existing atomicWrite/strict object/operation codecs and flock helpers; extract only the minimal reusable directory-sync helper needed. Do not add another recursive JSON scanner or widen operation/control schema limits. Unpublished audit directories are recognizable and cleaned only under control.lock, before a later reconciliation, never by admission. Unexpected or unreadable audit residue refuses rather than recursively deleting untrusted paths.

Writer-enforced retention: at most 128 audit bundles and 30 days, pruning oldest **completed** bundles only, before a new publication. A bundle still referenced by an unresolved/operations record is pinned; if no safe eviction is possible, refuse before mutating the tombstone. Each bundle costs at most 1 MiB+16 KiB, so the complete store is bounded near130 MiB. Basis: manual recovery is exceptional; 128 in a month is deliberately ample and the bound is remeasured by recorded bundle counts/bytes, not a per-launch workload assumption. Expired/evicted receipt means idempotency is no longer provable: absence then refuses instead of inventing success. Preserve future unknown members in metadata codecs; no replay rewrite occurs.

This is a user-triggered recovery path, O(record bytes + bounded audit count), not admission. Existing 3-second queue CLI deadline bounds gate wait and iteration. Admission adds zero scans/process probes. Counted tests at 8 and16 audit bundles require one bounded enumeration, linear reads, and prove ordinary TryAdmit performs none of this work. Oversize input, overcapacity pinned residue, cancellation and IO errors refuse safely. Network/UI/parallel fan-out: none.

## Implementation tasks (one atomic feature; one close review)

### 1. Pure request, receipt codec and transition contract

Files: create reconcile_model.go / reconcile_model_test.go under cmd/sdlc/internal/queuegate.
- [ ] RED table tests for request bounds, exact digest, legal effects and refused state/event pairs.
- [ ] Implement typed request/receipt and TransitionReconcile; strict required types/null/duplicates/version, capture Extra via existing decodeObject/encodeObject, retain original raw bytes separately.
- [ ] GREEN sequence tests: prepare/death/replay, publish/death/replay, unlink/death/replay, duplicate actor request, conflicting digest/note, expired evidence and late original-PID observations. Round-trip future outer/nested operation fields; original bytes including whitespace must remain identical. Test maximum-size valid operation fits the bundle without weakening queue codecs.
- [ ] Run go test ./cmd/sdlc/internal/queuegate -run 'TestReconcile(Model|Codec)' -count=1 and commit the tested unit.

### 2. Filesystem transaction and private conformance

Files: create reconcile_store.go / reconcile_store_test.go; modify store.go only for genuinely shared sync/retention helpers if needed.
- [ ] RED real-flock tests: success retains exact bytes/digest/actor/note/time; wrong digest, held tombstone, any operation file, live owner and uncertain probe leave all original bytes unchanged.
- [ ] Implement Store.Reconcile with injected owner observations, pure effects, per-step context checks and a test seam at actual durable transaction boundaries (not simulated successful effects).
- [ ] Adversarial tests inject write/sync/rename/unlink failure before and after each committed step, then rerun through the same production function. Require original evidence durable before deletion, no false confirmed success, preserved pause/interruption bytes and no duplicate audit on lost response. Change the record or create a conflicting lease between attempts; require refusal.
- [ ] Cover pending cancellation with and without retained interruption; status after reconciliation must not fabricate a new orphan unknown. Count at8/16 bundles, pin all capacity and verify safe refusal; preserve unknown/corrupt audit residue. Assert TryAdmit has zero process/audit scans.
- [ ] Run go test ./cmd/sdlc/internal/queuegate -count=1; commit the transaction and its controls.

### 3. CLI, operator guidance and acceptance

Files: modify cmd/sdlc/queuecmd.go; create cmd/sdlc/queue_reconcile.go / queue_reconcile_test.go; update atlas/workflow/sdlc-binary.md (queue section; index already links it).
- [ ] RED Cobra tests for exact --id/--digest/--note argv, existing steward/:0 authority, numbered worker refusal, bounded actor/note and conventional nonzero refusal. Test complete private snapshots before/after without pausing the production machine.
- [ ] Wire CLI to the existing portable Store and kernel probe. Output named audit path and reconciled/already-reconciled; errors name the safe recovery. Help explains operator attestation, finite replay retention, no descendant inference and no resume/restart. No new workflow/local-gates config.
- [ ] Run real child-process conformance for the owner probe: live child refuses, waited-for dead child permits, PID-incarnation mismatch permits, unsupported/error path refuses. Use only private queue roots and private records, never production pause owners.
- [ ] Test representative original unmanaged record shape end-to-end with the actual executable: compute raw SHA, reconcile, validate original bytes and receipt, observe running/complete/empty, repeat same request, then wrong request and live lease controls. Build in this slot only. No :0 stage without a grant.
- [ ] Run targeted queue/authority/legacy/cancellation and lint controls, push, recorded suite, central close review and land. Do not change code while review runs; no extra ad-hoc review.
- [ ] Hub rebuilds central binary. Operator performs real reconciliation only if a real unresolved record remains with attested descendant absence; otherwise use private executable evidence plus a read-only live status observation. Never seed an invented tombstone or force live fleet quiescence for acceptance.

## Coverage and live acceptance revision to approve

Covered: one same-boot unknown operation, strict authority, exact identity, lifetime/record exclusion, durable original/audit, partial write recovery, replay, bounded retention, cancellation preservation and no admission cost. Not covered: automatic proof of descendants, supervisor rollout, arbitrary corrupt-record repair, live process termination, clearing legacy overlay activity, or changes to queue status schema.

The originally named 10-20 unmanaged tombstone disappeared after the OS reboot, as existing status reconciliation permits. Current live status has active work/legacy-overlay uncertainty and therefore is not globally complete. Proposed acceptance revision: preserve the historical Done-when, prove exact same-boot recovery in the private executable fixture, and obtain read-only live status/operator acceptance without inventing that original tombstone or attributing unrelated active work to this issue. Ask the hub to approve this revision and the audit retention policy before change-code.

## Principles

ARCH-DRY: existing control.lock, codecs, authority, StartTime, interruption writer and portable roots. ARCH-PURE: request/receipt/transition versus file/process adapters. ARCH-PURPOSE: supported same-boot recovery plus executable consumer proof, not manual deletion. ARCH-MOCK: stateful process registry and real file/lock stack, checked against real child lifetimes. ARCH-CONSTRAINTS: explicit batch envelope and counted linear work. ARCH-SECURE: unsafe/malformed/unreadable/corrupt evidence refuses, no credentials or shared-process-group actions. ARCH-ORDER: durable effect sequence and crash/replay enumeration above. ARCH-FUNERAL: writer-enforced 30-day/128-bundle retention and pinned partial transactions. ARCH-EVOLVE: additive strict receipt codec and byte-identical original operation; no existing schema/version change.

## Revisions

### 2026-10-20 — hub-approved rare manual repair

Reason: authenticated hub receipt 9a890439-8430-48c5-b14b-97462e76a4bc approves the core but rejects the proposed transaction/replay model, two-file bundle and retention machinery. The implementation below replaces those portions of the original proposal; the earlier text remains only as design history. No flow pin: change-code infers the flow and close upgrades the review if the actual diff requires it.

**Operative design:** `sdlc queue reconcile --id ID --digest SHA256 --note TEXT`, authorized through existing operator/:0 logic. Validate bounded arguments and actor, then under control.lock lock/read the exact unknown tombstone, require matching raw SHA256, no operations/ID.json record/lease and original kernel PID incarnation absent. Unreadable/unknown/live evidence refuses. The explicit note attests descendant absence; never infer that from ps.

Append one JSON line to reconcile.jsonl: schema_version=1, ID, original raw bytes encoded as base64, digest, actor, note and UTC time. Fsync the log and root directory before unlinking the exact tombstone; fsync unresolved/ after unlink. Reuse a prior matching ID/digest audit entry instead of appending another. A crash after append but before unlink leaves both the logged attempt and tombstone; retry validates both and completes. A missing tombstone succeeds only with a valid matching audit and no active operation record. Conflicting/corrupt/unreadable audit evidence or a torn final log line refuses, preserving everything for inspection; never truncate or rewrite the append-only log. Each record is bounded to 2 MiB (base64 of an existing <=1 MiB operation plus metadata), parsed one line at a time. Unknown metadata fields are tolerated and remain byte-for-byte in the immutable log; strict required types/nulls/duplicates remain enforced using existing JSON helpers. No format change to existing queue records.

No transaction state model or new cleanup family. No audit retention or expiry, as explicitly directed by the hub for this rare manual repair; permanent cost is one audit line per distinct repaired digest and a streaming linear scan per manual invocation, with no admission cost. Existing pause and interruption evidence stays untouched; pending cancellation uses existing interruption preservation before tombstone retirement so ordinary orphan-intent reconciliation can finish without inventing unknown work. No live production pause or destructive repair during development.

**Current concepts:** Pure ReconcileRequest/immutable audit record validation and raw digest checks in internal/queuegate/reconcile.go. IO is Store.Reconcile using existing WithGate/flock, a kernel owner-incarnation callback and the append/fsync/unlink sequence. Cobra and existing authority/probe wiring live in queue_reconcile.go. Tests use the same portable private backend; real child-process checks conform the stateful PID/incarnation fake. All original model/bundle/retention entities and growth tests above are withdrawn.

**Current visible outcomes:** invalid/live/unknown evidence -> refused, inspect or wait as named; append/sync failure -> not reconciled, repair storage and rerun; durable audit plus remaining tombstone -> partial logged attempt, rerun exact digest; matching audit plus absent tombstone -> already reconciled. Pause owners and interrupted work never resume. Audit loss is not repaired by claiming a missing tombstone was reconciled.

**Current tasks:**
- [ ] RED the actual Cobra command with a private copy of the exact 10-20 unmanaged record shape; require recovered/complete private status and original audit bytes/actor/note/time. Controls: wrong digest, missing note, live operation lease, live owner, unknown owner and unsafe/corrupt record refuse unchanged; numbered worker authority refuses before storage.
- [ ] Implement minimal Store.Reconcile, immutable strict log codec and CLI. Use the shared kernel StartTime/ErrNoProcess seam; no external ps scan or background work. Replay is idempotent by ID/digest and preserves the first audit record. Reuse existing interruption writer when needed, preserving prior bytes.
- [ ] Crash a real helper after the audit is fsynced but before unlink, observe the surviving tombstone and one audit line, then retry through production code and require one unchanged line and completed recovery. Cover append/sync errors, malformed/torn audit and record changes between attempts. Verify raw unknown operation fields survive in base64 unchanged.
- [ ] Run targeted queue/authority/legacy/cancel/lint controls, build a binary in this slot and run private executable conformance. Read-only live acceptance is approved because reboot retired the original tombstone; never seed a fake production record. Update atlas/help; push, record suite, close and land normally. No separate ad-hoc review; no :0 staging without approval.

ARCH application: retain the core DRY/PURE/PURPOSE/MOCK/SECURE/EVOLVE boundaries above. ARCH-ORDER is served by the explicit append-fsync-before-unlink order and the real crash/retry control, without a new model. ARCH-CONSTRAINTS is a bounded-line streaming scan on an exceptional manual path, no admission scan/process cost. ARCH-FUNERAL retention is explicitly overridden by the hub: permanent append-only audit, one line per distinct repair, no retention or expiry rule.
