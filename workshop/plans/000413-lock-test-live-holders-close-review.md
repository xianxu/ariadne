# Boundary Review — ariadne#413 (whole-issue close)

| field | value |
|-------|-------|
| issue | 413 — Concurrent dead-holder lock test marks its live contenders dead and flakes with missing metadata |
| repo | ariadne |
| issue file | workshop/issues/000413-lock-test-live-holders.md |
| boundary | whole-issue close |
| milestone | — |
| window | 94aae475501975b79bff7772e787a79efa1900d4..80dc826eab72c75f93a56a9265ce4b65359b638e |
| command | sdlc close --issue 413 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T19:00:07-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes fulfill #413’s revised Spec and Plan. Reclamation now rejects stale observations of replacement holders, serializes reapers, and preserves cancellation and timeout limits. No blocking findings.

```findings
{}
```

1. **Strengths**
   - Reuses the existing flock helper and observation logic; the descriptor guard closes on every return (`repolock.go:241`).
   - Deterministic delayed-observation coverage verifies that a replacement live holder survives (`repolock_test.go:445`).
   - Tests cover busy guards, cancellation, deadline expiry, inode replacement, initialization, ownership and cleanup.
   - Atlas accurately documents the new recovery mechanism. No new user-facing command or configuration requires a README update.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage**
   - Reclamation package passed 30 repetitions and 10 race-enabled repetitions.
   - Targeted CLI integration tests passed.
   - An unchanged isolation test failed because the harness nests temporary directories inside the checkout; excluding that unrelated check resolved the selection.
   - Inspected the recorded full-suite log: all shards passed. Only issue prose changed afterward.

6. **Architecture**
   - **ARCH-DRY: Pass.** Existing flock and observation behavior reused.
   - **ARCH-PURE: Pass.** Filesystem coordination remains in the integration helper, with clock and liveness seams injected.
   - **ARCH-PURPOSE: Pass.** Addresses both the inaccurate fixture and the demonstrated production race.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0e1a598f-1725-4025-b8ba-9199cc55b4db","repository":"github.com/xianxu/ariadne","issue_id":"000413","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2bdcfa2d8847475f","reviewed_head":"80dc826eab72c75f93a56a9265ce4b65359b638e","slot":"ariadne:5","requested_at":"2026-10-17T01:58:35.345028Z"},"phase":"finished","pid":5501,"started_at":"2026-10-17T01:58:35.565402Z","finished_at":"2026-10-17T02:00:06.628961Z","session_id":"01a14795-1989-7ef0-b217-a3c133314b0a","outcome":"returned"}] -->
