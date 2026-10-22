---
collaboration-mode: discussion
---

# Bound suite grant overtaking by waiter age

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to choose execution. Steps use checkbox syntax.

**Goal:** A suite ticket waiting at least 60 minutes protects its FIFO prefix against further grant overtaking, with the same decision visible in queue status.

**Architecture:** Keep Go as grant-validity authority and Python as suite-admission authority. Stamp immutable monotonic enqueue time under the existing allocation lock; compute one pure aging order used by admission, waiter diagnostics and a bounded read-only helper inspection called by text queue status. Retain all existing capacity, pause and landing-line exclusion.

**Tech Stack:** Python scheduler/flock/process tests; Go CLI observer and existing stateful grant provider; disposable local queue resources.

## Approved policy

TL approved 60 minutes (delivery498c89c0-fcc5-44b5-9fc5-77287a49ee8b), rather than the initially proposed30. The machine routinely queues an hour deep;60 preserves useful priority while preventing today's2.5-hour overtaking.

1. Under Tickets allocation lock, stamp a new ticket's immutable enqueue monotonic time. Do not use suite process start, grant timestamps, file mtime or a renewable claim as age.
2. At each admission decision, find the last numeric-FIFO ticket whose elapsed age is at least60minutes, or whose age cannot be trusted. Preserve the entire prefix through that ticket in FIFO order.
3. Apply the existing valid-grant-first stable partition only to the younger suffix. Multiple grants retain FIFO there; grant renewals, replacements, revoke/regrant and train subjects never reset enqueue age.
4. Existing running suites continue. A protected heavy head waits for all its slots atomically; a lighter younger ticket cannot fill around it.
5. Missing, boolean, nonfinite, negative or future enqueue values are unknown and protect their prefix. A held ticket's flock cannot survive reboot, so no new boot registry is required for same-boot monotonic comparisons.

The contract bounds overtaking, not wall-clock start time: running suites, queue pause and stopped live FIFO heads may still delay admission. This issue does not modify repository landing-line GrantHold/Order or grant validity/expiry rules.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| TicketAge | scripts/test-lock.py | new validated elapsed-age projection; unknown is explicit |
| SuiteOrder | scripts/test-lock.py | new order plus protected boundary, ages and grant-override reason |
| effective_order | scripts/test-lock.py | modified to derive from SuiteOrder |

TicketAge is1:1 with a live ticket; SuiteOrder is1:N tickets for one immutable observation. Supply wall time for grant expiry and monotonic time for age as separate explicit inputs. Prefix selection and grant partition are computed once; diagnostics consume that value rather than reimplementing the decision (ARCH-PURE/ARCH-DRY). No extra business-state machine or durable fairness debt is introduced.

| Integration | Lives in | Status | Wraps |
|---|---|---|---|
| Tickets creation/election | scripts/test-lock.py | modified | existing allocation lock, ticket flock and slot scan |
| Suite inspection | scripts/test-lock.py | new read-only command | bounded ticket snapshot and existing GrantEvidence/provider |
| Suite admission observation | cmd/sdlc/queue_suite_status.go | new | workflowFile supplier resolution and bounded Python process |
| Queue text rendering | cmd/sdlc/queuecmd.go | modified | injected admission observation |
| GrantEvidence / suiteGrantSnapshot | existing Python/Go owners | reused | grant-log validity and expiry |

Stateful external coverage reuses TestSuiteQueuePriorityRealHelper and the production Go provider, real isolated flocks/processes and private machine-lock roots. No private replacement grant parser or Go scheduler is added (ARCH-MOCK).

## One policy setting and status contract

Use one named module-level setting, GRANT_MAX_WAIT_SECONDS=3600, in the supplying Python helper. Admission and helper inspection derive from it; Go has no second60-minute constant. No per-invocation environment override or new tuning CLI is proposed, since different waiter values would disagree on admission. Tests inject explicit policy/time inputs into the pure core.

The read-only inspection reports policy seconds, observation validity, ticket ages (or unknown), effective order, protected boundary and whether grant priority was overridden. Queue text status renders that projection and prints the setting. Per-waiter output names its own age, protected-prefix reason and effective position. Keep the existing frozen queue status --json schema_version1 unchanged; supplemental scheduling information is text-only in this issue.

