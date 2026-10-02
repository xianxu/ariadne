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
