# Boundary Review — ariadne#511 (whole-issue close)

| field | value |
|-------|-------|
| issue | 511 — landing trains: don't restart CI when main advances only by issue-detail commits (filings collided with train 197 twice on 10-20) |
| repo | ariadne |
| issue file | workshop/issues/000511-landing-trains-don-t-restart-ci-when-main-advances-only-by-issue-detail-commits-filings-collided-with-train-197-twice-on-10-20.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8666782f29dd3829d7493b6d89e4436ab4537adf..71d9a6ae50fe907a400ad899372ed073f02613be |
| command | sdlc close --issue 511 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T03:59:33-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation matches the revised Spec and Plan. Issue-record movement preserves the checked train head only after the shared history, readiness and landing-check proofs pass; other movement reintegrates normally. No blocking findings.

Done-when: current — it covers head/CI retention, card Markdown, fallback controls, late duplicates and verification before landing.

```findings
{}
```

1. **Strengths**
   - Reuses the shared movement classifier, readiness validator and landing-check authority (`cmd/sdlc/landingtrain.go:482`).
   - Rechecks destination gates before the expected-head merge (`cmd/sdlc/landingtrain.go:501`).
   - Real-Git two-car regressions verify publication count, exact CI heads, suite count and preservation of records and cars (`cmd/sdlc/landingtrain_issues_test.go:313`).
   - README and atlas explain the new behavior and tracker-configuration exclusion.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** Both targeted runs passed: new train regressions plus classifier/history controls in 72.467s; existing single-car, readiness, watched-document and docs-only reintegration controls in 11.164s. `git diff --check` passed. The full suite was not rerun during this review; its recorded pass remains required before landing.

6. **Architectural notes**
   - **ARCH-DRY: pass** — shared policy and proof helpers remain authoritative.
   - **ARCH-PURE: pass** — deterministic decisions remain separate from the conductor’s injected IO shell.
   - **ARCH-PURPOSE: pass** — preserves actual head and CI identity, with conservative fallbacks.
   - **ARCH-EVOLVE: pass** — no persisted format or schema changes.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cb85741f-64a7-4d25-89ca-6f4b3cd0cbcb","repository":"github.com/xianxu/ariadne","issue_id":"000511","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/b05e37563ee257c2","reviewed_head":"71d9a6ae50fe907a400ad899372ed073f02613be","slot":"ariadne:2","requested_at":"2026-10-21T10:57:37.185397Z"},"phase":"finished","pid":16335,"started_at":"2026-10-21T10:57:37.301209Z","finished_at":"2026-10-21T10:59:33.348194Z","session_id":"01a15e1c-0720-7960-b93d-4c52e5ed57f2","outcome":"returned"}] -->
