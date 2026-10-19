---
collaboration-mode: discussion
---

# Suite queue grant priority implementation plan

> For agentic workers: consult AGENTS.md §3 for the execution approach. Implementation starts only after TL approval and the sdlc change-code gate. This is one atomic delivery with one close boundary; Task headings are not milestone tags.

**Goal:** A suite for a currently granted issue gets the next sufficient free machine capacity before waiting normal suites, without preempting a running suite; each class remains FIFO.

**Architecture:** Retain the existing kernel-held suite slots and live ticket leases. Reelect the waiting head at each admission attempt under the existing allocation mutex. Obtain active grant evidence through one read-only, versioned Go helper protocol that reuses landingline.ValidGrants; cache its results in memory by grant-log identity and recheck expiry at admission. Do not add another durable grant store, priority grant verb, daemon, or slot pool.

**Tech stack:** Python/flock lock helper; existing Go sdlc and landingline package; real-helper subprocess tests using private lock paths and fixture grant logs.

## Core concepts

| Entity | Kind | Location | Status |
| --- | --- | --- | --- |
| Suite subject: repository common directory plus issue ID | PURE value | cmd/sdlc/suitepriority.go (new) and optional ticket fields in scripts/test-lock.py | new |
| ValidGrants | PURE policy, reused unchanged | cmd/sdlc/internal/landingline/order.go | existing |
| Effective suite order: stable priority/normal partition of numeric ticket order | PURE | scripts/test-lock.py | new |
| Grant snapshot: active requested issue IDs and deadlines, keyed by repository | PURE value | cmd/sdlc/suitepriority.go | new |
| Read-only priority protocol | INTEGRATION | cmd/sdlc/suitepriority.go; cmd/sdlc/landingcmd.go | new |
| Grant-log reader | INTEGRATION, existing JSONL parsing reused | cmd/sdlc/internal/landingline/log.go | modified to expose read-only access without creating a landing line |
| Machine queue election and weighted slot scan | INTEGRATION | scripts/test-lock.py | modified |
| Recorded-check subject and evaluator binding | INTEGRATION | cmd/sdlc/localcheck_run.go | modified |
| Helper capability preflight | INTEGRATION | cmd/sdlc/helpercapability.go | reused unchanged |
| Direct make test binding | INTEGRATION | Makefile.workflow | modified |

Each ticket references one repository and issue. Multiple tickets may reference the same issue; their machine ticket numbers still order them. The grant log remains the sole authority. No boolean supplied by a worker confers priority. Existing cooperative trust applies: this prevents mistakes, not a same-user process forging source records.

The Python predicate sees only validated deadlines, not a second implementation of grantor/revocation/max-duration rules (ARCH-DRY). Its expiry check prevents a cached snapshot outliving a grant. The new values and ordering can be tested without processes; actual admission uses real flock files and processes (ARCH-PURE).

## Required behavior

1. Keep numeric ticket allocation separate from effective order. Partition already-numbered live tickets into active-grant and normal classes, preserving ticket order within both.
2. Under .tickets/alloc, inspect live tickets, obtain their current grant evidence, elect the effective first ticket and attempt the existing .queue plus all-or-none weighted scan. Release .queue after an unsuccessful scan; it must not become a permanent reservation for a waiting normal head.
3. A successful scan hands slot descriptors to the existing watcher. Once admitted, a suite is outside waiting order and nothing can preempt it. Slot ownership remains one kernel resource for the owner's lifetime.
4. A priority weight-2 head with only one free slot waits without holding that slot. Normal light suites cannot backfill ahead of it. This preserves both weight and FIFO; next free capacity means enough capacity for the head's weight.
5. Grant changes are effective at the next admission attempt. Expiry removes priority even if the grant log did not change. Revocation/renewal/late grants invalidate cached snapshots through the log's file identity/mtime/size. An expired ticket stays queued as a normal ticket at its original numeric position.
6. Preserve reentry before queueing and during admission, cancellation, timeout, watched-owner exit, stopped-waiter FIFO and stale-ticket pruning. No second lock implementation or scan-by-process absence proof is introduced (ARCH-PURPOSE).

### Subject identity and grant evidence

