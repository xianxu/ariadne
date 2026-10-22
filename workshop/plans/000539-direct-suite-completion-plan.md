---
collaboration-mode: discussion
---

# Direct suite completion evidence implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A normally finished direct suite leaves no unresolved queue lifetime when the watcher positively observes exit 0 and confirms its private process group is empty; uncertain endings retain useful exit evidence.

**Architecture:** Keep the existing parse-time acquisition and detached watcher. Register a native exit observer inside the watcher before acknowledging acquisition, gather bounded process-group evidence once at exit, and feed those observations through one pure completion decision. Serialize cancellation and retirement under the existing queue mutex, preserve conservative recovery, and append a compact completion line to the queue audit for diagnosis.

**Tech Stack:** Python 3, Darwin kqueue, native process metadata, portable queue JSON/flock store, existing Go queue readers, real-process and filesystem tests.

## Contract and evidence

`QueueGate.retire` currently tombstones every admitted unmanaged operation, independent of how it ended. The watcher only polls `kill(pid, 0)` and therefore cannot distinguish clean completion. This is deterministic, not a Couch restart race. The original three tombstones contain no exit timestamp or status; preserve them unchanged as fixtures, with their audited hashes and provenance in the issue Log.

The hub approved this boundary in receipt 298daa02-f072-4415-a649-99a5c8cc6d5b: observed exit 0 **and** no live process in the make's process group authorize retirement. This is process-group completion, not proof that an arbitrary descendant did not deliberately escape into another session. No new command-wrapper contract or managed-run semantics are introduced. Require the recorded group to equal the make PID; a shared shell group cannot be assumed private. Capture that binding before admitting work and recheck the kernel incarnation around observer registration. Never signal or reconcile a group as part of this fix.

The watcher itself must detach into its own session before checking the make group, otherwise it would always count as a survivor. A watcher handshake must close all unused pipe ends and have a bounded deadline; acquisition must not announce success before the observer has registered. Observer failure does not imply suite completion. Managed operations continue to be owned and retired by Go.

Darwin's NOTE_EXITSTATUS supplies a wait status to an authorized nonparent observer. Decode with the OS wait-status helpers; an event lacking the status flag does not establish exit 0. Registration happens after fork because the kqueue is not inherited across fork. On hosts without an equivalent supported status source, retain today's conservative tombstone with an explicit unavailable-status reason. Do not infer zero from an absent PID, an empty log or a successful acquisition.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| DirectCompletionEvidence | scripts/test-lock.py | new |
| direct_retirement_decision(evidence, cancellation) | scripts/test-lock.py | new |

Evidence names the exact owner binding, observed exit time, known wait status or unknown status, private group binding, group observation (empty / survivors / unknown), and cancellation state. One operation owns one evidence value. The decision is the sole authority for selecting normal retirement versus unresolved retirement; it does no IO. Keep explicit tagged outcomes rather than independent booleans whose combinations silently imply success. Future platforms can supply the same evidence without changing retirement policy. Unit tests in `scripts/test/test-lock-exit.test.py` execute the production decision without IO mocks.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| Native owner-exit observer | scripts/test-lock.py | new | kqueue and exact PID/start binding |
| Process-group snapshot | scripts/test-lock.py | new | bounded native process inventory |
| Watcher acquisition handshake | scripts/test-lock.py | modified | fork, session detachment and pipe lifecycle |
| QueueGate retirement | scripts/test-lock.py | modified | existing mutex, atomic JSON and operation lease |
| Completion audit append | scripts/test-lock.py | new | compact JSONL beside existing reconciliation audit |

The IO shell supplies typed observations to the pure decision. Test doubles model an owner registry with incarnations, exit events and group membership that change across calls; they can return unavailable, partial and contradictory observations. Use the same seam for real native conformance tests, not a separate test-only completion path. The filesystem backend remains rooted in `WF_SDLC_QUEUE_ROOT`; no test may read or modify the machine queue. Reuse the existing strict JSON/raw-field preservation and atomic-write helpers (ARCH-DRY).

## State and ordering

