# Boundary Review — ariadne#491 (whole-issue close)

| field | value |
|-------|-------|
| issue | 491 — Suite queue position becomes fictitious on allocation contention |
| repo | ariadne |
| issue file | workshop/issues/000491-queue-contention-position.md |
| boundary | whole-issue close |
| milestone | — |
| window | 702b01324fa543038178e23b831d3e3346c1204a..0b60e8b7e74e9ce5fff56a4d8b5963515dd6f228 |
| command | sdlc close --issue 491 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T06:19:23-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned change fixes fictitious queue positions without changing admission decisions or lock behavior. Regression coverage exercises unavailable order, recovery, FIFO admission, stale diagnostics, and log stability. No blocking findings.

Done-when: current — its clauses still cover contention reporting, recovery using the same ticket, and suppression of fictitious position updates.

```findings
{}
```

1. **Strengths**

   - `scripts/test-lock.py:615–622` distinguishes missing evidence (`None`) from an observed empty predecessor list.
   - `scripts/test-lock.py:928–951` caches observed diagnostics, explicitly labels stale reuse, and excludes freshness from the reporting key.
   - `scripts/test/test-lock-pause.test.py:293` uses real allocation contention and verifies the same ticket subsequently retains FIFO admission.
   - `atlas/workflow/sdlc-binary.md:2330` accurately documents unknown and last-observed reporting.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - All 18 pause tests passed.
   - Go integration tests `TestQueueRealHelperPause` and `TestSuiteQueuePriorityRealHelper` passed.
   - Loading the base implementation against the new regression tests produced failures, confirming regression sensitivity.
   - Standalone priority execution lacked `bin/sdlc`; the successful Go harness supplied the required provider.
   - Range whitespace checks passed; the worktree remains clean. Full-suite execution was outside this targeted review.

6. **Architectural notes**

   - **ARCH-DRY: pass** — reuses the scheduler and reporting interval.
   - **ARCH-PURE: pass** — adds bounded diagnostic state at the existing IO boundary; admission is injected for renderer tests.
   - **ARCH-PURPOSE: pass** — both unobserved producer paths and the production renderer preserve uncertainty.
   - **ARCH-EVOLVE: pass** — no persisted-format changes.

   Atlas coverage is present. No new command, flag, configuration key, or usage step requires a README update.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e973fa4f-a54a-4c0f-baa2-e195fd0cac80","repository":"github.com/xianxu/ariadne","issue_id":"000491","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/8dd3f0f6b4730fb0","reviewed_head":"0b60e8b7e74e9ce5fff56a4d8b5963515dd6f228","slot":"ariadne:3","requested_at":"2026-10-20T13:17:26.899689Z"},"phase":"finished","pid":76461,"started_at":"2026-10-20T13:17:27.080968Z","finished_at":"2026-10-20T13:19:22.384872Z","session_id":"01a15975-b058-7001-bccd-e11cd159aa64","outcome":"returned"}] -->