- Bind recorded suite identity before creating/entering the detached run's machine ticket. Use the original checkout's Git common directory and issue branch, never the temporary sdlc-local directory name.
- Bind an issue only when that checkout's branch identifies the issue and its head is the tested commit. A main diagnostic (SkipRecord), non-suite local check, detached/unknown subject, or a different tested revision gets issue 0 and stays normal. Never infer the issue from a PR number or from another ticket.
- Direct make test resolves the same subject once through the provider from its real checkout. Recorded runs pass their explicit subject so the helper never guesses from the temporary worktree.
- Add a hidden landing suite-priority protocol to sdlc. One describe request resolves a direct caller's subject. One batch request accepts the unique repository/issue pairs in a queue scan and reads the existing grant logs; it returns a version, active IDs, expiry deadlines and a Go-derived refresh_at boundary at the earliest future Granted/Expires transition in the requested logs. Reuse the existing grant JSONL reader and ValidGrants. It must not acquire a landing turn, fetch, create directories, or write logs.
- The recorded runner supplies its own current sdlc executable as the evaluator. Makefile.workflow supplies the built sdlc beside the supplying workflow/helper source. Do not search arbitrary PATH entries or execute shell command strings from ticket files. A ticket carries identity data, never an executable path.
- A helper process uses its trusted evaluator for every ticket, with argv/stdin JSON, bounded output and timeout, and reaps it before returning/cancelling. No evaluator survives admission.
- Cache only in memory, per grant-log identity (path, inode, mtime_ns, size), for the life of that waiting helper. A changed source is reread; reaching the Go-derived refresh_at boundary invalidates the snapshot even with unchanged source metadata. Expiry is also checked against current UTC time on every admission attempt. This discovers activation of a previously future-dated grant without duplicating policy in Python. Missing/unreadable/invalid evidence confers no priority and produces a clear warning. Never cache a transient read error indefinitely.

### Two rollout decisions for TL approval

- **Unavailable evaluator:** propose preserving capacity/FIFO with an explicit priority-unavailable warning, rather than letting a missing or old sdlc break ordinary make test. Recorded current sdlc runs supply their own capable executable. All priority smoke tests require a capable provider. If the TL prefers a hard refusal for known issue suites, settle that before code.
- **Older sticky helper:** a pre-upgrade helper already holding .queue cannot safely be overtaken. Keep its capacity correctness and let it drain; never unlock or kill another waiter. The new priority guarantee applies once participating waiting helpers have upgraded. Stage/rollout must check for and drain old waiters; document this compatibility boundary rather than claiming preemption across versions.

## Visible states and recovery

Actions: start suite; wait; grant/renew/revoke through existing sdlc landing verbs; cancel own local check; retry unchanged; refresh/build supplying tools.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| No valid grant, legacy ticket, or issue 0 | Normal FIFO position; next waiter named | Wait | None |
| Active grant while slots occupied | Fleet-blocking, issue/deadline, effective position, running holders | Wait | None |
| Earlier priority ticket or insufficient weighted capacity | Names the priority waiter/capacity ahead | Wait | None |
| Grant expires or is revoked while waiting | Normal FIFO at original sequence; reason visible | Renew through existing grant verb if still warranted | Priority only; ticket/work retained |
| Grant appears or renews while waiting | Updated effective priority position | Wait | None |
| Grant/provider evidence unreadable or incompatible | Priority unavailable; normal admission remains bounded | Refresh/build supplying tools; retry if necessary | Priority until evidence is readable; never suite capacity |
| Own cancel, timeout, helper death or watched owner death | Existing cancellation/timeout; ticket no longer blocks | Retry unchanged | This unstarted attempt's queue position |
| Running suite | Existing slot-holder record | Wait for completion or cancel own run | Cancellation can lose partial test work; no preemption added |
| Old helper owns sticky queue token | Names legacy queue holder; priority upgrade cannot bypass it | Let old owner drain, then refresh its supplier | Priority latency during mixed-version transition |

Covered: grant validity changes, priority/normal FIFO, all-or-none weights, reentry, cancellation, process death and mixed-version capacity safety. Not covered: changing the machine suite cap while runs are live, protection against malicious same-user file edits, or preempting an already running suite. These are outside the dispatched contract.

## Cost and lifetime

Growth drivers are T live tickets, R distinct repository grant logs, and G changed-log entries. One admission pass scans each ticket once for records, stats each distinct grant source once, and partitions in O(T), in addition to the existing numeric ordering. Batch policy evaluation occurs only on initial/cache-miss source evidence; unchanged polling starts no evaluator/Git processes. Changed logs cost O(G) to fold with the existing validator. No network or workspace-wide scan is permitted.

Counted N/2N tests assert one metadata read per ticket, at most one grant read per distinct changed source, and one provider batch per cache miss; repeat an unchanged pass and assert zero provider calls. A clock-only expiry must alter order without a reread. Counting, not wall-clock speed, is the growth oracle. Admission callbacks have a bounded process/output budget using existing helper-query conventions; record those exact limits in the implementation and test timeout/oversized response behavior.

