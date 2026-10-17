# Boundary Review — ariadne#429 (whole-issue close)

| field | value |
|-------|-------|
| issue | 429 — tl-lead status check counts issue-filing/archive commits as landings, flagging Threads/Slots stale with nothing landed |
| repo | ariadne |
| issue file | workshop/issues/000429-tl-lead-status-check-counts-issue-filing-archive-commits-as-landings-flagging-threads-slots-stale-with-nothing-landed.md |
| boundary | whole-issue close |
| milestone | — |
| window | c3477aa9b0b1fb9069f0a301b011e46b4c2b4409..7c6717dc1c3c2f923bfc8b981b9d752ab1b2b61c |
| command | sdlc close --issue 429 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T13:10:30-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned diff satisfies both Done-when clauses: direct maintenance commits no longer count as landings, merged PRs do, and stale diagnostics identify the repository, merge commit, and time. No blocking defects or missing scope found.

```findings
{}
```

1. **Strengths**
   - `cmd/tl-lead/freshness.go:140` uses commit topology rather than message heuristics, matching both SDLC merge paths’ `--merge` behavior.
   - `cmd/tl-lead/freshness_test.go:128` exercises real Git histories, including direct commits, misleading subjects, nested merges, excluded repositories, and missing refs.
   - `cmd/tl-lead/freshness.go:72` renders the selected merge identity through the existing shared freshness path.
   - Help and `atlas/workflow/tl-skills.md` consistently describe the revised rule.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** `go test ./cmd/tl-lead -count=1` passed. Both landing/non-landing cases and diagnostic identity are covered; existing tests retain same-minute and error-reporting coverage. `git diff --check` passed. The recorded repository-wide suite was not independently rerun.

6. **Architectural notes**
   - **ARCH-DRY — pass:** Landing selection and stale rendering retain their existing shared owners; all status consumers use that path.
   - **ARCH-PURE — pass:** Git IO remains behind the injected landing seam; this diff adds no IO to freshness decisions.
   - **ARCH-PURPOSE — pass:** The change addresses the false-landing class and both documented acceptance clauses.
   - No new command, flag, configuration, or installation surface requires a README update. Atlas coverage is present.

7. **Plan revision recommendations:** None. No Core concepts table is present to reconcile.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5b726ef4-3677-4612-9ee1-413359ebe1e6","repository":"github.com/xianxu/ariadne","issue_id":"000429","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/83e8a7aa7ea68a2d","reviewed_head":"7c6717dc1c3c2f923bfc8b981b9d752ab1b2b61c","slot":"ariadne:5","requested_at":"2026-10-17T20:09:53.636149Z"},"phase":"finished","pid":4112,"started_at":"2026-10-17T20:09:53.759098Z","finished_at":"2026-10-17T20:10:29.384006Z","session_id":"01a14b7c-3698-7562-b2ed-a5b7cc9325dc","outcome":"returned"}] -->
