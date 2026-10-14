# Boundary Review — ariadne#366 (whole-issue close)

| field | value |
|-------|-------|
| issue | 366 — sdlc local-check: a slot whose ariadne substrate predates the weighted lock fails with a raw usage error; detect the version skew and name 'weave refresh' |
| repo | ariadne |
| issue file | workshop/issues/000366-sdlc-local-check-a-slot-whose-ariadne-substrate-predates-the-weighted-lock-fails-with-a-raw-usage-error-detect-the-version-skew-and-name-weave-refresh.md |
| boundary | whole-issue close |
| milestone | — |
| window | 407e48a39057503fab7cdecaf6c69ce338ce7631..5d13f3bc5d5a3d703dbb6b2a6edb2794daf22f60 |
| command | sdlc close --issue 366 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T10:15:57-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation covers stale-helper refusal and issue-branch substrate refresh well. Targeted SDLC tests and all targeted internal refresh tests passed, as did the three new CLI regressions. Approval remains blocked by missing README coverage and an existing CLI regression that cannot complete setup in this review environment.

```findings
findings:
  - id: new
    severity: Important
    family: user-facing-surface-documentation
    title: |
      README update is missing for weave refresh --substrate
    detail: |
      cmd/weave/refresh.go:67 introduces a user-facing flag, but README.md:214 still documents only ordinary refresh and --rebase. Add substrate-only usage, branch preservation, incompatible flags, and the source-owned-helper exception to satisfy the docs update gate.
  - id: new
    severity: Important
    kind: unverifiable
    family: required-regression-verification
    title: |
      Existing refresh retry regression cannot initialize under the review Git ceiling
    detail: |
      Running go test ./cmd/weave ./cmd/weave/internal/refresh -run '^(TestRefresh|TestSubstrate)' -count=1 -timeout 3m -v failed at cmd/weave/refresh_test.go:44: TestRefreshCLICompilesAndRetries calls compile before initializing its temporary repositories, and rev-parse --show-toplevel returns fatal: not a git repository under GIT_CEILING_DIRECTORIES. Provide passing host evidence for this regression and the required all-weave run at the pinned head, or make the fixture independent of enclosing repository discovery.
```

1. **Strengths**

   - Capability parsing distinguishes option declarations from prose and similarly named flags.
   - Preflight precedes worktree maintenance, acquisition, check execution, and GitHub publication.
   - Substrate-only refresh reuses the existing transaction and retains consumer revalidation.
   - Tests exercise deterministic races, private-supplier ownership, partial progress, retry, and transitive ordering.

2. **Critical findings**

   None.

3. **Important findings**

   The two findings above require resolution before admission.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Targeted SDLC tests passed: 24.958s.
   - Real-helper baseline and repeated-run reentry tests passed: 20.199s.
   - Targeted internal refresh tests passed: 61.294s.
   - All three new CLI regressions passed.
   - Existing CLI retry regression failed during fixture setup. The broader all-weave run was stopped without a completed result.
   - No independent full-suite result was established during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one helper preflight and one refresh transaction.
   - **ARCH-PURE — pass:** parsing remains pure; process and Git effects remain in integration code. Core-concept entries match the implementation.
   - **ARCH-PURPOSE — pass:** recovery works on divergent issue branches; source-owned helpers receive a distinct repair.
   - **ARCH-MOCK — pass:** portable Git fixtures, injected mutation hooks, and real isolated Python-helper checks exercise production boundaries.
   - **ARCH-CONSTRAINTS — pass:** help probes enforce time and combined-output bounds.
   - **ARCH-SECURE — pass:** shell quoting, resolved supplier identity, and credential-safe diagnostics are covered.
   - **ARCH-ORDER — pass:** existing transitions remain authoritative; race and partial-progress tests passed.
   - **ARCH-FUNERAL — pass:** no new durable artifact family; probe output and processes have bounded lifetimes.

7. **Plan revision recommendations**

   Add a `## Revisions` entry expanding Task 3 to include README documentation for `--substrate`. Record environment-qualified acceptance evidence for the existing refresh retry regression; the logged acceptance claim alone does not resolve the verification finding.
