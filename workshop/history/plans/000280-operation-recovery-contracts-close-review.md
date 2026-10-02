# Boundary Review — ariadne#280 (whole-issue close)

| field | value |
|-------|-------|
| issue | 280 — Publish tested operation recovery contracts |
| repo | ariadne |
| issue file | workshop/issues/000280-operation-recovery-contracts.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3b7315b7fde5083f036e377d41e8053193e52ef3..61713c0fd8ca4d3eb9d6515e8e0978127784c8ca |
| command | sdlc close --issue 280 |
| reviewer | claude |
| timestamp | 2026-10-02T11:28:59-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed. For BR-7, the atlas page now has a **Scope:** line, and `sdlc help recovery` renders `recovery.Scope` through `{{RECOVERY_SCOPE}}` → `recovery.ScopeText()`. For BR-14, the plan's Core concepts now say `Verbs`, the DRY rationale no longer mentions JSON, and the issue's M1 row names `attachRecoveryContracts`. I checked every backticked identifier in the plan's Core concepts and Plan rows against the tree: all of them, including the 20+ proof test names, exist in Go sources. The code is small and correct. `cardPublish` is a faithful replacement: `UpdateCard` delegates to `UpdateCardWithTrailers` with nil trailers, and no direct `repo.UpdateCard(` callers remain outside tests. The contract test derives its required set from the command tree. The scheduling example runs from the same data that renders it. One new Minor finding remains: a value that only holds for the test fixture appears in the user-facing example as an expected value.

1. **Strengths**
   - `cmd/sdlc/cardpublish.go`: one seam and one wording for a lost acknowledgement. Reclaim's inline message is gone, and the guidance comes from one place.
   - `cmd/sdlc/recovery_contract_test.go:57`: the required verbs come from the repo-lock annotation, not a hand-kept list. The test also catches stale exemptions, missing help sections and proofs that name a test no file declares (found by an AST scan).
   - `cmd/sdlc/recovery_example_test.go`: steps the registry marks as convergent retries are delivered twice, and an unknown actor fails the test. The tracker is restored as soon as the offline step finishes.
   - `main.go` `attachRecoveryContracts`: one pass over the command tree, so no verb's help can miss its section.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `cmd/sdlc/internal/recovery/page.go` (the `checkpoints.flow.kind = quick` expectation): the example tells a coordinator to expect `quick`. Real delegated work with a durable plan reports `full`, and a coordinator following the example literally would take that for "no branch yet". The expectation should either be about presence or say that either flow kind is fine.

5. **Test coverage:** all proofs resolve to declared tests. The lost-acknowledgement reruns for claim and set-status, the second close starting a new generation, and the landing that leaves a reopened card alone are covered. The log records mutation checks for both the contract test and the example.

6. **Architecture**
   - **ARCH-DRY:** pass. One registry feeds the per-verb help, the topic and the tests.
   - **ARCH-PURE:** pass. `recovery` is pure (render, validate, `Lookup`); the IO stays in cmd tests.
   - **ARCH-PURPOSE:** pass. Every derived verb has a contract or a reasoned exemption, and the example is executed rather than only described.
   - **ARCH-MOCK:** pass. Tests use real git fixtures and a stubbed judge.
   - **ARCH-CONSTRAINTS:** pass. Help is rendered once per tree build, so cost is negligible.
   - **ARCH-SECURE:** pass. `Lookup` degrades to `(_, false)` on input it can't resolve.
   - **ARCH-ORDER:** pass. Nothing new holds state across events; the example tests a duplicate delivery and an unreachable tracker in order.
   - **ARCH-FUNERAL:** pass. Nothing durable is created outside test temp dirs.

7. **Plan revisions:** none needed; the plan matches the code.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      atlas/workflow/recovery-contracts.md has a Scope line; helptext/recovery.md renders {{RECOVERY_SCOPE}} via recovery.ScopeText() (page.go:20), and the per-verb Section shows the scope.
  - id: BR-14
    disposition: addressed
    note: |
      Plan Core concepts say `Verbs`, the DRY rationale drops JSON, and the issue M1 row names attachRecoveryContracts; every backticked identifier in Core concepts and Plan rows was grepped and resolves in the tree.