Tickets, allocation/queue flocks and slot watchers retain existing cleanup. Optional subject fields add fixed-size data to existing ticket files; no new durable family is created. Grant snapshots/caches die with the helper. The helper protocol has no background server and no on-disk derived priority index (ARCH-FUNERAL).

## Task 1 — Pin behavior red through the real helper

Files: scripts/test/test-lock.test.sh; scripts/test/test-lock-priority.test.py (new); cmd/sdlc/suitepriority_test.go (new); cmd/sdlc/localcheck_reentry_test.go as needed.

Unit-test surfaces and adversarial strategy:

| Production function | Input class and mechanical guard |
| --- | --- |
| suiteSubject(root, testedSHA) | Detached/unrelated branch/revision identities → real temporary Git repository; only the tested issue head binds. |
| suiteGrantSnapshot(subjects, now) | Grant lifecycle and malformed log records → real JSONL reader plus ValidGrants and controllable clock, no directory creation; unchanged-source activation and expiry must change authority. |
| query_priority(provider, request) | Absent/old/slow/oversized/malformed protocol → bounded/reaped subprocess, strict version/schema validation and visible normal-FIFO fallback. |
| GrantEvidence.deadlines(records, now) | Changing sources and clock-only validity transitions → injected clock/stat/query counters; reread on refresh_at as well as source change, no provider calls on stable pre-boundary polls. |
| effective_order(records, deadlines, now) | Arbitrary arrival IDs and grant deadlines → stable partition property, original numeric FIFO within each class, expired evidence cannot confer priority. |
| Tickets.admission(fds, queue, want, evidence) | Adversarial interleaving of live waiters and weighted capacity → real helper subprocesses and private real flocks, diagnostic/ticket barriers and bounded waits, owner survival and no partial holds. |

The real-helper integration harness tests the dispatched next-free-capacity rule against an older normal head, plus FIFO, lifecycle and weight properties under controlled process schedules. In particular an expired priority ticket returns to its original normal FIFO position and the older normal waiter proceeds. Exercise actual production Go grant validation over fixture logs: a test-executable launcher may enter the production protocol handler, never fake its priority decision. Existing reentry, cancellation/death and mixed-helper controls remain the lifetime regression boundary.

- [x] Write and run the next-free-slot regression red through the real helper; preserve its expected failure.
- [x] Add the named unit strategies as their functions are introduced, with bounded synchronization helpers that explicitly fail on missed deadlines.

## Task 2 — Add one grant-evidence boundary and reelected admission

Files: cmd/sdlc/suitepriority.go (new), cmd/sdlc/landingcmd.go, cmd/sdlc/internal/landingline/log.go, scripts/test-lock.py, Makefile.workflow, cmd/sdlc/localcheck_run.go and focused tests; cmd/sdlc/helpercapability.go is reused unchanged.

- [x] Implement the versioned read-only subject/batch protocol and the named Go unit strategies. Pass explicit tested-suite subject and trusted evaluator from recorded runs; direct make uses its supplying provider. Preserve make's existing stdout states.
- [x] Implement GrantEvidence and effective_order with source-aware and temporal invalidation; verify counted growth and protocol failure fallback.
- [x] Implement Tickets.admission as one election/token/weighted-scan critical section under the existing allocation mutex. Failed capacity scans release the queue token; watcher lifetime is unchanged.
- [x] Report effective class/issue/deadline and predecessor/capacity, including class/position changes; preserve recognizable waiting/acquired diagnostics. Run the real-helper integration boundary green and commit.

## Task 3 — Verify rollout and land one atomic feature

Files: atlas/workflow/sdlc-binary.md and atlas/workflow/workspace-branching.md if its suite guidance needs a pointer; relevant embedded local-check/landing help; issue Log. Update README only if the protocol becomes a user-facing command/flag.

- [x] Document priority, expiry/revocation, FIFO, weights, unknown evidence, and the legacy-head rollout boundary. Preserve local-gates configuration; if required scope touches one, stop for operator verification/label through the TL.
- [x] Run the focused family: bash scripts/test/test-lock.test.sh; go test ./cmd/sdlc/internal/landingline -count=1; go test ./cmd/sdlc -run 'Test(SuitePriority|SuiteQueuePriority|LocalCheck.*(Reentry|Signal)|.*Helper.*)' -count=1 -timeout 10m, adjusted to actual discovered test names before reporting.
- [x] Commit/push; run sdlc local-check run --name suite. Do not manually chase newer main after a branch pass; close the tested branch and let merge own final integration.
- [x] Per TL rollout revision below, build in this worktree and run the isolated real-helper harness against that binary with private lock/grant paths. Never stage or perturb live suites. The TL rebuilds :0 after landing.
- [x] Complete close review with measured actuals and concrete real-helper evidence; round 3 is SHIP, all findings disposed, PR #312 is open and final-code full suite passes.

