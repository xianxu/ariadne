# Boundary Review — ariadne#452 (whole-issue close)

| field | value |
|-------|-------|
| issue | 452 — suite queue priority: a fleet-blocking landing grant also jumps the machine suite queue |
| repo | ariadne |
| issue file | workshop/issues/000452-suite-queue-priority-a-fleet-blocking-landing-grant-also-jumps-the-machine-suite-queue.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3a0012b449ff8488eb53b938ec591ca996ad46aa..eb44436e9e8781ac40a4888c228e2734f7ebd1bb |
| command | sdlc close --issue 452 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T18:58:17-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The priority mechanism matches the main admission contract, and the focused and legacy tests pass. Two reproduced crashes block shipping: concurrent ticket removal can abort another waiter, and malformed persisted subject data can abort admission instead of degrading to normal FIFO.

```findings
findings:
  - id: new
    severity: Critical
    family: queue-ticket-lifecycle-serialization
    title: |
      Concurrent ticket removal can crash another suite's admission
    detail: |
      scripts/test-lock.py:144-145 opens paths returned by live_tickets without handling disappearance, while Tickets.release at lines 165-171 unlinks tickets without the allocation mutex. An admitted, cancelled, or timed-out waiter can therefore remove its ticket between another waiter's scan and open. I reproduced FileNotFoundError by scheduling release at that boundary. Serialize removal with election or tolerate vanished tickets, and add a deterministic regression covering this interleaving (ARCH-ORDER).
  - id: new
    severity: Critical
    family: persisted-subject-validation
    title: |
      Malformed ticket subjects crash admission instead of falling back to FIFO
    detail: |
      scripts/test-lock.py:224-232 accepts an absolute common_dir containing an embedded NUL; GrantEvidence.deadlines then passes it to os.stat at line 265, whose ValueError escapes the OSError handler. I reproduced this with common_dir="/tmp/review\u0000bad". A malformed live ticket can consequently crash every participating waiter that reads it. Validate persisted subject paths before filesystem use and visibly discard invalid evidence; add malformed-ticket regressions (ARCH-SECURE).
```

1. **Strengths**

   - Grant authority remains centralized in `landingline.ValidGrants`; Python does not duplicate authorization or revocation policy.
   - Memory-only caching handles source changes and time boundaries, with counted tests confirming stable polls avoid provider calls.
   - Real-helper tests exercise priority admission, expiry, revocation, weighted capacity, cancellation, and direct-make binding using private resources.

2. **Critical findings**

   - `scripts/test-lock.py:145`: restore safe handling of disappearing tickets and cover concurrent release.
   - `scripts/test-lock.py:265`: reject malformed persisted paths and preserve visible FIFO fallback.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Focused Go family, including the real-helper harness: `cmd/sdlc` 44.629s.
   - Full `landingline` package: 1.318s.
   - Existing `scripts/test/test-lock.test.sh`: all passed.
   - Pinned-range `git diff --check`.

   The passing tests omit both reproduced failures. Add a controlled scan/release interleaving and malformed persisted-subject cases. I did not rerun the full suite; the tracker explicitly requires its final-code pass before landing.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing grant authority and JSONL reader reused.
   - **ARCH-PURE — pass:** ordering stays pure; filesystem and process work remain at integration boundaries.
   - **ARCH-PURPOSE — pass:** both recorded suites and direct make receive subject/provider bindings.
   - **ARCH-MOCK — pass:** real lock/provider behavior is exercised through private fixtures and production boundaries.
   - **ARCH-CONSTRAINTS — pass:** provider timeout/output limits and cache-call counts are tested.
   - **ARCH-SECURE — flag:** persisted subject validation permits an admission crash.
   - **ARCH-ORDER — flag:** ticket removal bypasses election serialization.
   - **ARCH-FUNERAL — pass:** existing ticket cleanup remains; caches are process-local and providers are reaped.

   Atlas and embedded help cover the new behavior. The hidden protocol and internal metadata introduce no additional user-facing command or configuration requiring a README update.

