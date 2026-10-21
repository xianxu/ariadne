# Boundary Review — ariadne#503 (whole-issue close)

| field | value |
|-------|-------|
| issue | 503 — flake: TestVouchingStartsWhenTrainsTurnOn |
| repo | ariadne |
| issue file | workshop/issues/000503-flake-testvouchingstartswhentrainsturnon.md |
| boundary | whole-issue close |
| milestone | — |
| window | e5c65cacc9942b245562e4cf9d9b00ab4aad8a86..a99969209d1387ede6d2d71e297bff9c4b218abb |
| command | sdlc close --issue 503 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T13:15:11-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The change fixes the fixture’s cancellation race by separating callback entry from persisted publication. Both cancellation orderings are exercised, production cancellation behavior remains intact, and cleanup joins the worker on every exit. No blocking findings.

```findings
{}
```

1. **Strengths**

   - Channels force the callback/publication gap deterministically (`cmd/sdlc/landingcar_test.go:340`).
   - The successful path reads persisted readiness before cancelling; the early-cancellation path verifies readiness remains unpublished (`cmd/sdlc/landingcar_test.go:371`).
   - Deferred cleanup cancels, releases the callback, and joins the worker before releasing its ticket (`cmd/sdlc/landingcar_test.go:356`).
   - Reuses `watchWriter` and `Line.Readiness` without adding production hooks.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Focused regression passed **100 repetitions**.
   - Regression, waiting-loop integration, and shared-schedule controls passed **25 repetitions under `-race`**.
   - `git diff --check` passed; checkout remained clean.
   - Verified the trains-off assertion and both cancellation orderings against production code. The related counter-pattern sweep found no sibling requiring the same repair.
   - Full recorded suite was not run during this review; its passing result remains the documented landing requirement.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing synchronization and persistence readers are reused.
   - **ARCH-PURE — pass:** integration controls remain in the test fixture; no business logic moved into IO.
   - **ARCH-PURPOSE — pass:** addresses the publication race and preserves cancellation before publication.
   - **ARCH-EVOLVE — pass:** no persisted formats or rewrite paths changed.
   - No new architectural or user-facing surface requires atlas or README updates.

7. **Plan revision recommendations:** None; the current Plan and revisions match the implementation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e08bc937-4388-4396-8004-d689ef2ea188","repository":"github.com/xianxu/ariadne","issue_id":"000503","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/557be9722904614e","reviewed_head":"a99969209d1387ede6d2d71e297bff9c4b218abb","slot":"ariadne:5","requested_at":"2026-10-20T20:13:56.129913Z"},"phase":"finished","pid":64051,"started_at":"2026-10-20T20:13:56.258097Z","finished_at":"2026-10-20T20:14:56.366313Z","session_id":"01a15af2-fdfd-7e52-98fe-8297c3d535d2","outcome":"returned"}] -->
