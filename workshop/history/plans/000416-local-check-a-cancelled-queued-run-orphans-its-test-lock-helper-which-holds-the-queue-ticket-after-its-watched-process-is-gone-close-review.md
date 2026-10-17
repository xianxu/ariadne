# Boundary Review — ariadne#416 (whole-issue close)

| field | value |
|-------|-------|
| issue | 416 — local-check: a cancelled queued run orphans its test-lock helper, which holds the queue ticket after its watched process is gone |
| repo | ariadne |
| issue file | workshop/issues/000416-local-check-a-cancelled-queued-run-orphans-its-test-lock-helper-which-holds-the-queue-ticket-after-its-watched-process-is-gone.md |
| boundary | whole-issue close |
| milestone | — |
| window | aa856d15a3b5037eab593919a089578df56f2a4f..f27e0f72d125f70f5aa0a02c7eeab1fdc3226f33 |
| command | sdlc close --issue 416 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T22:00:29-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The queued cancellation and dead-watch fixes pass their focused regressions. One broader regression blocks shipping: taking signal ownership for the entire command makes PID-targeted SIGTERM ineffective while existing context-free subprocess calls block.

Done-when: current — the revision clarifies verification timing without changing the required cancellation behaviors.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: signal-ownership-requires-cancellable-io
    title: |
      Command-wide signal ownership suppresses termination during context-free subprocess calls
    detail: |
      cmd/sdlc/main.go:83 enables NotifyContext for all of local-check run, but its production Git/network preamble still uses subprocesses that ignore that context. Using the pinned-head test binary's real CLI entry point with a private Git fixture blocked in fetch, the CLI remained alive after SIGTERM and swallowed a second SIGTERM; the fixture required SIGKILL cleanup. The affected family comprises RepoTopLevel and detectRepo queries, the localcheck_cmd.go Git closure and its consumers (declaration/default-suite reads, revision/branch checks, log-directory discovery, worktree add/remove/prune, and trailer reads), workspace.Resolve's execGitRunner binding, and PRListForBranch. Make these subprocess seams cancellable through existing bounded context-aware helpers, give cleanup a separate bounded context, and add production-command signal regressions. The new signal test replaces RunE at localcheck_signal_unix_test.go:42, bypassing the demonstrated preamble failure.
```

1. **Strengths:** The helper checks liveness before allocation and during queue polling, reusing existing ticket cleanup (`scripts/test-lock.py:231,258`). Helper cancellation reuses process-group termination and bounded pipe draining (`cmd/sdlc/localcheck_run.go:289`). Tests cover both SIGINT and SIGTERM, successor admission, and preservation of an unrelated lock holder.

2. **Critical findings:** The finding above originates at `cmd/sdlc/main.go:83`. Signal ownership must cover the command’s blocking subprocess seams before this broader interception is safe.

3. **Important findings:** None. Atlas documents the lifecycle change. No new command, flag, or configuration surface requires a README addition.

4. **Minor findings:** None.

5. **Test coverage:** Passed:
   - `go test -trimpath ./cmd/sdlc -run '^Test(LocalCheckSignal|RunLocalCheck|CLISignal)' -count=1 -timeout 5m`
   - `bash scripts/test/test-lock.test.sh`

   Independently reproduced the SIGTERM regression through the production command entry point. The recorded full-suite and CI claims were not independently rerun.

6. **Architecture:** **ARCH-DRY: pass** — existing process-group helpers and ticket cleanup are reused. **ARCH-PURE: pass** — changes stay in the IO boundary. **ARCH-PURPOSE: flag** — queue cancellation works, but command-wide signal ownership leaves sibling subprocess consumers unable to unwind.

7. **Plan revision recommendation:** Add a `## Revisions` entry covering cancellation across the enumerated subprocess seams, bounded cleanup after cancellation, and regression tests through the unmodified production command handler.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b3e20790-c4ed-48bf-8d7f-971ec8f5ad52","repository":"github.com/xianxu/ariadne","issue_id":"000416","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/b8b4eb203c1d0a28","reviewed_head":"f27e0f72d125f70f5aa0a02c7eeab1fdc3226f33","slot":"ariadne:5","requested_at":"2026-10-17T04:58:52.058548Z"},"phase":"finished","pid":82469,"started_at":"2026-10-17T04:58:52.163541Z","finished_at":"2026-10-17T05:00:27.548053Z","session_id":"01a1483a-24a6-7953-bd97-284b15bc9b98","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T00:40:55-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 416 — local-check: a cancelled queued run orphans its test-lock helper, which holds the queue ticket after its watched process is gone |
| repo | ariadne |
| issue file | workshop/issues/000416-local-check-a-cancelled-queued-run-orphans-its-test-lock-helper-which-holds-the-queue-ticket-after-its-watched-process-is-gone.md |
| boundary | whole-issue close |
| milestone | — |
| window | d1a4e571ebdc8e709b4d22607de0e53810e02cb3..412b0dc4db97e5531f54ebd596b37627b9b40459 |
| command | sdlc close --issue 416 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T00:40:55-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes satisfy the cancellation requirements and revised plan. BR-1 is addressed: production-handler regressions pass, and temporarily undoing signal scoping and bounded cleanup makes the relevant tests fail. No new findings. The checkout was restored and is clean.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Signal ownership is limited to queue/suite execution; all three production constructors use independent bounded cleanup. Production-handler tests pass. Restoring command-wide signal ownership reproduces swallowed SIGINT/SIGTERM in both fetch and trailer reads; removing bounded cleanup reproduces cancellation exceeding five seconds.
```

1. **Strengths**
   - Signal ownership follows resource lifetime (`cmd/sdlc/localcheck_run.go:108,148`).
   - Dead-watch checks reuse existing ticket cleanup (`scripts/test-lock.py:231,258`).
   - Private fixtures verify successor admission and preservation of unrelated holders.
   - Cleanup reuses the existing bounded Git runner across all production constructors.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Passed the focused `LocalCheckSignal`, `RunLocalCheck`, and `CLISignal` Go tests (65.980s), plus the complete `scripts/test/test-lock.test.sh` suite. Mutation verification produced the expected five failures. Full-suite and CI claims were not independently rerun.

6. **Architecture:** **ARCH-DRY: pass**—shared process-group, cleanup, and liveness helpers. **ARCH-PURE: pass**—changes remain in the IO shell with injected Git seams. **ARCH-PURPOSE: pass**—both cancellation paths, ownership preservation, successor admission, and default termination outside the owned phase are covered. Atlas is updated; no new README-facing surface was introduced.

7. **Plan revisions:** None required; the existing BR-1 revision matches the implementation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"86f7b099-0944-4c63-949c-4d7979270099","repository":"github.com/xianxu/ariadne","issue_id":"000416","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/ec385bd069c88189","reviewed_head":"412b0dc4db97e5531f54ebd596b37627b9b40459","slot":"ariadne:5","requested_at":"2026-10-17T07:38:21.485816Z"},"phase":"finished","pid":35597,"started_at":"2026-10-17T07:38:21.661626Z","finished_at":"2026-10-17T07:40:54.598834Z","session_id":"01a148cc-2a2c-7701-ab08-82b21a857815","outcome":"returned"}] -->
