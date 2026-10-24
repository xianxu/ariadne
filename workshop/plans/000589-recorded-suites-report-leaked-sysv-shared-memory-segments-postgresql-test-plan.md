# Recorded suite SysV leak warnings — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Record newly observed unattached SysV segments with creator evidence, without blaming concurrent suites or changing their pass/fail outcome.

**Architecture:** One observation brackets each admitted host command batch, after queue acquisition and before result hashing. Pure parsers and a delta classifier consume bounded platform snapshots and optional owned-group witnesses. The existing hashed log retains warnings; no new attempt schema or resource cleanup authority is introduced.

**Tech Stack:** Go; native Darwin ipcs; Linux /proc/sysvipc/shm; existing process-group runner and saved-log proof; stateful fake observation seam.

## Decisions for TL approval

- Advisory only: preserve the command exit, known-flake classification, pass/fail status, queue semantics and existing cleanup behavior. Never run ipcrm, remove pre-existing segments, signal a creator, or claim to prevent leaks.
- Default wording: **new segments appeared during this run, not attributed**. Only a creator witnessed in the managed command's own process group gets a suite-group witness label and a known process basename. A dead CPID is still named numerically, with command/ownership unknown. PID proximity, same UID, time overlap and process-name guesses are not attribution.
- Use the existing run log as the durable result. Append a recognizable diagnostic footer before hashing, with run/check/head identity, owner segment counts, candidate IDs, CPIDs, attachment counts and witness/unknown reason. A final compact summary stays in the bounded tail and supports a **SysV warning** status-description prefix on both fresh publication and saved-result republish. No new SuiteAttempt fields or versions.
- Read Darwin and Linux host IPC only; VM/runner checks do not imply guest IPC coverage. Unsupported platforms, missing tools, malformed or unreadable snapshots and expired probe budgets produce a visible **observation unavailable** note; they never prevent a test or turn green/red by themselves.
- Deterministic fakes first. Native one-segment qualification waits for the operator to free capacity, as directed by ops:0. The real machine currently has all32 slots occupied; none will be removed by this work.

Alternatives: making leaks a failure would change landing policy and misclassify concurrent observations; automatic cleanup would introduce destructive ownership decisions. A periodic descendant/IPC monitor could capture more short-lived creators, but adds a new lifetime and hot-path scan. Two boundary snapshots plus one optional terminal group witness keep that limit explicit. No global monitor lock is proposed: unmanaged processes and peer runners would not participate, so it cannot establish causal attribution.

## Findings and sources

2026-10-24 read-only native capture:32 segments, each56 bytes and NATTCH0; most creation-time displays00:43:21–00:43:47. ipcs -m -p supplies CPID, and ipcs -m -a supplies NATTCH/size/owner. Preserve that capture as a parser fixture, with no claim that these existing segments belong to this run.