7. **Plan revision recommendations**

   Add a `## Revisions` entry documenting ticket-removal synchronization and the malformed-subject fallback policy, with deterministic regression evidence for both.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f5c17809-02e2-4770-9fba-e8fe1f55f378","repository":"github.com/xianxu/ariadne","issue_id":"000452","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/bde00c2a014aaac8","reviewed_head":"eb44436e9e8781ac40a4888c228e2734f7ebd1bb","slot":"ariadne:3","requested_at":"2026-10-19T01:55:58.79483Z"},"phase":"finished","pid":26372,"started_at":"2026-10-19T01:55:58.922231Z","finished_at":"2026-10-19T01:58:17.126507Z","session_id":"01a151df-6c7d-76d2-a333-c1c4b3edc0b2","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T19:07:04-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 452 — suite queue priority: a fleet-blocking landing grant also jumps the machine suite queue |
| repo | ariadne |
| issue file | workshop/issues/000452-suite-queue-priority-a-fleet-blocking-landing-grant-also-jumps-the-machine-suite-queue.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3a0012b449ff8488eb53b938ec591ca996ad46aa..cbd3a63740a3c87de7c8657da7a7fda0e1d14cf3 |
| command | sdlc close --issue 452 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T19:07:04-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Both prior defects are fixed, with regressions that pass at the pinned head and fail against the pre-fix implementation. Focused and legacy tests pass. The remaining blocker is a Core concepts table contradiction; this review’s explicit table-consistency rule classifies it as Critical. No additional production defect was found.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      scripts/test/test-lock-priority.test.py:302 deterministically releases an older ticket after enumeration. It passes at head and raises FileNotFoundError with the pre-fix implementation loaded in memory.
  - id: BR-2
    disposition: addressed
    note: |
      scripts/test/test-lock-priority.test.py:330 exercises malformed persisted subjects, verifies visible fallback and prevents filesystem access. It passes at head and raises ValueError for the embedded-NUL case with the pre-fix implementation.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: core-concept-change-traceability
    title: |
      Core concepts table claims a modification absent from the pinned range
    detail: |
      workshop/plans/000452-suite-queue-priority-a-fleet-blocking-landing-grant-also-jumps-the-machine-suite-queue-plan.md:26 lists cmd/sdlc/helpercapability.go under a modified entity, but the pinned diff contains no change to that file. The later revision correctly explains that existing capability checks remain unchanged. Correct the table to distinguish the modified localcheck_run.go binding from unchanged capability infrastructure, and add a Revisions entry reconciling it. This is a prose-only correction; no runtime change or wording-presence test is required.
```

1. **Strengths**

   - Grant validity reuses `landingline.ValidGrants`, including authorization, revocation and expiry.
   - Admission reelection preserves weighted, all-or-none capacity acquisition without reserving the queue token between attempts.
   - Prior-finding tests reproduce the actual failure boundaries rather than merely asserting implementation details.
   - Provider execution has explicit timeout, output bounds and cancellation cleanup.

2. **Critical findings**

   - The plan table mismatch above. Correct the table; changing capability code is unnecessary.

3. **Important findings**

   - None.

4. **Minor findings**

   - None.

5. **Test coverage notes**

   - Focused Go family passed: `cmd/sdlc` in 85.935s; `landingline` in 0.169s, including the real-helper priority harness.
   - Complete `landingline` package tests passed in 5.116s.
   - Legacy `scripts/test/test-lock.test.sh` passed.
   - Both prior regressions independently failed against the pre-fix implementation and passed at head.
   - Pinned-range `git diff --check` passed; checkout remained clean.
   - A full recorded suite was not independently rerun during this review; its final-code pass remains required before landing.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing grant authority and JSONL reader are reused.
   - **ARCH-PURE — pass:** stable ordering is pure; filesystem and subprocess work remain in integration code.
   - **ARCH-PURPOSE — pass:** priority reaches both recorded checks and direct make, including live grant changes.
   - **ARCH-MOCK — pass:** no new external service dependency; tests exercise production protocol logic against private fixtures.
   - **ARCH-CONSTRAINTS — pass:** provider bounds and cache behavior have adversarial and counted coverage.
   - **ARCH-SECURE — pass:** persisted subjects are validated before path use; tickets cannot select executables.
   - **ARCH-ORDER — pass:** election and weighted scan share the allocation critical section; controlled tests cover release, stopped waiters, expiry and cancellation.
   - **ARCH-FUNERAL — pass:** caches die with helpers; provider groups are reaped; existing ticket cleanup is retained.

   Atlas and embedded help cover the feature. The hidden protocol and internal binding metadata introduce no additional user-facing command or configuration requiring a README update.

7. **Plan revision recommendation**

   Add a `## Revisions` entry stating that subject/provider binding changed in `localcheck_run.go`, while `helpercapability.go` remains reused unchanged because optional environment metadata preserves the existing capability contract. Update the Core concepts row accordingly.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a0b91489-4dad-4250-8aab-73d1fc311e20","repository":"github.com/xianxu/ariadne","issue_id":"000452","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a54cf661db7e568a","reviewed_head":"cbd3a63740a3c87de7c8657da7a7fda0e1d14cf3","slot":"ariadne:3","requested_at":"2026-10-19T02:04:21.661143Z"},"phase":"finished","pid":22517,"started_at":"2026-10-19T02:04:21.991358Z","finished_at":"2026-10-19T02:06:47.499899Z","session_id":"01a151e7-1a59-7613-bf29-30afc1f04833","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T19:11:18-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 452 — suite queue priority: a fleet-blocking landing grant also jumps the machine suite queue |
