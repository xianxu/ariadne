# Boundary Review — ariadne#494 (whole-issue close)

| field | value |
|-------|-------|
| issue | 494 — Train car's own merge warns 'active issue missing without confirmed archive' after the train already archived it |
| repo | ariadne |
| issue file | workshop/issues/000494-train-car-s-own-merge-warns-active-issue-missing-without-confirmed-archive-after-the-train-already-archived-it.md |
| boundary | whole-issue close |
| milestone | — |
| window | dc76211e27578aca7aa62d434db9ca2bd43eed5b..56a0a9c84276d7bb914673cec39fc6983424f55d |
| command | sdlc close --issue 494 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T06:57:47-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned change fulfills the revised Spec: a validated train outcome enables archive discovery, while the shared verifier supplies the archive proof. Both focused test runs passed. No blocking findings; the separately required recorded-suite pass remains a landing prerequisite.

```findings
{}
```

1. **Strengths**
   - Retains the validated waiter outcome through cleanup (`cmd/sdlc/landing.go:420`).
   - Reuses the complete archive-transition verifier across both provenance paths (`cmd/sdlc/landingarchive.go:731`).
   - Exercises two real merge callers with deterministic cleanup ordering, including output, resting-branch preservation, and ref cleanup (`cmd/sdlc/landingtrain_archive_test.go:21`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Passed archive proof controls and the two-car regression (77.200s); passed no-outcome recovery and tracked serial landing checks across merge, squash, and rebase (53.289s). `git diff --check` passed. The full recorded suite was not run during this review.

6. **Architecture:** ARCH-DRY passes through verifier reuse. ARCH-PURE passes: history lookup remains in the existing injectable integration boundary. ARCH-PURPOSE passes: the actual waiter failure and safe cleanup are exercised. ARCH-EVOLVE passes: no persisted format changes. The atlas documents the flow; no new runnable surface requires a README change.

7. **Plan revisions:** None required.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3de28267-02fc-40b6-88b1-b4c7a977d59e","repository":"github.com/xianxu/ariadne","issue_id":"000494","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/290a1e60eb02aaa2","reviewed_head":"56a0a9c84276d7bb914673cec39fc6983424f55d","slot":"ariadne:1","requested_at":"2026-10-20T13:54:47.73622Z"},"phase":"finished","pid":11561,"started_at":"2026-10-20T13:54:47.903612Z","finished_at":"2026-10-20T13:57:45.832965Z","session_id":"01a15997-e170-7552-803f-7f29c8b22307","outcome":"returned"}] -->
