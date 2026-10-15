# Boundary Review — ariadne#384 (whole-issue close)

| field | value |
|-------|-------|
| issue | 384 — sdlc: fixed-then-ship re-close after a post-close fix can't measure actuals, and correct-actual refuses (ambiguous close generation) |
| repo | ariadne |
| issue file | workshop/issues/000384-sdlc-fixed-then-ship-re-close-after-a-post-close-fix-can-t-measure-actuals-and-correct-actual-refuses-ambiguous-close-generation.md |
| boundary | whole-issue close |
| milestone | — |
| window | f113e1e9fc9b7eac98bf7b3a644fc97f1095a360..cf2845c0d334b28818c54eb998cbd0b2b69f35b1 |
| command | sdlc close --issue 384 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T05:23:40-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes satisfy #384’s Spec and Plan. Re-close measures the current candidate while historical consumers retain strict replay. Ambiguous correction now provides eligible targets and concrete recovery commands. No blocking defects found.

```findings
{}
```

1. **Strengths**
   - Candidate/replay separation preserves repository validation and historical pin checks (`cmd/sdlc/actualinputs.go:112`).
   - Explicit row selection remains constrained by effective-row eligibility (`cmd/sdlc/actualcorrection.go:93`).
   - Production regressions exercise repeated measured re-closes, publication, rejected reviews, and conflict-safe correction recovery.
   - README, atlas, and help explain the new behavior and selector.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**
   - Lifecycle, legacy recovery, selector, and refusal regressions passed: 113.053s.
   - Historical mismatch, milestone replay, exclusion, and correction retirement controls passed: 42.858s.
   - Internal actualcorrection, estimate, and recovery packages passed.
   - Range whitespace checks passed. Full suite and mutation checks were not independently rerun.

6. **Architectural notes**
   - **ARCH-DRY — pass:** shared attribution engine and eligibility predicate.
   - **ARCH-PURE — pass:** deterministic selection tested directly; IO remains in integration callers.
   - **ARCH-PURPOSE — pass:** both re-close measurement and ambiguous repair paths delivered.
   - **ARCH-MOCK — pass:** portable fixtures and existing injected seams exercise production paths.
   - **ARCH-CONSTRAINTS — pass:** no new concurrency, fan-out, or workload mechanism.
   - **ARCH-SECURE — pass:** selectors cannot bypass provenance eligibility; generated commands use shell quoting.
   - **ARCH-ORDER — pass:** durable targets remain fixed across pending and completed retries.
   - **ARCH-FUNERAL — pass:** existing supersession, archival, and receipt cleanup own the retained artifacts.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3b46e019-8776-4462-a22b-d0d9cdaffdef","repository":"github.com/xianxu/ariadne","issue_id":"000384","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2a495e1fb5459dbb","reviewed_head":"cf2845c0d334b28818c54eb998cbd0b2b69f35b1","slot":"ariadne:5","requested_at":"2026-10-15T12:20:40.287247Z"},"phase":"finished","pid":77241,"started_at":"2026-10-15T12:20:40.706972Z","finished_at":"2026-10-15T12:23:39.369306Z","session_id":"01a13f81-ea60-77c3-8bbd-33e4d38ca3a6","outcome":"returned"}] -->
