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

---

## Re-review — 2026-10-21T03:24:57-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 510 — sdlc local-check cancel: select one run by id (not only --name); notice when queuing a second run of the same check on a different head |
| repo | ariadne |
| issue file | workshop/issues/000510-sdlc-local-check-cancel-select-one-run-by-id-not-only-name-notice-when-queuing-a-second-run-of-the-same-check-on-a-different-head.md |
| boundary | whole-issue close |
| milestone | — |
| window | af071ef20b003892e2c8c43766ffc964607a3e81..dcc50b776875f0dd40a0d2639a5fed3a79063d5b |
| command | sdlc close --issue 510 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T03:24:57-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff correctly resolves both documentation conflicts, preserving main’s landing-scope guidance and exact run-ID cancellation instructions. No executable code or contract sections changed. No defects found within this review window.

Done-when: current — it still covers exact cancellation, sibling FIFO preservation, foreign-ID refusal, and duplicate notices.

```findings
{}
```

1. **Strengths**
   - [README.md:606](/Volumes/TestScratch/review-checkouts/91ebb1512fd7d054/README.md:606) retains landing-check guidance alongside exact cancellation instructions.
   - [local-check.md:262](/Volumes/TestScratch/review-checkouts/91ebb1512fd7d054/cmd/sdlc/helptext/local-check.md:262) correctly describes `--all` as head-scoped and preserves the `cancel --run` example. Both match the command implementation.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Five focused tests passed in 3.493s, covering landing selection/round-trip, exact cancellation with sibling preservation, and duplicate-notice controls/linear reads. `git diff --check` passed. Full suite not run.

6. **Architecture:** ARCH-DRY passes: redundant conflict text removed. ARCH-PURE passes: no logic/IO changes. ARCH-PURPOSE passes: both sides’ documentation intent survives. ARCH-EVOLVE passes: no persisted-format changes. No new surface requires an atlas update.

7. **Plan revisions:** None required. No prior findings to dispose of.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7314d786-9a09-4a4e-9f7f-b264a73789bf","repository":"github.com/xianxu/ariadne","issue_id":"000510","recipe":"fix-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/91ebb1512fd7d054","reviewed_head":"dcc50b776875f0dd40a0d2639a5fed3a79063d5b","slot":"ariadne:2","requested_at":"2026-10-21T10:24:06.77887Z"},"phase":"finished","pid":33867,"started_at":"2026-10-21T10:24:07.062397Z","finished_at":"2026-10-21T10:24:56.406469Z","session_id":"01a15dfd-5b8d-7e71-8e41-6525e6ba1bad","outcome":"returned"}] -->
