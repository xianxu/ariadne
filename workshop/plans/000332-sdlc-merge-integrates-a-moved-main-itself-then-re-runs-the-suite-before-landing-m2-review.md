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

---

## Re-review — 2026-10-14T01:02:21-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 332 — sdlc merge integrates a moved main itself, then re-runs the suite before landing |
| repo | ariadne |
| issue file | workshop/issues/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9ea8e8a37b74f56317b1e4b0547411426b2f259e..5fb37ac030d00755217c474413e0d60a5f273993 |
| command | sdlc milestone-close --issue 332 --milestone M2 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T01:02:21-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M2’s integration, conflict recovery, lease protection, retry bound, and exact-head suite behavior passed the targeted tests. BR-14 is addressed. BR-12 and BR-13 remain partially open: the new deadlines do not reach every production subprocess. Two disposable regression probes reproduced those gaps. The repository remains unchanged; disposable copies were removed.

```findings
dispose:
  - id: BR-12
    disposition: not-addressed
    note: |
      Git/GitHub read deadlines and late-green rejection are corrected, with mutation-sensitive tests. However, prepareCIGateFacts at cmd/sdlc/cigate_wait.go:145 reaches probeMakeTestFacts, which creates a background context at cmd/sdlc/localchecks_make.go:84. A real Makefile probe took 1.07 seconds with only 40 ms remaining in the CI budget. Propagate the remaining budget through the complete facts-probe path, retaining separate bounded cleanup.
  - id: BR-13
    disposition: not-addressed
    note: |
      Merge/push deadlines, descendant cleanup, and uncertain-push recovery have meaningful regression evidence. The integration call-graph sweep remains incomplete: publish replay calls gitx.RunGit, whose runner uses context.Background at cmd/sdlc/internal/gitx/window.go:28; runEnv does likewise at rebasedpatch.go:17. A cancelled runLandingPublishGate still spent 1.06 seconds in its first Git read. Ownership also reaches archiveRead at landingarchive.go:126, which invokes the unbounded landingArchiveRunner directly. Apply cancellation and command ceilings across the entire enumerated graph.
  - id: BR-14
    disposition: addressed
    note: |
      Production integration outcomes now pass through transitionIntegration in cmd/sdlc/integrationstate.go, owning head, lease base, PR, resume state, phase, and attempts. Removing retry-count advancement made both state regression tests fail. Production retry, recovery, and cancelled-push tests passed; Plan revision R6 describes the delivered model.
```

1. **Strengths**

   - Real bare-remote tests verify integration, conflict abort, replay rollback, concurrent-push protection, and reintegration limits.
   - Interrupted integration validates both parents, trailer, and reconstructed tree before publication (`landingintegrate.go:47`).
   - Cancelled push tests establish remote publication and verify recovery without duplicate integration.
   - README, help, and atlas document the new flow and remaining server-side main race.

2. **Critical findings**

   None.

3. **Important findings**

   - **BR-12 — ARCH-CONSTRAINTS:** [`localchecks_make.go:84`](/Users/xianxu/Library/Caches/sdlc/review-checkouts/0d4e5e0686e65731/cmd/sdlc/localchecks_make.go:84) gives the Makefile probe an independent background deadline. Budget expiry is detected afterward, but cannot interrupt the probe. Pass the operation context and remaining budget through facts preparation, default-suite derivation, and probe execution.
   - **BR-13 — ARCH-CONSTRAINTS, ARCH-ORDER, ARCH-PURPOSE:** [`window.go:28`](/Users/xianxu/Library/Caches/sdlc/review-checkouts/0d4e5e0686e65731/cmd/sdlc/internal/gitx/window.go:28) discards operation cancellation; [`landingarchive.go:126`](/Users/xianxu/Library/Caches/sdlc/review-checkouts/0d4e5e0686e65731/cmd/sdlc/landingarchive.go:126) bypasses the bounded runner. Complete the existing `unbounded-subprocess` class sweep, covering replay, ownership, probes, and cleanup—not only merge and push.

4. **Minor findings**

   None.

5. **Test coverage notes**

   The targeted M2 set passed: `cmd/sdlc` in **138.918s**, `gitx` in **3.280s**. Removing read-budget enforcement, retry advancement, and descendant cleanup produced regression failures; restored tests passed. The two additional deadline/cancellation probes failed consistently on pinned head. The full suite and authenticated live conformance were not run.

