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

---

## Re-review — 2026-10-20T03:22:54-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 481 — sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain |
| repo | ariadne |
| issue file | workshop/issues/000481-sdlc-queue-pause-resume-cancel-running-and-the-tl-worker-drain-rule-for-couch-drain.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 1179590a9230dc0f8eb7263aeeb1549ed09e411c..afb94a8e2be9d2477fbbe7b53d7b8eb7f52152c7 |
| command | sdlc milestone-close --issue 481 --milestone M2 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T03:22:54-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned M2 implementation satisfies the reviewed admission and interruption contracts. BR-5 is addressed across both admission consumers, with regression tests independently confirmed to fail when each fix is reverted. No new blocking findings.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      Go landing admission and Python weighted admission now acquire secondary locks nonblockingly under the global gate. Contention tests cover responsive pause, waiter cancellation and retained FIFO position. Independently reverting each fix in temporary copies made its regression test fail; both pass at the pinned head.
```

1. **Strengths**

   - Global-before-local ordering is preserved at `cmd/sdlc/internal/landingline/line.go:234` and `scripts/test-lock.py:624`.
   - Completion and cancellation serialize through `Lease.RetireOutcome`; tests cover both orderings.
   - Managed cancellation tests verify descendant termination, retained diagnostics, absence of failed-suite evidence and same-head retry.
   - README, atlas and worker help cover the new runtime and drain behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed queuegate and landingline tests normally and with `-race`; all `TestQueue` tests normally and with `-race`; 15 Python pause tests; 16 priority tests through the production-provider wrapper; selected saved-pass and repeated-run compatibility tests; worker/helptext tests; and diff whitespace checks.

   The standalone priority invocation lacked `bin/sdlc`; the repository wrapper supplied the provider and passed. No full suite or external pair#722 conformance rerun was performed during this review. Temporary mutation copies were removed; the checkout remains clean.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass: existing schedulers, locks and process-group cleanup reused. |
   | ARCH-PURE | Pass: pure models remain separated from runtime IO adapters. |
   | ARCH-PURPOSE | Pass: suite, landing, ordinary merge and cleanup admissions are covered. |
   | ARCH-MOCK | Pass: portable stores, real flocks and stateful GitHub fixtures exercise production seams. |
   | ARCH-CONSTRAINTS | Pass: secondary contention releases the global gate; counted tests protect admission costs. |
   | ARCH-SECURE | Pass: cancellation uses exact operation/incarnation bindings; reasons remain data. |
   | ARCH-ORDER | Pass: deterministic barriers cover contention and completion/cancellation ordering; cancellation readers are joined. |
   | ARCH-FUNERAL | Pass: leases, watchers, intents and diagnostic retention have explicit lifecycles. |
   | ARCH-EVOLVE | Pass: strict codecs and additive-member preservation remain covered, including Python rewrites. |

7. **Plan revision recommendations**

   None. The existing BR-5 revision states the nonblocking-secondary-lock rule and identifies both consumers.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b5fb9a33-5894-4f92-90f4-745b409f477c","repository":"github.com/xianxu/ariadne","issue_id":"000481","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/bf4f905406079241","reviewed_head":"afb94a8e2be9d2477fbbe7b53d7b8eb7f52152c7","slot":"ariadne:3","requested_at":"2026-10-20T10:18:49.928213Z"},"phase":"finished","pid":47961,"started_at":"2026-10-20T10:18:50.062487Z","finished_at":"2026-10-20T10:22:53.839615Z","session_id":"01a158d2-28b3-7f71-8175-c0c69ed43865","outcome":"returned"}] -->
