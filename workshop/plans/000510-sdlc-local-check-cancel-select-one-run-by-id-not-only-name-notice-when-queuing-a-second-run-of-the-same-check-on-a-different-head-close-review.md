# Boundary Review — ariadne#510 (whole-issue close)

| field | value |
|-------|-------|
| issue | 510 — sdlc local-check cancel: select one run by id (not only --name); notice when queuing a second run of the same check on a different head |
| repo | ariadne |
| issue file | workshop/issues/000510-sdlc-local-check-cancel-select-one-run-by-id-not-only-name-notice-when-queuing-a-second-run-of-the-same-check-on-a-different-head.md |
| boundary | whole-issue close |
| milestone | — |
| window | fb6c4e165c9f3dd37e43ad248e8fa57d343f04fd..8c3b5c407ca83d5ebc2e8538104b06c9aa65b8cf |
| command | sdlc close --issue 510 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T02:52:26-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range delivers exact run-ID cancellation and advisory duplicate notices while retaining checkout ownership and PID-generation safeguards. No new defects found. README, help, and atlas cover the new surface.

Done-when: current — its clauses still cover exact cancellation, sibling preservation, foreign-ID refusal, and duplicate notices after the discovery-premise correction.

```findings
{}
```

1. **Strengths:** Shared lease reader avoids parallel cancellation/notice implementations (`cmd/sdlc/localcheck_lease.go:113`). Empty and conflicting selectors refuse (`cmd/sdlc/localcheck_cmd.go:52`). The regression verifies sibling lease and FIFO-ticket identity, contents, and subsequent admission (`cmd/sdlc/localcheck_runid_unix_test.go:187`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** The new queue regression, duplicate-notice controls, linear-read checks, and existing PID-generation, stale-lease, publication, and lifetime checks passed. A broader selection exceeded my 120-second timeout in the unchanged cleanup of `TestLocalCheckCancelOnlyOwnRun`; the stack confirms cleanup was waiting on an already-consumed child-result channel. Full suite not run.

6. **Architecture:** ARCH-DRY passes through shared lease observation and classification. ARCH-PURE passes: deterministic ownership classification remains separate from lease/process IO. ARCH-PURPOSE passes: exact selection preserves the sibling’s queue position, and notices offer an explicit replacement action. ARCH-EVOLVE passes: persisted lease fields and meanings are unchanged; no rewrite path was introduced.

7. **Plan revisions:** None required. No prior findings to dispose of.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"38eff78f-b6ff-4e49-ab94-80e8c3989090","repository":"github.com/xianxu/ariadne","issue_id":"000510","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/b4e237cce6cabd40","reviewed_head":"8c3b5c407ca83d5ebc2e8538104b06c9aa65b8cf","slot":"ariadne:2","requested_at":"2026-10-21T09:49:24.485176Z"},"phase":"finished","pid":45875,"started_at":"2026-10-21T09:49:24.590385Z","finished_at":"2026-10-21T09:52:25.598394Z","session_id":"01a15ddd-9400-7443-b9f6-7ea03f1bd1b1","outcome":"returned"}] -->
