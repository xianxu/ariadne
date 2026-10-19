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
| Recorded-check subject and evaluator binding | INTEGRATION | cmd/sdlc/localcheck_run.go; cmd/sdlc/helpercapability.go | modified |
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
- Add a hidden landing suite-priority protocol to sdlc. One describe request resolves a direct caller's subject. One batch request accepts the unique repository/issue pairs in a queue scan and reads the existing grant logs; it returns a version, active IDs and expiry deadlines. Reuse the existing grant JSONL reader and ValidGrants. It must not acquire a landing turn, fetch, create directories, or write logs.
- The recorded runner supplies its own current sdlc executable as the evaluator. Makefile.workflow supplies the built sdlc beside the supplying workflow/helper source. Do not search arbitrary PATH entries or execute shell command strings from ticket files. A ticket carries identity data, never an executable path.
- A helper process uses its trusted evaluator for every ticket, with argv/stdin JSON, bounded output and timeout, and reaps it before returning/cancelling. No evaluator survives admission.
- Cache only in memory, per grant-log identity (path, inode, mtime_ns, size), for the life of that waiting helper. A changed source is reread; expiry is checked against current UTC time on every admission attempt. Missing/unreadable/invalid evidence confers no priority and produces a clear warning. Never cache a transient read error indefinitely.

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

Files: scripts/test/test-lock.test.sh; new cmd/sdlc/suitepriority_test.go; cmd/sdlc/localcheck_reentry_test.go as needed.

- [ ] Extend bounded ticket/diagnostic waits so a missed synchronization deadline fails explicitly; use private WF_TEST_LOCK_FILE and controlled watched-owner processes.
- [ ] Add a red case with both slots held, an earlier normal waiter already reporting next, then a granted priority waiter. Release one slot; priority must acquire it while the original running owner survives. Release priority/remaining owner and verify normal follows.
- [ ] Add P1/P2 FIFO after an older normal, normal-only FIFO, expired-at-start and expires-while-waiting controls. Use grant-log replacement/revocation under held capacity, not a long sleep.
- [ ] Add weight-2 priority, normal light backfill refusal, priority cancellation/death and nested/repeated-run cases. Keep the mixed-old-helper safety test.
- [ ] Exercise the real Python helper and actual production Go grant-reader/validator against fixture grant files. A test-executable launcher may route the hidden command into its production handler; it must not fake active-priority decisions.
- [ ] Run bash scripts/test/test-lock.test.sh and focused new Go priority tests; retain expected red output, then commit the regressions.

## Task 2 — Add one grant-evidence boundary and reelected admission

Files: cmd/sdlc/suitepriority.go (new), cmd/sdlc/landingcmd.go, cmd/sdlc/internal/landingline/log.go, scripts/test-lock.py, Makefile.workflow, cmd/sdlc/localcheck_run.go, cmd/sdlc/helpercapability.go and their focused tests.

- [ ] Implement/read-test the versioned read-only subject/batch protocol with existing ValidGrants. Cover wrong repository/issue, revoked, expired, future-dated, unauthorized, malformed and missing grants; do not duplicate validity policy.
- [ ] Pass explicit tested-suite subject and evaluator from recorded runs. Keep main diagnostics/non-suite checks normal. Update helper capability preflight and its canonical fake-helper fixture for the added argv contract before any waiter can be queued incorrectly.
- [ ] Wire direct make test to its supplying provider and verify branch identity through the real helper. Preserve make's acquired/reentrant/skipped/timeout stdout protocol.
- [ ] Implement in-memory source-aware evidence caching and pure stable partition, including counted growth tests.
- [ ] Change admission to one election/queue-token/weighted-scan critical section under the existing allocation mutex. Failed capacity scans release the queue token. Slot descriptors and watcher behavior remain unchanged.
- [ ] Report effective class/issue/deadline and what it waits behind; report a class/position transition promptly and preserve existing recognizable waiting/acquired lines used by consumers.
- [ ] Run all new real-helper scenarios green, existing shell FIFO/weight/lifetime regressions, landingline unit tests, and recorded-run reentry/signal/helper-capability controls. Commit the implementation once the class sweep is green.

## Task 3 — Verify rollout and land one atomic feature

Files: atlas/workflow/sdlc-binary.md and atlas/workflow/workspace-branching.md if its suite guidance needs a pointer; relevant embedded local-check/landing help; issue Log. Update README only if the protocol becomes a user-facing command/flag.

- [ ] Document priority, expiry/revocation, FIFO, weights, unknown evidence, and the legacy-head rollout boundary. Preserve local-gates configuration; if required scope touches one, stop for operator verification/label through the TL.
- [ ] Run the focused family: bash scripts/test/test-lock.test.sh; go test ./cmd/sdlc/internal/landingline -count=1; go test ./cmd/sdlc -run 'Test(SuitePriority|SuiteQueuePriority|LocalCheck.*(Reentry|Signal)|.*Helper.*)' -count=1 -timeout 10m, adjusted to actual discovered test names before reporting.
- [ ] Commit/push; run sdlc local-check run --name suite. Do not manually chase newer main after a branch pass; close the tested branch and let merge own final integration.
- [ ] Ask the TL for a :0 stage grant before any staging. Stage/build capable source and have the TL coordinate one isolated real-helper smoke using private lock/grant paths. No tests may reorder or cancel live machine suites merely for demonstration. Clear/rebuild/verify under the TL's staging instructions.
- [ ] Close with measured actuals and concrete real-helper/full-suite evidence. The binary owns the single fresh-context close review. On SHIP, sdlc pr then sdlc merge --yes, verify tracker/archive/clean resting branch, report PR and Git-derived merge hash to hub, stop.

## Approval checkpoint

No implementation or regression edits have started. Await TL approval of the design, including the two rollout decisions above; then run sdlc change-code on the full flow. Add timestamped Revisions for subsequent design deltas, and stop/re-plan at a third delta on the same mechanism.