findings:
  - id: new
    severity: Minor
    family: fixture-value-as-guidance
    title: |
      Scheduling example tells coordinators to expect checkpoints.flow.kind = quick, which only the fixture produces
    detail: |
      Real delegated work with a durable plan is the full flow, so a coordinator following the rendered example would read a correct state as a mismatch. Assert presence (or accept quick or full) in recovery.Example, and let the test check the fixture's specific value separately.
```

---

## Re-review — 2026-10-02T11:31:09-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 280 — Publish tested operation recovery contracts |
| repo | ariadne |
| issue file | workshop/issues/000280-operation-recovery-contracts.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3b7315b7fde5083f036e377d41e8053193e52ef3..563c32ef7b32b5085c81a920a60f67a15eeaf82c |
| command | sdlc close --issue 280 |
| reviewer | claude |
| timestamp | 2026-10-02T11:31:09-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

BR-15 is fixed. In `cmd/sdlc/internal/recovery/page.go:71` the scheduling example now tells coordinators to expect `checkpoints.state = present` instead of `checkpoints.flow.kind = quick`. Its `Otherwise` text says the flow will read quick or full once `change-code` records it. The test (`cmd/sdlc/recovery_example_test.go:110-125`) still checks that this fixture recorded `quick`, but as the fixture's own fact rather than guidance. A `flowChecked` guard fails the test if no step reaches the checkpoints. `TestSchedulingExampleRuns` and the recovery package tests pass at HEAD.

One sibling of the same class remains at `page.go:76`. The close-observation step expects `checkpoints.reviews[close].verdict = SHIP`, which only holds because the test fixture stubs the judge to SHIP (`recovery_example_test.go:47`). A FIX-THEN-SHIP close also completes; the close code handles it at `close.go:1406`, and `catalog.go:102` covers it. So a coordinator could see a correct, completed close as a mismatch. The `Otherwise` text then sends them to `sdlc issue recovery reconcile`, which is wrong advice. This is Minor and does not block the gate.

1. **Strengths**
   - The BR-15 fix keeps the two concerns apart: guidance says what any successful run produces, and the test checks the fixture's own value separately (`recovery_example_test.go:110-116`).
   - The `flowChecked` guard keeps the fixture check reachable. If the step is ever reworded and stops looking at the checkpoints, the test fails instead of quietly passing.
   - The new `Otherwise` text explains what a coordinator should expect from the flow without hard-coding a value.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** the `page.go:76` close-verdict expectation of `SHIP` (detailed in the findings block).

5. **Test coverage notes:** The example test runs one fixture, so every `Expect` value in it is only proven for that one run. The structural fix is to run the example against each legal variant (quick and full flow, SHIP and FIX-THEN-SHIP verdict). A value that only one variant produces would then fail the test.

6. **Architecture:**
   - **ARCH-DRY:** pass.
   - **ARCH-PURE:** pass. `Example` is pure data, and the test drives it through in-process commands.
   - **ARCH-PURPOSE:** flagged under the BR-15 class. The previous round fixed the instance it named, but an enumerable sibling remains (the close verdict).
   - **ARCH-MOCK:** pass. The fixture uses a stubbed judge behind the existing seam.
   - **ARCH-CONSTRAINTS:** pass.
   - **ARCH-SECURE:** not applicable. This window reads no untrusted input and touches no secrets.
   - **ARCH-ORDER:** pass. Example steps are ordered and every observation is checked.
   - **ARCH-FUNERAL:** not applicable. This window creates nothing durable beyond review artifacts that are archived with the issue.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-15
    disposition: addressed
    note: |
      page.go:71 now expects checkpoints.state=present; the test checks the fixture's quick flow separately, guarded by flowChecked; TestSchedulingExampleRuns passes at HEAD.
findings:
  - id: new
    severity: Minor
    family: fixture-value-as-guidance
    title: |
      Close-observation step expects checkpoints.reviews[close].verdict = SHIP, which only the stubbed judge produces
    detail: |
      This is the 2nd finding in family fixture-value-as-guidance. Rule: every Expect value in recovery.Example must hold for every successful run, never just the fixture's run. Enforce it by running TestSchedulingExampleRuns over each legal variant (quick and full flow, SHIP and FIX-THEN-SHIP verdict), so a value that holds for only one variant fails. A FIX-THEN-SHIP close completes (close.go:1406, catalog.go:102), yet page.go:76 would read it as a mismatch and send the coordinator to reconcile. Expect a verdict that is present or non-blocking, and keep SHIP as a check on the fixture only.
```
