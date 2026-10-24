# Recorded suite SysV leak warnings — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Record newly observed unattached SysV segments with creator evidence, without blaming concurrent suites or changing their pass/fail outcome.

**Architecture:** One observation brackets each admitted host command batch, after queue acquisition and before result hashing. Pure parsers and a delta classifier consume bounded platform snapshots without claiming creator causality. The existing hashed log retains warnings; no new attempt schema or resource cleanup authority is introduced.

**Tech Stack:** Go; native Darwin ipcs; Linux /proc/sysvipc/shm; existing process-group runner and saved-log proof; stateful fake observation seam.

## Approved decisions

- Advisory only: preserve the command exit, known-flake classification, pass/fail status, queue semantics and existing cleanup behavior. Never run ipcrm, remove pre-existing segments, signal a creator, or claim to prevent leaks.
- Default wording: **new segments appeared during this run, not attributed**. Kernel creator PIDs are always numeric and unattributed, with command/relationship unknown; a current PID incarnation does not prove allocation-time identity. PID proximity, same UID, time overlap and process-name guesses are not attribution.
- Use the existing run log as the durable result. Append a recognizable diagnostic footer before hashing, with run/check/head identity, owner segment counts, candidate IDs, CPIDs, attachment counts and explicit unknown relationship. A final compact summary stays in the bounded tail and supports a **SysV warning** status-description prefix on both fresh publication and saved-result republish. No new SuiteAttempt fields or versions.
- Read Darwin and Linux host IPC only; VM/runner checks do not imply guest IPC coverage. Unsupported platforms, missing tools, malformed or unreadable snapshots and expired probe budgets produce a visible **observation unavailable** note; they never prevent a test or turn green/red by themselves.
- Deterministic fakes first. Native one-segment qualification waits for the operator to free capacity, as directed by ops:0. The initial capture had all32 slots occupied; after explicit clearance the single owned-segment qualification passed, as recorded in the issue Log.

Alternatives: making leaks a failure would change landing policy and misclassify concurrent observations; automatic cleanup would introduce destructive ownership decisions. A periodic descendant/IPC monitor could capture more short-lived creators, but adds a new lifetime and hot-path scan. Two boundary snapshots keep that limit explicit. No global monitor lock is proposed: unmanaged processes and peer runners would not participate, so it cannot establish causal attribution.

## Findings and sources

2026-10-24 read-only native capture:32 segments, each56 bytes and NATTCH0; most creation-time displays00:43:21–00:43:47. ipcs -m -p supplies CPID, and ipcs -m -a supplies NATTCH/size/owner. Preserve that capture as a parser fixture, with no claim that these existing segments belong to this run.