Landing protocol: sdlc merge --yes owns main integration and any required integration suite; verify tracker/archive/clean resting branch, report PR and Git-derived merge hash to hub, then stop.

## Approval checkpoint

No implementation or regression edits have started. Await TL approval of the design, including the two rollout decisions above; then run sdlc change-code on the full flow. Add timestamped Revisions for subsequent design deltas, and stop/re-plan at a third delta on the same mechanism.

## Revisions

2026-10-18 — TL approved design (Couch receipt 3b55ef84-675d-4f81-834d-80f67f178da1). Resolve both rollout questions: a missing, old or unreadable provider always emits a visible priority-unavailable warning and falls back to normal FIFO, including known-issue suites; priority must never prevent a suite running. Old sticky waiters drain naturally, never killed or unlocked. Priority lasts only while a grant is valid. The expiry-while-waiting regression must prove an expired priority ticket returns to its original normal FIFO position and the older normal waiter proceeds. No perpetual priority bit is retained.

2026-10-18 — Plan gate round 1: address PQ-1 by including a Go-derived refresh_at boundary covering future activation and expiry, invalidating unchanged-source evidence at that time; Python retains no independent grant-validity policy. Address PQ-2 by naming production test surfaces and replacing prose case inventories with adversarial strategies. This is the first mechanism design delta.

2026-10-18 — Implementation binding uses optional WF_TEST_PRIORITY_PROVIDER / WF_TEST_PRIORITY_SUBJECT environment metadata, preserving older helper capacity argv instead of requiring a new capability flag. This implements the approved always-run fallback policy without changing the grant authority or admission mechanism. Focused and legacy real-helper tests pass; full recorded suite and rollout/close remain.

2026-10-18 — Rollout revision, TL receipt aabb93a7-728d-4289-b2fe-5a3dcaec9128: no :0 staging; it would expose every live slot to the new tools. Build in this worktree and run the harness against that binary with private resources. Completed: make sdlc-build; bin/sdlc revision 4e9ba7e62802d49e11997a8150999bdda3c4d64e, modified=false, trimpath=true; all 13 harness tests pass against the actual binary in 24.615s. Real-fleet observation happens naturally after landing when TL rebuilds :0. Recorded branch suite is queued behind an existing pair suite.

2026-10-18 — Verification sequencing: tl-worker permits close review alongside a queued recorded suite, provided the final-code recorded pass precedes landing. Initial full suite reached all shards/packages and failed only the existing #448 TestConstructScripts/plain timeout. Targeted unchanged diagnostic passed; one unchanged issue-backed full rerun is queued on e8b34c96c6be091b199fe68f6dda37e6eef6327d. Start close review now; do not land before that pass. No production design or acceptance criteria changed.

2026-10-18 — Boundary round 1: BR-1 restores disappearance tolerance between a live-ticket snapshot and open; release remains compatible with older helpers, with no new mutex contract. BR-2 centralizes caller/persisted subject path validation in valid_subject before stat/provider use, emitting normal-FIFO fallback on invalid metadata. A real-flock controlled release interleaving and malformed-subject sweep fail before and pass after. Focused family passes (63.600s / 0.069s); queued unchanged rerun cancelled before execution to replace it with the corrected code.

2026-10-18 — BR-3 prose correction: the Core concepts table and Task 2 file inventory now distinguish the modified localcheck_run.go subject/provider binding from helpercapability.go, which is reused unchanged. Optional environment metadata preserves its existing capability contract; no capability code change is needed or claimed. Round 2 disposed BR-1 and BR-2 as addressed; runtime code and queued suite candidate remain unchanged. Actual-binary smoke against the corrected runtime passed all 16 tests in 32.322s.

2026-10-18 — Final verification completed: recorded final-code suite on cbd3a63740a3c87de7c8657da7a7fda0e1d14cf3 passed all 1657 cmd/sdlc tests and 70 other packages, zero failures, 674.6s; attempt 33012b468ed484ca9808f38f207b87eb. Runtime code is unchanged since that head. Close SHIP/PR #312 preparation is complete; landing remains the binary-owned protocol above, rather than an unchecked implementation task. Earlier #448 failure and final passing attempt are both retained.