Reuse workflowFile to resolve the actual supplying helper. Add an inspect command to the existing helper, keeping acquire's accepted argv/required watch contract. Inspection never creates resources, allocates tickets, prunes dead files, acquires suite slots, changes grant logs or launches watchers. Use a nonblocking allocation lock to take/revalidate its snapshot, querying slow grant evidence outside that lock. A changed snapshot or old/unavailable helper produces explicitly unavailable order rather than a guessed rank. Reuse the existing status context deadline and bounded process-group runner; keep all concurrent observation work joined before return.

## State and action table

Actions: wait, revoke/let grant expire, resume a paused queue, retry status observation, cancel one's own queued run through sdlc.

| Internal condition | Visible state | Next action | Possible loss |
|---|---|---|---|
| All known ages below60m | Grant FIFO followed by normal FIFO, ages and limit shown | Wait or revoke grant | None |
| Aged ticket exists | FIFO protected through ticketN; aging override named where applicable | Wait for protected prefix | None |
| Legacy/invalid age | Age unknown; FIFO prefix conservatively protected | Wait | Priority benefit only |
| Protected heavy head lacks capacity | Protected next head, required slots still unavailable | Wait for existing holders | None; no partial capacity held |
| Queue paused | Existing pause reason plus observed aging information | Resume named pause owner | None; age not reset |
| Provider/snapshot/helper unavailable | Order unavailable or existing documented FIFO fallback, never invented priority | Retry observation | Priority benefit only |
| Ticket owner dies | Existing liveness pruning removes its ticket | Wait/retry own cancelled operation | Only dead run's queue place |

Allocation lock is the single ordering resource. The queue token and global admission gate still serialize admission; no negative observation is used as replacement authority. Time is resampled at election after revalidation, so crossing the boundary needs no file write (ARCH-ORDER).

## Compatibility and lifetime

The monotonic enqueue member is additive to existing ticket/slot JSON. Dict copy/rewrite paths retain unknown members; known age types are validated before use. No existing field changes meaning, and older metadata is protected instead of denied admission. Unknown nested metadata survives ticket-to-slot propagation. Inspection caps each record read at64KiB; existing ticket/slot writers do not enforce that size, and adding writer bounds is outside this issue. Files are removed by current ticket release/dead-owner pruning; no new durable history or per-grant bypass counter exists (ARCH-EVOLVE/ARCH-FUNERAL).

Already-running older helper processes do not hot-reload Python code; a new helper protects their missing-age tickets, but cannot retrofit an old process's scheduling algorithm. Do not cancel/reorder live work. Live conformance uses a homogeneous private cohort; rollout is effective on new helper launches. The TL chooses whether any separate operator-controlled fleet cutover is needed. Status explicitly reports unavailable inspection when the supplying helper predates it.

## Operating envelope and bounds

N live tickets and I subject issue references drive admission cost. Preserve existing numeric ordering and grant cache; aging adds one pass and stable partition, O(N+I) work and O(N) transient output, with no new provider call per poll beyond existing evidence refresh. Add counted N/2N tests for age/subject visits; timing alone is not the oracle.

Inspection is a human status read: cap4096 ticket records,64KiB per existing record and1MiB total response, and remain within the existing3-second status context. An exceeded bound or busy/changed snapshot is explicit incomplete scheduling observation. Grant provider calls retain existing2-second/64KiB ceilings and occur outside allocation/global locks. Admission polling and running-suite capacity limits are unchanged. No new CPU workers, network probes or unbounded background process is introduced (ARCH-CONSTRAINTS).

## Implementation and verification

### 1. Pure order and red regression

Files: scripts/test-lock.py; scripts/test/test-lock-priority.test.py.

- [ ] Write and run the failing pure-order regression, then implement TicketAge/SuiteOrder and route effective_order through the shared projection. Verify using the function-level strategies below.
- [ ] Commit the verified pure core.

### 2. Ticket lifecycle and actual admission

