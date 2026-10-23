# Boundary Review — ariadne#569 (whole-issue close)

| field | value |
|-------|-------|
| issue | 569 — Diagnose legacy actual-window refusal blocking pair#162 |
| repo | ariadne |
| issue file | workshop/issues/000569-legacy-actual-window.md |
| boundary | whole-issue close |
| milestone | — |
| window | b10a950e2693b79fe69b1d4279f1b752c9d48162..96cd72ed248907d4cdee3a026054be87cbd47a69 |
| command | sdlc close --issue 569 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T02:07:20-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation matches the revised scope: exact old unmarked imports can establish temporal disjointness without certifying their historical completion or actual. No blocking defects found. README, help, and atlas describe the behavior and its limits.

Done-when: current — the clauses cover labelled disjointness, strict refusal controls, hybrid regression coverage, and the landing/readiness sequence.

```findings
{}
```

1. **Strengths**
   - Shared source and generation validation avoids a second history walker (`cmd/sdlc/legacycompletion.go:36`).
   - Bound-only mode returns before completion authority is populated (`cmd/sdlc/legacycompletion.go:197`).
   - The exclusion path requires strict temporal separation and emits an issue/end warning (`cmd/sdlc/actualexclusion.go:323`).
   - Real-Git tests cover restored generations and verify that disjointness leaves refs and checkout unchanged (`cmd/sdlc/legacydisjoint_test.go:59`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**
   - New disjointness and hybrid measurement regressions passed normally and under race detection.
   - Strict damaged-evidence, unmarked-import, history-limit, history-read-failure, and pure generation-selection controls passed.
   - `git diff --check` passed; checkout remained clean.
   - The full recorded suite was not rerun during this review.

6. **Architectural notes**
   - **ARCH-DRY: pass** — immutable history validation remains shared.
   - **ARCH-PURE: pass** — temporal comparison and generation selection remain pure; Git reads stay in the integration layer.
   - **ARCH-PURPOSE: pass** — covers the full earliest prefix/engagement boundary and preserves overlap refusal.
   - **ARCH-EVOLVE: pass** — no persisted format changes or new rewrite paths.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"76de1c9f-4d21-4ad9-8008-975baebf257e","repository":"github.com/xianxu/ariadne","issue_id":"000569","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/38cb31914bca19b2","reviewed_head":"96cd72ed248907d4cdee3a026054be87cbd47a69","slot":"ariadne:3","requested_at":"2026-10-23T09:05:06.531615Z"},"phase":"finished","pid":52738,"started_at":"2026-10-23T09:05:06.625098Z","finished_at":"2026-10-23T09:07:20.249297Z","session_id":"01a16801-bd55-7ad2-9952-0cd55678e92a","outcome":"returned"}] -->