| State / event | Result and effects |
|---|---|
| Waiting / owner disappears or acquisition fails | Existing waiting retirement; no claim that a suite ran |
| Admitted / observer registration acknowledged | Return acquired; watcher owns operation and slot FDs |
| Admitted / exact exit event, status 0, private group empty, no cancel | Append and fsync completion audit, then persist finished phase, then unlink operation and release FDs |
| Admitted / nonzero, signal, missing status, identity mismatch, shared/unknown group, survivors or failed scan | Persist unknown tombstone with all known observations, then remove operation and release FDs |
| Admitted / valid cancellation | Preserve existing fenced owner signal and interruption envelope; cancellation wins over a later clean-exit observation |
| Any / malformed cancellation or operation | Retain uncertainty; never erase the corrupt evidence or reinterpret it as no cancellation |
| Watcher dies before durable finished phase | Existing Go recovery retains unknown; no automatic clearing on dead PID |
| Watcher dies after durable finished phase | Existing Go observer may remove terminal operation residue; completion audit already durable |
| Couch restarts while watcher survives | No protocol transition; independent watcher continues to completion |
| Duplicate retirement / crash around audit or unlink | Same operation ID and binding; no second completion claim, no loss of existing unknown fields |

Read cancellation and validate the current operation binding while holding `control.lock` at the final decision/write. A cancellation written first prevents normal completion; once the finished transition is committed, later cancellation cannot resurrect the retired run. Keep process probes outside the global mutex, then validate that their operation binding still matches inside it. Failed or partial process enumeration is unknown, never empty. A recycled group ID is not authority to signal or delete anything; finding any live member prevents retirement.

## Persisted evidence and resource lifetime

Add optional evidence members to operation JSON and completion audit: `exit_observed_at` (UTC wall time when observation occurred), `exit_status` only when positively known, `process_group`, `group_observed_at`, `survivor_pids`, and a typed observation/reason indicating unknown or partial evidence. Omission means unknown, never zero. Include a concise human explanation in existing `detail`, so current Go status readers expose the useful evidence without requiring a second authoritative parser. Existing Go readers preserve additive members through `Extra`; exercise that round trip. Known evidence members must be strict in the Python boundary, duplicate keys remain invalid, and nested unknown fields must survive rewrites.

On normal retirement append one compact line to `completion.jsonl` beside `reconcile.jsonl`, under the existing mutex, and fsync before committing finished phase. Record operation ID, PID, incarnation, `exit_observed_at`, known status 0 and confirmed empty group. Do not append another record type to `reconcile.jsonl`: older reconciliation readers strictly require an original unknown operation, so mixing completion records would break retries. This follows the existing append-only audit lifecycle and line-bound model, with a stricter 1-KiB completion-line bound; no per-operation receipt directory, retention registry or diagnostic sweep. Its explicit lasting cost is at most 1 KiB per normal direct-suite completion; document this operator-approved append-only audit growth. It is diagnostic evidence, never suite-pass or admission authority. A torn append or failed fsync prevents normal retirement; preserve uncertainty. Unknown endings reuse existing bounded unresolved/interrupted stores and operator reconciliation.

The watcher owns its exit observer, handshake end, operation FD and slot FDs until retirement; close each on every exit path. A parent that cannot confirm watcher readiness must not return acquired, and must not accidentally release a still-running admitted owner without retaining uncertainty. No new perpetual sweeper or supervisor is introduced. Tests clean up only PIDs and groups they created and joined.

## Operating envelope

This is batch teardown, not a per-keystroke path. Preserve the current 100-ms cancellation polling cadence. Exit observation adds no fleet inventory on each tick: perform one bounded group scan after exit. Bound native enumeration to 65,536 entries and 5 seconds; these are conservative initial machine limits, not claims that a timed-out scan proves emptiness. Overflow, permissions, partial enumeration or timeout yield unknown. Record only a bounded survivor list (256 PIDs) with an explicit truncated marker; any truncation refuses normal retirement. Readiness handshake deadline is 5 seconds; timeout is acquisition infrastructure failure. All bounds get boundary tests and named constants.

Measure scan call counts at N and 2N fake process entries: one enumeration, linear filtering, no per-process subprocess and no scan on idle watcher ticks. Existing 4096-operation/1-MiB registry bounds remain. Network is N/A: no network dependency or credential. Co-tenancy matters: use native metadata or a single bounded snapshot command, never repeated ps invocations per PID. Native process-list errors must remain visible. If a command is necessary, route it through the typed snapshot seam, with bounded output and timeout and a stateful fake.

## Risky functions and executable test strategies