6. **Architectural notes**

   - **ARCH-DRY — pass:** immutable gate facts and existing gate derivation are reused.
   - **ARCH-PURE — pass:** transition tests exercise the core without IO.
   - **ARCH-PURPOSE — flag:** BR-13’s promised call-graph sweep remains incomplete.
   - **ARCH-MOCK — pass:** stateful GitHub fixtures, real portable Git remotes, and existing live-conformance hooks are present.
   - **ARCH-CONSTRAINTS — flag:** BR-12 and BR-13 leave deadline gaps.
   - **ARCH-SECURE — pass:** resume validation and leases preserve identity and concurrent work.
   - **ARCH-ORDER — flag:** state ownership is corrected, but replay subprocesses still ignore operation cancellation.
   - **ARCH-FUNERAL — pass:** integration commits use branch lifecycle; temporary resources have cleanup paths.

7. **Plan revision recommendations**

   Add **“R7 — Complete context propagation across integration IO”** under `## Revisions`. Enumerate the facts probe, publish replay, ownership reads, transport, and cleanup; identify each context-bearing seam and regression test. Correct R6’s claim that the entire integration Git graph is already bounded and cancellable.

---

## Re-review — 2026-10-14T01:38:07-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 332 — sdlc merge integrates a moved main itself, then re-runs the suite before landing |
| repo | ariadne |
| issue file | workshop/issues/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9ea8e8a37b74f56317b1e4b0547411426b2f259e..96d1b05bb868cf36b6db5779702b104ad724b331 |
| command | sdlc milestone-close --issue 332 --milestone M2 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T01:38:07-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

M2’s integration, conflict recovery, leased push, bounded CI wait, retry limit, and interrupted-push recovery match the documented milestone. Both open findings are addressed, with regression tests independently shown to fail when fixes are removed. No new blocking findings. M3 failure classification remains outside this boundary.

```findings
dispose:
  - id: BR-12
    disposition: addressed
    note: |
      cigate_wait.go:79 enforces remaining deadlines and rejects late results; localchecks_make.go:70 propagates the budget into the real Makefile probe. Deadline tests pass. Removing probe context propagation makes both budget and cancellation regressions fail at approximately 1.53 seconds.
  - id: BR-13
    disposition: addressed
    note: |
      landinggit.go:61 bounds production Git processes and descendants; contextual helpers cover replay, archive ownership, and tracker discovery. Cancellation and uncertain-push recovery tests pass. Removing propagation independently fails archive ownership, tracker workspace/discovery, and all four replay-command regressions.
  - id: BR-14
    disposition: addressed
    note: |
      Previously accepted disposition remains valid. Production integration state advances through transitionIntegration; direct pure state tests and landing sequence tests pass.
```

1. **Strengths**

   - Resume validates parents, trailer, and reconstructed merge tree rather than trusting a commit message (`cmd/sdlc/landingintegrate.go:48`).
   - Push cancellation preserves possibly published work and supports reconciliation on retry (`cmd/sdlc/landingintegrate.go:227`).
   - Tests exercise real Git merges, bare remotes, suite subprocesses, and deterministic main movement.
   - README, help, atlas, and recovery contracts document the new behavior and the remaining server-side main race.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Focused deadline/process tests passed; M2 landing and legacy sequence tests passed.
   - Integration-state, publish, review-window, and landing-archive checks passed.
   - Full gitx package passed in 44.525 seconds.
   - Mutation checks failed as expected; disposable checkouts were removed.
   - Five broader archive failures reproduced at the pinned base. They resulted from repository-nested `TMPDIR`; all five passed with temporary fixtures outside the repository.
   - Full repository suite and live GitHub conformance were not run.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared facts preparation, replay primitives, and pure window decisions are reused.
   - **ARCH-PURE — pass:** integration transitions run directly without IO; effects remain in the shell.
   - **ARCH-PURPOSE — pass:** M2 delivers the integration workflow and completes the previously missed cancellation paths.
   - **ARCH-MOCK — pass:** stateful GitHub fakes and real isolated Git remotes exercise production seams; existing conformance coverage remains available.
   - **ARCH-CONSTRAINTS — pass:** polling deadlines include inflight work; subprocess ceilings and retry limits are tested.
   - **ARCH-SECURE — pass:** resume evidence and leases protect against fabricated markers and concurrent pushes; cancellation cannot authorize a late success.
   - **ARCH-ORDER — pass:** typed transitions govern integration; deterministic sequence tests cover main movement and uncertain publication.
   - **ARCH-FUNERAL — pass:** temporary probes/worktrees have cleanup paths; no new persistent journal is introduced.

7. **Plan revision recommendations:** None required for M2. Existing revisions describe the delivered entities and corrections.
