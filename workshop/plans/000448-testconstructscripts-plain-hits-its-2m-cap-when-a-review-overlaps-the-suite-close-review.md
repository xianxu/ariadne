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