| Production function | Adversarial strategy and independent mechanical guard |
|---|---|
| `direct_retirement_decision(evidence, cancellation)` | Generate the finite evidence-state product and event sequences; assert the independent invariant that only exact observed zero plus confirmed empty private group with no cancellation can finish. |
| `decode_exit_evidence(fields)` | Seed a generative malformed/partial JSON corpus from the real incidents; strict rejection and raw unknown-field round trips guard against fabricated completion. |
| `OwnerExitObserver.register()` / `.poll()` / `.close()` | Stateful owner/incarnation/event double with controlled timing plus real sibling-process conformance; a barrier proves registration before target release and independent parent wait checks status. |
| `process_group_snapshot(pgid, owner)` | Stateful complete/partial native snapshots with counted N/2N entry filtering plus real private-group conformance; uncertainty must never equal emptiness and actual surviving children are independent witnesses. |
| `start_owner_watcher(gate, watch)` | Inject failure at handshake boundaries and compare against real fork/session behavior; acquired output is forbidden before readiness and private FD/child cleanup is observed externally. |
| `QueueGate.retire(evidence)` | Stateful filesystem crash/cancel interleavings through existing mutex; a fresh Go observer and independent physical-slot acquisition check durable terminal versus unknown behavior. |
| `append_completion_audit(root, operation, evidence)` | Inject torn writes, size violations and fsync failures; strict whole-line reads and the operation's retained lifetime prove failed durability cannot announce normal completion. |

## Implementation and verification

### Task 1: Preserve incident evidence and reproduce normal completion

**Files:** new `scripts/test/fixtures/direct-suite-completion/{65dbdc31d6444c45a8692456d1e0a595,f70ce4520db34fdaa8208eb03bf778c0,b54088501a804c8eab8d7bab2de27e79}.json`; new `scripts/test/test-lock-exit.test.py`; existing `scripts/test/test-lock-pause.test.py` private fixture where reusable.

- [x] Copy the original operation bytes out of the three audited base64 payloads, verify hashes against the Log, and record provenance beside fixtures. Do not reconstruct or normalize these JSON bytes. Assert old records still decode as unknown and never become completion evidence by themselves.
- [x] Build a real isolated make fixture in a private process group, with a recipe barrier that lets the test observe admission before releasing a successful recipe. Await bounded observable state, not a sleep standing in for readiness. After make exit 0, assert no active/unresolved record, a valid completion audit line with bound PID and observation time, and reacquirable physical slots.
- [x] Run `python3 scripts/test/test-lock-exit.test.py -v`. Record the red failure showing the current tombstone. A separate oracle must inspect actual queue files and acquire the slot; do not merely call the new decision function and assert its return.

### Task 2: Observe exact owner exits and fence the decision

**Files:** modify `scripts/test-lock.py`; test `scripts/test/test-lock-exit.test.py`.

- [x] Write decision and evidence-parser tests using the strategies above and the state-transition table as the contract; run them red before implementation.
- [x] Register native exit observation in the forked watcher, detach its session, and acknowledge readiness through a bounded pipe handshake before returning acquired. Capture group/incarnation before registration, verify afterward, and fail conservatively on mismatch. Remove or reuse the currently unused `wait_for_exit` instead of leaving a second exit implementation.
- [x] Implement `OwnerExitObserver` and `process_group_snapshot` through the named seams; execute their stateful and real-process conformance strategies above.
- [x] Feed exact exit and group observations into `direct_retirement_decision`; preserve managed borrowing and fenced cancellation. Apply its independent completion invariant through the actual watcher path.
- [x] Run the new file and `python3 scripts/test/test-lock-pause.test.py -v`; keep existing recycled-PID, surviving-child, mutex-reopen, cancellation, nested and managed controls.

### Task 3: Persist diagnostic evidence and cover crashes

**Files:** modify `scripts/test-lock.py`; tests in `scripts/test/test-lock-exit.test.py`, `cmd/sdlc/internal/queuegate/codec_test.go`, `cmd/sdlc/queue_runtime_test.go`.

