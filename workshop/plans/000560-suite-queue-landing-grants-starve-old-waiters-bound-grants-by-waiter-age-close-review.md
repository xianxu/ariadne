# Boundary Review — ariadne#560 (whole-issue close)

| field | value |
|-------|-------|
| issue | 560 — Suite queue: landing grants starve old waiters; bound grants by waiter age |
| repo | ariadne |
| issue file | workshop/issues/000560-suite-queue-landing-grants-starve-old-waiters-bound-grants-by-waiter-age.md |
| boundary | whole-issue close |
| milestone | — |
| window | 25024b5ed8de7bc930f357fbabb3a6ecf1771566..532929d548a9badfc18ccf1b433180af93f27037 |
| command | sdlc close --issue 560 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T15:37:56-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The aging policy, weighted admission, shared status projection, and documentation match the approved design. Targeted tests pass. One subprocess-lifetime defect blocks shipping: cancelling status can leave its grant provider running outside the observer’s process group.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: subprocess-cancellation-ownership
    title: |
      Status cancellation leaves the grant provider running
    detail: |
      cmd/sdlc/queue_suite_status.go:101 kills only the inspector's process group, but query_priority at scripts/test-lock.py:919–920 launches the provider in a separate session. Killing the inspector bypasses its provider-cleanup finally block, leaving the provider and potentially its descendants alive after status returns. A private fixture using the actual inspect command reproduced distinct helper/provider groups and confirmed provider survival after the observer-equivalent SIGKILL. Preserve bounded ownership across both groups and add a regression through observeSuiteQueue using the real inspector with a blocked provider; cancellation and deadline return must leave neither provider nor descendants running. ARCH-ORDER, ARCH-FUNERAL.
```

1. **Strengths**

   - `suite_order` centralizes protected-prefix ordering for admission, inspection, and diagnostics.
   - Enqueue age is stamped under the allocation lock and retained through final watcher publication, including unknown nested metadata.
   - Real flock tests verify that an aged heavy waiter blocks younger grants without partial slot acquisition or preemption.
   - README and atlas explain the policy, unavailable observations, frozen JSON compatibility, and rollout limits.

2. **Critical findings**

   - **Provider escapes cancellation**, at `cmd/sdlc/queue_suite_status.go:101`, interacting with `scripts/test-lock.py:919–920`. Ensure inspector cancellation cleans up its separately grouped provider before returning. Simply adding a Python signal handler is insufficient while Go immediately uses SIGKILL.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed 22 Python aging/resource/evidence tests.
   - Passed 18 Python pause tests.
   - Passed `go test ./cmd/sdlc -run 'TestSuiteQueue|TestHelper.*' -count=1`, including the real helper/provider harness.
   - Existing descendant tests cover children remaining in the inspector’s group; they miss the production provider’s separate session.
   - The private cancellation reproduction confirmed the uncovered defect. Checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one scheduling projection and policy owner.
   - **ARCH-PURE — pass:** deterministic ordering core; IO remains in integration functions.
   - **ARCH-PURPOSE — pass:** actual admission and observable aging behavior are delivered.
   - **ARCH-MOCK — pass:** private stateful grant fixtures exercise the production provider seam.
   - **ARCH-CONSTRAINTS — pass:** counted linear work and bounded inspection reads/output are covered.
   - **ARCH-SECURE — pass:** hostile files and invalid observations fail visibly.
   - **ARCH-ORDER — flag:** cancellation drops ownership of an in-flight provider.
   - **ARCH-FUNERAL — flag:** that provider lacks cleanup when its inspector is killed.
   - **ARCH-EVOLVE — pass:** additive age metadata and unknown nested members survive publication; status JSON remains unchanged.

   Core-concept table entries map to the documented projections and integration paths.

7. **Plan revision recommendations**

   Add a `## Revisions` entry extending the observer verification contract to cancellation/deadline during a real inspector’s provider call, including separately grouped provider descendants. Retain the existing requirement that all observation work ends before return.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1393cf07-8a1f-484f-9881-ff17abc9cade","repository":"github.com/xianxu/ariadne","issue_id":"000560","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/c8a72681fe917250","reviewed_head":"532929d548a9badfc18ccf1b433180af93f27037","slot":"ariadne:1","requested_at":"2026-10-22T22:35:57.133123Z"},"phase":"finished","pid":28161,"started_at":"2026-10-22T22:35:57.315073Z","finished_at":"2026-10-22T22:37:55.750063Z","session_id":"01a165c1-bb70-7232-bccf-8e3580a32ec0","outcome":"returned"}] -->
