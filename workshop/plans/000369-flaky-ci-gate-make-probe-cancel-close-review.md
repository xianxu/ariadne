# Boundary Review — ariadne#369 (whole-issue close)

| field | value |
|-------|-------|
| issue | 369 — TestWaitCIGateBoundsRealMakeProbe/cancel is flaky on main: cancel sometimes returns after ~1.05s (the make child's sleep) instead of under 700ms; also TestSetupLeaseSurvivesRunnerParentDeath failed once under the full suite |
| repo | ariadne |
| issue file | workshop/issues/000369-flaky-ci-gate-make-probe-cancel.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9f600c11c3e73bfcf1d3b5085084720bbfd67116..b97208e5484273f0cc54d44be9b8257d5b0cb584 |
| command | sdlc close --issue 369 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T13:48:10-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range satisfies the issue’s Spec and Plan. Cancellation retains the existing timing limits, the regressions detect the targeted failures, and the recorded full-suite success covers all changed code. No blocking findings.

```findings
{}
```

1. **Strengths**

   - Bounded repeat signalling handles a successful first signal that misses a pipe-holding descendant (`cmd/sdlc/internal/processgroup/process_unix.go:37`).
   - Successful cancellation preserves the nil result required by `exec.Cmd.Cancel`; dying-group behavior remains covered.
   - Lease fixtures keep the child alive while proving resource release, preventing exit cleanup from masking an omitted close (`cmd/weave/internal/weavefs/runner_setup_test.go:127`).
   - Suite summaries retain assertion output belonging to failed subtests (`scripts/test-shard.py:143`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Both make-probe modes passed 50 repeats, preserving the 700ms assertion.
   - Both setup-lease modes and the inherited-lease test passed 50 repeats.
   - Complete processgroup, procgroup, and judge package tests passed.
   - Disposable-copy mutations made all four targeted regressions fail: single-pass termination, incorrect cancellation return, discarded subtest diagnostics, and omitted lease closes.
   - Verified `sdlc-local/suite` success on `b74440ee`, including its recorded log hash. The remaining range changes only issue documentation.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Shared termination continues through `processgroup` and the existing kernel-result interpreter.
   - **ARCH-PURE — pass:** Process signalling and clock use remain in the integration boundary, with an injected signalling seam.
   - **ARCH-PURPOSE — pass:** The fix addresses surviving descendants without loosening production timeouts. Both lease modes and the related inherited-lease fixture are covered.
   - No new architectural or user-facing surface requires atlas or README updates.

7. **Plan revision recommendations:** None. The checked Plan items match the implementation and verified evidence.
