# Boundary Review — ariadne#391 (whole-issue close)

| field | value |
|-------|-------|
| issue | 391 — sdlc actuals fail machine-wide on a Qoder subagent transcript ('incomplete qoder session metadata') after #385 |
| repo | ariadne |
| issue file | workshop/issues/000391-sdlc-actuals-fail-machine-wide-on-a-qoder-subagent-transcript-incomplete-qoder-session-metadata-after-385.md |
| boundary | whole-issue close |
| milestone | — |
| window | 888e65ede88198060166e3b2e2cd0766e4b07e47..6bdd2b646b30f7c880c369c876a6eebb2483f615 |
| command | sdlc close --issue 391 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T14:07:52-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned diff fulfills #391’s skip-with-diagnostic approach. Unclassifiable Qoder sources no longer block unrelated measurements, while positively relevant damage and cancellation remain fatal. The implementation, tests, and atlas agree; no blocking findings surfaced.

```findings
{}
```

1. **Strengths**

   - Actuals and source selection share the correction through discovery, avoiding separate fixes in callers (`cmd/sdlc/internal/transcripts/discovery.go:54`).
   - File and directory failures retain path/reason diagnostics (`discovery.go:60`, `discovery.go:98`).
   - Native nested sidechain fixtures verify parent preservation, identity isolation, and actual/close adoption.
   - The ownership matrix covers malformed input and line/record limits both before and after workspace evidence.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Independently passed:

   - All `TestQoder` tests in `cmd/sdlc`, `internal/transcripts`, and `internal/agentid`.
   - The complete `internal/transcripts` suite.
   - `git diff --check`.

   Reviewed assertions cover successful adoption, known-workspace refusal, unreadable sources/directories, and cancellation. The recorded full suite and host-specific `actual --issue 387` result were not independently rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Both consumers retain shared discovery and classification.
   - **ARCH-PURE — pass:** The change stays within the existing filesystem boundary; it adds no business logic to command handlers.
   - **ARCH-PURPOSE — pass:** The fix covers the broader unclassifiable-source family, including traversal failures, rather than special-casing one subagent filename.

   The atlas documents the changed behavior. No new command, flag, configuration key, or installation step requires a README update. No Core concepts table is present.

7. **Plan revision recommendations:** None; the checked plan matches the delivered changes.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ddd4ec44-d44e-4d69-81b0-762489153feb","repository":"github.com/xianxu/ariadne","issue_id":"000391","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/0fd7ed945858b317","reviewed_head":"6bdd2b646b30f7c880c369c876a6eebb2483f615","slot":"ariadne:3","requested_at":"2026-10-15T21:06:33.4679Z"},"phase":"finished","pid":25528,"started_at":"2026-10-15T21:06:33.573301Z","finished_at":"2026-10-15T21:07:51.922788Z","session_id":"01a14163-5ecf-7c91-9d04-0d3ff68bc334","outcome":"returned"}] -->
