# Boundary Review — ariadne#332 (milestone M2)

| field | value |
|-------|-------|
| issue | 332 — sdlc merge integrates a moved main itself, then re-runs the suite before landing |
| repo | ariadne |
| issue file | workshop/issues/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9ea8e8a37b74f56317b1e4b0547411426b2f259e..77a9668453d70378d18a8c13874810e973553d4c |
| command | sdlc milestone-close --issue 332 --milestone M2 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T23:48:50-07:00 |
| verdict | unknown |

## Review

## Attempt 1 (ended without a verdict)



## Attempt 2 (retry; also without a verdict)


---

## Re-review — 2026-10-14T00:01:36-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 332 — sdlc merge integrates a moved main itself, then re-runs the suite before landing |
| repo | ariadne |
| issue file | workshop/issues/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9ea8e8a37b74f56317b1e4b0547411426b2f259e..99b0ee0916d97cfeb3182c570f6b88b9494cfb52 |
| command | sdlc milestone-close --issue 332 --milestone M2 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T00:01:36-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The principal M2 scenarios pass: integration, conflict rollback, replay refusal, interrupted-push recovery, concurrent-push protection, reintegration limits, and legacy refusal. README and atlas changes cover the new surface. The boundary still needs enforceable execution bounds and a transition model that owns integration state.

1. **Strengths**
   - Resume validates both parents and the reconstructed merge tree, rejecting forged trailers and modified trees (`cmd/sdlc/landingintegrate.go:84`).
   - Publish replay runs before publication; refusal restores the branch. Lease protection preserves concurrent remote work (`cmd/sdlc/landingintegrate.go:222`).
   - Tests reproduce main movement during the suite through a controlled hook and real disposable Git remotes.
   - Typed CI refusals distinguish retryable states; immutable configuration is cached across polls.

2. **Critical findings**
   - None.

3. **Important findings**
   - **CI wait does not enforce its advertised deadline** — `cmd/sdlc/cigate_wait.go:86`. Each complete gate evaluation receives the original context. The budget is checked only afterward, and only for pending/unregistered results. A poll started at the deadline can spend additional minutes reading GitHub and ancestor statuses; a late green result bypasses the elapsed-time check entirely. Apply the remaining budget to polling IO, while keeping suite execution separately timed. Test slow reads and green completion after expiry. **ARCH-CONSTRAINTS**.
   - **New integration subprocesses have no lifetime bound** — `cmd/sdlc/landingintegrate.go:199,246`. Merge and push use `execGitRunner`, whose `exec.Command` calls have neither context cancellation nor deadlines (`cmd/sdlc/runner.go:37`). A stalled transport, hook, or merge driver can hang indefinitely. **This is the 3rd finding in family `unbounded-subprocess`.** State and enforce the rule across the integration call graph, including recovery probes and cleanup, rather than patching only these two calls. Test cancellation, hanging descendants, and uncertain push completion. **ARCH-CONSTRAINTS, ARCH-ORDER**.
   - **The transition function does not own state transitions** — `cmd/sdlc/landingintegrate.go:34` and `cmd/sdlc/landing.go:430`. The IO shell directly updates authoritative head, PR, attempts, and resume state. The pure helper returns an action from independent observations; its `lease-refused` event has no production caller, and its `integrationGates` result is not explicitly dispatched. Introduce an owned state/event transition model returning next state and effects, and route production outcomes through it. **ARCH-ORDER**.

4. **Minor findings**
   - None requiring a separate finding.

5. **Test coverage notes**
   - Two focused runs passed: **30.941s** and **11.181s**, including M2 behavior, production gate wiring, instance-window handling, and recovery-contract proofs.
   - Existing timeout tests advance time during pauses; they do not exercise slow polling IO or late green responses.
   - Full suite and live conformance were not run. No prior findings required disposition or mutation verification.
   - Checkout remained unchanged. `git diff --check` reported one trailing blank line in the M2 review artifact.

6. **Architectural notes**
   - **ARCH-DRY — pass:** existing publish, suite, and GitHub seams reused.
   - **ARCH-PURE — pass:** decision helper is directly tested without IO; transition ownership remains the separate ORDER finding.
   - **ARCH-PURPOSE — pass for M2:** planned integration behavior is present; failure classification belongs to M3.
   - **ARCH-MOCK — pass:** disposable Git remotes and stateful GitHub doubles share production seams; existing live-conformance hooks are present.
   - **ARCH-CONSTRAINTS — flag:** deadline and subprocess findings above.
   - **ARCH-SECURE — pass:** resume verifies evidence structurally; pushes use explicit leases.
   - **ARCH-ORDER — flag:** shell-owned state bypasses the promised transition boundary.
   - **ARCH-FUNERAL — pass:** no new durable artifact family without an existing lifecycle.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry defining integration state ownership, typed events, effects, and uncertain-push reconciliation.
   - Add an entry enumerating subprocess bounds and specifying how the CI deadline applies to IO while excluding suite runtime.
   - M2 entity locations are accounted for by R5; no additional core-concept classification correction is needed.

```findings
findings:
  - id: new
    severity: Important
    family: polling-deadlines-cover-inflight-work
    title: |
      CI polling can exceed its advertised deadline and accept a late green result
    detail: |
      cmd/sdlc/cigate_wait.go:86-93 passes the original context through a complete gate evaluation and checks elapsed time only for pending/unregistered results. ARCH-CONSTRAINTS: enforce the remaining budget on polling IO, separately from suite runtime, and test slow reads and green completion after expiry.
  - id: new
    severity: Important
    family: unbounded-subprocess
    title: |
      Integration merge and push subprocesses have no deadline or cancellation
    detail: |
      cmd/sdlc/landingintegrate.go:199,246 uses the unbounded execGitRunner in cmd/sdlc/runner.go:37-44. This is the 3rd finding in family unbounded-subprocess: state the rule and sweep the integration call graph, including probes and cleanup, rather than fixing only these instances. ARCH-CONSTRAINTS and ARCH-ORDER require bounded process and descendant lifetimes plus tests for cancellation and uncertain push completion.
  - id: new
    severity: Important
    family: authoritative-state-transitions-owned-by-core
    title: |
      Integration state remains mutable outside the pure transition model
    detail: |
      cmd/sdlc/landing.go:430-435 updates head, PR, attempts and resume state directly; cmd/sdlc/landingintegrate.go:43 only selects actions, and its lease-refused event is test-only. ARCH-ORDER: introduce an owned state/event transition model returning next state and effects, route production outcomes through it, and revise the plan to describe the enforced boundary.
```
