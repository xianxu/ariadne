---
collaboration-mode: discussion
---

# Owner-scoped queue drain implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3. Use the executing-plans skill for the shared admission integration; bounded independent tests can be delegated. Steps use checkbox syntax.

**Goal:** Implement the frozen pair#722 queue contract: owner-scoped pause/resume, trustworthy global status, interruption rather than failure, and the TL-worker drain instruction.

**Architecture:** One machine-wide gate serializes pause mutation with admission and active lease publication. A lifetime flock, rather than a PID observation, proves each new operation exists. The existing Python weighted FIFO and repository landing lines retain their ordering; both consult this gate at their actual admission boundary. Read-only legacy observations can add unknown activity but never authorize cancellation or erase lease evidence.

**Tech Stack:** Go, Python 3, JSON, Unix flock, existing processgroup and workspace identity adapters. No service, network dependency, timer-based resume, workflow edit, or schema change to SuiteAttempt.

## Contract and decisions

- The issue's frozen Contract supersedes the older Spec about owner tokens and drained reports. Do not copy the supervisor's complete notice template into help.
- CLI argv and required status fields are exactly those frozen in #481. Pause/resume require a nonempty bounded `--owner`; reason is opaque text. The commands reuse `landingline.Authorize` with current checkout and COUCH_TREE identities. Read-only status needs no mutation authority.
- `pause` adds one owner idempotently. Repeating the same owner can upgrade graceful to cancel-running, never downgrade that intent. It does not replace other owners. Repeating a token with a different reason retains its original reason; a new drain uses a new token.
- `resume` removes only the named owner and is idempotent when absent. Owners never expire. Removing the last owner allows the existing queues to admit again; it never restarts interrupted work.
- Hub clarification, 2026-10-20: graceful pause permits already-running work to finish, but an admitted merge stops with interruption if it reaches a new suite boundary while paused. It must not stay active indefinitely waiting for permission to start that suite.
- Hub clarification: old/unreadable activity is unknown, never absent. Until every pre-gate runner has ended, status is incomplete and names it. New workers use central rebuilt sdlc; no staging or touching the live machine pause state during tests.
- pair:5 confirmed its `sdlcDrainQueue` uses inherited environment unchanged, verified steward cwd, 5-second calls and 1 MiB combined output. Nonzero mutation exit is uncertain, not proof of rollback. Status stdout is JSON only. Production protocol remains unchanged.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Control` / `PauseOwner` / `Transition` | `cmd/sdlc/internal/queuegate/model.go` | implemented in M1 |
| `Operation` / `TransitionOperation` | `cmd/sdlc/internal/queuegate/model.go` | implemented in M1 |
| `Snapshot` / `BuildSnapshot` / `ValidateSnapshot` | `cmd/sdlc/internal/queuegate/status.go` | implemented in M1 |
| `InterruptionRecord` | `cmd/sdlc/internal/queuegate/interruption.go` | M1 boundary correction |

- **Control / PauseOwner:** `Control` holds the versioned array of 0..128 `PauseOwner` values. `Transition(Control, Action) (Control, error)` accepts pause/resume actions and returns a detached next value; it does not return interruption IDs. The exported Go values are mutable. Persistence is constrained by `Store.Pause`/`Store.Resume`, which invoke the pure transition and validating codec under the gate. Fast pause's separate store scan writes cancellation intents for exact held operations. Unrelated owners and opaque future members survive rewrites.
- **Operation / TransitionOperation:** `Operation` holds one opaque operation ID, workspace (possibly empty), head (optional), kind(s), process incarnation and a string `Phase`. `TransitionOperation(Operation, event) (Operation, error)` validates legal transitions among waiting, running, interrupting, unknown, finished and interrupted. There is no separate phase type; interruption diagnostics use the separate envelope described below. Enrolled/waiting is excluded from public active entries, while its held lease lets process observations recognize a new waiter. Multiple kinds remain separate status entries keyed by operation ID plus kind, not by PID.
- **Snapshot:** `BuildSnapshot` derives running/paused/unknown, completeness, required arrays and detail from `Control` and operation observations; `ValidateSnapshot` checks state/owner combinations, bounds and diagnostics. `observeLegacyQueue` supplies the conservative legacy overlay. A malformed record never becomes an authoritative empty result.
- **InterruptionRecord:** the original hypothetical interruption struct becomes a separate write-once diagnostic envelope: schema version, ID, `outcome:"interrupted"`, embedded `Operation`, reason and recorded timestamp. `Store.retainInterruption` records this evidence even when the embedded operation remains unknown or interrupting; recording a cancellation does not claim every descendant has finished or change unknown into completed. M2 still must connect real runner cancellation and retained logs to this distinct evidence family. It must never become a failed SuiteAttempt or a successful local-check status, poison flake/main-baseline classification, or block a same-head retry.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `Store` / `Lease` | `cmd/sdlc/internal/queuegate/store.go`, `lease.go`, `codec.go` | new | portable files, atomic rename and flock |
| `NewQueueCmd` | `cmd/sdlc/queuecmd.go`, `cmd/sdlc/main.go`, `cmd/sdlc/helptext/queue.md` | new | CLI, existing authority and snapshot IO |
| `observeLegacyQueue` | `cmd/sdlc/queue_legacy.go` | implemented in M1 | bounded local topology, held old locks and process observations |
| Python queue gate adapter | `scripts/test-lock.py` | modified | same root/mutex/lease protocol at weighted admission |
| Go admission adapters | `cmd/sdlc/localcheck_run.go`, `landingqueue.go`, `merge.go`, `landing.go`, `landingtrain.go`, `cmd/sdlc/internal/landingline/line.go` | modified | existing suite execution and all merge/landing paths |
| TL-worker help | `cmd/tl-worker/help.md`, `cmd/tl-worker/testdata/rules.tsv` | modified | embedded worker instruction |
| `queueRuntime` / `startQueueOperation` | `cmd/sdlc/queue_runtime.go` | new | exact-ID lease, workspace identity and cancellation reader |
| `queueInterruptedError` / `retainQueueInterruption` | `cmd/sdlc/queue_interruption.go` | new | interruption error and separate raw log retention |
| `QueueGate` | `scripts/test-lock.py` | new | portable disk gate and borrowed/owned suite lifetime |

- Reuse the existing flock implementation, `processgroup.StartTime`, managed suite signal/context cleanup, landing authorization, workspace resolution and queue order. Do not duplicate the weighted scheduler or grant calculation.
- Python only needs a small reader/admission lease adapter for the shared disk protocol; Go alone owns pause-set mutation. Cross-language fixtures test the reader against the Go writer, including malformed and additive JSON. No external process is started on every admission poll.
- Stateful tests use real portable folders and real flocks. pair#722's existing FakeDrainQueue remains the peer's model; actual executable conformance compares our CLI results through its decoder/adapter, rather than inventing a second nominally compatible protocol.

## Shared storage and lifetime

`WF_SDLC_QUEUE_ROOT` selects a private portable state directory. Default: `$XDG_STATE_HOME/ariadne/queue`, falling back to `~/.local/state/ariadne/queue`. Go and Python resolve this identically. A test supplying `WF_TEST_LOCK_FILE` without an explicit queue root derives a sibling private queue root, so existing private lock fixtures never consult or mutate machine owners. Tests also use a private fleet discovery scope; neither fixture can enumerate live work.

| Artifact | Creator / last reader / end | Bound |
|----------|-----------------------------|-------|
| `control.lock` | Store; every short admission/mutation/snapshot transaction; fixed reusable file | one |
| `pause.json` | authorized pause/resume; admission and status; empty owner set after matching resumes | 128 owners, bounded reason/token text, <=1 MiB |
| `operations/<id>.json` (flock-held) | admitted/enrolled operation; status/cancel owner; owner removes on retirement, next locked scan reaps proven dead leases | <=4096 live entries; no truncating success |
| cancellation intents tied to exact IDs | fast pause; owning run; retirement removes together with lease | <=4096 live intents |
| `interrupted/<id>.json` and optional log | interrupted owner; operator/retry diagnostics; prune after 30 days, cap 4096 newest records | bounded JSON/log retention, never ordinary failure store |

- Publish every lease already locked, under the control mutex, before making an admission visible. Mutating a lease phase or kind happens under that same mutex. Snapshot holds it while reading/probing, so it cannot see the new admission without its lease or a partial rewrite.
- Lock order: global control mutex, then existing queue allocation/line mutex; lifetime locks are nonblocking probes. Never hold the global mutex during queue waits, a provider call, Git/network IO, running work or cancellation completion.
- Landing's `tryAdmit` must check the gate inside its admission transaction; a stale Options.Admission callback is insufficient. The handle owns the resulting lifetime lease until merge cleanup finishes, not merely until head.lock is released.
- Machine helper checks pause, takes weighted slots and publishes its active lease atomically with respect to pause under the same mutex. Its existing watcher holds the lease and slot fds until the watched owner really ends. A Go local-check controller supplies its exact safe cancellation identity; nested/reentrant makes reuse the enclosing operation instead of multiplying active suites.
- Fast pause durably records the owner and cancellation intent first. The running owner handles its own intent using managed cancellation and existing processgroup cleanup; a same-ID/locked-lease plus kernel start-time check fences any signal. The pause call returns promptly; status reports interrupting until actual lifetime resources release. No kill based on a process snapshot, no arbitrary process-group signal, no unlinking another live holder's lock.
- A direct make holder must remain interrupting/unknown until its watched lifetime is actually released. Its helper must not report quiescence merely because a termination request was sent. Real nested-child tests verify this boundary.
- Process death after pause publication cannot lose the hold. Process death before/after interruption publication is reconciled from the held lease/intent; stale cleanup preserves interruption evidence before removing residue. Resume never clears those records or implies restart.

## State/event enumeration (ARCH-ORDER)

Pure transitions own legal mutations. Filesystem effects carry outcomes back; uncertain writes require status reconciliation, not assumed rollback.

| Prior state | Event | Next state / effect |
|-------------|-------|---------------------|
| no owners | pause A | paused by A; no new admission |
| A present | repeated pause A | same hold; fast can upgrade intent; original reason retained |
| A present | pause B | both holds; still paused |
| A+B present | resume A | only B remains |
| A present | resume A | no owners; queues may elect existing FIFO heads |
| any | resume absent owner | unchanged |
| enrolled/waiting | paused admission | retain ticket and ID; print pause reason; no active entry |
| enrolled/waiting | unpaused admission | running; publish lease and take resource in same transaction |
| running suite | graceful pause | running until completion; no successor starts |
| running suite | fast pause | interrupting; owner cancellation; retain interruption on retirement |
| admitted merge | paused next-suite boundary | interrupted; release turn; preserve prior successful head evidence |
| admitted merge | suite already passed on exact head | retain evidence; finish current merge step without launching new work |
| running | completion races fast pause | serialize completion/cancel request; completed-before-request stays completed, otherwise interrupted conservatively |
| interrupting | additional owner / resume | does not undo cancellation or restart work |
| held lease | unreadable metadata / unavailable probe | unknown/incomplete; do not signal or infer absent |
| old holder | status | unknown/incomplete naming old owner; no cancellation authority |
| pause write has uncertain outcome | repeated pause / status | token reconciles idempotently; no automatic compensating resume |

### User/automation state table

Actions: pause, pause --cancel-running, inspect status, resume the named owner, checkpoint/ack a drain, repair evidence/roll out the rebuilt binary, rerun the interrupted exact head.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|-------------------------------|---------------|-----------------|---------------|
| zero owners, complete observation | running | pause | no loss |
| one or more owners, waiters only | paused; active [] | resume named owner | queue ordering preserved |
| paused with running suite/merge | paused; active running | inspect status until current work finishes | no completed evidence lost |
| fast cancellation requested, owner still alive | paused; active interrupting | inspect status until retirement | unfinished work may need rerun; logs retained |
| merge reaches blocked new-suite boundary | interrupted record; merge exits | rerun exact head after resume | successful prior suite retained; partial uncompleted suite needs rerun |
| interruption complete | paused; no corresponding active entry | rerun exact head after resume | interrupted attempt never becomes a failure or flake |
| legacy runner / malformed held state / observation timeout / overflow | unknown or incomplete, with detail and unknown entry as applicable | repair evidence/finish rollout | no false quiescence; drain may remain stuck |
| pause/resume response lost after effect | next status shows owner set | inspect status | caller may not yet know whether mutation applied |
| worker receives authenticated graceful/fast notice | checkpoint owed until durable; then ack | checkpoint/ack drain ID | fast preserves unfinished edits, not completed-work claims |

Covered: multiple concurrent owners, lost replies, graceful and fast transitions, races at both admission points, all three active kinds, same-head retries, unknown legacy activity and process/lease retirement. Not covered: Couch board/ack/notice rendering (pair#722 owns it), preemption/queue reordering, owner TTLs, silently clearing damaged state, restarting cancelled work or operator disk/permission changes.

## Legacy rollout and status trust

- New global leases are authoritative. Read-only legacy discovery is an overlay because old ordinary merges have no continuous lifetime lock. Do not claim that an empty new registry proves the whole machine is quiet.
- Resolve the steward's fleet topology locally using bounded `fleet.NormalizeVantage`, then `project.FleetRepoDirs` plus Git common-dir deduplication only. Do not call full CollectInventory: it can fetch trackers and inspect every worktree.
- Probe held legacy machine slot locks and each discovered common directory's landing/head.lock. Match a complete new operation binding where possible. Unmatched/unreadable holders produce stable observational IDs (resource identity plus incarnation/ticket), unknown workspace when unproved, state unknown and complete=false. Weighted slots deduplicate by complete owner identity, never PID alone.
- Observe same-user merge processes under a bounded platform adapter to catch old ordinary and cleanup paths without a landing lock. New enrolled/waiting lease bindings exclude new waiters from active; unknown or ambiguous old processes remain unknown. PID+kernel start is only observational identity, never permission to signal.
- Failed/incomplete topology or process inspection, unsupported platform or exhausted bounds gives complete=false with reason. No cached earlier empty snapshot proves current quiescence. If inspection cannot establish old-runner absence, drain stays visibly stuck pending rollout/repair.
- All observations are bounded and local. Tests control the process/topology adapters and real legacy flocks. No test pauses the live fleet. After landing the hub rebuilds :0; older running binaries finish/relaunch before a complete snapshot can be trusted.

## Operating envelope / trust

- This is command/admission/shutdown work, not a per-byte or UI path. Mutation/snapshot target <4 seconds, below pair's 5-second ceiling. Mutex acquisition and legacy observation use context deadlines; inability to observe returns unknown/incomplete rather than a hanging call.
- Admission poll cost is O(P + Q) for bounded owner set P and existing FIFO ticket count Q; do not scan fleet/processes during admission. Registry snapshot is one O(A) lease pass plus one bounded legacy process/topology pass. Count reader calls for 8 and 16 records/tickets and reject quadratic rescans; timing alone is not the guard.
- Bounds: 128 owners, 4096 active entries/leases, <=1 MiB status. Check output size before writing stdout; oversized results become a small unknown/incomplete snapshot, never truncated JSON or truncated success. Bound token/reason/workspace/detail fields so the decoder's proposed limits remain satisfied.
- Persisted pause and lease JSON are untrusted: reject duplicate keys, wrong known types, missing required members, malformed IDs, invalid enum combinations and unsupported versions. Decode unknown additive members into raw JSON and preserve them on rewrites. Incomplete/unsupported control data blocks admission visibly; it never resumes by treating a parse failure as no owners. No format-version refusal without an explicit unsafe-to-ignore justification.
- No credentials touched. Reason text stays data, is encoded with JSON and printed safely as quoted text; never used as shell or executable input. File paths derive from validated IDs, never reason/owner text.
- Architecture lenses: ARCH-DRY reuses schedulers and centralizes pause ownership; ARCH-PURE/ORDER isolate transitions; ARCH-MOCK uses portable real stores and peer fake conformance; ARCH-CONSTRAINTS bounds output/work; ARCH-SECURE makes unknown explicit; ARCH-FUNERAL names cleanup; ARCH-EVOLVE preserves additive fields; ARCH-PURPOSE covers both admission families and every merge path.

## M1 — Core model, portable backend and CLI protocol

Files: new `cmd/sdlc/internal/queuegate/{model,status,codec,store,lease}.go` with colocated tests; new `cmd/sdlc/queuecmd.go`, `queuecmd_test.go`, `queue_legacy.go`, `queue_legacy_test.go`, `helptext/queue.md`; root command registration and registry/help tests. Add backend fixture files under `cmd/sdlc/internal/queuegate/testdata/` as needed.

- [x] `Transition` / `TransitionOperation` / `BuildSnapshot`: TDD pure event sequences against the independently stated state-table invariants; duplicate/reordered events cannot release another owner, invent an active waiter or convert uncertainty to completion.
- [x] `StrictObject` / `DecodeControl` / `DecodeOperation` / `ValidateSnapshot`: malformed JSON and fuzz seeds exercise strict typed known fields, bounded depth/bytes and opaque unknown-member round trips. `StrictObject` owns structural validation shared by persisted metadata and the legacy observer.
- [x] `Store.WithGate` / `Store.Pause` / `Store.Resume` / `Store.Enroll` / `Lease.TryAdmit` / `Lease.Retire`: real-flock barriers force ordering around publication and death; count 8/16 lease reads to reject superlinear scans, and use failed-write/retry sequences to preserve applied owner effects. Write tests first, observe RED, implement the portable backend, then GREEN.
- [x] `NewQueueCmd` / `runQueueMutation` / `runQueueStatus`: run the real private Store through argv with the existing authority adapter; peer strict decoding, output caps and loss-after-effect reconciliation protect framing and idempotence.
- [x] `observeLegacyQueue`: controlled process/topology sequences plus real held legacy flocks force partial/ambiguous evidence; the mechanical guard is complete=false and no signaling from observations. Production stays local and bounded; no full CollectInventory.
- [x] Commit and close M1 with targeted evidence through `sdlc milestone-close`; it owns the fresh review. M1 establishes the backend/CLI boundary; M2 is required before issue close or landing, so no fleet rollout occurs with unwired consumers.

## M2 — Admission, interruption, drain help and peer conformance

Files: `scripts/test-lock.py`; new `scripts/test/test-lock-pause.test.py`; Go `localcheck_run.go`, `localcheck_lease.go` as needed, `landingqueue.go`, `merge.go`, `landing.go`, `landingtrain.go`, `internal/landingline/line.go` and their integration tests. New `queue_interruption.go` / tests keeps interrupted records out of SuiteAttempt readers. Update `cmd/tl-worker/help.md`, rule TSV, README and atlas/workflow/sdlc-binary.md; link atlas index if a new map is needed.

- [x] Python `Tickets.admission` / `acquire` / `watch_queue_owner`: real-lock and child-lifetime barriers drive weighted admission and cancellation; Go-writer fixtures guard cross-language state semantics. Preserve existing FIFO/grant/reentrant scheduling while adding atomic gate publication. Observe RED before wiring, then GREEN.
- [x] `Line.tryAdmit` / `enterLandingLine` / `runMerge` / `runDurableMerge` / `maybeConductTrain`: barrier-driven real repository/line integration forces pause at each admission family; active leases cover cleanup and bypass paths. A counted `TestQueueAdmissionNeverScansLegacy` requires zero process/topology calls at 8/16 attempts. Keep existing vouch tasks joined on all exits.
- [x] `runUnderSlot` / `runLocalCheck` / `retainQueueInterruption`: real owned child groups and the stateful fake GitHub store drive cancellation/setup/retry boundaries; failure-store/status assertions reject poisoning, and saved exact-head evidence prevents unnecessary reruns. The distinct interruption family stays outside SuiteAttempt and flakes.
- [x] `watch_queue_owner` / `consumeQueueCancellation`: unmanaged owner-death and surviving-child barriers prove that an uncertain descendant boundary remains unknown after make exits. No shared group signal; only a held exact lease plus matching kernel identity can authorize stopping a single owner. This branch cannot produce complete quiescence from make-PID death alone.
- [x] TL-worker's existing help renderer plus `testdata/rules.tsv`: phrase guards and a durable-checkpoint/ack example reject a separate drained report, premature ack and discarded unfinished edits, while retaining pair#722 as notice formatter owner.
- [x] Document CLI, portable override, unknown/legacy rollout, interruption evidence, no TTL, matching-owner resume and same-head merge recovery in help/README/atlas together. Update the issue Log with pair#722 hand-off and observer limitations.
- [x] Build only in this slot. Run actual binary pause/resume/status inside private queue/lock/fleet paths and supply pair:5 that exact override and executable path. Ask pair:5 to run its adapter + strict decoder against this binary, verifying two owners, lost reply reconciliation, active IDs/kinds, interrupted retirement and resume semantics. Commit observed evidence, not canned JSON. No stage or live pause.
- [ ] Run relevant race tests and package tests after code stabilizes; create PR via sdlc, run one recorded full suite on pushed final code, close M2 and the issue through binary-owned reviews, fix genuine findings, enter `sdlc merge --yes` on SHIP. Report final PR/merge to hub; hub rebuilds and owns fleet rollout.

### Verification commands

Use targeted tests while iterating; production functions and adversarial strategies above own coverage.

```text
go test ./cmd/sdlc/internal/queuegate -count=1
go test ./cmd/sdlc/internal/queuegate -race -count=1
go test ./cmd/sdlc/internal/landingline -run 'Test.*(Pause|Queue|FIFO|Train)' -count=1
python3 scripts/test/test-lock-pause.test.py
python3 scripts/test/test-lock-priority.test.py
go test ./cmd/sdlc -run '^TestQueue' -count=1
go test ./cmd/sdlc -run '^TestQueue' -race -count=1
go test ./cmd/tl-worker ./pkg/golint -count=1
git diff --check
sdlc local-check run --name suite
```

Real helper tests use private locks, deterministic readiness barriers and cleanup/join every child before teardown. Never pause the operator's queue as an acceptance shortcut.

## Revisions

### 2026-10-20 — initial design from frozen contract and hub clarifications

The original Spec predates owner composition and exact ack text. The frozen Contract remains unchanged. Hub explicitly approved the shared gate/lifetime shape, graceful merge interruption at a new-suite boundary, and unknown legacy rollout handling. Implementation approval is requested on this durable plan before change-code.

### 2026-10-20 — implementation approval and admission cost guard

Hub approved the full plan and both review boundaries. Legacy process/topology observation is status-only: add `TestQueueAdmissionNeverScansLegacy` with counted observer adapters, exercise 8 and 16 suite/landing admission attempts, and assert exactly zero process/topology calls. The state-table recoveries remain running/pause, paused/resume named owner, interrupting/inspect until retirement, interrupted/rerun after resume, unknown/repair evidence. Match pair#722's implemented text limits: owner/id/head 256 bytes, reason/workspace/detail 4096 bytes. The private discovery override is `WF_SDLC_QUEUE_FLEET_ROOT`; fixtures scope the process observer to their private registered operations/fixture process set so they never inspect or pause the production fleet.

### 2026-10-20 — PQ-1/PQ-2: explicit unmanaged ownership boundary and function strategies

Verified PQ-1: the existing Python watcher observes make's PID only (`scripts/test-lock.py`, `wait_for_exit` and `acquire`); managed descendant cancellation exists in Go `runUnderSlot`, not in that watcher. Hub approved the simpler honest-unknown boundary rather than a new direct-make supervisor. This paragraph supersedes the earlier broad claim that the existing watcher alone establishes all descendant completion.

`consumeQueueCancellation` is the managed Go owner's intent consumer; it cancels the existing owned context and waits for processgroup cleanup before releasing active leases. The helper's new `watch_queue_owner` consumes unmanaged direct-make intents: it may SIGTERM only that exact single make process after matching a held lease and kernel incarnation. It never signals a shared process group and never claims its children are gone from make exit. Such a run remains visible as unknown/incomplete, with an unresolved evidence record naming the operation and workspace, even after the watch PID disappears. Resume and ordinary dead-lease cleanup cannot erase that uncertainty. The operator must reconcile the remaining work; a confirmed different machine boot is sufficient evidence that no earlier descendants survive. Unavailable boot evidence retains unknown. Unresolved records are bounded at the registry capacity and never aged into false absence; reaching the cap refuses further registration visibly until repair.

Owner-death/child-survival tests pin a real child at a barrier, end the raw make owner, and require unknown while the child still exists; PID reuse and unavailable kernel identity forbid signaling. Managed-run controls prove the actual owned process group has stopped and an interrupted record is retained without failed evidence. The worker rule reinforces recorded sdlc local-check full suites and targeted go test -run; direct full make runs should be rare. A future direct-make supervisor is explicitly deferred unless this visible stuck state becomes a real operational problem.

PQ-2 is addressed by compressing the M1/M2 inventories into named production functions with one adversarial strategy and mechanical guard per seam. State tables remain the approved transition/visible-state contract, not a prose test-case inventory.

### 2026-10-20 09:02 UTC — BR-2 semantic traceability and BR-4 shared JSON boundary

The M1 boundary review correctly found that the concept inventory still described proposed types and guarantees the implementation did not provide. The implemented pure symbols are `Control`, `PauseOwner`, `Transition`, `Operation`, `TransitionOperation`, `Snapshot`, `BuildSnapshot` and `ValidateSnapshot`; phases are validated string values. `Transition` returns the next control and an error, while store operations persist that value and separately request interruptions. Detached exported Go values remain mutable; validation at persistence and transition boundaries enforces the model. The BR-1 companion correction introduces the actual `InterruptionRecord` envelope in `interruption.go`: its diagnostic outcome does not require the embedded operation's unknown lifetime to become completed. This revision corrects the table, prose and checked M1 function names to those actual semantics without changing the approved contract or scope.

BR-4 identified duplicate recursive JSON validation in the legacy observer. The shared exported `queuegate.StrictObject` now owns structural validation for both the typed codec and `queueLegacyFields`, preserving opaque raw member values and rejecting duplicate/nested malformed JSON through one boundary. A direct adversarial helper test was added before implementation. M2 admission integration, runner interruption and logs, worker help and peer executable conformance remain pending.

### 2026-10-20 09:17 UTC — BR-1/BR-3 cancellation recovery and transaction budgets

The exact-ID intent is authoritative before its lease phase catches up. Both the owner and status reconcile that crash window. Retiring or recovering a cancelled owner first retains an immutable `InterruptionRecord`; descendant uncertainty remains independently unresolved. Removing an intent requires retained evidence and either owner retirement or recovery. An orphan intent with matching interruption evidence is collected; missing or malformed proof stays named unknown. Prior-boot reconciliation removes obsolete unresolved activity and its intent without erasing diagnostic history. Enrollment refuses reuse of an ID still present in any evidence family. Interruption retention reserves capacity before publishing, so writer and reader limits agree.

Contexts apply throughout scans, recovery and durable cancellation publication, not only mutex acquisition. A deadline after the hold is written leaves that hold applied and reports incomplete publication; retrying the same owner completes remaining intents. Deterministic cancellation at the first/second record proves loop exit independently of wall time. A representative maximum registry (4,096 rows, two held running suites) completed fast-pause publication in 374 ms with exactly 4,096 reads on this machine; setup is excluded. Timing is evidence, while bounded scans and cancellation checks are the mechanical guard. No supervisor, protocol change or new admission mechanism is introduced.

### 2026-10-20 — M1 SHIP and M2 lifetime wiring checkpoint

M1 round 2 verified all four findings addressed and closed SHIP. M2 uses the same resource: Go enrolls a waiting `queueRuntime`; line admission or Python's weighted admission publishes running under the existing global mutex. Python borrows the managed controller lease via `WF_SDLC_QUEUE_ID`; raw head/slot bindings name its checkout root, not the state directory. An admitted merge sends `WF_SDLC_QUEUE_STOP_IF_PAUSED=1` at a new suite boundary. The watcher reopens the mutex after fork, retaining only the intended lifetime/resource descriptors. Direct make reports an unknown workspace rather than inferring authority from path spelling.

`Lease.Read` observes only its exact lease; `RetireOutcome` returns cancellation's serialized outcome to prevent a passing result after an intent wins. Optional raw interruption logs reuse the existing per-head log retention (latest per check, newest 20 heads); immutable JSON interruption history remains separately bounded at 30 days/4096 records. Neither family is SuiteAttempt evidence. This refines the earlier combined JSON/log retention wording to the actual two existing lifetimes, without changing the frozen queue schema or adding a supervisor.

### 2026-10-20 — BR-5: enforce nonblocking secondary acquisition

M2 round 1 found that both real admission callbacks violated the existing `Lease.TryAdmit` nonblocking-resource contract: `Line.tryAdmit` called a blocking line mutex, and Python's weighted callback blocked on allocation. Any stopped local lock owner could thereby pin the global gate. This is the second finding in the transaction-deadline family; the correction sweeps every nested flock in both admission consumers, rather than extending the timeout.

Global-before-local order stays unchanged. Under the global gate, `Line.withMutexMode(true, ...)` and `Tickets.locked(blocking=False)` try the secondary lock and return no admission on contention; the caller releases the global gate before polling again. Fresh unpublished Python lifetime locks also use nonblocking acquisition, matching Go enrollment. Blocking setup/preview/writer transactions remain outside the global gate. No new lock, state, supervisor or protocol is introduced. Real-flock barriers retain the exact FIFO ticket, prove pause remains responsive while the secondary lock is held, and cancel/join the waiter before that lock is released. The same ticket admits after resume/retry.