Files: scripts/test-lock.py; scripts/test/test-lock-priority.test.py; scripts/test/test-lock-pause.test.py; existing Go helper harness tests.

- [ ] Add the lifecycle/admission regressions, observe red, then stamp enqueue time under alloc and propagate the immutable record through election and acquisition.
- [ ] Verify through the existing real helper/provider harness using the function-level strategies below; capture a waiting diagnostic and commit.

### 3. Shared inspection and queue status

Files: scripts/test-lock.py; cmd/sdlc/queue_suite_status.go; cmd/sdlc/queue_suite_status_test.go; cmd/sdlc/queuecmd.go; cmd/sdlc/queuecmd_test.go.

- [ ] Add inspection/status regressions and observe red, then implement the bounded read-only inspection and inject its contextual observer into queueCommandDeps.
- [ ] Render the shared projection in text status, verify with the function-level strategies below, and commit.

### 4. Docs, conformance and landing

Files: atlas/workflow/sdlc-binary.md; README.md if its queue entry needs the new policy; issue Log and plan.

- [ ] Document the60-minute overtaking bound, unknown-age protection, running-heavy-head limit and current-helper rollout scope. Keep Go grant validity / landing queue contracts unchanged.
- [ ] Run targeted Python/Go suites through the existing real-provider harness. Propose a built-binary read-only queue status plus private controlled mixed-weight admission as live conformance; no real grant edits, ticket rewrites, cancellations or fleet stage without hub authorization.
- [ ] Push/checkpoint; run the recorded suite, mandatory boundary review and sdlc merge. Report landing and clear any authorized stage. No implementation before full-plan approval.

## Risk/function conformance table

| Risky function | Adversarial strategy | Independent mechanical guard |
|---|---|---|
| ticket_age (TicketAge projection) | Arbitrary JSON ages and independent wall/monotonic clock movement | Explicit unknown/elapsed results and immutable-input assertions |
| suite_order (SuiteOrder projection), effective_order | Boundary-crossing FIFO cohorts with changing grant evidence | Expected protected permutation; counted age/subject visits at N/2N reject superlinear growth |
| Tickets.__init__ | Delayed allocation and caller-supplied record members | Controlled allocation lock/clock proves stamp occurs under alloc; record round-trip preserves unknown members |
| Tickets.elect | Adversarial weighted cohorts with occupied slots | Independent real flock admission oracle proves order and all-or-none capacity; holders remain untouched |
| Tickets.admission | Pause and provider-triggered snapshot churn | Real gate/snapshot fixture proves revalidation before admission and unchanged enqueue stamp |
| inspect_queue | Concurrent ticket changes, dead records and slow provider | Explicit unavailable observation plus byte/identity comparison proves no queue mutation; enforce record/count/output ceilings |
| observeSuiteQueue | Old or malformed helper and process cancellation | Bounded real subprocess fixture proves supplier resolution, deadline/output enforcement and child reaping |
| format_suite_wait, renderSuiteQueue | Shared projection with overridden grants or unknown age | Rendered position/reason matches independently observed next admission; frozen status JSON fixture remains unchanged |

## Revisions

### 2026-10-22 — plan-quality verification contract

PQ-1 requested function names and one adversarial strategy/mechanical guard per risky function. Consolidated repeated test inventories into the table and named the planned pure, inspection, observer and rendering functions. PQ-2 corrected the distinction between bounded reads and unchecked existing writes; writer enforcement remains outside scope. Approved scheduling policy and rollout are unchanged.

## Architectural lenses

ARCH-DRY/PURE: one Python policy/setting/projection owner, Go only renders observed scheduling facts. ARCH-PURPOSE: include queue status and actual weighted admission, not just a comparator. ARCH-MOCK: reuse stateful Go grant provider and private real process/flock cohort. ARCH-CONSTRAINTS: linear extra work and bounded observation. ARCH-SECURE: inspection never mutates capacity/grants; malformed ages cannot grant priority. ARCH-ORDER: immutable age plus existing allocation/global gates, explicit unknown observations. ARCH-FUNERAL: existing ticket/slot lifetimes, no new persistent family. ARCH-EVOLVE: additive age metadata, unknown members retained, frozen status JSON untouched.