| repo | ariadne |
| issue file | workshop/issues/000452-suite-queue-priority-a-fleet-blocking-landing-grant-also-jumps-the-machine-suite-queue.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3a0012b449ff8488eb53b938ec591ca996ad46aa..0e36a07b0b8ff23f565224549d685ef68a7c4e36 |
| command | sdlc close --issue 452 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T19:11:18-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the Spec’s priority, FIFO, expiry, weight, and lifecycle requirements. BR-3’s documentation correction matches the code; no new findings remain. This review passes the close boundary. The plan’s recorded full-suite pass remains required before landing.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The controlled ticket-removal regression passes at head and fails with FileNotFoundError against the pre-fix helper loaded in memory.
  - id: BR-2
    disposition: addressed
    note: |
      The malformed-subject regression passes at head and fails with embedded-NUL ValueError against the pre-fix helper loaded in memory; validation prevents invalid identities reaching filesystem access.
  - id: BR-3
    disposition: addressed
    note: |
      Plan lines 26–27 now distinguish the modified localcheck_run.go binding from unchanged helpercapability.go; Task 2 and the Revisions entry at line 141 agree with the pinned implementation diff. This is a prose-only correction.
```

1. **Strengths**

   - Grant authority remains centralized in `landingline.ValidGrants`; the new provider reads existing logs without creating a landing line.
   - `Tickets.admission` ree elects under the allocation mutex and releases the queue token after unsuccessful scans, preserving weighted admission without reserving capacity.
   - Real-helper tests control queue arrivals, capacity release, stopped waiters, cancellation, and grant changes.
   - Atlas and embedded help document fallback and mixed-version rollout. No new public command or flag requires a README update.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Focused Go family passed: `cmd/sdlc` in 48.829s.
   - Complete `landingline` package passed in 0.406s.
   - Legacy lock-helper suite passed.
   - Six Python unit regressions passed.
   - BR-1 and BR-2 both demonstrated failure without their fixes.
   - Reviewed-range whitespace check passed; tracked checkout remained clean.
   - The recorded full suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — Pass:** existing grant policy and JSONL reader reused.
   - **ARCH-PURE — Pass:** deterministic ordering separated from process and filesystem operations; Core concepts classifications match.
   - **ARCH-PURPOSE — Pass:** both recorded suites and direct make bind subjects and consume the shared authority.
   - **ARCH-MOCK — Pass:** owned provider and helper run through production boundaries against portable, isolated fixtures.
   - **ARCH-CONSTRAINTS — Pass:** counted cache tests, bounded provider transport, and all-or-none weighted scans exercise the declared envelope.
   - **ARCH-SECURE — Pass:** persisted subjects validated; tickets cannot select executables; damaged evidence degrades visibly.
   - **ARCH-ORDER — Pass:** serialized election and controlled sequence tests cover grant changes, expiry, death, cancellation, and stopped waiters.
   - **ARCH-FUNERAL — Pass:** caches die with helpers, provider groups are reaped, and ticket/slot cleanup retains existing ownership.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"76215204-6e74-4de4-8cec-b8672a76154f","repository":"github.com/xianxu/ariadne","issue_id":"000452","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/268c07b459af1e60","reviewed_head":"0e36a07b0b8ff23f565224549d685ef68a7c4e36","slot":"ariadne:3","requested_at":"2026-10-19T02:09:21.910567Z"},"phase":"finished","pid":78552,"started_at":"2026-10-19T02:09:22.091911Z","finished_at":"2026-10-19T02:11:17.553782Z","session_id":"01a151eb-ae31-70f1-afed-5955ac4e6690","outcome":"returned"}] -->
