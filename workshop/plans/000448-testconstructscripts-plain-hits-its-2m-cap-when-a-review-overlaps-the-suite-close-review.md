# Boundary Review — ariadne#448 (whole-issue close)

| field | value |
|-------|-------|
| issue | 448 — TestConstructScripts/plain hits its 2m cap when a review overlaps the suite |
| repo | ariadne |
| issue file | workshop/issues/000448-testconstructscripts-plain-hits-its-2m-cap-when-a-review-overlaps-the-suite.md |
| boundary | whole-issue close |
| milestone | — |
| window | 724addd9dcb79045ff15ecf934ee07daa86c52ba..3df9c09a2d08a9de0d01c24715758ee512e89a78 |
| command | sdlc close --issue 448 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T20:05:30-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation fixes the cumulative deadline scope while preserving canonical selection, process-group cancellation, and bounded pipe draining. Both changed packages pass. Two Important gaps remain: regression coverage for independent hang budgets and README documentation for the new flag.

Done-when: current — it covers per-script bounds, shared selection, compatibility, and lint enforcement. The recorded full-suite pass remains required before landing.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: behavioral-contract-regression-coverage
    title: |
      Independent script budgets lack a regression test
    detail: |
      scripts/test/scripts_test.go:49-68 implements the central fix, but construct_runner_test.go:50-89 exercises only the shell runner. No test distinguishes fresh per-script contexts from a context shared across the batch; the real scripts can pass under either implementation. The complete affected family is budget independence across selected scripts in plain, no_color_ci, and forced_color_terminal. Add a deterministic regression through an injected command/context seam that fails when the context is shared, while retaining the production two-minute ceiling.
  - id: new
    severity: Important
    kind: defect
    family: user-facing-surface-documentation
    title: |
      README update appears missing for construct runner --list
    detail: |
      scripts/test/construct-scripts.test.sh:15-21 introduces the --list flag, but README.md is unchanged in the pinned range and contains no usage for it. This is the sole new command-line flag in this window. Document its NUL-delimited, non-executing output and preserve the distinction from default batch execution; atlas/workflow/sdlc-binary.md already documents the architectural flow.
```

1. **Strengths**

   - Selection and exclusions remain owned by one shell runner.
   - Fixtures exercise actual shell behavior, including spaces/newlines, exclusions, listing without execution, and batch success/failure.
   - The extracted command helper retains the unchanged ceiling, process-group SIGKILL, and five-second pipe-drain bound.
   - Lint tests cover the original invocation shapes, import aliases, and justified suppression.

2. **Critical findings:** None.

3. **Important findings:** The two findings above require regression coverage and a README update.

4. **Minor findings:** None.

5. **Test coverage notes**

   `go test ./scripts/test ./pkg/golint -count=1 -v -timeout 10m` passed, including all eighteen script/environment combinations. The catalog scan passed over 1,196 tracked Go files. `git diff --check` passed. These targeted results do not establish the recorded full-suite acceptance criterion.

6. **Architectural notes**

   - **ARCH-DRY: pass** — Go consumes shell selection; CI retains the same selector.
   - **ARCH-PURE: pass** — AST matching is pure, and command execution remains a small IO helper.
   - **ARCH-PURPOSE: pass for implementation** — every selected script receives its own context; regression evidence needs the addition identified above.

7. **Plan revision recommendations**

   Add a `## Revisions` entry committing to deterministic verification of independent budgets across all three environments. Preserve the outstanding recorded final-head suite requirement.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a9d58fc7-6b85-409e-97b8-f2cf48c40bfc","repository":"github.com/xianxu/ariadne","issue_id":"000448","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/11ca7eb0cadd83e8","reviewed_head":"3df9c09a2d08a9de0d01c24715758ee512e89a78","slot":"ariadne:3","requested_at":"2026-10-19T03:03:41.795207Z"},"phase":"finished","pid":52599,"started_at":"2026-10-19T03:03:41.916844Z","finished_at":"2026-10-19T03:05:29.79983Z","session_id":"01a1521d-6b8a-7fc2-9a4b-3e8e0a45e1a6","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T20:11:29-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 448 — TestConstructScripts/plain hits its 2m cap when a review overlaps the suite |
| repo | ariadne |
| issue file | workshop/issues/000448-testconstructscripts-plain-hits-its-2m-cap-when-a-review-overlaps-the-suite.md |
| boundary | whole-issue close |
| milestone | — |
| window | 724addd9dcb79045ff15ecf934ee07daa86c52ba..25bf575191a753a127daf2bd1833cd5b1e27a6ec |
| command | sdlc close --issue 448 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T20:11:29-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the revised contract, and both prior findings are addressed. No new findings emerged. Targeted tests and the shared-context mutation check passed their expected outcomes. The recorded final-head full-suite pass remains a separate landing requirement.

Done-when: current — it covers independent unchanged budgets, process-group cancellation, canonical selection, batch compatibility, and lint enforcement.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TestConstructScriptBudgetsIndependent exercises the production loop across all three environments. A temporary Go overlay moving context ownership to the batch made it fail in every environment: three budgets instead of nine, with prior budgets still live. The pinned implementation passes.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:370-374 now documents --list as non-executing NUL-delimited selection and distinguishes default batch execution. This matches construct-scripts.test.sh:18-41 and its real-shell fixture tests.
```

1. **Strengths**

   - Selection and exclusions retain one owner; Go consumes the shell runner’s listing (`scripts/test/scripts_test.go:60`).
   - Each command gets its own unchanged ceiling; process-group cancellation and bounded draining remain intact (`scripts/test/scripts_test.go:75-102`).
   - Real-shell fixtures cover unusual paths, exclusions, listing without execution, batch success/failure, unknown arguments, and empty discovery.
   - README and atlas document the new flag and execution flow.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   `go test ./scripts/test ./pkg/golint -count=1 -timeout 10m` passed: scripts 78.522s; golint 0.479s. The shared-context mutation failed as expected. `git diff --check` passed. These checks do not establish the recorded full-suite acceptance criterion.

6. **Architectural notes**

   - **ARCH-DRY: pass** — Go and CI use the canonical shell selection; lint reuses existing AST helpers.
   - **ARCH-PURE: pass** — AST matching stays pure; command execution is injected through a small IO seam.
   - **ARCH-PURPOSE: pass** — every selected script in all three environments receives an independent budget, with regression evidence distinguishing the original defect.

7. **Plan revision recommendations:** None. Existing revisions accurately describe the implementation and outstanding landing verification.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cace75ec-953b-45f7-8b04-ef4b26eb7a74","repository":"github.com/xianxu/ariadne","issue_id":"000448","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/984533fb2d66b289","reviewed_head":"25bf575191a753a127daf2bd1833cd5b1e27a6ec","slot":"ariadne:3","requested_at":"2026-10-19T03:09:40.239746Z"},"phase":"finished","pid":25703,"started_at":"2026-10-19T03:09:40.415693Z","finished_at":"2026-10-19T03:11:26.419898Z","session_id":"01a15222-e413-7bd1-959e-f23bc8e18e9a","outcome":"returned"}] -->