The local macOS ipcs manual provides owner filtering via -u UID and warns that the changing kernel table is not a consistent transaction. Linux exposes UID, CPID and attachment metadata in one [kernel /proc/sysvipc/shm table](https://github.com/torvalds/linux/blob/master/ipc/shm.c); do not reuse Darwin flags on util-linux ipcs, whose meanings differ ([primary source](https://github.com/util-linux/util-linux/blob/master/sys-utils/ipcs.c)). Missing/conflicting observation is not an empty table.

echo#32 already exists, is owned by echo:2, codecomplete/SHIP and not landed at inspection. It owns the harness fix; do not file a duplicate or edit its checkout. Its findings supersede the older claim that mmap eliminates all SysV use: PostgreSQL still allocates a56-byte interlock, and postmaster.pid line7 identifies it. Documentation must require clean shutdown and an owned no-cleanup-path guardian, then verification that that specific segment is gone. Do not teach a machine-wide ipcrm sweep.

## Core concepts

### PURE entities

| Name | Kind | Lives in | Status |
|---|---|---|---|
| Segment / Snapshot / Witness / Candidate / Report | PURE data | `cmd/sdlc/internal/sysvshm/segment.go` | new |
| ParseDarwin / ParseLinux | PURE decoding | `cmd/sdlc/internal/sysvshm/parse.go` | new |
| Delta | PURE classification | `cmd/sdlc/internal/sysvshm/report.go` | new |
| writeSysVReport / boundedSysVText / hasSysVWarning | PURE rendering and extraction | `cmd/sdlc/sysvwarning.go` | new |
| localRunRecord.statusDescription | PURE rendering | `cmd/sdlc/localchecks.go` | modified |
| savedAttemptRecord | PURE validation of supplied attempt, digest and tail | `cmd/sdlc/localcheck_republish.go` | modified |

Segment holds ID, key, owner UID, creator PID, attachment count, size and available generation/change-time evidence. Snapshot distinguishes complete-empty from unavailable/partial. Delta warns only for owner-matching new identities left unattached; a pre-existing segment becoming unattached is not a new allocation. Contradictory observations cannot establish a clean delta. Report separates observed candidates from witnessed creators. Rendering accepts an explicit writer; no observation or cleanup is performed by the pure policy functions (ARCH-DRY).

### Integration points

| Name | Kind | Lives in | Status | Wraps |
|---|---|---|---|---|
| Observer / Native / NewNative | INTEGRATION | `cmd/sdlc/internal/sysvshm/observe.go` | new | bounded Darwin command / Linux proc table reads |
| Native.WitnessGroup / kernelWitness | INTEGRATION | `observe.go`, `witness_darwin.go`, `witness_linux.go`, `witness_other.go` | new | process-group snapshot and kernel identity |
| fakeSysVObserver | INTEGRATION fake | `cmd/sdlc/sysvwarning_test.go` | new | stateful private segment/process table |
| localRunDeps / runUnderResourceBatch / runManagedCommandBatch | INTEGRATION | `cmd/sdlc/localcheck_run.go` | modified | admitted host command lifetime and terminal witness |
| captureSysV / sysVRunReport | INTEGRATION | `cmd/sdlc/sysvwarning.go` | new | caller-derived bounded observation and log writer |
| loadSavedAttempt | INTEGRATION | `cmd/sdlc/localcheck_republish.go` | existing | saved log reads and digest verification before pure replay |

Inject Observer through localRunDeps. Darwin uses a bounded ipcs -m -a -u numericUID command under a fixed C locale. Linux reads its bounded proc table directly and filters effective UID numerically. Other platforms explicitly report unsupported coverage. Both adapters feed the same strict required-column projection, rejecting malformed numbers, repeated IDs/columns and incomplete rows while tolerating unrelated future columns.

Before occurs only after admission and scratch/log setup, immediately before the batch starts. After occurs immediately after runManagedCommandBatch returns, including command failure, timeout, managed cancellation and cleanup-unconfirmed returns, before any log close/hash/publication. Use one small independent observation budget on cancellation so the canceled run context does not silently skip the report; observation failure is itself recorded. A hard-killed sdlc cannot finish observation: its beginning marker is incomplete evidence, never a clean result.

One optional bounded process snapshot is taken before the terminal command's group cleanup, using the group already created by processgroup.Configure. It names only actually witnessed group members. No periodic scans, sibling-workspace inference or global ancestry traversal. Earlier successful commands in a multi-package batch have already been joined; short-lived/escaped/earlier creators without a witness remain unattributed. Validate available PID incarnation evidence and discard changed/uncertain witnesses. If this witness cannot be established cheaply, the honest fallback is numeric CPID plus unknown name/ownership, not a new supervisor.

The stateful fake has allocate/attach/detach/remove, creator exit/reuse, foreign concurrent allocation, probe failure and cleanup events. Native reads and command execution are separate effects; pure delta tests require no mock IO. Actual monitor storage is the existing run log, removed by existing attempt/log retention. In-memory snapshots/witnesses end with the batch; no new background process or durable sidecar exists (ARCH-FUNERAL).

## Visible states and recovery

Actions: inspect the named run/creator; fix the harness; rerun for fresh evidence; ask the operator to restore observation/capacity.

| Internal condition | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Complete before/after, no new unattached owner segments | Observed; no new candidate | None | Short-lived allocations that were removed are outside the leak report |
| New unattached segments, known own-group witness | SysV warning; witnessed creator IDs/names | Fix the harness | Segments may continue consuming capacity |
| New unattached segments, no proven creator relationship | SysV warning; not attributed | Inspect the named creator/run | Causal attribution unavailable; never blame this suite |
| Missing/truncated/bounded-out before or after | Observation unavailable | Restore observation and rerun for fresh evidence | Leak coverage unavailable; test result unchanged |
| Cancellation/timeout with observer completion | Original interrupted/failed outcome plus warning/note | Follow existing rerun recovery | Warning cannot turn interruption into pass |
| sdlc killed before after-snapshot | Incomplete run/observation | Inspect retained evidence | No completed leak report; never infer absence |
| Saved result republished | Original warning/status marker retained | Inspect its original log | No fresh machine observation is implied |

## Envelope and counted cost

This is batch diagnostics, not admission policy. Growth drivers: S segment rows, P process rows for the optional witness, and K commands in a batch. Exactly two IPC captures per started host batch, at most one terminal process capture, and one map-based delta pass: O(S+P+K) work/memory, no per-segment subprocesses or per-event rescans. Count 8/16 segments and1/8 commands to reject extra probes or quadratic comparisons. Monitor bounds:1MiB/4096 segment rows and a2-second budget per boundary; optional process capture has its own2-second bound and1MiB cap. Bounds are conservative operating assumptions, not permission to treat overflow as an empty table. No extra CPU-heavy suite slot or VM lease is acquired.

## Implementation and function-level adversarial strategies

One coherent close/review boundary, no artificial milestone tags. Use @superpowers-test-driven-development after plan approval and change-code.

### 1. Pure parsers, delta and warning representation

Files: create the internal/sysvshm files above with colocated *_test.go and testdata native captures; create sysvwarning.go / sysvwarning_test.go.

- [x] ParseDarwin/ParseLinux: write red tables for empty/populated native formats,32-row incident, unrelated columns, duplicate IDs/headers, wrong types, truncated output, owner filtering, over-budget input and changing/replaced IDs. Compare projected native fixture fields to independent expected records. Retain primary-source Linux table provenance.
- [x] Delta: red sequence tests cover pre-existing unattached segments, new attached vs unattached, clean removal, foreign concurrent allocation, owner mismatch, dead/reused creator, invalid before/after and mixed attributed/unattributed candidates. Independent invariant: no pre-existing or attached segment is named as a new leak; no foreign/unproven creator is blamed. Count linear work at8/16 rows.
- [x] writeSysVReport/hasSysVWarning: red rendering/replay tests verify IDs/CPIDs, explicit not-attributed language, unknown names, owner counts, final summary, bounded basename text and unchanged exit/outcome. Diagnostic bytes must not become Go failures or typed flake proof.
- [x] Run `go test ./cmd/sdlc/internal/sysvshm ./cmd/sdlc -run 'SysV|Sysv' -count=1`, observe intended failures, implement the pure functions, rerun green and commit.

### 2. Platform readers and one managed lifetime

Files: platform Observer files; localcheck_run.go; localcheck_republish.go; localchecks.go; existing localcheck_run_test.go / localcheck_republish_test.go and new sysvwarning integration tests.

- [x] Write a stateful fake-backed managed-run regression that allocates then abandons a segment; assert warning in stderr and the saved hashed log, and a warning marker in fresh/saved status descriptions while pass/fail remains unchanged. Foreign allocation must report not-attributed, not this-run ownership.
- [x] runUnderResourceBatch: inject before/after failures, cancellation, timeout, command-start failure, early package failure, cleanup-unconfirmed and log publication failure. Verify after is attempted before log hash on every reachable return after begin; no probe runs during queue waiting, proof-only republish, or VM/runner checks. Existing interrupted/environment evidence still decides whether any completed record exists.
- [x] runManagedCommandBatch: barrier-controlled final cleanup proves witness capture precedes group removal, is at most once for the batch, and unavailable/changed PID proof stays unattributed. Test multiple commands and early terminal errors without adding a background sampler.
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
