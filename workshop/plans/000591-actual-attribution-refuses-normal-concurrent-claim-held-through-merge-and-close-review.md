# Boundary Review — ariadne#591 (whole-issue close)

| field | value |
|-------|-------|
| issue | 591 — Actual attribution refuses normal concurrent claim held through merge and resting train checkout |
| repo | ariadne |
| issue file | workshop/issues/000591-actual-attribution-refuses-normal-concurrent-claim-held-through-merge-and.md |
| boundary | whole-issue close |
| milestone | — |
| window | 01ff8c994db8cb9b20300544ed1598550ce3a7c2..c701ea2fe03b33401640ba1e203a3e5ab38d26c1 |
| command | sdlc close --issue 591 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T02:06:10-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation matches the revised Spec: witnessed resting intervals remain unassigned, own/foreign attribution stays disjoint, and uncertain concurrent history still refuses. No blocking findings. The recorded pre-close suite remains a separate landing prerequisite; this review does not certify its queued run.

Done-when: current — its fixture, attribution-safety, verification, and documentation clauses still cover the revised contract.

```findings
{}
```

1. **Strengths**
   - Selector ordering preserves sole-reservation planning before resting exclusion (`cmd/sdlc/internal/activetime/branch.go:153`).
   - Fresh and replay paths share warning formatting that distinguishes branch-window minutes from focused activity (`cmd/sdlc/actualbranch.go:297`).
   - Saved-proof validation rejects non-resting exclusions and prevents excluded notes from supporting selected windows.
   - Three captured incidents exercise the actual reflog parser and selector.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage**
   - Complete workspace, activetime, and gatestate package tests passed.
   - Focused race checks and broader owned-actual/replay controls passed.
   - All three recorded incidents failed against the base selector through a temporary Go overlay, reproducing the expected refusal.
   - Real-Git partitioning, retained replay, corruption mutations, and visible excluded minutes passed. Checkout remained clean.

6. **Architecture**
   - **ARCH-DRY — pass:** shared resting predicate and warning formatter.
   - **ARCH-PURE — pass:** attribution remains pure; existing IO boundaries are unchanged.
   - **ARCH-PURPOSE — pass:** exclusion, audit validation, fresh measurement, replay, and output are delivered together.
   - **ARCH-EVOLVE — pass:** nested unknown fields survive JSON/YAML round trips; strict validation remains intact. Previous readers reject the new kind explicitly.
   - Atlas documents the semantics and compatibility limit. No new command, flag, configuration key, or usage step requires a README update.

7. **Plan revisions:** None required.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3906262d-8fbf-438e-9449-e1dd32c0add6","repository":"github.com/xianxu/ariadne","issue_id":"000591","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/839c63a668040214","reviewed_head":"c701ea2fe03b33401640ba1e203a3e5ab38d26c1","slot":"ariadne:3","requested_at":"2026-10-24T09:02:42.826601Z"},"phase":"finished","pid":7755,"started_at":"2026-10-24T09:02:42.921363Z","finished_at":"2026-10-24T09:06:10.16159Z","session_id":"01a16d25-e80e-7da3-8ce5-8b337463a459","outcome":"returned"}] -->