The local macOS ipcs manual provides owner filtering via -u UID and warns that the changing kernel table is not a consistent transaction. Linux exposes UID, CPID and attachment metadata in one [kernel /proc/sysvipc/shm table](https://github.com/torvalds/linux/blob/master/ipc/shm.c); do not reuse Darwin flags on util-linux ipcs, whose meanings differ ([primary source](https://github.com/util-linux/util-linux/blob/master/sys-utils/ipcs.c)). Missing/conflicting observation is not an empty table.

echo#32 already exists, is owned by echo:2, codecomplete/SHIP and not landed at inspection. It owns the harness fix; do not file a duplicate or edit its checkout. Its findings supersede the older claim that mmap eliminates all SysV use: PostgreSQL still allocates a56-byte interlock, and postmaster.pid line7 identifies it. Documentation must require clean shutdown and an owned no-cleanup-path guardian, then verification that that specific segment is gone. Do not teach a machine-wide ipcrm sweep.

## Core concepts

### PURE entities

| Name | Kind | Lives in | Status |
|---|---|---|---|
| Segment / Snapshot / Candidate / Report | PURE data | `cmd/sdlc/internal/sysvshm/segment.go` | new |
| ParseDarwin / ParseLinux | PURE decoding | `cmd/sdlc/internal/sysvshm/parse.go` | new |
| Delta | PURE classification | `cmd/sdlc/internal/sysvshm/report.go` | new |
| writeSysVReport / boundedSysVText / hasSysVWarning | PURE rendering and extraction | `cmd/sdlc/sysvwarning.go` | new |
| localRunRecord / statusDescription | PURE diagnostic projection/rendering | `cmd/sdlc/localchecks.go` | modified |
| resolvedSavedRecord | PURE verified parent/child warning union | `cmd/sdlc/localcheck_resume_saved.go` | modified |
| savedAttemptRecord | PURE validation of supplied attempt, digest and tail | `cmd/sdlc/localcheck_republish.go` | existing, reused |

Segment holds ID, key, owner UID, creator PID, attachment count, size and mutable change-time metadata. Snapshot distinguishes complete-empty from unavailable/partial. Delta warns only for owner-matching newly observed numeric IDs left unattached; a pre-existing segment becoming unattached is not a new allocation. Contradictory observations cannot establish a clean delta. Report names numeric creator PIDs without attributing them to a process. Same-ID metadata changes and replacement guesses do not produce a candidate. Rendering accepts an explicit writer; no observation or cleanup is performed by the pure policy functions (ARCH-DRY).

### Integration points

| Name | Kind | Lives in | Status | Wraps |
|---|---|---|---|---|
| Observer / Native / NewNative | INTEGRATION | `cmd/sdlc/internal/sysvshm/observe.go` | new | bounded Darwin command / Linux proc table reads |
| fakeSysVObserver | INTEGRATION fake | `cmd/sdlc/sysvwarning_test.go` | new | stateful private segment table |
| localRunDeps / runUnderResourceBatch | INTEGRATION | `cmd/sdlc/localcheck_run.go` | modified | admitted host command lifetime and execution timeout factory |
| captureSysV / sysVRunReport | INTEGRATION | `cmd/sdlc/sysvwarning.go` | new | caller-derived bounded observation and log writer |
| loadSavedAttempt | INTEGRATION | `cmd/sdlc/localcheck_republish.go` | existing | saved log reads and digest verification before pure replay |

Inject Observer through localRunDeps. Darwin uses a bounded ipcs -m -a -u numericUID command under a fixed C locale. Linux reads its bounded proc table directly and filters effective UID numerically. Other platforms explicitly report unsupported coverage. Both adapters feed the same strict required-column projection, rejecting malformed numbers, repeated IDs/columns and incomplete rows while tolerating unrelated future columns.

Before occurs only after admission and scratch/log setup, immediately before the batch starts. After occurs immediately after runManagedCommandBatch returns, including command failure, timeout, managed cancellation and cleanup-unconfirmed returns, before any log close/hash/publication. Use one small independent observation budget on cancellation so the canceled run context does not silently skip the report; observation failure is itself recorded. A hard-killed sdlc cannot finish observation: its beginning marker is incomplete evidence, never a clean result.

No process snapshot or creator-incarnation lookup runs. A before/after table cannot link a reused numeric PID to the allocating process. Unknown attribution is the deliberate coverage boundary, rather than a new supervisor.

The stateful fake has allocate/attach/detach/remove, creator exit/reuse, foreign concurrent allocation, probe failure and cleanup events. Native reads and command execution are separate effects; pure delta tests require no mock IO. Actual monitor storage is the existing run log, removed by existing attempt/log retention. In-memory snapshots end with the batch; no new background process or durable sidecar exists (ARCH-FUNERAL).

## Visible states and recovery

Actions: inspect the named run/creator; fix the harness; rerun for fresh evidence; ask the operator to restore observation/capacity.

| Internal condition | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Complete before/after, no new unattached owner segments | Observed; no new candidate | None | Short-lived allocations that were removed are outside the leak report |
| New unattached segments, no proven creator relationship | SysV warning; not attributed | Inspect the named creator/run | Causal attribution unavailable; never blame this suite |
| Missing/truncated/bounded-out before or after | Observation unavailable | Restore observation and rerun for fresh evidence | Leak coverage unavailable; test result unchanged |
| Cancellation/timeout with observer completion | Original interrupted/failed outcome plus warning/note | Follow existing rerun recovery | Warning cannot turn interruption into pass |
| sdlc killed before after-snapshot | Incomplete run/observation | Inspect retained evidence | No completed leak report; never infer absence |
| Saved result republished | Original warning/status marker retained | Inspect its original log | No fresh machine observation is implied |

## Envelope and counted cost

This is batch diagnostics, not admission policy. Growth drivers: S segment rows and K commands in a batch. Exactly two IPC captures per started host batch, one map-based delta pass: O(S+K) work/memory, no per-segment subprocesses or per-event rescans. Count 8/16 segments and1/8 commands to reject extra probes or quadratic comparisons. Monitor bounds:1MiB/4096 segment rows and a2-second budget per boundary. Bounds are conservative operating assumptions, not permission to treat overflow as an empty table. No extra CPU-heavy suite slot or VM lease is acquired.

## Implementation and function-level adversarial strategies

One coherent close/review boundary, no artificial milestone tags. Use @superpowers-test-driven-development after plan approval and change-code.

### 1. Pure parsers, delta and warning representation

Files: create the internal/sysvshm files above with colocated *_test.go and testdata native captures; create sysvwarning.go / sysvwarning_test.go.

- [x] ParseDarwin/ParseLinux: write red tables for empty/populated native formats,32-row incident, unrelated columns, duplicate IDs/headers, wrong types, truncated output, owner filtering, over-budget input and changing/replaced IDs. Compare projected native fixture fields to independent expected records. Retain primary-source Linux table provenance.
- [x] Delta: red sequence tests cover pre-existing unattached segments, new attached vs unattached, clean removal, foreign concurrent allocation, owner mismatch, numeric creator PID reuse, invalid before/after and all candidates explicitly unattributed. Independent invariant: no pre-existing or attached segment is named as a new leak; no foreign/unproven creator is blamed. Count linear work at8/16 rows.
- [x] writeSysVReport/hasSysVWarning: red rendering/replay tests verify IDs/CPIDs, explicit not-attributed language, unknown names, owner counts, final summary, bounded diagnostic text and unchanged exit/outcome. Diagnostic bytes must not become Go failures or typed flake proof.
- [x] Run `go test ./cmd/sdlc/internal/sysvshm ./cmd/sdlc -run 'SysV|Sysv' -count=1`, observe intended failures, implement the pure functions, rerun green and commit.

### 2. Platform readers and one managed lifetime

Files: observe.go; localcheck_run.go; localcheck_resume_saved.go; localchecks.go; existing localcheck_run_test.go / localcheck_republish_test.go and new sysvwarning integration tests.

- [x] Write a stateful fake-backed managed-run regression that allocates then abandons a segment; assert warning in stderr and the saved hashed log, and a warning marker in fresh/saved status descriptions while pass/fail remains unchanged. Foreign allocation must report not-attributed, not this-run ownership.
- [x] runUnderResourceBatch: inject before/after failures, cancellation, timeout, command-start failure, early package failure, cleanup-unconfirmed and log publication failure. Verify after is attempted before log hash on every reachable return after begin; no probe runs during queue waiting, proof-only republish, or VM/runner checks. Existing interrupted/environment evidence still decides whether any completed record exists.
- [x] runUnderResourceBatch: the controlled execution-context factory proves the before-probe precedes timer creation and an after-probe advancing the completed timer cannot change success; command timeout controls still fail. Existing command-group cleanup is reused without observation hooks.
- [x] savedAttemptRecord/statusDescription: mutate footer/log bytes and require existing digest refusal; valid replay preserves the SysV warning marker. Existing old logs without a footer remain readable and imply no new coverage. No SuiteAttempt schema change, unknown-field rewrite or new gate outcome.
- [x] Native adapters use fake command/file backends for deterministic Darwin/Linux tests. Native allocation test is opt-in, uses unix.SysvShmGet(IPC_PRIVATE) in a child and removes only that exact test-owned ID through unix.SysvShmCtl(IPC_RMID) in parent cleanup. Default suites allocate no real IPC; global capacity exhaustion must never be treated as a product regression or trigger machine cleanup.
- [x] Run focused/race/native-process controls and the tracked Go catalog; commit implementation and evidence.

### 3. Harness guidance and qualification

Files: update `atlas/workflow/sdlc-binary.md` and its existing index link; add only necessary local-check help text in `cmd/sdlc/helptext/local-check.md` (confirmed embedded help source).

- [x] Document observation scope/limits, warning-only semantics, creator uncertainty and log/status location. Explain PostgreSQL's persistent interlock even with mmap; require owned graceful shutdown on all paths, guardian for killed test owners, specific-segment verification and bounded escalation. Link existing echo#32; notify echo:2 of the shared guidance, with no peer edits or duplicate filing.
- [x] After the operator clears capacity, perform the opt-in single owned-segment qualification and record its exact output. No staging or live suite mutation; ask the hub to approve private executable qualification plus post-land primary rebuild as rollout. Do not claim native allocation qualification while capacity is unavailable.
- [x] Push and complete the recorded `sdlc local-check run --pre-close` (passed, as recorded in the issue Log). Mandatory close review and central landing remain pending; report the landing to the hub. #590 has landed; #564 is now approved and its revised plan is committed separately.

## Revisions

- 2026-10-24: Initial design, incorporating ops:0's approved attribution boundary: unproven concurrent deltas are not attributed. Fakes precede native qualification, which waits for operator-cleared slots. Await TL plan approval before change-code.

- 2026-10-24: Ops:0 approved the full plan (delivery 38e93c9b-c77c-42ff-9f67-fedbfc22b3af): boundary snapshots, group-witness-only attribution, hashed-log warning and status marker, no schema/outcome/cleanup changes, fakes before operator-cleared native qualification. The retained change-code completed successfully (inferred quick flow; close upgrades review if implementation exceeds its shell).

- 2026-10-24 — recorded suite context guard: derive the independently bounded snapshot/witness observation contexts from WithoutCancel(caller), preserving caller values instead of minting Background. Added a stateful normal/canceled-run regression observed red before the fix. The existing context guard remains strict; no policy, timeout or schema change.

- 2026-10-24 — native qualification cleared and passed: ops:0 receipt b2809aee-c5d4-4c49-b816-91dcddc8f5c4 relayed operator-cleared capacity. One owned child segment was observed and removed by exact-ID verified cleanup; independent before/after ipcs tables were empty. Updated completed implementation/verification checkboxes and aligned the concept tables with actual functions and explicit PURE/INTEGRATION kinds. Close and landing remain pending behind the retained #594 merge; no runtime change accompanies this bookkeeping.

- 2026-10-24 — close preparation: canonicalized the durable plan filename and ticked the issue verification step after the recorded suite and authorized native pass. Mandatory close and landing remain pending.

### 2026-10-24 — boundary review re-plan, awaiting hub approval

Round1 returned REWORK with BR-1 through BR-6. Inspection confirms the behavior gaps: the timeout context starts before the before-snapshot and is read after the after-snapshot; Delta uses mutable CTIME and accepts any nonempty start string for a numeric CPID match; resumed status uses only the parent footer despite already verifying the child log. savedAttemptRecord is existing, reused code, and README needs the usage pointer. No close finalized. Canceled only the still-queued obsolete attempt a86c82b2c9bafba095f03084c59fd64b through local-check cancel; its runner exited before admission. No capacity or peer process changed.

This crosses three mechanisms, so stop and bring one revised design instead of incremental patches:

1. **Outcome isolation (BR-1):** move execution timeout creation after the before-snapshot; freeze execution deadline state and duration immediately after command/group completion, cancel that timer, then perform the independently bounded after-snapshot. No diagnostic work determines timeout/lane-overrun. The command's existing cancellation and cleanup policy remains authoritative. Remove the optional terminal witness hook under choice2, so no observation sits inside the execution batch. Regression strategies: controlled observers delay before longer than the execution timeout while the eventual command succeeds, and gate the after-snapshot until the completed command's old timer expires; both must retain success, plus genuine command timeout still fails. Tests use explicit barriers and generous enclosing deadlines, no CPU stress or increased production timeout.
2. **Conservative creator reporting (BR-2):** remove process-group creator attribution and its WitnessGroup/native ps/kernel-witness machinery. Report the kernel's numeric CPID with process name and relationship unknown; every candidate remains not attributed. A current matching PID/start cannot prove it was the allocating incarnation. The simplest alternative is to keep a witnessed-current-process label, but users would still read it as creator attribution; proving allocation-time incarnation would need a monitor or new generation evidence disproportionate to an advisory. Pure PID-reuse fixture must remain unattributed even when a same-number current group process exists. Native owned-child qualification still proves only the fixture's exact allocation/cleanup; it does not grant production attribution. This deliberately reduces optional attribution coverage.
3. **Stable observed identity (BR-3):** compare numeric segment IDs present in the before table, without treating CTIME, permissions/UID, attachments or size mutations as new allocation. Any same-ID change remains outside the new-ID candidate set; same-ID replacement cannot be proved by these boundary tables and is an explicit coverage limit. Word reports as newly observed segments, never a proof of allocation causality. Tests mutate CTIME and all mutable metadata on existing IDs and keep them out; genuinely absent-before owner IDs remain candidates, foreign owners and attached rows remain excluded. Count N/2N map work as before.
4. **Resume warning union (BR-5):** reuse the existing parent/child digest validation in resolveResumeProof and the single resolvedSavedRecord path used by fresh and saved publication. After proof succeeds, derive an in-memory warning bit from the verified child footer and combine it with the parent warning in statusDescription. No persisted record field/schema or child-log rewrite. Tests exercise clean parent plus child-only warning, parent-only warning, neither, child digest mutation refusal, and fresh/republication descriptions through the same saved proof seam.
5. **Documentation (BR-4/6):** mark savedAttemptRecord existing/reused (corrected above); add README warning/qualification pointer and explain honest creator/identity limits in help/atlas/plan. No wording-only tests; inspect source referents and run existing meaningful controls. Update lessons for allocation metadata vs identity and observational deadlines.

All new regressions must fail before the corresponding fix. Run one targeted command at a time, count1, then one newly recorded pre-close through the capacity queue and the next mandatory close review. No new native allocation is necessary: the exact owned fixture path is unchanged; rerun only if that path changes and authorization/capacity still hold. No schema, pass/fail policy, IPC cleanup authority or supervisor is introduced. Wait for the hub's approval before changing these mechanisms.

- 2026-10-24 — consolidated re-plan APPROVED in verified submitted ops:0 receiptdcf4a2bc-390c-41a0-b00b-4f93dcbcca96: separate diagnostic budgets and freeze execution outcome, numeric creator PID always unattributed with witness machinery removed, conservative ID comparison, verified resume-warning union, and documentation corrections. No schema, policy or cleanup change. Proceed with regression-first implementation.

- 2026-10-24 — implemented approved review correction: removed witness IO and types; numeric-ID-only delta and always-unattributed CPID rendering; timer starts after the before-probe and execution outcome is frozen before after-probe. A defaulted execution-context factory in the existing IO dependency provides a controlled-clock regression, avoiding subsecond startup assumptions under fleet load. Parent/child warning union uses the existing verified proof on fresh and saved publication. README/help/atlas and current concept table now describe this reduced mechanism; savedAttemptRecord remains existing/reused. Historical Revisions above retain the superseded design trail.

### 2026-10-24 — round2 outcome-family re-plan, awaiting hub approval

Round2 independently reversed the fixes and disposed BR-1/3/4/5/6, but retained BR-2 for regression sensitivity and added BR-7 in the advisory-outcome-isolation family. A timer snapshot alone is not a complete outcome boundary: caller/signal/queue cancellation is still read after diagnostic IO. This is the third delta on that mechanism, so stop code changes and bring one final class-wide rule.

**Rule/resource:** the existing managed command/group lifetime and queue lease establish execution completion. After the command group is joined with confirmed cleanup, freeze command deadline state, finish the existing cancellation-consumer/signal lifecycle, and retire the queue lease through its existing RetireOutcome under control.lock *before* the after-probe. RetireOutcome is the one serialized cancellation-versus-completion authority: a durable request before retirement yields interrupted; a request after retirement cannot target the ended operation. Release suite capacity and its run lease at this same existing completion boundary. Then collect the bounded after-probe and hash/publish the already decided outcome. No new lock, state, schema or cancellation protocol. Cleanup-unconfirmed remains unknown/retained; no retirement or success is invented for it.

The simpler alternative—freeze only ownedCtx.Err but leave the durable queue lease live—fails because a late queue cancellation can still change RetireOutcome after the diagnostic probe. Merely moving caller checks enumerates symptoms and leaves signal/queue races. Retiring once under the existing lifetime lock closes all three inputs together (ARCH-ORDER/ARCH-FUNERAL). Caller/signal consumers are joined before retirement as their existing contract requires; RetireOutcome still sees a request arriving between that join and its lock transaction. Commands canceled before or during execution remain interrupted. Late cancellation during diagnostic-only work cannot revoke a completed test.

Regressions: explicit before/during/after barriers cross success and exit7 outcomes; controlled caller cancellation and durable private queue requests on each side; the existing queuegate retirement race test proves serialized request ordering. The after-probe sees completed operation absent and captures cancellation attempts without changing the frozen result. A cleanup-unconfirmed control retains unknown state. No sleeps, stress loops or live queue mutation.

BR-2 regression: populate the former optional Witness field via a reflection-only test helper when that field exists, then assert numeric CPID/unknown relationship rendering. Current simplified types omit the field; restoring the former Witness-bearing Candidate and unsafe conditional renderer via a private Go overlay must make the same committed test fail. This directly addresses the reviewer's restoration counterexample without resurrecting a production witness API or adding a persisted format. Document the measured overlay red and normal green.

After approval, fix this class once, run targeted serial count1 controls and catalog, refresh private executable proof, then one fresh recorded pre-close and close round3. Canceled only the obsolete queued attempt600eccc730443e406ad96fb76da71c81 (before admission); no peer work changed.

- 2026-10-24 — completion-boundary re-plan APPROVED by verified submitted ops:0 receipt342f86e2-5c7d-4748-abb7-349fefe1fc3b. Join cancellation readers and retire through existing RetireOutcome before after-probe; all pre-retirement requests remain interrupted, unknown cleanup stays retained. BR-2 must fail against the unsafe renderer. Integrate #599 from main after it lands and verify composition. A fourth design delta cuts scope to a separate post-publish warning step; no further expansion of outcome machinery.