- [x] Write exit evidence on unknown outcomes and a compact diagnostic audit line for confirmed completion. Serialize final cancellation/binding validation and writes using `control.lock`. Preserve all additive raw fields. Apply the named parser, retirement and audit adversarial strategies above.
- [x] Inject failures before audit, before/after finished-phase persistence, and before/after unlink. A fresh Go observer must retain unknown before the committed terminal transition and remove terminal residue after it; no record from the three incidents is retroactively treated as finished.
- [x] Kill the watcher mid-run in a private fixture; verify unknown remains and no fabricated exit timestamp/status appears. Model a Couch/supervisor exit while the independent watcher survives; release the recipe and verify normal completion. Make clear that losing the watcher is different from restarting Couch. Test that late cancellation cannot rewrite a committed finished outcome.
- [x] Verify current Go strict decoding and rewriting preserve new evidence through `Extra`, and `detail` makes exit observations/survivors visible in a real private queue-status invocation. Add the new Python file to the existing Go helper integration runner so the recorded suite actually runs it.
- [x] Run `go test ./cmd/sdlc/internal/queuegate -count=1` and the targeted `go test ./cmd/sdlc -run 'TestQueueRealHelper' -count=1 -timeout 10m`; run `bash scripts/test/test-lock.test.sh` and `python3 scripts/test/test-lock-priority.test.py -v` for shared Makefile/helper compatibility. Check exact existing runner names before selecting the final commands.

### Task 4: Document, close and land

**Files:** modify `atlas/workflow/sdlc-binary.md`; update issue Log and this plan.

- [x] Update the existing direct-suite paragraph: observed clean exit plus empty private group permits retirement; unknown status, survivors, cancellation and observer failure remain unresolved. Document observation timestamps, completion audit location and per-run growth, platform fallback and the managed-path distinction. Do not imply a historical exit time can be recovered from old tombstones.
- [ ] Record red/green evidence, native conformance results, count/limit tests and any platform limits. Check `git diff --check`, commit with issue reference and author trailer, push through SDLC, and run the recorded suite on the pushed head with `sdlc local-check run --name suite`.
- [ ] Close with the mandatory fresh-context SDLC review, resolve findings, then merge using the retained SDLC runner. No live machine cancellation/reconciliation and no :0 deployment are part of this plan. Report the outcome and PR to the hub.

## Architecture review

- ARCH-DRY: reuse queue mutex/atomic JSON/unknown-field codec and managed completion ownership; one exit observer and one decision owner.
- ARCH-PURE: evidence-to-decision matrix is pure; process and persistence seams supply observations.
- ARCH-PURPOSE: fix real clean direct completion, preserve all three original captures, record future observed exit times; do not claim old absent timestamps were recovered.
- ARCH-MOCK: stateful owner/group observer double plus real sibling-process and actual make conformance; private portable filesystem queue.
- ARCH-CONSTRAINTS: one bounded exit-time scan, counted N/2N coverage, explicit startup/scan deadlines and unknown-on-overload.
- ARCH-SECURE: exact PID/start/private-group binding, strict optional evidence, no machine-state test access or credentials, no signaling expansion.
- ARCH-ORDER: one explicit decision matrix; handshake before acquire returns; cancellation and durable retirement serialize; crash cut points tested independently.
- ARCH-FUNERAL: detached watcher closes every handle, completion audit line bounded with explicit per-run lasting cost, unresolved records retain existing reconciliation lifecycle.
- ARCH-EVOLVE: additive schema-1 evidence and opaque-field preservation through old readers; no new field can independently authorize an old reader to delete unknown evidence.

## Revisions

- 2026-10-21: TL approved the plan in receipt c86c6fec-3912-4277-9ea1-f519c2aa7f25, cutting the proposed 256-file diagnostic receipt store. Replaced it with one compact completion audit line beside reconciliation, with no receipt registry or sweeper. Kept the separate log because existing reconciliation readers require original unknown-operation records. Corrected the earlier outgoing checkpoint message: actual plan commit is `35fb28672`, not its accidental placeholder; the correction was also sent immediately through Couch.

- 2026-10-21: PQ-1 refined the test contract: named the production decision, validation, observer, snapshot, handshake, retirement and audit functions, each with one adversarial strategy and independent mechanical guard; compressed prose input inventories into this table. Behavioral state table and approved scope are unchanged.

- 2026-10-21: Implementation checkpoint: Tasks 1–3 verified with actual make, stateful native observation/snapshot and audit failure seams, real handshake failure/timeout, production retirement crash cuts, fresh Go reader recovery, original bytes and current-provider compatibility. Supplemental pure/integration Python test file is `scripts/test/test-lock-decision.test.py`, registered alongside the real-process file by the Go suite. Group membership is captured for shared groups too, while the pure decision still refuses retirement unless pgid equals the owner PID.
