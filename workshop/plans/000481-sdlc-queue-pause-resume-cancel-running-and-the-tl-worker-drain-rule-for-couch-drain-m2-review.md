# Boundary Review — ariadne#481 (milestone M2)

| field | value |
|-------|-------|
| issue | 481 — sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain |
| repo | ariadne |
| issue file | workshop/issues/000481-sdlc-queue-pause-resume-cancel-running-and-the-tl-worker-drain-rule-for-couch-drain.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 1179590a9230dc0f8eb7263aeeb1549ed09e411c..7c307ed8b35875f36791fd65d6256a08ad2a5644 |
| command | sdlc milestone-close --issue 481 --milestone M2 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T03:02:26-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M2 wires both admission families, retains separate interruption evidence, and updates README, atlas, and worker instructions. Targeted tests pass, including race checks. One blocking defect remains: admission can hold the machine-wide gate indefinitely while waiting for a secondary mutex, preventing pause/resume and stalling unrelated repositories.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: transaction-deadline-enforcement
    title: |
      Admission waits indefinitely on secondary locks while holding the global gate
    detail: |
      scripts/test-lock.py:620 takes Tickets.locked's blocking flock at line 559 inside QueueGate.try_admit's global transaction. cmd/sdlc/internal/landingline/line.go:234 similarly calls withMutex's blocking flock at line 559 inside gate(resource). A stopped or delayed allocation-lock holder therefore monopolizes the machine gate: pause/resume cannot acquire it, and unrelated admissions stall. A private foreground reproduction using the production Python admission path confirmed that admission remained blocked and control.lock remained held after 4.2 seconds; releasing the allocation lock immediately completed admission. This is the 2nd finding in family `transaction-deadline-enforcement`. Earlier rounds fixed instances. Do NOT fix only this instance: enforce the rule that every secondary lock acquired inside the global gate is nonblocking, releasing the gate and retrying on contention. Sweep both Go landing and Python weighted admission, preserve global-before-local ordering, and add deterministic contention tests proving pause remains responsive and cancellation terminates the waiter. ARCH-CONSTRAINTS / ARCH-ORDER.
```

1. **Strengths**

   - Actual landing admission and head publication share the global transaction; slow grant inputs are prepared outside it.
   - `Lease.RetireOutcome` serializes completion against cancellation, and managed-suite tests verify descendant cleanup.
   - Interruption tests verify retained diagnostics, no SuiteAttempt/status poisoning, and successful same-head retry.
   - README and atlas changes cover the new runtime and interruption-log surface.

2. **Critical findings**

   - The finding above covers both [Python admission](/Volumes/TestScratch/review-checkouts/d96a99e0fc0e7a54/scripts/test-lock.py:620) and [landing admission](/Volumes/TestScratch/review-checkouts/d96a99e0fc0e7a54/cmd/sdlc/internal/landingline/line.go:234). Replace blocking secondary acquisition with a nonblocking attempt and retry after releasing the global gate.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Queuegate and landingline package tests.
   - All `TestQueue` tests, normally and with `-race`.
   - All 13 private Python pause tests.
   - All 16 priority tests through `TestSuiteQueuePriorityRealHelper`.
   - TL-worker, golint, and diff whitespace checks.

   The initial standalone priority run lacked `bin/sdlc`; the repository wrapper supplied the provider and passed. Existing tests do not exercise secondary-lock contention while admission owns the global gate. No full suite was run during this review.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass: existing schedulers and cancellation machinery reused. |
   | ARCH-PURE | Pass: pure transitions remain separate from runtime IO. |
   | ARCH-PURPOSE | Pass: suite, landing, ordinary merge, and cleanup admission paths are wired. |
   | ARCH-MOCK | Pass: portable stores, real flocks, and stateful GitHub fixtures exercise integration seams. |
   | ARCH-CONSTRAINTS | **Flag:** secondary-lock waits violate the declared transaction envelope. |
   | ARCH-SECURE | Pass: exact operation identities fence cancellation; reasons remain data. |
   | ARCH-ORDER | **Flag:** a secondary-lock holder can indefinitely extend the global transaction. |
   | ARCH-FUNERAL | Pass: leases, cancellation readers, diagnostics, and unresolved evidence have explicit lifecycles. |
   | ARCH-EVOLVE | Pass: additive operation members survive Python rewrites; strict boundary tests pass. |

7. **Plan revision recommendations**

   Append a `## Revisions` entry establishing nonblocking secondary-lock acquisition under the global gate. Enumerate both admission consumers and require contention barriers that verify responsive pause, bounded cancellation, and retained FIFO position.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"028e317a-5938-4738-8a59-97ee102a17b1","repository":"github.com/xianxu/ariadne","issue_id":"000481","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/d96a99e0fc0e7a54","reviewed_head":"7c307ed8b35875f36791fd65d6256a08ad2a5644","slot":"ariadne:3","requested_at":"2026-10-20T09:57:26.103159Z"},"phase":"finished","pid":57488,"started_at":"2026-10-20T09:57:26.258262Z","finished_at":"2026-10-20T10:02:22.183238Z","session_id":"01a158be-9200-7e20-b3c8-1340920e89cc","outcome":"returned"}] -->
